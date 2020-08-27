package util

import (
	"context"
	"github.com/rpccloud/rpc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoDatabaseConfig struct {
	URI            string
	DataBase       string
	MaxConnections uint
}

func WithMongoClient(
	uri string,
	timeout time.Duration,
	fn func(client *mongo.Client, ctx context.Context) error,
) (err error) {
	ctx, cancel := context.WithDeadline(context.Background(), rpc.TimeNow().Add(timeout))
	defer cancel()

	if client, e := mongo.Connect(ctx, options.Client().ApplyURI(uri)); e != nil {
		return e
	} else {
		defer func() {
			if e = client.Disconnect(ctx); e != nil && err != nil {
				err = nil
			}
		}()
		return fn(client, ctx)
	}
}

//type mongoDBManagerClientPool struct {
//  ch chan *mongo.Client
//}
//
//type MongoDBManager struct {
//
//}
//
//func (p *MongoDBManager) WithClient(
//  uri string,
//  timeout time.Duration,
//  fn func(client *mongo.Client, ctx context.Context) error,
