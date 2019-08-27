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

func getTestWebSocketServer() *WebSocketServer {
	server := NewWebSocketServer()
	server.AddService("user", newServiceMeta().
		Echo("sayHello", true, func(
			ctx *rpcContext,
			name string,
		) *rpcReturn {
			return ctx.OK("hello " + name)
		}))

	go func() {
		_ = server.Start("0.0.0.0", 10000, "/ws")
	}()
	return server
}

func BenchmarkNewWebSocketServer(b *testing.B) {
	server := getTestWebSocketServer()
	time.Sleep(3 * time.Second)

	b.ReportAllocs()
	b.N = 100000
	b.SetParallelism(500)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		client := (*WebSocketClient)(nil)
		for client == nil {
			client = NewWebSocketClient(
				"ws://127.0.0.1:10000/ws",
				16000,
				5*1024*1024,
			)
		}
		for pb.Next() {
			client.SendMessage("$.user:sayHello", "tianshuo")
		}
		_ = client.Close()
	})

	_ = server.Close()
}
