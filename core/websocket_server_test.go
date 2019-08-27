package core

import (
	"fmt"
	"testing"
	"time"
)

func TestNewWebSocketServer(t *testing.T) {
	server := NewWebSocketServer()
	server.AddService("user", newServiceMeta().
		Echo("sayHello", true, func(
			ctx *rpcContext,
			name string,
		) *rpcReturn {
			return ctx.OK("hello " + name)
		}))

	go func() {
		time.Sleep(2 * time.Second)
		client := NewWebSocketClient(
			"ws://127.0.0.1:10000/ws",
			16000,
			5*1024*1024,
		)
		v, err := client.SendMessage("$.user:sayHello", "tianshuo")
		fmt.Println("OK<", v, err)

		if err := client.Close(); err != nil {
			fmt.Println(err)
		}
		time.Sleep(1 * time.Second)
		if err := server.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	_ = server.Start("0.0.0.0", 10000, "/ws")
}
