package main

import (
	"fmt"
	"github.com/rpccloud/rpc"
	"github.com/rpccloud/rpc/app/system"
	"github.com/rpccloud/rpc/app/user"
	"github.com/rpccloud/rpc/app/util"
	"time"
)

func test() {
	client, err := rpc.Dial("ws://127.0.0.1:8080/")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	fmt.Println(client.SendMessage(
		5*time.Second,
		"#.user.phone:Create",
		"+86",
		"13011112248",
	))
}

func main() {
	mongoDatabaseConfig := &util.MongoDatabaseConfig{
		URI: "mongodb://dev:World2019@" +
			"192.168.1.61:27011," +
			"192.168.1.61:27012," +
			"192.168.1.61:27013/dev?replicaSet=rs0",
		DataBase:       "dev",
		MaxConnections: 1000,
	}

	go func() {
		time.Sleep(5000 * time.Millisecond)
		test()
	}()

	rpc.NewServer().SetNumOfThreads(1000000).
		AddService("system", system.Service, mongoDatabaseConfig).
		AddService("user", user.Service, mongoDatabaseConfig).
		ListenWebSocket("0.0.0.0:8080").
		Serve()
}
