package main

import (
	"github.com/rpccloud/rpc"
	"github.com/rpccloud/rpc/app/system"
	"github.com/rpccloud/rpc/app/util"
)

func main() {
	seedServiceConfig := &system.SeedServiceConfig{
		Collection: "seedtest05",
		Obscure:    false,
		MongoDatabaseConfig: util.MongoDatabaseConfig{
			DataBase:    "dev",
			Host:        "192.168.1.61",
			Port:        27017,
			Username:    "dev",
			Password:    "World2019",
			ExtraParams: "w=majority",
		},
	}

	rpc.NewServer().
		AddService("system", system.SeedService, seedServiceConfig).
		ListenWebSocket("0.0.0.0:8080").
		Serve()
}
