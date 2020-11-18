package client

import (
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/server"
	"testing"
	"time"
)

func TestClient_Debug(t *testing.T) {
	rpcServer := server.NewServer().ListenWebSocket("0.0.0.0:28888")

	go func() {
		rpcServer.Serve()
	}()

	time.Sleep(3000 * time.Millisecond)

	rpcClient, err := newClient("ws://0.0.0.0:28888")

	if err != nil {
		panic(err)
	}

	for i := 0; i < 10000; i++ {
		rpcClient.SendMessage(20*time.Second, "#.test:SayHello", i)
	}

	rpcServer.Close()
}

func BenchmarkClient_Close(b *testing.B) {
	rpcServer := server.NewServer().ListenWebSocket("0.0.0.0:28888")
	rpcServer.AddService(
		"test",
		core.NewService().On("SayHello", func(rt core.Runtime) core.Return {
			return rt.Reply(true)
		}),
		nil,
	).SetNumOfThreads(4096)
	go func() {
		rpcServer.Serve()
	}()

	time.Sleep(1000 * time.Millisecond)

	rpcClient, err := newClient("ws://0.0.0.0:28888")

	if err != nil {
		panic(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.N = 100000

	for i := 0; i < b.N; i++ {
		rpcClient.SendMessage(10*time.Second, "#.test:SayHello")
	}

	rpcServer.Close()
	// rpcClient.Close()
}
