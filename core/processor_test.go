package core

import (
	"fmt"
	"github.com/huandu/go-tls/g"
	"os"
	"runtime/pprof"
	"testing"
	"time"
)

var service = newServiceMeta().
	Echo("sayHello", true, func(
		ctx Context,
		i int64,
	) Return {
		return ctx.OK(i)
	}).
	Echo("sayGoodBye", true, func(
		ctx Context,
		name rpcString,
	) Return {
		return nil
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

		fmt.Println(g.G())
		for pb.Next() {
			stream := NewRPCStream()
			byte16 := make([]byte, 16, 16)
			stream.WriteBytes(byte16)
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteInt64(3)
			processor.put(stream)
		}
	})

	pprof.StopCPUProfile()
}

func BenchmarkRpcProcessor_G(b *testing.B) {
	//c := 0
	b.ResetTimer()
	b.ReportAllocs()
	b.N = 500000000
	b.SetParallelism(12)
	// count := uint64(0)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			//if &c == nil {
			//  fmt.Println("hi")
			//  //atomic.AddUint64(&count, 1)
			//}
		}
	})
	//fmt.Println(count)
}
