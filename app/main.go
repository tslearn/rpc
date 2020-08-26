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
		"13011112229",
	))
}

func main() {
	mongoDatabaseConfig := &util.MongoDatabaseConfig{
		URI: "mongodb+srv://dev:World2019@cluster0.epusb.mongodb.net/dev" +
			"?retryWrites=true&w=majority",
		DataBase: "dev",
	}

	go func() {
		time.Sleep(6 * time.Second)
		test()
	}()

	rpc.NewServer().
		AddService("system", system.Service, mongoDatabaseConfig).
		AddService("user", user.Service, mongoDatabaseConfig).
		ListenWebSocket("0.0.0.0:8080").
		Serve()
}
