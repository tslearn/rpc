package internal

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewProcessor(t *testing.T) {
	assert := NewAssert(t)

	// Test(1) onReturnStream is nil
	assert(NewProcessor(true, 1, 1, 1, nil, 5*time.Second, nil, nil)).IsNil()

	// Test(2) numOfThreads <= 0
	helper2 := newTestProcessorReturnHelper()
	assert(
		NewProcessor(true, 0, 1, 1, nil, 5*time.Second, nil, helper2.GetFunction()),
	).IsNil()
	_, _, panicArray2 := helper2.GetReturn()
	assert(len(panicArray2)).Equals(1)
	assert(panicArray2[0].GetKind()).Equals(ErrorKindKernelPanic)
	assert(panicArray2[0].GetMessage()).Equals("rpc: numOfThreads is wrong")
	assert(strings.Contains(panicArray2[0].GetDebug(), "NewProcessor")).IsTrue()

	// Test(3) maxNodeDepth <= 0
	helper3 := newTestProcessorReturnHelper()
	assert(
		NewProcessor(true, 1, 0, 1, nil, 5*time.Second, nil, helper3.GetFunction()),
	).IsNil()
	_, _, panicArray3 := helper3.GetReturn()
	assert(len(panicArray3)).Equals(1)
	assert(panicArray3[0].GetKind()).Equals(ErrorKindKernelPanic)
	assert(panicArray3[0].GetMessage()).Equals("rpc: maxNodeDepth is wrong")
	assert(strings.Contains(panicArray3[0].GetDebug(), "NewProcessor")).IsTrue()

	// Test(4) maxCallDepth <= 0
	helper4 := newTestProcessorReturnHelper()
	assert(
		NewProcessor(true, 1, 1, 0, nil, 5*time.Second, nil, helper4.GetFunction()),
	).IsNil()
	_, _, panicArray4 := helper4.GetReturn()
	assert(len(panicArray4)).Equals(1)
	assert(panicArray4[0].GetKind()).Equals(ErrorKindKernelPanic)
	assert(panicArray4[0].GetMessage()).Equals("rpc: maxCallDepth is wrong")
	assert(strings.Contains(panicArray4[0].GetDebug(), "NewProcessor")).IsTrue()

	// Test(5) mount service error
	helper5 := newTestProcessorReturnHelper()
	assert(NewProcessor(
		true,
		1,
		1,
		1,
		nil,
		5*time.Second,
		[]*rpcChildMeta{nil},
		helper5.GetFunction(),
	)).IsNil()
	_, _, panicArray5 := helper5.GetReturn()
	assert(len(panicArray5)).Equals(1)
	assert(panicArray5[0].GetKind()).Equals(ErrorKindKernelPanic)
	assert(panicArray5[0].GetMessage()).Equals("rpc: nodeMeta is nil")
	assert(strings.Contains(panicArray5[0].GetDebug(), "NewProcessor")).IsTrue()

	// Test(6) OK
	helper6 := newTestProcessorReturnHelper()
	processor6 := NewProcessor(
		true,
		65535,
		2,
		3,
		nil,
		5*time.Second,
		[]*rpcChildMeta{{
			name: "test",
			service: NewService().Reply("Eval", func(ctx Context) Return {
				time.Sleep(time.Second)
				return ctx.OK(true)
			}),
			fileLine: "",
		}},
		helper6.GetFunction(),
	)
	for i := 0; i < 65536; i++ {
		stream := NewStream()
		stream.WriteString("#.test:Eval")
		stream.WriteUint64(3)
		stream.WriteString("")
		processor6.PutStream(stream)
	}
	assert(processor6).IsNotNil()
	assert(processor6.isDebug).IsTrue()
	assert(len(processor6.repliesMap)).Equals(1)
	assert(len(processor6.servicesMap)).Equals(2)
	assert(processor6.maxNodeDepth).Equals(uint64(2))
	assert(processor6.maxCallDepth).Equals(uint64(3))
	assert(len(processor6.threads)).Equals(65536)
	assert(len(processor6.freeCHArray)).Equals(freeGroups)
	assert(processor6.readThreadPos).Equals(uint64(65536))
	assert(processor6.fnError).IsNotNil()
	processor6.Close()
	assert(processor6.writeThreadPos).Equals(uint64(65536))
	sumFrees := 0
	for _, freeCH := range processor6.freeCHArray {
		sumFrees += len(freeCH)
	}
	assert(sumFrees).Equals(65536)
}

func TestProcessor_Close(t *testing.T) {
	assert := NewAssert(t)

	// Test(1) p.panicSubscription == nil
	processor1 := getFakeProcessor(true)
	assert(processor1.Close()).IsFalse()

	// Test(2)
	replyFileLine2 := ""
	helper2 := newTestProcessorReturnHelper()
	processor2 := NewProcessor(
		true,
		1024,
		2,
		3,
		nil,
		time.Second,
		[]*rpcChildMeta{{
			name: "test",
			service: NewService().Reply("Eval", func(ctx Context) Return {
				replyFileLine2 = ctx.getThread().GetExecReplyFileLine()
				time.Sleep(2 * time.Second)
				return ctx.OK(true)
			}),
			fileLine: "",
		}},
		helper2.GetFunction(),
	)
	for i := 0; i < 1; i++ {
		stream := NewStream()
		stream.WriteString("#.test:Eval")
		stream.WriteUint64(3)
		stream.WriteString("")
		processor2.PutStream(stream)
	}
	assert(processor2.Close()).IsFalse()
	assert(helper2.GetReturn()).Equals([]Any{}, []Error{}, []Error{
		NewReplyPanic(
			"rpc: the following replies can not close: \n\t" +
				replyFileLine2 + " (1 goroutine)",
		),
	})

	// Test(3)
	replyFileLine3 := ""
	helper3 := newTestProcessorReturnHelper()
	processor3 := NewProcessor(
		true,
		1024,
		2,
		3,
		nil,
		time.Second,
		[]*rpcChildMeta{{
			name: "test",
			service: NewService().Reply("Eval", func(ctx Context) Return {
				replyFileLine3 = ctx.getThread().GetExecReplyFileLine()
				time.Sleep(2 * time.Second)
				return ctx.OK(true)
			}),
			fileLine: "",
		}},
		helper3.GetFunction(),
	)
	for i := 0; i < 2; i++ {
		stream := NewStream()
		stream.WriteString("#.test:Eval")
		stream.WriteUint64(3)
		stream.WriteString("")
		processor3.PutStream(stream)
	}
	assert(processor3.Close()).IsFalse()
	assert(helper3.GetReturn()).Equals([]Any{}, []Error{}, []Error{
		NewReplyPanic(
			"rpc: the following replies can not close: \n\t" +
				replyFileLine3 + " (2 goroutines)",
		),
	})
}

func TestProcessor_Start_Stop(t *testing.T) {
	//assert := NewAssert(t)
	//
	//processor := NewProcessor(true, 8192, 16, 32, nil, nil)
	//assert(processor.Close()).IsNotNil()
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
	//assert(processor.Close()).IsNil()
	//assert(processor.isRunning).IsFalse()
	//for i := 0; i < len(processor.threadPools); i++ {
	//	assert(processor.threadPools[i]).IsNil()
	//}
	//assert(processor.Close()).IsNotNil()
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
//	assert(processor.repliesMap["$:testOK"].meta.name).Equals("testOK")
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
		&testFuncCache{},
		5*time.Second,
		[]*rpcChildMeta{&rpcChildMeta{
			name: "user",
			service: NewService().
				Reply("sayHello", func(ctx Context, name String) Return {
					return ctx.OK(name)
				}),
			fileLine: "",
		}},
		func(stream *Stream) {
			stream.SetReadPosToBodyStart()

			if kind, ok := stream.ReadUint64(); ok && kind == uint64(ErrorKindNone) {
				atomic.AddUint64(&success, 1)
			} else {
				atomic.AddUint64(&failed, 1)
			}

			stream.Release()
		},
	)

	b.ReportAllocs()
	b.N = 100000000
	b.SetParallelism(1024)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stream := NewStream()
			stream.WriteString("#.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("")
			stream.WriteString("")
			atomic.AddUint64(&total, 1)
			processor.PutStream(stream)
		}
	})

	fmt.Println(processor.Close())
	fmt.Println(total, success, failed)
}
