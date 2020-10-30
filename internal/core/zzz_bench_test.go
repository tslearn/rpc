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
	testCount int,
	b *testing.B,
	args ...interface{},
) bool {
	if stream, err := MakeRequestStream("#.test:bench", "", args...); err == nil {
		defer stream.Release()

		if processor, err := NewProcessor(
			false,
			numOfThreads,
			64,
			64,
			threadBufferSize,
			fnCache,
			5*time.Second,
			[]*ServiceMeta{{
				name:     "test",
				service:  NewService().On("bench", handler),
				fileLine: "",
				data:     serviceData,
			}},
			func(s *Stream) { s.Release() },
		); err == nil {
			sendBuffer := stream.GetBuffer()
			runtime.GC()
			b.ResetTimer()
			b.ReportAllocs()
			b.N = testCount
			b.SetParallelism(32)

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					stream := NewStream()
					stream.PutBytesTo(sendBuffer, 0)
					processor.PutStream(stream)
				}
			})
			return processor.Close()
		}
	}

	return false
}

func BenchmarkRPC_basic(b *testing.B) {
	testWithRPCBenchmark(
		8192*24,
		2048,
		&testFuncCache{},
		func(rt Runtime) Return {
			return rt.Reply(true)
		},
		nil,
		20000000,
		b,
	)
}

func BenchmarkRPC_string(b *testing.B) {
	testWithRPCBenchmark(
		8192*24,
		2048,
		&testFuncCache{},
		func(rt Runtime, rtName RTValue) Return {
			if name, err := rtName.ToString(); err != nil {
				panic("error")
			} else if name != "kitty" {
				panic("error")
			} else {
				return rt.Reply(true)
			}
		},
		nil,
		20000000,
		b,
		"kitty",
	)
}

func BenchmarkRPC_array(b *testing.B) {
	testWithRPCBenchmark(
		8192*24,
		2048,
		&testFuncCache{},
		func(rt Runtime, rtArray RTArray) Return {
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
		},
		nil,
		20000000,
		b,
		Array{"hello", "world", true},
	)
}

func BenchmarkRPC_map(b *testing.B) {
	testWithRPCBenchmark(
		8192*24,
		2048,
		&testFuncCache{},
		func(rt Runtime, rtMap RTMap) Return {
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
		},
		nil,
		20000000,
		b,
		Map{"name": "kitty", "age": int64(12)},
	)
}

func BenchmarkRPC_call(b *testing.B) {
	testWithRPCBenchmark(
		1024,
		2048,
		&testFuncCache{},
		func(rt Runtime, v int64) Return {
			if v == 0 {
				return rt.Reply(true)
			}
			return rt.Reply(rt.Call("#.test:bench", v-1))
		},
		nil,
		1000000,
		b,
		int64(60),
	)
}
