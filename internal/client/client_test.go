package client

import (
	"fmt"
	"github.com/rpccloud/rpc"
	"github.com/rpccloud/rpc/internal/server"
	"testing"
	"time"
)

func TestClient_Debug(t *testing.T) {
	rpcServer := server.NewServer().ListenTCP("0.0.0.0:28888")

	go func() {
		rpcServer.Serve()
	}()

	time.Sleep(3000 * time.Millisecond)

	rpcClient, err := newClient("tcp://0.0.0.0:28888")

	if err != nil {
		panic(err)
	}

	for i := 0; i < 100; i++ {
		fmt.Println(rpcClient.SendMessage(20*time.Second, "#.test:SayHello", i))
	}

	rpcServer.Close()
}

type testFuncCache struct{}

func (p *testFuncCache) Get(fnString string) rpc.ActionCacheFunc {
	switch fnString {
	case "":
		return func(rt rpc.Runtime, stream *rpc.Stream, fn interface{}) int {
			if !stream.IsReadFinish() {
				return -1
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(rpc.Runtime) rpc.Return)(rt)
				return 0
			}
		}
	default:
		return nil
	}
}

func BenchmarkClient_Debug(b *testing.B) {
	//rpcServer := server.NewServer().ListenTCP("0.0.0.0:28888")
	//rpcServer.AddService(
	//	"test",
	//	core.NewService().On("SayHello", func(rt core.Runtime) core.Return {
	//		return rt.Reply(true)
	//	}),
	//	nil,
	//).SetNumOfThreads(4096).SetActionCache(&testFuncCache{})
	//go func() {
	//	rpcServer.Serve()
	//}()
	//
	//time.Sleep(1000 * time.Millisecond)

	rpcClient, err := newClient("tcp://0.0.0.0:28888")
	time.Sleep(1000 * time.Millisecond)

	if err != nil {
		panic(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.N = 1000000

	for i := 0; i < b.N; i++ {
		rpcClient.SendMessage(10*time.Second, "#.test:SayHello")
	}

	//	rpcServer.Close()
	// rpcClient.Close()
}
