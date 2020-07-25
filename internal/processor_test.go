package internal

import (
	"fmt"
	"sync/atomic"
	"testing"
)

func TestNewProcessor(t *testing.T) {
	//assert := NewAssert(t)
	//
	//processor := getNewProcessor()
	//assert(processor).IsNotNil()
	//assert(len(processor.repliesMap)).Equals(0)
	//assert(len(processor.servicesMap)).Equals(1)
	//assert(processor.isDebug).IsTrue()
	//assert(processor.maxNodeDepth).Equals(uint64(16))
	//assert(processor.maxCallDepth).Equals(uint64(32))
	//assert(processor.Stop()).IsNotNil()
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

//
//func TestProcessor_AddService(t *testing.T) {
//	assert := NewAssert(t)
//
//	processor := getNewProcessor()
//	assert(processor.AddService("test", nil, "DebugMessage")).
//		Equals(NewProtocolError("Service is nil").AddDebug("DebugMessage"))
//
//	service := NewService()
//	assert(processor.AddService("test", service, "")).IsNil()
//}
//
//func TestProcessor_BuildCache(t *testing.T) {
//	assert := NewAssert(t)
//	_, file, _, _ := runtime.Caller(0)
//
//	processor0 := getNewProcessor()
//	assert(processor0.BuildCache(
//		"pkgName",
//		path.Join(path.Dir(file), "_tmp_/processor-build-cache-0.go"),
//	)).IsNil()
//	assert(readStringFromFile(
//		path.Join(path.Dir(file), "_snapshot_/processor-build-cache-0.snapshot"),
//	)).Equals(readStringFromFile(
//		path.Join(path.Dir(file), "_tmp_/processor-build-cache-0.go")))
//
//	processor1 := getNewProcessor()
//	_ = processor1.AddService("abc", NewService().
//		Reply("sayHello", func(ctx *ContextObject, name string) *ReturnObject {
//			return ctx.OK("hello " + name)
//		}), "")
//	assert(processor1.BuildCache(
//		"pkgName",
//		path.Join(path.Dir(file), "_tmp_/processor-build-cache-1.go"),
//	)).IsNil()
//	assert(readStringFromFile(
//		path.Join(path.Dir(file), "_snapshot_/processor-build-cache-1.snapshot"),
//	)).Equals(readStringFromFile(
//		path.Join(path.Dir(file), "_tmp_/processor-build-cache-1.go")))
//
//	_ = os.RemoveAll(path.Join(path.Dir(file), "_tmp_"))
//}
//
//func TestProcessor_mountNode(t *testing.T) {
//	assert := NewAssert(t)
//
//	processor := getNewProcessor()
//
//	assert(processor.mountNode(rootName, nil).GetMessage()).
//		Equals("rpc: mountNode: nodeMeta is nil")
//	assert(processor.mountNode(rootName, nil).GetDebug()).
//		Equals("")
//
//	assert(processor.mountNode(rootName, &rpcChildMeta{
//		name:     "+",
//		service:  NewService(),
//		fileLine: "DebugMessage",
//	})).Equals(NewBaseError(
//		"Service name \"+\" is illegal",
//	).AddDebug("DebugMessage"))
//
//	assert(processor.mountNode(rootName, &rpcChildMeta{
//		name:     "abc",
//		service:  nil,
//		fileLine: "DebugMessage",
//	})).Equals(NewBaseError(
//		"Service is nil",
//	).AddDebug("DebugMessage"))
//
//	assert(processor.mountNode("123", &rpcChildMeta{
//		name:     "abc",
//		service:  NewService(),
//		fileLine: "DebugMessage",
//	})).Equals(NewBaseError(
//		"rpc: mountNode: parentNode is nil",
//	).AddDebug("DebugMessage"))
//
//	processor.maxNodeDepth = 0
//	assert(processor.mountNode(rootName, &rpcChildMeta{
//		name:     "abc",
//		service:  NewService(),
//		fileLine: "DebugMessage",
//	})).Equals(NewBaseError(
//		"Service path depth $.abc is too long, it must be less or equal than 0",
//	).AddDebug("DebugMessage"))
//	processor.maxNodeDepth = 16
//
//	_ = processor.mountNode(rootName, &rpcChildMeta{
//		name:     "abc",
//		service:  NewService(),
//		fileLine: "DebugMessage",
//	})
//	assert(processor.mountNode(rootName, &rpcChildMeta{
//		name:     "abc",
//		service:  NewService(),
//		fileLine: "DebugMessage",
//	})).Equals(NewBaseError(
//		"Service name \"abc\" is duplicated",
//	).AddDebug("Current:\n\tDebugMessage\nConflict:\n\tDebugMessage"))
//
//	// mount reply error
//	service := NewService()
//	service.Reply("abc", nil)
//	assert(processor.mountNode(rootName, &rpcChildMeta{
//		name:     "test",
//		service:  service,
//		fileLine: "DebugMessage",
//	}).GetMessage()).Equals("Reply handler is nil")
//
//	// mount children error
//	service1 := NewService()
//	service1.AddChildService("abc", NewService())
//	assert(len(service1.children)).Equals(1)
//	service1.children[0] = nil
//	assert(processor.mountNode(rootName, &rpcChildMeta{
//		name:     "003",
//		service:  service1,
//		fileLine: "DebugMessage",
//	}).GetMessage()).Equals("rpc: mountNode: nodeMeta is nil")
//
//	// OK
//	service2 := NewService()
//	service2.AddChildService("user", NewService().
//		Reply("sayHello", func(ctx *ContextObject) *ReturnObject {
//			return ctx.OK(true)
//		}))
//	assert(processor.mountNode(rootName, &rpcChildMeta{
//		name:     "system",
//		service:  service2,
//		fileLine: "DebugMessage",
//	})).IsNil()
//}
//
//func TestProcessor_mountReply(t *testing.T) {
//	assert := NewAssert(t)
//
//	processor := NewProcessor(
//		true,
//		8192,
//		16,
//		16,
//		&testFuncCache{},
//	)
//	rootNode := processor.servicesMap[rootName]
//
//	// check the node is nil
//	assert(processor.mountReply(nil, nil)).Equals(NewBaseError(
//		"rpc: mountReply: node is nil",
//	))
//
//	// check the rpcReplyMeta is nil
//	assert(processor.mountReply(rootNode, nil)).Equals(NewBaseError(
//		"rpc: mountReply: rpcReplyMeta is nil",
//	))
//
//	// check the name
//	assert(processor.mountReply(rootNode, &rpcReplyMeta{
//		name:     "###",
//		handler:  nil,
//		fileLine: "DebugMessage",
//	})).Equals(NewBaseError(
//		"Reply name ### is illegal",
//	).AddDebug("DebugMessage"))
//
//	// check the reply path is not occupied
//	_ = processor.mountReply(rootNode, &rpcReplyMeta{
//		name:     "testOccupied",
//		handler:  func(ctx *ContextObject) *ReturnObject { return ctx.OK(true) },
//		fileLine: "DebugMessage",
//	})
//	assert(processor.mountReply(rootNode, &rpcReplyMeta{
//		name:     "testOccupied",
//		handler:  func(ctx *ContextObject) *ReturnObject { return ctx.OK(true) },
//		fileLine: "DebugMessage",
//	})).Equals(NewBaseError(
//		"Reply name testOccupied is duplicated",
//	).AddDebug("Current:\n\tDebugMessage\nConflict:\n\tDebugMessage"))
//
//	// check the reply handler is nil
//	assert(processor.mountReply(rootNode, &rpcReplyMeta{
//		name:     "testReplyHandlerIsNil",
//		handler:  nil,
//		fileLine: "DebugMessage",
//	})).Equals(NewBaseError(
//		"Reply handler is nil",
//	).AddDebug("DebugMessage"))
//
//	// Check reply handler is Func
//	assert(processor.mountReply(rootNode, &rpcReplyMeta{
//		name:     "testReplyHandlerIsFunction",
//		handler:  make(chan bool),
//		fileLine: "DebugMessage",
//	})).Equals(NewBaseError(
//		"Reply handler must be func(ctx rpc.ContextObject, ...) rpc.ReturnObject",
//	).AddDebug("DebugMessage"))
//
//	// Check reply handler arguments types
//	assert(processor.mountReply(rootNode, &rpcReplyMeta{
//		name:     "testReplyHandlerArguments",
//		handler:  func(ctx bool) *ReturnObject { return nilReturn },
//		fileLine: "DebugMessage",
//	})).Equals(NewBaseError(
//		"Reply handler 1st argument type must be rpc.ContextObject",
//	).AddDebug("DebugMessage"))
//
//	assert(processor.mountReply(rootNode, &rpcReplyMeta{
//		name:     "testReplyHandlerArguments",
//		handler:  func(ctx *ContextObject, ch chan bool) *ReturnObject { return nilReturn },
//		fileLine: "DebugMessage",
//	})).Equals(NewBaseError(
//		"Reply handler 2nd argument type <chan bool> not supported",
//	).AddDebug("DebugMessage"))
//
//	// Check return type
//	assert(processor.mountReply(rootNode, &rpcReplyMeta{
//		name:     "testReplyHandlerReturn",
//		handler:  func(ctx *ContextObject) (*ReturnObject, bool) { return nilReturn, true },
//		fileLine: "DebugMessage",
//	})).Equals(NewBaseError(
//		"Reply handler return type must be rpc.ReturnObject",
//	).AddDebug("DebugMessage"))
//
//	assert(processor.mountReply(rootNode, &rpcReplyMeta{
//		name:     "testReplyHandlerReturn",
//		handler:  func(ctx ContextObject) bool { return true },
//		fileLine: "DebugMessage",
//	})).Equals(NewBaseError(
//		"Reply handler return type must be rpc.ReturnObject",
//	).AddDebug("DebugMessage"))
//
//	// ok
//	assert(processor.mountReply(rootNode, &rpcReplyMeta{
//		name:     "testOK",
//		handler:  func(ctx *ContextObject, _ bool, _ Map) *ReturnObject { return nilReturn },
//		fileLine: AddFileLine("", 0),
//	})).IsNil()
//
//	assert(processor.repliesMap["$:testOK"].replyMeta.name).Equals("testOK")
//	assert(processor.repliesMap["$:testOK"].reflectFn).IsNotNil()
//	assert(processor.repliesMap["$:testOK"].callString).
//		Equals("$:testOK(rpc.ContextObject, rpc.Bool, rpc.Map) rpc.ReturnObject")
//	assert(
//		strings.Contains(processor.repliesMap["$:testOK"].GetDebug(), "$:testOK"),
//	).IsTrue()
//	assert(processor.repliesMap["$:testOK"].argTypes[0]).
//		Equals(reflect.ValueOf(nilContext).Type())
//	assert(processor.repliesMap["$:testOK"].argTypes[1]).Equals(boolType)
//	assert(processor.repliesMap["$:testOK"].argTypes[2]).Equals(mapType)
//	assert(processor.repliesMap["$:testOK"].indicator).IsNotNil()
//}
//
//func TestProcessor_OutPutErrors(t *testing.T) {
//	assert := NewAssert(t)
//
//	processor := getNewProcessor()
//
//	// Service is nil
//	assert(processor.AddService("", nil, "DebugMessage")).
//		Equals(NewBaseError(
//			"Service is nil",
//		).AddDebug("DebugMessage"))
//
//	assert(processor.AddService("abc", (*Service)(nil), "DebugMessage")).
//		Equals(NewBaseError(
//			"Service is nil",
//		).AddDebug("DebugMessage"))
//
//	// Service name %s is illegal
//	assert(processor.AddService("\"\"", NewService(), "DebugMessage")).
//		Equals(NewBaseError(
//			"Service name \"\"\"\" is illegal",
//		).AddDebug("DebugMessage"))
//
//	processor.maxNodeDepth = 0
//	assert(processor.AddService("abc", NewService(), "DebugMessage")).
//		Equals(NewBaseError(
//			"Service path depth $.abc is too long, it must be less or equal than 0",
//		).AddDebug("DebugMessage"))
//	processor.maxNodeDepth = 16
//
//	_ = processor.AddService("abc", NewService(), "DebugMessage")
//	assert(processor.AddService("abc", NewService(), "DebugMessage")).
//		Equals(NewBaseError(
//			"Service name \"abc\" is duplicated",
//		).AddDebug("Current:\n\tDebugMessage\nConflict:\n\tDebugMessage"))
//}

func BenchmarkRpcProcessor_Execute(b *testing.B) {
	total := uint64(0)
	success := uint64(0)
	failed := uint64(0)
	processor := NewProcessor(
		false,
		8192*24,
		16,
		16,
		nil,
	)
	_ = processor.Start(
		func(stream *Stream) {
			if stream.GetStreamKind() == StreamKindResponseOK {
				atomic.AddUint64(&success, 1)
			} else {
				stream.SetReadPosToBodyStart()
				fmt.Println(stream.ReadUint64())
				fmt.Println(stream.ReadString())
				fmt.Println(stream.ReadString())
				atomic.AddUint64(&failed, 1)
			}
			stream.Release()
		},
	)
	_ = processor.AddService(
		"user",
		NewService().
			Reply("sayHello", func(ctx Context, name String) Return {
				return ctx.OK(name)
			}),
		"",
	)

	b.ReportAllocs()
	b.N = 50000000
	b.SetParallelism(1024)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stream := NewStream()
			stream.SetStreamKind(StreamKindRequest)
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("")
			atomic.AddUint64(&total, 1)
			processor.PutStream(stream)
		}
	})

	fmt.Println(processor.Stop())
	fmt.Println(total, success, failed)
}
