package main

import (
	"fmt"
	"github.com/rpccloud/rpc"
	"github.com/rpccloud/rpc/app/system"
	"github.com/rpccloud/rpc/app/util"
	"time"
)

// "mongodb://%s:%s@%s:%d/%s?%s",
//			p.Username,
//			p.Password,
//			p.Host,
//			p.Port,
//			p.DataBase,
//			p.ExtraParams,

func test() {
	client, err := rpc.Dial("ws://127.0.0.1:8080/")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	fmt.Println(client.SendMessage(
		5*time.Second,
		"#.system.seed:GetSeed",
	))
}

func main() {
	mongoDatabaseConfig := &util.MongoDatabaseConfig{
		URI:      "mongodb://dev:World2019@192.168.1.61:27017/dev?w=majority",
		DataBase: "dev",
	}

	go func() {
		time.Sleep(time.Second)
		test()
	}()

	rpc.NewServer().
		AddService("system", system.Service, mongoDatabaseConfig).
		ListenWebSocket("0.0.0.0:8080").
		Serve()
}
