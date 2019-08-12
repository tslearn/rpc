package core

import (
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
		name RPCString,
	) Return {
		return nil
	})

func getProcessor() *rpcProcessor {
	return newProcessor(
		NewLogger(),
		16,
		16,
		func(stream *RPCStream, success bool) {
			stream.Release()
		},
	)
}

func TestRpcProcessor_Execute(t *testing.T) {
	processor := getProcessor()
	processor.start()
	processor.AddService("user", service)

	stream := NewRPCStream()
	byte16 := make([]byte, 16, 16)
	stream.WriteBytes(byte16)
	stream.WriteString("$.user:sayHello")
	stream.WriteUint64(3)
	stream.WriteString("#")
	stream.WriteInt64(3)

	processor.put(stream)

	time.Sleep(10000 * time.Second)
	processor.stop()
}

func BenchmarkRpcProcessor_Execute(b *testing.B) {

	processor := getProcessor()
	processor.start()
	processor.AddService("user", service)

	time.Sleep(2 * time.Second)

	b.ResetTimer()
	b.ReportAllocs()
	b.N = 50000000
	b.SetParallelism(4096)
	b.RunParallel(func(pb *testing.PB) {
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
}
