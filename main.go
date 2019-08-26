package main

import (
	"fmt"
	"github.com/tslearn/rpcc/core"
	"time"
)

func main() {
	server := core.NewWebSocketServer()
	go newClientGoroutine(server)
	//server.SetReadSizeLimit(500 * 1024 * 1024)
	err := server.Start("0.0.0.0", 10000, "/ws")

	if err != nil {
		fmt.Println(err)
	}
}

func newClientGoroutine(server *core.WebSocketServer) {
	time.Sleep(2 * time.Second)
	client := core.NewWebSocketClient(
		"ws://127.0.0.1:10000/ws",
		16000,
		5*1024*1024,
	)

	v, err := client.SendMessage("$.user.sayHello", "tianshuo")
	fmt.Println(v, err)

	if err := client.Close(); err != nil {
		fmt.Println(err)
	}

	time.Sleep(5 * time.Second)
	if err := server.Close(); err != nil {
		fmt.Println(err)
	}
}
