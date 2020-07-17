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

func TestNewProcessor(t *testing.T) {
	assert := NewAssert(t)

	processor := getNewProcessor()
	assert(processor).IsNotNil()
	assert(len(processor.repliesMap)).Equals(0)
	assert(len(processor.servicesMap)).Equals(1)
	assert(processor.isDebug).IsTrue()
	assert(processor.maxNodeDepth).Equals(uint64(16))
	assert(processor.maxCallDepth).Equals(uint64(32))
	assert(processor.Stop()).IsNotNil()
}

func TestProcessor_Start_Stop(t *testing.T) {
	//assert := NewAssert(t)
	//
	//processor := NewProcessor(true, 8192, 16, 32, nil, nil)
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

func TestProcessor_PutStream(t *testing.T) {
	//assert := NewAssert(t)
	//processor := NewProcessor(16, 32, nil, nil)
	//assert(processor.PutStream(NewStream())).IsFalse()
	//processor.Start()
	//assert(processor.PutStream(NewStream())).IsTrue()
	//for i := 0; i < len(processor.threadPools); i++ {
	//	processor.threadPools[i].stop()
	//}
	//assert(processor.PutStream(NewStream())).IsFalse()
}

func TestProcessor_AddService(t *testing.T) {
	assert := NewAssert(t)

	processor := getNewProcessor()
	assert(processor.AddService("test", nil, "DebugMessage")).
		Equals(NewError("Service is nil").AddDebug("DebugMessage"))

	service := NewService()
	assert(processor.AddService("test", service, "")).IsNil()
}

func TestProcessor_BuildCache(t *testing.T) {
	assert := NewAssert(t)
	_, file, _, _ := runtime.Caller(0)

	processor0 := getNewProcessor()
	assert(processor0.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/processor-build-cache-0.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/processor-build-cache-0.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/processor-build-cache-0.go")))

	processor1 := getNewProcessor()
	_ = processor1.AddService("abc", NewService().
		Reply("sayHello", func(ctx *Context, name string) *Return {
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

func TestProcessor_mountNode(t *testing.T) {
	assert := NewAssert(t)

	processor := getNewProcessor()

	assert(processor.mountNode(rootName, nil).GetMessage()).
		Equals("rpc: mountNode: nodeMeta is nil")
	assert(processor.mountNode(rootName, nil).GetDebug()).
		Equals("")

	assert(processor.mountNode(rootName, &rpcChildMeta{
		name:    "+",
		service: NewService(),
		debug:   "DebugMessage",
	})).Equals(NewError(
		"Service name \"+\" is illegal",
	).AddDebug("DebugMessage"))

	assert(processor.mountNode(rootName, &rpcChildMeta{
		name:    "abc",
		service: nil,
		debug:   "DebugMessage",
	})).Equals(NewError(
		"Service is nil",
	).AddDebug("DebugMessage"))

	assert(processor.mountNode("123", &rpcChildMeta{
		name:    "abc",
		service: NewService(),
		debug:   "DebugMessage",
	})).Equals(NewError(
		"rpc: mountNode: parentNode is nil",
	).AddDebug("DebugMessage"))

	processor.maxNodeDepth = 0
	assert(processor.mountNode(rootName, &rpcChildMeta{
		name:    "abc",
		service: NewService(),
		debug:   "DebugMessage",
	})).Equals(NewError(
		"Service path depth $.abc is too long, it must be less or equal than 0",
	).AddDebug("DebugMessage"))
	processor.maxNodeDepth = 16

	_ = processor.mountNode(rootName, &rpcChildMeta{
		name:    "abc",
		service: NewService(),
		debug:   "DebugMessage",
	})
	assert(processor.mountNode(rootName, &rpcChildMeta{
		name:    "abc",
		service: NewService(),
		debug:   "DebugMessage",
	})).Equals(NewError(
		"Service name \"abc\" is duplicated",
	).AddDebug("Current:\n\tDebugMessage\nConflict:\n\tDebugMessage"))

	// mount reply error
	service := NewService()
	service.Reply("abc", nil)
	assert(processor.mountNode(rootName, &rpcChildMeta{
		name:    "test",
		service: service,
		debug:   "DebugMessage",
	}).GetMessage()).Equals("Reply handler is nil")

	// mount children error
	service1 := NewService()
	service1.AddChild("abc", NewService())
	assert(len(service1.children)).Equals(1)
	service1.children[0] = nil
	assert(processor.mountNode(rootName, &rpcChildMeta{
		name:    "003",
		service: service1,
		debug:   "DebugMessage",
	}).GetMessage()).Equals("rpc: mountNode: nodeMeta is nil")

	// OK
	service2 := NewService()
	service2.AddChild("user", NewService().
		Reply("sayHello", func(ctx *Context) *Return {
			return ctx.OK(true)
		}))
	assert(processor.mountNode(rootName, &rpcChildMeta{
		name:    "system",
		service: service2,
		debug:   "DebugMessage",
	})).IsNil()
}

func TestProcessor_mountReply(t *testing.T) {
	assert := NewAssert(t)

	processor := NewProcessor(
		true,
		8192,
		16,
		16,
		&testFuncCache{},
		func(tag string, err Error) {

		},
		func(v interface{}, debug string) {

		},
	)
	rootNode := processor.servicesMap[rootName]

	// check the node is nil
	assert(processor.mountReply(nil, nil)).Equals(NewError(
		"rpc: mountReply: node is nil",
	))

	// check the rpcReplyMeta is nil
	assert(processor.mountReply(rootNode, nil)).Equals(NewError(
		"rpc: mountReply: rpcReplyMeta is nil",
	))

	// check the name
	assert(processor.mountReply(rootNode, &rpcReplyMeta{
		"###",
		nil,
		"DebugMessage",
	})).Equals(NewError(
		"Reply name ### is illegal",
	).AddDebug("DebugMessage"))

	// check the reply path is not occupied
	_ = processor.mountReply(rootNode, &rpcReplyMeta{
		"testOccupied",
		func(ctx *Context) *Return { return ctx.OK(true) },
		"DebugMessage",
	})
	assert(processor.mountReply(rootNode, &rpcReplyMeta{
		"testOccupied",
		func(ctx *Context) *Return { return ctx.OK(true) },
		"DebugMessage",
	})).Equals(NewError(
		"Reply name testOccupied is duplicated",
	).AddDebug("Current:\n\tDebugMessage\nConflict:\n\tDebugMessage"))

	// check the reply handler is nil
	assert(processor.mountReply(rootNode, &rpcReplyMeta{
		"testReplyHandlerIsNil",
		nil,
		"DebugMessage",
	})).Equals(NewError(
		"Reply handler is nil",
	).AddDebug("DebugMessage"))

	// Check reply handler is Func
	assert(processor.mountReply(rootNode, &rpcReplyMeta{
		"testReplyHandlerIsFunction",
		make(chan bool),
		"DebugMessage",
	})).Equals(NewError(
		"Reply handler must be func(ctx rpc.Context, ...) rpc.Return",
	).AddDebug("DebugMessage"))

	// Check reply handler arguments types
	assert(processor.mountReply(rootNode, &rpcReplyMeta{
		"testReplyHandlerArguments",
		func(ctx bool) *Return { return nilReturn },
		"DebugMessage",
	})).Equals(NewError(
		"Reply handler 1st argument type must be rpc.Context",
	).AddDebug("DebugMessage"))

	assert(processor.mountReply(rootNode, &rpcReplyMeta{
		"testReplyHandlerArguments",
		func(ctx *Context, ch chan bool) *Return { return nilReturn },
		"DebugMessage",
	})).Equals(NewError(
		"Reply handler 2nd argument type <chan bool> not supported",
	).AddDebug("DebugMessage"))

	// Check return type
	assert(processor.mountReply(rootNode, &rpcReplyMeta{
		"testReplyHandlerReturn",
		func(ctx *Context) (*Return, bool) { return nilReturn, true },
		"DebugMessage",
	})).Equals(NewError(
		"Reply handler return type must be rpc.Return",
	).AddDebug("DebugMessage"))

	assert(processor.mountReply(rootNode, &rpcReplyMeta{
		"testReplyHandlerReturn",
		func(ctx Context) bool { return true },
		"DebugMessage",
	})).Equals(NewError(
		"Reply handler return type must be rpc.Return",
	).AddDebug("DebugMessage"))

	// ok
	assert(processor.mountReply(rootNode, &rpcReplyMeta{
		"testOK",
		func(ctx *Context, _ bool, _ Map) *Return { return nilReturn },
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

func TestProcessor_OutPutErrors(t *testing.T) {
	assert := NewAssert(t)

	processor := getNewProcessor()

	// Service is nil
	assert(processor.AddService("", nil, "DebugMessage")).
		Equals(NewError(
			"Service is nil",
		).AddDebug("DebugMessage"))

	assert(processor.AddService("abc", (*Service)(nil), "DebugMessage")).
		Equals(NewError(
			"Service is nil",
		).AddDebug("DebugMessage"))

	// Service name %s is illegal
	assert(processor.AddService("\"\"", NewService(), "DebugMessage")).
		Equals(NewError(
			"Service name \"\"\"\" is illegal",
		).AddDebug("DebugMessage"))

	processor.maxNodeDepth = 0
	assert(processor.AddService("abc", NewService(), "DebugMessage")).
		Equals(NewError(
			"Service path depth $.abc is too long, it must be less or equal than 0",
		).AddDebug("DebugMessage"))
	processor.maxNodeDepth = 16

	_ = processor.AddService("abc", NewService(), "DebugMessage")
	assert(processor.AddService("abc", NewService(), "DebugMessage")).
		Equals(NewError(
			"Service name \"abc\" is duplicated",
		).AddDebug("Current:\n\tDebugMessage\nConflict:\n\tDebugMessage"))
}

func BenchmarkRpcProcessor_Execute(b *testing.B) {
	total := uint64(0)
	success := uint64(0)
	failed := uint64(0)
	processor := NewProcessor(
		true,
		8192*24,
		16,
		16,
		&testFuncCache{},
		func(tag string, err Error) {

		},
		func(v interface{}, debug string) {

		},
	)
	_ = processor.Start(
		func(stream *Stream, ok bool) {
			if ok {
				atomic.AddUint64(&success, 1)
			} else {
				atomic.AddUint64(&failed, 1)
			}
			stream.Release()
		},
	)
	_ = processor.AddService(
		"user",
		NewService().
			Reply("sayHello", func(
				ctx *Context,
				name String,
			) *Return {
				return ctx.OK(name)
			}),
		"",
	)

	time.Sleep(3 * time.Second)
	b.ReportAllocs()
	b.N = 50000000
	b.SetParallelism(1024)

	// file, _ := os.Create("../cpu.prof")
	// _ = pprof.StartCPUProfile(file)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("")
			atomic.AddUint64(&total, 1)
			processor.PutStream(stream)
		}
	})
	b.StopTimer()
	// pprof.StopCPUProfile()

	time.Sleep(3 * time.Second)

	sumFrees := 0
	for _, ch := range processor.freeThreadsCHGroup {
		sumFrees += len(ch)
	}
	fmt.Println(sumFrees)
	fmt.Println(processor.Stop())
	fmt.Println(total, success, failed)
}
