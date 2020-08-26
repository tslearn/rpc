package util

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoDatabaseConfig struct {
	URI      string
	DataBase string
}

func WithMongoClient(
	uri string,
	timeout time.Duration,
	fn func(client *mongo.Client, ctx context.Context) error,
) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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
