package core

import (
	"testing"
)

func TestWebSocketClient_basic(t *testing.T) {
	assert := newAssert(t)

	server := NewWebSocketServer()
	server.AddService("user", newServiceMeta().
		Echo("sayHello", true, func(
			ctx *rpcContext,
			name string,
		) *rpcReturn {
			return ctx.OK("hello " + name)
		}))
	server.StartBackground("0.0.0.0", 12345, "/")

	client := NewWebSocketClient("ws://127.0.0.1:12345/ws")
	assert(client.SendMessage(
		"$.user:sayHello",
		"world",
	)).Equals("hello world", nil)

	_ = client.Close()
}
