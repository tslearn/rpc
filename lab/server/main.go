package main

import (
	"fmt"
	"github.com/rpccloud/rpc"
)

func main() {
	fmt.Println("Starting ....")

	userService := rpc.NewService().
		On("SayHello", func(rt rpc.Runtime, name rpc.String) rpc.Return {
			fmt.Println("OK .... return ")
			return rt.Reply("Hello, " + name)
		})

	rpc.NewServer().
		Listen("ws", "127.0.0.1:8888", nil).
		AddService("user", userService, nil).
		Open()
}
