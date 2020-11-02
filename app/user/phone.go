package user

import (
	"context"
	"errors"
	"github.com/rpccloud/rpc"
	"github.com/rpccloud/rpc/app/util"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"time"
)

type mongoDBUserPhone struct {
	ID          int64  `bson:"_id"`
	GlobalPhone string `bson:"globalPhone"`
	Password    string `bson:"password"`
	Active      bool   `bson:"active"`
}

var phoneService = rpc.NewServiceWithOnMount(onPhoneServiceMount).
	Reply("Create", create).
	Reply("GetCode", getCode)

func onPhoneServiceMount(service *core.Service, data interface{}) error {
	if cfg, ok := data.(*util.MongoDatabaseConfig); ok && cfg != nil {
		return util.WithMongoClient(cfg, 3*time.Second,
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
}

func getGlobalPhone(zone string, phone string) string {
	return base.ConcatString(zone, " ", phone)
}

func getCode(ctx rpc.Runtime, zone string, phone string) rpc.Return {
	if err := getCheckManager().SendCheckCode(
		getGlobalPhone(zone, phone),
	); err != nil {
		return ctx.Error(err)
	} else {
		return ctx.OK(true)
	}
}

func create(
	ctx rpc.Runtime,
	zone string,
	phone string,
	checkCode string,
) rpc.Return {
	globalPhone := getGlobalPhone(zone, phone)
	if !getCheckManager().CheckCode(globalPhone, checkCode) {
		return ctx.Error(errors.New("check code error"))
	}

	cfg, ok := ctx.GetServiceConfig().(*util.MongoDatabaseConfig)
	if !ok || cfg == nil {
		return ctx.Error(errors.New("config error"))
	}

	if ret, err := ctx.Call("#.system.seed:GetSeed"); err != nil {
		return ctx.Error(err)
	} else if uid, ok := ret.(int64); !ok || uid <= 0 {
		return ctx.Error(errors.New("internal error"))
	} else {
		insertPhone := mongoDBUserPhone{
			ID:          uid,
			GlobalPhone: globalPhone,
			Password:    "",
			Active:      false,
		}
		insertUser := mongoDBUser{
			ID:         uid,
			SecurityL1: rpc.GetRandString(256),
			SecurityL2: rpc.GetRandString(256),
			SecurityL3: rpc.GetRandString(256),
		}

		if err := util.WithMongoClient(cfg, 3*time.Second,
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
			return ctx.OK(insertUser.SecurityL1)
		}
	}
}
