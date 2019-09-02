package core

import (
	"os"
	"runtime/pprof"
	"testing"
	"time"
)

func BenchmarkRpcProcessor_Execute(b *testing.B) {
	processor := newProcessor(
		NewLogger(),
		16,
		16,
		func(stream *rpcStream, success bool) {
			stream.Release()
		},
	)
	processor.start()
	processor.AddService("user", newServiceMeta().
		Echo("sayHello", true, func(
			ctx *rpcContext,
			name string,
		) *rpcReturn {
			return ctx.OK(name)
		}))
	file, _ := os.Create("../cpu.prof")

	time.Sleep(10000 * time.Millisecond)
	_ = pprof.StartCPUProfile(file)

	b.ReportAllocs()
	b.N = 50000000
	b.SetParallelism(1024)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stream := newRPCStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("tianshuo")
			processor.put(stream)
		}
	})
	b.StopTimer()

	pprof.StopCPUProfile()
}
