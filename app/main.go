package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync/atomic"
	"time"
)

type blockItem struct {
	ID   int64 `bson:"_id"`
	Seed int64 `bson:"seed"`
}

func getBlock() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if client, err := mongo.Connect(
		ctx,
		options.Client().ApplyURI(
			"mongodb://dev:World2019@192.168.1.61:27017/dev?w=majority",
		),
	); err != nil {
		fmt.Println(err)
		return 0, err
	} else {
		defer func() {
			if err = client.Disconnect(ctx); err != nil {
				panic(err)
			}
		}()

		collection := client.Database("dev").Collection("seedBlock")
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result := blockItem{}

		if err := collection.FindOneAndUpdate(
			ctx,
			bson.M{"_id": 3},
			bson.M{"$inc": bson.M{"seed": 1}},
		).Decode(&result); err != nil {
			return 0, err
		}

		//fmt.Println(result.ID)
		return result.Seed, nil
	}
}

func test() {
	sum := int64(0)

	ch := make(chan bool)

	for i := 0; i < 100; i++ {
		go func() {
			if v, err := getBlock(); err != nil {
				panic(err)
			} else {
				atomic.AddInt64(&sum, v)
			}

			ch <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-ch
	}

	fmt.Println(sum)

	//
	//server := rpc.NewServer().SetNumOfThreads(1).
	//  AddService("user", userService).
	//  ListenWebSocket("127.0.0.1:8080")
}

func main() {
	for i := 0; i < 1000; i++ {
		test()
	}
}
