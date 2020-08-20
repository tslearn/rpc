package util

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoDatabaseConfig struct {
	DataBase    string
	Host        string
	Port        uint16
	Username    string
	Password    string
	ExtraParams string
}

func (p *MongoDatabaseConfig) GetURI() string {
	if len(p.Username) > 0 {
		return fmt.Sprintf(
			"mongodb://%s:%s@%s:%d/%s?%s",
			p.Username,
			p.Password,
			p.Host,
			p.Port,
			p.DataBase,
			p.ExtraParams,
		)
	} else {
		return fmt.Sprintf(
			"mongodb://%s:%d/%s?%s",
			p.Host,
			p.Port,
			p.DataBase,
			p.ExtraParams,
		)
	}
}

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
