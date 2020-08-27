package util

import (
	"context"
	"errors"
	"fmt"
	"github.com/rpccloud/rpc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

type MongoDatabaseConfig struct {
	URI            string
	DataBase       string
	MaxConnections int
}

//func WithMongoClient(
//  uri string,
//  timeout time.Duration,
//  fn func(client *mongo.Client, ctx context.Context) error,
//) (err error) {
//  ctx, cancel := context.WithDeadline(context.Background(), rpc.TimeNow().Add(timeout))
//  defer cancel()
//
//  if client, e := mongo.Connect(ctx, options.Client().ApplyURI(uri)); e != nil {
//    return e
//  } else {
//    defer func() {
//      if e = client.Disconnect(ctx); e != nil && err != nil {
//        err = nil
//      }
//    }()
//    return fn(client, ctx)
//  }
//}

type mongoDBConn struct {
	client       *mongo.Client
	lastUsedTime time.Time
}

type mongoDBManagerPool struct {
	uri      string
	ch       chan *mongoDBConn
	currSize int
	maxSize  int
	sync.Mutex
}

func (p *mongoDBManagerPool) getConn(
	ctx context.Context,
) (*mongoDBConn, error) {
	select {
	case ret := <-p.ch:
		return ret, nil
	default:
		conn, err, isFull := (*mongoDBConn)(nil), error(nil), false
		func() {
			p.Lock()
			defer p.Unlock()
			if p.currSize < p.maxSize {
				client, err := mongo.Connect(ctx, options.Client().ApplyURI(p.uri))
				if err == nil {
					p.currSize++
					conn = &mongoDBConn{client: client}
				}
			} else {
				isFull = true
			}
		}()
		if !isFull {
			return conn, err
		}
		return p.waitConn(ctx)
	}
}

func (p *mongoDBManagerPool) waitConn(
	ctx context.Context,
) (*mongoDBConn, error) {
	select {
	case ret := <-p.ch:
		return ret, nil
	case <-ctx.Done():
		return nil, errors.New("timeout")
	}
}

func (p *mongoDBManagerPool) onTimer() {

}

type withClientType = func(
	cfg *MongoDatabaseConfig,
	timeout time.Duration,
	fn func(client *mongo.Client, ctx context.Context) error,
) error

var WithMongoClient = func() withClientType {
	mp := sync.Map{}

	go func() {
		mp.Range(func(k, v interface{}) bool {
			v.(*mongoDBManagerPool).onTimer()
			return true
		})
	}()

	return func(
		cfg *MongoDatabaseConfig,
		timeout time.Duration,
		fn func(client *mongo.Client, ctx context.Context) error,
	) (ret error) {
		pool := (*mongoDBManagerPool)(nil)
		if cfg == nil {
			return errors.New("config error")
		} else if v, ok := mp.Load(cfg.URI); !ok {
			pool = &mongoDBManagerPool{
				uri:      cfg.URI,
				ch:       make(chan *mongoDBConn, cfg.MaxConnections),
				currSize: 0,
				maxSize:  cfg.MaxConnections,
			}
			mp.Store(cfg.URI, pool)
		} else {
			pool = v.(*mongoDBManagerPool)
		}

		ctx, cancel := context.WithDeadline(
			context.Background(),
			rpc.TimeNow().Add(timeout),
		)
		defer cancel()

		if conn, err := pool.getConn(ctx); err != nil {
			return err
		} else {
			defer func() {
				if v := recover(); v != nil {
					ret = fmt.Errorf("runtime error: %s", v)
				}
				conn.lastUsedTime = rpc.TimeNow()
				pool.ch <- conn
			}()
			return fn(conn.client, ctx)
		}
	}
}()
