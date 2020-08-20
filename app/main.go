package main

import (
	"github.com/rpccloud/rpc"
	"github.com/rpccloud/rpc/app/system"
)

func main() {
	seedServiceConfig := &system.SeedServiceConfig{
		URI:        "mongodb://dev:World2019@192.168.1.61:27017/dev?w=majority",
		DataBase:   "dev",
		Collection: "seedtest04",
		Obscure:    false,
	}

	rpc.NewServer().
		AddService("system", system.SeedService, seedServiceConfig).
		ListenWebSocket("0.0.0.0:8080").
		Serve()
}
