package user

import (
	"context"
	"errors"
	"github.com/rpccloud/rpc"
	"github.com/rpccloud/rpc/app/util"
	"github.com/rpccloud/rpc/internal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"time"
)

type mongoDBUserPhone struct {
	ID          int64  `bson:"_id"`
	GlobalPhone string `bson:"globalPhone"`
	Code        string `bson:"code"`
	SendTimeMS  int64  `bson:"sendTimeMS"`
}

var phoneService = rpc.NewServiceWithOnMount(
	func(service *internal.Service, data interface{}) error {
		if cfg, ok := data.(*util.MongoDatabaseConfig); ok && cfg != nil {
			return util.WithMongoClient(cfg.URI, 3*time.Second,
				func(client *mongo.Client, ctx context.Context) error {
					collection := client.Database(cfg.DataBase).Collection("user_phone")
					opts := options.Index().SetUnique(true).SetName("user_phone_index")
					if name, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
						Keys: bson.M{"globalPhone": 1}, Options: opts,
					}); err != nil {
						return err
					} else if name != "user_phone_index" {
						return errors.New("internal error")
					} else {
						return nil
					}
				},
			)
		} else {
			return errors.New("config error")
		}
	},
).Reply("Create", create)

func create(
	ctx rpc.Context,
	zone string,
	phone string,
) rpc.Return {
	cfg, ok := ctx.GetServiceData().(*util.MongoDatabaseConfig)
	if !ok || cfg == nil {
		return ctx.Error(errors.New("config error"))
	}

	if ret, err := ctx.Call("#.system.seed:GetSeed"); err != nil {
		return ctx.Error(err)
	} else if uid, ok := ret.(int64); !ok || uid <= 0 {
		return ctx.Error(errors.New("internal error"))
	} else {
		globalPhone := internal.ConcatString(zone, " ", phone)
		insertPhone := mongoDBUserPhone{
			ID:          uid,
			GlobalPhone: globalPhone,
			Code:        "",
			SendTimeMS:  0,
		}
		insertUser := mongoDBUser{
			ID:         uid,
			SecurityL1: rpc.GetRandString(256),
			SecurityL2: rpc.GetRandString(256),
			SecurityL3: rpc.GetRandString(256),
		}

		if err := util.WithMongoClient(
			cfg.URI,
			3*time.Second,
			func(client *mongo.Client, ctx context.Context) error {
				// Create collections.
				wcMajority := writeconcern.New(
					writeconcern.WMajority(),
					writeconcern.WTimeout(3*time.Second),
				)
				opts := options.Collection().SetWriteConcern(wcMajority)
				users := client.Database(cfg.DataBase).Collection("user", opts)
				phones := client.Database(cfg.DataBase).Collection("user_phone", opts)

				// Step 1: Define the callback that specifies the sequence of operations
				// to perform inside the transaction.
				callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
					// Important: You must pass sessCtx as the Context parameter to the
					// operations for them to be executed in the transaction.
					if _, err := users.InsertOne(sessCtx, insertUser); err != nil {
						return nil, err
					}
					if _, err := phones.InsertOne(sessCtx, insertPhone); err != nil {
						return nil, err
					}
					return nil, nil
				}

				// Step 2: Start a session and run the callback using WithTransaction.
				session, err := client.StartSession()
				if err != nil {
					return err
				}
				defer session.EndSession(ctx)
				if _, err = session.WithTransaction(ctx, callback); err != nil {
					return err
				}

				return nil
			},
		); err != nil {
			return ctx.Error(err)
		} else {
			return ctx.OK(true)
		}
	}
}
