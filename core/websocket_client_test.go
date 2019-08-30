package core

import (
	"fmt"
	"testing"
	"time"
)

func TestWebSocketClient_basic(t *testing.T) {
	//	assert := newAssert(t)

	server := NewWebSocketServer()
	server.AddService("user", newServiceMeta().
		Echo("sayHello", true, func(
			ctx *rpcContext,
			name string,
		) *rpcReturn {
			return ctx.OK("hello " + name)
		}))
	server.StartBackground("0.0.0.0", 12345, "/")

	for i := 0; i < 10; i++ {
		client := NewWebSocketClient("ws://127.0.0.1:12345/ws")

		fmt.Println(client.SendMessage(
			"$.user:sayHello",
			"world",
		))

		_ = client.Close()
	}

	time.Sleep(30 * time.Second)

	_ = server.Close()
}
