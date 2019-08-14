package core

import (
	"os"
	"runtime/pprof"
	"testing"
	"time"
)

var service = newServiceMeta().
	Echo("sayHello", true, func(
		ctx Context,
		name string,
	) Return {
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
	time.Sleep(2 * time.Second)

	b.ResetTimer()
	b.ReportAllocs()
	b.N = 20000000
	b.SetParallelism(1024)

	file, _ := os.Create("../cpu.prof")
	pprof.StartCPUProfile(file)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stream := NewRPCStream()
			byte16 := make([]byte, 16, 16)
			stream.WriteBytes(byte16)
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("tianshuo")
			processor.put(stream)
		}
	})

	pprof.StopCPUProfile()
}
