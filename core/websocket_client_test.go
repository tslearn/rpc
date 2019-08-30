package core

import (
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
			server.logger.Info("Start")
			time.Sleep(6 * time.Second)
			return ctx.OK("hello " + name)
		}))
	server.SetReadTimeoutMS(1500)
	server.StartBackground("0.0.0.0", 12345, "/")

	for i := 0; i < 10; i++ {
		client := NewWebSocketClient("ws://127.0.0.1:12345/ws")

		server.logger.Info(client.SendMessage(
			"$.user:sayHello",
			"world",
		))

		_ = client.Close()
	}

	_ = server.Close()
}
