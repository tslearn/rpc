package system

import (
	"context"
	"errors"
	"fmt"
	"github.com/rpccloud/rpc"
	"github.com/rpccloud/rpc/app/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const seedManagerBlockSize = 1 << 20

type SeedServiceConfig struct {
	Collection string
	Obscure    bool
	util.MongoDatabaseConfig
}

type getBlockByMongoDBKind = func(cfg *SeedServiceConfig) (*seedBlock, *seedError)

var getBlockByMongoDB = (func() getBlockByMongoDBKind {
	type mongoDBItem struct {
		ID   int64 `bson:"_id"`
		Seed int64 `bson:"seed"`
	}

	isExist := false
	mu := sync.Mutex{}

	fnCreateIfNotExist := func(cfg *SeedServiceConfig) {
		mu.Lock()
		defer mu.Unlock()

		if !isExist {
			_, _ = util.WithMongoClient(
				cfg.GetURI(),
				2*time.Second,
				func(client *mongo.Client, ctx context.Context) (interface{}, error) {
					collection := client.Database(cfg.DataBase).Collection(cfg.Collection)
					_, _ = collection.InsertOne(ctx, bson.M{"_id": 0, "seed": 1})
					cur, err := collection.Find(ctx, bson.M{"_id": 0})
					if err != nil {
						return nil, err
					}
					isExist = cur.RemainingBatchLength() == 1
					_ = cur.Close(ctx)

					return nil, nil
				},
			)
		}
		return
	}

	return func(cfg *SeedServiceConfig) (*seedBlock, *seedError) {
		fnCreateIfNotExist(cfg)
		if ret, err := util.WithMongoClient(
			cfg.GetURI(),
			2*time.Second,
			func(client *mongo.Client, ctx context.Context) (interface{}, error) {
				collection := client.Database(cfg.DataBase).Collection(cfg.Collection)
				result := mongoDBItem{}

				if err := collection.FindOneAndUpdate(
					ctx,
					bson.M{"_id": 0},
					bson.M{"$inc": bson.M{"seed": 1}},
				).Decode(&result); err != nil {
					return int64(-1), err
				} else {
					return result.Seed, nil
				}
			},
		); err != nil {
			return nil, &seedError{message: err.Error(), time: time.Now()}
		} else if blcokID := ret.(int64); blcokID < 0 {
			return nil, &seedError{message: "internal error", time: time.Now()}
		} else {
			return &seedBlock{blockID: ret.(int64), innerID: 0}, nil
		}
	}
})()

type seedBlock struct {
	blockID int64
	innerID int64
}

type seedError struct {
	message string
	time    time.Time
}

func (p *seedBlock) getSeed() (ret int64) {
	if inner := atomic.AddInt64(&p.innerID, 1); inner < seedManagerBlockSize {
		return p.blockID*seedManagerBlockSize + inner
	}
	return -1
}

type seedManager struct {
	config    *SeedServiceConfig
	currBlock unsafe.Pointer
	nextBlock unsafe.Pointer
	time      time.Time
	sync.Mutex
}

func newSeedManager(config *SeedServiceConfig) *seedManager {
	return &seedManager{
		config:    config,
		currBlock: nil,
		nextBlock: nil,
		time:      time.Now().Add(-time.Hour),
	}
}

//const currBlockIsOK = 0
//const currBlockIsNil = 1
//const currBlockIsExhausted = 2

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
				if block, seedErr := getBlockByMongoDB(p.config); seedErr == nil {
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

	getManager := func(ctx rpc.Context) (*seedManager, error) {
		if ptr := atomic.LoadPointer(&manager); ptr != nil {
			return (*seedManager)(ptr), nil
		} else if cfg, ok := ctx.GetServiceData().(*SeedServiceConfig); !ok {
			return nil, errors.New("config error")
		} else {
			mu.Lock()
			defer mu.Unlock()
			if manager == nil {
				atomic.StorePointer(&manager, unsafe.Pointer(newSeedManager(cfg)))
			}
			return (*seedManager)(manager), nil
		}
	}

	return func(ctx rpc.Context) rpc.Return {
		fmt.Println("request")
		if mgr, err := getManager(ctx); err != nil {
			fmt.Println(err)
			return ctx.Error(err)
		} else if ret, err := mgr.getSeed(); err != nil {
			fmt.Println(err)
			return ctx.Error(err)
		} else {
			fmt.Println(ret)
			return ctx.OK(ret)
		}
	}
}

var SeedService = rpc.NewService().
	Reply("GetSeed", getSeedWrapper())
