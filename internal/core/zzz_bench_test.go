package core

import (
	"runtime"
	"testing"
	"time"
)

func testWithRPCBenchmark(
	numOfThreads int,
	threadBufferSize uint32,
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
			threadBufferSize,
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
		b.N = 20000000
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
		2048,
		&testFuncCache{},
		handler,
		nil,
		onReady,
	)
}

func BenchmarkRPC_string(b *testing.B) {
	handler := func(rt Runtime, rtName RTValue) Return {
		if name, err := rtName.ToString(); err != nil {
			panic("error")
		} else if name != "kitty" {
			panic("error")
		} else {
			return rt.Reply(true)
		}
	}

	onReady := func(processor *Processor, sendBuffer []byte) {
		runtime.GC()
		b.ResetTimer()
		b.ReportAllocs()
		b.N = 20000000
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
		2048,
		&testFuncCache{},
		handler,
		nil,
		onReady,
		"kitty",
	)
}

func BenchmarkRPC_array(b *testing.B) {
	handler := func(rt Runtime, rtArray RTArray) Return {
		if rtArray.Size() != 3 {
			panic("error")
		}
		if v, err := rtArray.Get(0).ToString(); err != nil || v != "hello" {
			panic("error")
		}
		if v, err := rtArray.Get(1).ToString(); err != nil || v != "world" {
			panic("error")
		}
		if v, err := rtArray.Get(2).ToBool(); err != nil || v != true {
			panic("error")
		}
		return rt.Reply(true)
	}

	onReady := func(processor *Processor, sendBuffer []byte) {
		runtime.GC()
		b.ResetTimer()
		b.ReportAllocs()
		b.N = 20000000
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
		2048,
		&testFuncCache{},
		handler,
		nil,
		onReady,
		Array{"hello", "world", true},
	)
}

func BenchmarkRPC_map(b *testing.B) {
	handler := func(rt Runtime, rtMap RTMap) Return {
		if rtMap.Size() != 2 {
			panic("error")
		}
		if v, err := rtMap.Get("name").ToString(); err != nil || v != "kitty" {
			panic("error")
		}
		if v, err := rtMap.Get("age").ToInt64(); err != nil || v != 12 {
			panic("error")
		}
		return rt.Reply(true)
	}

	onReady := func(processor *Processor, sendBuffer []byte) {
		runtime.GC()
		b.ResetTimer()
		b.ReportAllocs()
		b.N = 20000000
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
		2048,
		&testFuncCache{},
		handler,
		nil,
		onReady,
		Map{"name": "kitty", "age": int64(12)},
	)
}
