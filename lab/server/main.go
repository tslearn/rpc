package main

import (
	"fmt"
	"github.com/rpccloud/rpc"
	"time"
)

func main() {
	fmt.Println("Starting ....")

	userService := rpc.NewService().
		On("SayHello", func(rt rpc.Runtime, name rpc.String) rpc.Return {
			return rt.Reply("Hello, " + name)
		})

	go runClient()

	rpc.NewServer().
		Listen("ws", "127.0.0.1:8888", nil).
		AddService("user", userService, nil).
		Open()
}

func runClient() {
	time.Sleep(5 * time.Second)

	client, e := rpc.Dial("ws://127.0.0.1:8888")
	if e != nil {
		return
	}

	fmt.Println(client.Send(3*time.Second, "#.user:SayHello", "world"))
	client.Close()
}
