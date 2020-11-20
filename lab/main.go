package main

import (
	"github.com/rpccloud/rpc"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/server"
)

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

func main() {
	rpcServer := server.NewServer().ListenTCP("0.0.0.0:28888")
	rpcServer.AddService(
		"test",
		core.NewService().On("SayHello", func(rt core.Runtime) core.Return {
			return rt.Reply(true)
		}),
		nil,
	).SetNumOfThreads(4096).SetActionCache(&testFuncCache{})
	rpcServer.Serve()
}
