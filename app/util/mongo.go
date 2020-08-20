package util

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func WithMongoClient(
	uri string,
	timeout time.Duration,
	fn func(client *mongo.Client, ctx context.Context) (interface{}, error),
) (ret interface{}, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if client, e := mongo.Connect(ctx, options.Client().ApplyURI(uri)); e != nil {
		return nil, e
	} else {
		defer func() {
			if e = client.Disconnect(ctx); e != nil && err != nil {
				ret, err = nil, e
			}
		}()
		return fn(client, ctx)
	}
}
