package main

import (
	"github.com/rpccloud/rpc"
)

func main() {
	userService := rpc.NewService().
		On("SayHello", func(rt rpc.Runtime, name rpc.String) rpc.Return {
			return rt.Reply("Hello, " + name)
		})

	rpc.NewServer().
		Listen("ws", "127.0.0.1:8888", nil).
		AddService("user", userService, nil).
		Open()
}
