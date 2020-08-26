package user

import (
	"context"
	"errors"
	"github.com/rpccloud/rpc"
	"github.com/rpccloud/rpc/app/util"
	"github.com/rpccloud/rpc/internal"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
	"time"
)

var Service = rpc.NewServiceWithOnMount(
	func(service *internal.Service, data interface{}) error {
		if cfg, ok := data.(*util.MongoDatabaseConfig); !ok || cfg == nil {
			return errors.New("config error")
		} else if err := util.WithMongoClient(cfg.URI, 3*time.Second,
			func(client *mongo.Client, ctx context.Context) error {
				return client.Database(cfg.DataBase).CreateCollection(ctx, "user")
			},
		); err != nil && !strings.Contains(err.Error(), "already exists") {
			return err
		} else {
			service.AddChildService("phone", phoneService, data)
			return nil
		}
	},
)
