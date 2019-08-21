package core

import (
	"os"
	"runtime/pprof"
	"testing"
	"time"
)

var service = newServiceMeta().
	Echo("sayHello", true, func(
		ctx *rpcContext,
		name string,
	) *rpcReturn {
		return ctx.OK(name)
	})

func getProcessor() *rpcProcessor {
	return newProcessor(
		NewLogger(),
		16,
		16,
		func(stream *rpcStream, success bool) {
			stream.Release()
		},
	)
}

func BenchmarkRpcProcessor_Execute(b *testing.B) {
	processor := getProcessor()
	processor.start()
	processor.AddService("user", service)
	file, _ := os.Create("../cpu.prof")

	time.Sleep(2000 * time.Millisecond)
	pprof.StartCPUProfile(file)

	b.ReportAllocs()
	b.N = 100000000
	b.SetParallelism(1024)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stream := newRPCStream()
			byte16 := make([]byte, 16, 16)
			stream.WriteBytes(byte16)
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
