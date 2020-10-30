package core

import (
	"runtime"
	"testing"
	"time"
)

func testWithRPCBenchmark(
	numOfThreads int,
	fnCache ActionCache,
	handler interface{},
	serviceData Map,
	onReady func(processor *Processor, sendBuffer []byte),
	args ...interface{},
) bool {
	if stream, err := MakeRequestStream("#.test:bench", "", args...); err == nil {
		defer stream.Release()

		if processor, err := NewProcessor(
			false,
			numOfThreads,
			16,
			16,
			fnCache,
			5*time.Second,
			[]*ServiceMeta{{
				name:     "test",
				service:  NewService().On("bench", handler),
				fileLine: "",
				data:     serviceData,
			}},
			func(stream *Stream) { stream.Release() },
		); err == nil {
			onReady(processor, stream.GetBuffer())
			return processor.Close()
		}
	}

	return false
}

func BenchmarkRPC_basic(b *testing.B) {
	handler := func(rt Runtime) Return {
		return rt.Reply(true)
	}

	onReady := func(processor *Processor, sendBuffer []byte) {
		runtime.GC()
		b.ResetTimer()
		b.ReportAllocs()
		b.N = 100000000
		b.SetParallelism(32)

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				stream := NewStream()
				stream.PutBytesTo(sendBuffer, 0)
				processor.PutStream(stream)
			}
		})
	}

	testWithRPCBenchmark(
		8192*24,
		&testFuncCache{},
		handler,
		nil,
		onReady,
	)
}
