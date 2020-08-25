package user

//
//type UserServiceConfig struct {
//  Collection string
//  Obscure    bool
//  util.MongoDatabaseConfig
//}
//
//var UserService = rpc.NewService().
//  Reply("CreateUserByNames", createUserByPhone).
//  Reply("DeleteUser", deleteUser)
//
//type mongoDBSecurity struct {
//  ID         int64  `bson:"_id"`
//  SecurityL1 string `bson:"security_l1"` // light user
//  SecurityL2 string `bson:"security_l2"` // normal user
//  SecurityL3 string `bson:"security_l3"` // admin user
//}
//
//type mongoDBUser struct {
//  ID       int64  `bson:"_id"`
//  name     string `bson:"name"`
//  password string `bson:"password"`
//}
//
//type mongoDBUserPhone struct {
//  ID          int64  `bson:"_id"`
//  globalPhone string `bson:"globalPhone"`
//  code        string `bson:"code"`
//  sendTimeMS  uint64 `bson:"sendTimeMS"`
//}
//
//
//func createUserByPhoneWrapper() interface{} {
//  return func(ctx rpc.Context) rpc.Return {
//    if mgr, err := getManager(ctx); err != nil {
//      return ctx.Error(err)
//    } else if ret, err := mgr.getSeed(); err != nil {
//      return ctx.Error(err)
//    } else {
//      return ctx.OK(ret)
//    }
//  }
//}
//
//
//
//func createUserByPhone(
//  ctx rpc.Context,
//  zone string,
//  phone string,
//) rpc.Return {
//  globalPhone := internal.ConcatString(zone, " ", phone)
//
//  if cfg, ok := ctx.GetServiceData().(*util.MongoDatabaseConfig); !ok {
//    return ctx.Error(errors.New("config error"))
//  } else {
//    canCreate, err := util.WithMongoClient(cfg.URI, 3*time.Second,
//      func(client *mongo.Client, ctx context.Context) (interface{}, error) {
//        collection := client.Database(cfg.DataBase).Collection("user_phone")
//        cur, err := collection.Find(ctx, bson.M{"globalPhone": globalPhone})
//        if err != nil {
//          return false, err
//        }
//        defer func() {
//          _ = cur.Close(ctx)
//        }()
//
//        return cur.RemainingBatchLength() == 0, nil
//      },
//    )
//  }
//
//  // check if exist
//
//  _, _ = util.WithMongoClient(
//    cfg.URI,
//    3*time.Second,
//    func(client *mongo.Client, ctx context.Context) (interface{}, error) {
//      collection := client.Database(cfg.DataBase).Collection("system_seed")
//      _, _ = collection.InsertOne(ctx, mongoDBItem{ID: 0, Seed: 1})
//      cur, err := collection.Find(ctx, bson.M{"_id": 0})
//      if err != nil {
//        return nil, err
//      }
//      isExist = cur.RemainingBatchLength() == 1
//      _ = cur.Close(ctx)
//
//      return nil, nil
//    },
//  )
//}
//
//
//return ctx.OK("createUserByPhone")
//}
//
//func deleteUser(
//  ctx rpc.Context,
//  security string,
//  userID uint64,
//) rpc.Return {
//  return ctx.OK("deleteUser")
//}
