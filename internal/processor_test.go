package internal

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewRPCProcessor(t *testing.T) {
	assert := NewAssert(t)

	processor := NewRPCProcessor(true, 8192, 16, 32, nil)
	assert(processor).IsNotNil()
	assert(len(processor.repliesMap)).Equals(0)
	assert(len(processor.nodesMap)).Equals(1)
	assert(processor.isDebug).IsTrue()
	assert(processor.maxNodeDepth).Equals(uint64(16))
	assert(processor.maxCallDepth).Equals(uint64(32))
	processor.Stop()
}

func TestRPCProcessor_Start_Stop(t *testing.T) {
	//assert := NewAssert(t)
	//
	//processor := NewRPCProcessor(true, 8192, 16, 32, nil, nil)
	//assert(processor.Stop()).IsNotNil()
	//assert(processor.isRunning).IsFalse()
	//for i := 0; i < len(processor.threadPools); i++ {
	//	assert(processor.threadPools[i]).IsNil()
	//}
	//assert(processor.Start()).IsTrue()
	//assert(processor.isRunning).IsTrue()
	//for i := 0; i < len(processor.threadPools); i++ {
	//	assert(processor.threadPools[i]).IsNotNil()
	//}
	//assert(processor.Start()).IsFalse()
	//assert(processor.isRunning).IsTrue()
	//for i := 0; i < len(processor.threadPools); i++ {
	//	assert(processor.threadPools[i]).IsNotNil()
	//}
	//assert(processor.Stop()).IsNil()
	//assert(processor.isRunning).IsFalse()
	//for i := 0; i < len(processor.threadPools); i++ {
	//	assert(processor.threadPools[i]).IsNil()
	//}
	//assert(processor.Stop()).IsNotNil()
	//assert(processor.isRunning).IsFalse()
	//for i := 0; i < len(processor.threadPools); i++ {
	//	assert(processor.threadPools[i]).IsNil()
	//}
}

func TestRPCProcessor_PutStream(t *testing.T) {
	//assert := NewAssert(t)
	//processor := NewRPCProcessor(16, 32, nil, nil)
	//assert(processor.PutStream(NewRPCStream())).IsFalse()
	//processor.Start()
	//assert(processor.PutStream(NewRPCStream())).IsTrue()
	//for i := 0; i < len(processor.threadPools); i++ {
	//	processor.threadPools[i].stop()
	//}
	//assert(processor.PutStream(NewRPCStream())).IsFalse()
}

func TestRPCProcessor_AddService(t *testing.T) {
	assert := NewAssert(t)

	processor := NewRPCProcessor(true, 8192, 16, 32, nil)
	assert(processor.AddService("test", nil, "DebugMessage")).
		Equals(NewRPCErrorByDebug(
			"Service is nil",
			"DebugMessage",
		))

	service := NewService()
	assert(processor.AddService("test", service, "")).IsNil()
}

func TestRPCProcessor_BuildCache(t *testing.T) {
	assert := NewAssert(t)
	_, file, _, _ := runtime.Caller(0)

	processor0 := NewRPCProcessor(true, 8192, 16, 32, nil)
	assert(processor0.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/processor-build-cache-0.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/processor-build-cache-0.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/processor-build-cache-0.go")))

	processor1 := NewRPCProcessor(true, 8192, 16, 32, nil)
	_ = processor1.AddService("abc", NewService().
		Reply("sayHello", func(ctx *RPCContext, name string) *RPCReturn {
			return ctx.OK("hello " + name)
		}), "")
	assert(processor1.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/processor-build-cache-1.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/processor-build-cache-1.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/processor-build-cache-1.go")))

	_ = os.RemoveAll(path.Join(path.Dir(file), "_tmp_"))
}

func TestRPCProcessor_mountNode(t *testing.T) {
	assert := NewAssert(t)

	processor := NewRPCProcessor(true, 8192, 16, 16, nil)

	assert(processor.mountNode(rootName, nil).GetMessage()).
		Equals("rpc: mountNode: nodeMeta is nil")
	assert(processor.mountNode(rootName, nil).GetDebug()).
		Equals("")

	assert(processor.mountNode(rootName, &rpcAddChildMeta{
		name:    "+",
		service: NewService(),
		debug:   "DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Service name \"+\" is illegal",
		"DebugMessage",
	))

	assert(processor.mountNode(rootName, &rpcAddChildMeta{
		name:    "abc",
		service: nil,
		debug:   "DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Service is nil",
		"DebugMessage",
	))

	assert(processor.mountNode("123", &rpcAddChildMeta{
		name:    "abc",
		service: NewService(),
		debug:   "DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"rpc: mountNode: parentNode is nil",
		"DebugMessage",
	))

	processor.maxNodeDepth = 0
	assert(processor.mountNode(rootName, &rpcAddChildMeta{
		name:    "abc",
		service: NewService(),
		debug:   "DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Service path depth $.abc is too long, it must be less or equal than 0",
		"DebugMessage",
	))
	processor.maxNodeDepth = 16

	_ = processor.mountNode(rootName, &rpcAddChildMeta{
		name:    "abc",
		service: NewService(),
		debug:   "DebugMessage",
	})
	assert(processor.mountNode(rootName, &rpcAddChildMeta{
		name:    "abc",
		service: NewService(),
		debug:   "DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Service name \"abc\" is duplicated",
		"Current:\n\tDebugMessage\nConflict:\n\tDebugMessage",
	))

	// mount reply error
	service := NewService()
	service.Reply("abc", nil)
	assert(processor.mountNode(rootName, &rpcAddChildMeta{
		name:    "test",
		service: service,
		debug:   "DebugMessage",
	}).GetMessage()).Equals("Reply handler is nil")

	// mount children error
	service1 := NewService()
	service1.AddChild("abc", NewService())
	assert(len(service1.children)).Equals(1)
	service1.children[0] = nil
	assert(processor.mountNode(rootName, &rpcAddChildMeta{
		name:    "003",
		service: service1,
		debug:   "DebugMessage",
	}).GetMessage()).Equals("rpc: mountNode: nodeMeta is nil")

	// OK
	service2 := NewService()
	service2.AddChild("user", NewService().
		Reply("sayHello", func(ctx *RPCContext) *RPCReturn {
			return ctx.OK(true)
		}))
	assert(processor.mountNode(rootName, &rpcAddChildMeta{
		name:    "system",
		service: service2,
		debug:   "DebugMessage",
	})).IsNil()
}

func TestRPCProcessor_mountReply(t *testing.T) {
	assert := NewAssert(t)

	processor := NewRPCProcessor(true, 8192, 16, 16, &testFuncCache{})
	rootNode := processor.nodesMap[rootName]

	// check the node is nil
	assert(processor.mountReply(nil, nil)).Equals(NewRPCErrorByDebug(
		"rpc: mountReply: node is nil",
		"",
	))

	// check the replyMeta is nil
	assert(processor.mountReply(rootNode, nil)).Equals(NewRPCErrorByDebug(
		"rpc: mountReply: replyMeta is nil",
		"",
	))

	// check the name
	assert(processor.mountReply(rootNode, &rpcReplyMeta{
		"###",
		nil,
		"DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Reply name ### is illegal",
		"DebugMessage",
	))

	// check the reply path is not occupied
	_ = processor.mountReply(rootNode, &rpcReplyMeta{
		"testOccupied",
		func(ctx *RPCContext) *RPCReturn { return ctx.OK(true) },
		"DebugMessage",
	})
	assert(processor.mountReply(rootNode, &rpcReplyMeta{
		"testOccupied",
		func(ctx *RPCContext) *RPCReturn { return ctx.OK(true) },
		"DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Reply name testOccupied is duplicated",
		"Current:\n\tDebugMessage\nConflict:\n\tDebugMessage",
	))

	// check the reply handler is nil
	assert(processor.mountReply(rootNode, &rpcReplyMeta{
		"testReplyHandlerIsNil",
		nil,
		"DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Reply handler is nil",
		"DebugMessage",
	))

	// Check reply handler is Func
	assert(processor.mountReply(rootNode, &rpcReplyMeta{
		"testReplyHandlerIsFunction",
		make(chan bool),
		"DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Reply handler must be func(ctx rpc.Context, ...) rpc.Return",
		"DebugMessage",
	))

	// Check reply handler arguments types
	assert(processor.mountReply(rootNode, &rpcReplyMeta{
		"testReplyHandlerArguments",
		func(ctx bool) *RPCReturn { return nilReturn },
		"DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Reply handler 1st argument type must be rpc.Context",
		"DebugMessage",
	))

	assert(processor.mountReply(rootNode, &rpcReplyMeta{
		"testReplyHandlerArguments",
		func(ctx *RPCContext, ch chan bool) *RPCReturn { return nilReturn },
		"DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Reply handler 2nd argument type <chan bool> not supported",
		"DebugMessage",
	))

	// Check return type
	assert(processor.mountReply(rootNode, &rpcReplyMeta{
		"testReplyHandlerReturn",
		func(ctx *RPCContext) (*RPCReturn, bool) { return nilReturn, true },
		"DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Reply handler return type must be rpc.Return",
		"DebugMessage",
	))

	assert(processor.mountReply(rootNode, &rpcReplyMeta{
		"testReplyHandlerReturn",
		func(ctx RPCContext) bool { return true },
		"DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Reply handler return type must be rpc.Return",
		"DebugMessage",
	))

	// ok
	assert(processor.mountReply(rootNode, &rpcReplyMeta{
		"testOK",
		func(ctx *RPCContext, _ bool, _ RPCMap) *RPCReturn { return nilReturn },
		GetStackString(0),
	})).IsNil()

	assert(processor.repliesMap["$:testOK"].replyMeta.name).Equals("testOK")
	assert(processor.repliesMap["$:testOK"].reflectFn).IsNotNil()
	assert(processor.repliesMap["$:testOK"].callString).
		Equals("$:testOK(rpc.Context, rpc.Bool, rpc.Map) rpc.Return")
	assert(
		strings.Contains(processor.repliesMap["$:testOK"].debugString, "$:testOK"),
	).IsTrue()
	assert(processor.repliesMap["$:testOK"].argTypes[0]).
		Equals(reflect.ValueOf(nilContext).Type())
	assert(processor.repliesMap["$:testOK"].argTypes[1]).Equals(boolType)
	assert(processor.repliesMap["$:testOK"].argTypes[2]).Equals(mapType)
	assert(processor.repliesMap["$:testOK"].indicator).IsNotNil()
}

func TestRPCProcessor_OutPutErrors(t *testing.T) {
	assert := NewAssert(t)

	processor := NewRPCProcessor(true, 8192, 16, 16, nil)

	// Service is nil
	assert(processor.AddService("", nil, "DebugMessage")).
		Equals(NewRPCErrorByDebug(
			"Service is nil",
			"DebugMessage",
		))

	assert(processor.AddService("abc", (*Service)(nil), "DebugMessage")).
		Equals(NewRPCErrorByDebug(
			"Service is nil",
			"DebugMessage",
		))

	// Service name %s is illegal
	assert(processor.AddService("\"\"", NewService(), "DebugMessage")).
		Equals(NewRPCErrorByDebug(
			"Service name \"\"\"\" is illegal",
			"DebugMessage",
		))

	processor.maxNodeDepth = 0
	assert(processor.AddService("abc", NewService(), "DebugMessage")).
		Equals(NewRPCErrorByDebug(
			"Service path depth $.abc is too long, it must be less or equal than 0",
			"DebugMessage",
		))
	processor.maxNodeDepth = 16

	_ = processor.AddService("abc", NewService(), "DebugMessage")
	assert(processor.AddService("abc", NewService(), "DebugMessage")).
		Equals(NewRPCErrorByDebug(
			"Service name \"abc\" is duplicated",
			"Current:\n\tDebugMessage\nConflict:\n\tDebugMessage",
		))
}

func BenchmarkRpcProcessor_Execute(b *testing.B) {
	total := uint64(0)
	success := uint64(0)
	failed := uint64(0)
	processor := NewRPCProcessor(
		true,
		8192*24,
		16,
		16,
		&testFuncCache{},
	)
	processor.Start(
		func(stream *RPCStream, ok bool) {
			if ok {
				atomic.AddUint64(&success, 1)
			} else {
				atomic.AddUint64(&failed, 1)
			}
			stream.Release()
		},
		func(v interface{}, debug string) {
			//
		},
	)
	_ = processor.AddService(
		"user",
		NewService().
			Reply("sayHello", func(
				ctx *RPCContext,
				name string,
			) *RPCReturn {
				return ctx.OK(name)
			}),
		"",
	)
	//file, _ := os.Create("../cpu.prof")
	//_ = pprof.StartCPUProfile(file)

	b.ReportAllocs()
	b.N = 100000000
	b.SetParallelism(1024)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stream := NewRPCStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("world")
			atomic.AddUint64(&total, 1)
			if !processor.PutStream(stream) {
				time.Sleep(10 * time.Millisecond)
			}
		}
	})
	b.StopTimer()

	time.Sleep(time.Second)

	//pprof.StopCPUProfile()

	fmt.Println(processor.Stop())
	fmt.Println(total, success, failed)
}
