package seed

import (
	"context"
	"errors"
	"github.com/rpccloud/rpc"
	"github.com/rpccloud/rpc/app/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

var Service = rpc.NewServiceWithOnMount(
	func(_ *rpc.Service, data interface{}) error {
		if cfg, ok := data.(*util.MongoDatabaseConfig); ok && cfg != nil {
			return util.WithMongoClient(cfg, 3*time.Second,
				func(client *mongo.Client, ctx context.Context) error {
					_, err := client.
						Database(cfg.DataBase).
						Collection("system_seed").
						UpdateOne(
							ctx,
							bson.M{"_id": 0},
							bson.M{"$inc": bson.M{"seed": int64(1)}},
							options.Update().SetUpsert(true),
						)
					return err
				},
			)
		} else {
			return errors.New("config error")
		}
	},
).Reply("GetSeed", getSeedWrapper())

const seedManagerBlockSize = 1 << 20

type mongoDBSeedItem struct {
	ID   int64 `bson:"_id"`
	Seed int64 `bson:"seed"`
}

func getBlockByMongoDB(cfg *util.MongoDatabaseConfig) (*seedBlock, error) {
	blockID := int64(0)
	if err := util.WithMongoClient(cfg, 3*time.Second,
		func(client *mongo.Client, ctx context.Context) error {
			collection := client.Database(cfg.DataBase).Collection("system_seed")
			result := mongoDBSeedItem{}
			if err := collection.FindOneAndUpdate(
				ctx,
				bson.M{"_id": 0},
				bson.M{"$inc": bson.M{"seed": int64(1)}},
				options.FindOneAndUpdate().SetUpsert(true),
			).Decode(&result); err != nil {
				return err
			} else {
				blockID = result.Seed
				return nil
			}
		},
	); err != nil {
		return nil, err
	} else if blockID <= 0 {
		return nil, errors.New("internal database error")
	} else {
		return &seedBlock{blockID: blockID, innerID: 0}, nil
	}
}

type seedBlock struct {
	blockID int64
	innerID int64
}

func (p *seedBlock) getSeed() (ret int64) {
	if inner := atomic.AddInt64(&p.innerID, 1); inner < seedManagerBlockSize {
		return p.blockID*seedManagerBlockSize + inner
	}
	return -1
}

type seedManager struct {
	config    *util.MongoDatabaseConfig
	currBlock unsafe.Pointer
	nextBlock unsafe.Pointer
	time      time.Time
	sync.Mutex
}

func newSeedManager(config *util.MongoDatabaseConfig) *seedManager {
	return &seedManager{
		config:    config,
		currBlock: nil,
		nextBlock: nil,
		time:      time.Now().Add(-time.Hour),
	}
}

func (p *seedManager) getSeedInner() (int64, *seedBlock) {
	if block := (*seedBlock)(atomic.LoadPointer(&p.currBlock)); block != nil {
		return block.getSeed(), block
	}

	return -1, nil
}

func (p *seedManager) getSeed() (int64, error) {
	if seed, block := p.getSeedInner(); seed >= 0 {
		return seed, nil
	} else if block == nil {
		p.Lock()
		defer p.Unlock()

		if p.currBlock == nil {
			now := time.Now()
			if now.Sub(p.time) > 800*time.Millisecond {
				p.time = now
				currBlock, err1 := getBlockByMongoDB(p.config)
				nextBlock, err2 := getBlockByMongoDB(p.config)
				if err1 == nil && err2 == nil {
					atomic.StorePointer(&p.currBlock, unsafe.Pointer(currBlock))
					atomic.StorePointer(&p.nextBlock, unsafe.Pointer(nextBlock))
				}
			}
		}
	} else {
		p.Lock()
		defer p.Unlock()

		// try to update nextBlock
		if atomic.CompareAndSwapPointer(
			&p.currBlock,
			unsafe.Pointer(block),
			atomic.LoadPointer(&p.nextBlock),
		) {
			atomic.StorePointer(&p.nextBlock, nil)
			go func() {
				if block, err := getBlockByMongoDB(p.config); err == nil {
					atomic.StorePointer(&p.nextBlock, unsafe.Pointer(block))
				}
			}()
		}
	}

	// get seed again
	if seed, _ := p.getSeedInner(); seed >= 0 {
		return seed, nil
	} else {
		return -1, errors.New("it is temporarily unavailable")
	}
}

func getSeedWrapper() interface{} {
	manager := unsafe.Pointer(nil)
	mu := sync.Mutex{}

	getManager := func(ctx rpc.Runtime) (*seedManager, error) {
		if ptr := atomic.LoadPointer(&manager); ptr != nil {
			return (*seedManager)(ptr), nil
		} else if cfg, ok := ctx.GetServiceData().(*util.MongoDatabaseConfig); !ok {
			return nil, errors.New("config data error")
		} else {
			mu.Lock()
			defer mu.Unlock()
			if manager == nil {
				atomic.StorePointer(&manager, unsafe.Pointer(newSeedManager(cfg)))
			}
			return (*seedManager)(manager), nil
		}
	}

	return func(ctx rpc.Runtime) rpc.Return {
		if mgr, err := getManager(ctx); err != nil {
			return ctx.Error(err)
		} else if ret, err := mgr.getSeed(); err != nil {
			return ctx.Error(err)
		} else {
			return ctx.OK(ret)
		}
	}
}
