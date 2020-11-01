package core

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"testing"
	"time"
	"unsafe"
)

var (
	fnEvalBack    = func(stream *Stream) {}
	fnEvalFinish  = func(thread *rpcThread) {}
	testProcessor = getFakeProcessor(true)
)

func TestNewRPCThreadFrame(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 100; i++ {
			v := newRPCThreadFrame()
			assert(v).IsNotNil()
			v.Release()
		}
	})
}

func TestRpcThreadFrame_Reset(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := newRPCThreadFrame()
		v.stream = NewStream()
		v.actionNode = unsafe.Pointer(&rpcActionNode{})
		v.from = "#"
		v.depth = 13
		v.cacheArrayItemsPos = 16
		v.cacheMapItemsPos = 16
		v.cacheArrayEntryPos = 7
		v.cacheMapEntryPos = 7
		v.retStatus = 1
		v.lockStatus = 82737243243
		v.parentRTWritePos = 100
		v.next = &rpcThreadFrame{}
		v.Reset()
		assert(v.stream).Equal(nil)
		assert(v.actionNode).Equal(nil)
		assert(v.from).Equal("")
		assert(v.depth).Equal(uint16(13))
		assert(v.cacheArrayItemsPos).Equal(uint32(0))
		assert(v.cacheMapItemsPos).Equal(uint32(0))
		assert(v.cacheArrayEntryPos).Equal(uint32(0))
		assert(v.cacheMapEntryPos).Equal(uint32(0))
		assert(v.retStatus).Equal(uint32(1))
		assert(v.lockStatus).Equal(uint64(82737243243))
		assert(v.parentRTWritePos).Equal(streamPosBody)
		assert(v.next).Equal(nil)
		v.Release()
	})
}

func TestRpcThreadFrame_Release(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		mp := map[string]bool{}
		for i := 0; i < 1000; i++ {
			v := newRPCThreadFrame()
			assert(v.cacheArrayItemsPos).Equal(uint32(0))
			v.cacheArrayItemsPos = 32
			mp[fmt.Sprintf("%p", v)] = true
			v.Release()
		}
		assert(len(mp) < 1000).IsTrue()
	})
}

func TestNewRTMap(t *testing.T) {
	t.Run("thread is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(newRTMap(Runtime{}, 2)).Equal(RTMap{})
	})

	t.Run("size is less than zero", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		assert(newRTMap(testRuntime, -1)).Equal(RTMap{})
	})

	t.Run("size is zero", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := newRTMap(testRuntime, 0)
		assert(v.rt).Equal(testRuntime)
		assert(v.items != nil)
		assert(len(*v.items), cap(*v.items), *v.length).Equal(0, 0, uint32(0))
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 0; i < 100; i++ {
			testRuntime.thread.Reset()
			v := newRTMap(testRuntime, i)
			assert(v.rt).Equal(testRuntime)
			assert(len(*v.items), cap(*v.items), *v.length).Equal(0, i, uint32(0))
		}
	})
}

func TestRTArrayNewRTArray(t *testing.T) {
	t.Run("thread is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(newRTArray(Runtime{}, 2)).Equal(RTArray{})
	})

	t.Run("size is less than zero", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		assert(newRTArray(testRuntime, -1)).Equal(RTArray{})
	})

	t.Run("size is zero", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := newRTArray(testRuntime, 0)
		assert(v.rt).Equal(testRuntime)
		assert(v.items != nil)
		assert(len(*v.items), cap(*v.items)).Equal(0, 0)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 0; i < 100; i++ {
			testRuntime.thread.Reset()
			v := newRTArray(testRuntime, i)
			assert(v.rt).Equal(testRuntime)
			assert(len(*v.items), cap(*v.items)).Equal(0, i)
		}
	})
}

func TestNewThread(t *testing.T) {
	t.Run("processor is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(newThread(nil, 5*time.Second, 2048, fnEvalBack, nil)).IsNil()
	})

	t.Run("onEvalBack is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(newThread(testProcessor, 5*time.Second, 2048, nil, fnEvalFinish)).
			IsNil()
	})

	t.Run("onEvalFinish is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(newThread(testProcessor, 5*time.Second, 2048, fnEvalBack, nil)).
			IsNil()
	})

	t.Run("test ok (timeout 1s)", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 32; i++ {
			v := newThread(testProcessor, time.Second, 2048, fnEvalBack, fnEvalFinish)
			assert(v.processor).Equal(testProcessor)
			assert(v.inputCH).IsNotNil()
			assert(v.closeCH).IsNotNil()
			assert(v.closeTimeout).Equal(3 * time.Second)
			assert(v.top).Equal(&v.rootFrame)
			assert(v.sequence > 0).Equal(true)
			assert(v.sequence % 2).Equal(uint64(0))
			assert(v.rtStream).IsNotNil()
			assert(len(v.cacheEntry)).Equal(8)
			assert(len(v.cacheArrayItems), cap(v.cacheArrayItems)).Equal(128, 128)
			assert(len(v.cacheMapItems), cap(v.cacheMapItems)).Equal(32, 32)
			v.Close()
			assert(v.closeCH).IsNil()
		}
	})

	t.Run("test ok (timeout 5s)", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 32; i++ {
			v := newThread(
				testProcessor,
				5*time.Second,
				2048,
				fnEvalBack,
				fnEvalFinish,
			)
			assert(v.processor).Equal(testProcessor)
			assert(v.inputCH).IsNotNil()
			assert(v.closeCH).IsNotNil()
			assert(v.closeTimeout).Equal(5 * time.Second)
			assert(v.top).Equal(&v.rootFrame)
			assert(v.sequence > 0).Equal(true)
			assert(v.sequence % 2).Equal(uint64(0))
			assert(v.rtStream).IsNotNil()
			assert(len(v.cacheEntry)).Equal(8)
			assert(len(v.cacheArrayItems), cap(v.cacheArrayItems)).Equal(128, 128)
			assert(len(v.cacheMapItems), cap(v.cacheMapItems)).Equal(32, 32)
			v.Close()
			assert(v.closeCH).IsNil()
		}
	})

	t.Run("test eval", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 0; i < 100000; i++ {
			chBack := make(chan bool, 1)
			chFinish := make(chan bool, 1)
			v := newThread(
				testProcessor,
				5*time.Second,
				2048,
				func(stream *Stream) {
					chBack <- true
				},
				func(thread *rpcThread) {
					chFinish <- true
				},
			)
			v.PutStream(NewStream())
			assert(<-chBack).Equal(true)
			assert(<-chFinish).Equal(true)
			assert(v.sequence > 0).Equal(true)
			assert(v.sequence % 2).Equal(uint64(0))
			assert(v.top).Equal(&v.rootFrame)
			v.Close()
			assert(v.top.stream).Equal(nil)
		}
	})
}

func TestRpcThread_Reset(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := newThread(
			getFakeProcessor(true),
			5*time.Second,
			2048,
			func(stream *Stream) {},
			func(thread *rpcThread) {},
		)

		v.rtStream.Write(3)
		v.rootFrame.stream = NewStream()
		v.cacheEntry[0].mapLength = 3

		assert(v.rtStream.GetWritePos() == streamPosBody).IsFalse()
		assert(v.rootFrame.stream == nil).IsFalse()
		assert(v.cacheEntry[0].mapLength == 0).IsFalse()
		v.Reset()
		assert(v.rtStream.GetWritePos() == streamPosBody).IsTrue()
		assert(v.rootFrame.stream == nil).IsTrue()
		assert(v.cacheEntry[0].mapLength == 0).IsTrue()
		v.Close()
	})
}

func TestRpcThread_Close(t *testing.T) {
	t.Run("close twice", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := newThread(testProcessor, 3*time.Second, 2048, fnEvalBack, fnEvalFinish)
		assert(v.Close()).IsTrue()
		assert(v.Close()).IsFalse()
	})

	t.Run("cannot close", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(ParseResponseStream(
			testWithReply(false, nil, nil, func(rt Runtime, testThread bool) Return {
				if testThread {
					v := newThread(
						rt.thread.processor,
						3*time.Second,
						2048,
						fnEvalBack,
						fnEvalFinish,
					)
					s, _ := MakeRequestStream("#.test:Eval", "", false)
					v.PutStream(s)
					assert(v.Close()).IsFalse()
				} else {
					time.Sleep(3500 * time.Millisecond)
				}
				return rt.Reply(true)
			}, true),
		)).Equal(true, nil)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := newThread(testProcessor, 3*time.Second, 2048, fnEvalBack, fnEvalFinish)
		assert(v.Close()).IsTrue()
	})
}

func TestRpcThread_GetActionNode(t *testing.T) {
	t.Run("node is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := newThread(testProcessor, 3*time.Second, 2048, fnEvalBack, fnEvalFinish)
		assert(v.GetActionNode()).Equal(nil)
	})

	t.Run("node is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(ParseResponseStream(
			testWithReply(false, nil, nil, func(rt Runtime) Return {
				assert(rt.thread.GetActionNode()).IsNotNil()
				return rt.Reply(true)
			}),
		)).Equal(true, nil)
	})
}

func TestRpcThread_GetExecActionNodePath(t *testing.T) {
	t.Run("node is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := newThread(testProcessor, 3*time.Second, 2048, fnEvalBack, fnEvalFinish)
		assert(v.GetExecActionNodePath()).Equal("")
	})

	t.Run("node is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(ParseResponseStream(
			testWithReply(false, nil, nil, func(rt Runtime) Return {
				assert(rt.thread.GetExecActionNodePath()).Equal("#.test:Eval")
				return rt.Reply(true)
			}),
		)).Equal(true, nil)
	})
}

func TestRpcThread_GetExecActionDebug(t *testing.T) {
	t.Run("node is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := newThread(testProcessor, 3*time.Second, 2048, fnEvalBack, fnEvalFinish)
		assert(v.GetExecActionDebug()).Equal("")
	})

	t.Run("node is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(ParseResponseStream(
			testWithReply(false, nil, nil, func(rt Runtime) Return {
				(*rpcActionNode)(rt.thread.top.actionNode).meta.fileLine = "/file:001"
				assert(rt.thread.GetExecActionDebug()).Equal("#.test:Eval /file:001")
				return rt.Reply(true)
			}),
		)).Equal(true, nil)
	})
}

func TestRpcThread_Write(t *testing.T) {
	t.Run("")

}

//func TestRpcThread_WriteError(t *testing.T) {
//	assert := base.NewAssert(t)
//
//	// Test(1) ok
//	source1 := ""
//	assert(testRunWithProcessor(true, nil,
//		func(rt Runtime, name string) Return {
//			ret, source := rt.Error(errors.New("error")), base.GetFileLine(0)
//			source1 = source
//			return ret
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.WriteString("world")
//			return stream
//		},
//		nil,
//	)).Equal(nil, base.NewActionError("error").AddDebug("#.test:Eval "+source1), nil)
//}
//
//func TestRpcThread_WriteOK(t *testing.T) {
//	assert := base.NewAssert(t)
//
//	// Test(1) value is endless loop
//	source1 := ""
//	assert(testRunWithProcessor(true, nil,
//		func(rt Runtime, name string) Return {
//			v := make(Map)
//			v["v"] = v
//			ret, source := rt.OK(v), base.GetFileLine(0)
//			source1 = source
//			return ret
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.WriteString("world")
//			return stream
//		},
//		nil,
//	)).Equal(
//		nil,
//		nil,
//		base.NewActionPanic("value[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"]"+
//			"[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"]"+
//			"[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"]"+
//			"[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"]"+
//			"[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"]"+
//			"[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"]"+
//			"[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"] write overflow").
//			AddDebug("#.test:Eval "+source1),
//	)
//
//	// Test(2) value is not support
//	source2 := ""
//	assert(testRunWithProcessor(true, nil,
//		func(rt Runtime, name string) Return {
//			ret, source := rt.OK(make(chan bool)), base.GetFileLine(0)
//			source2 = source
//			return ret
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.WriteString("world")
//			return stream
//		},
//		nil,
//	)).Equal(
//		nil,
//		nil,
//		base.NewActionPanic("value type is not supported").
//			AddDebug("#.test:Eval "+source2),
//	)
//
//	// Test(3) ok
//	assert(testRunWithProcessor(true, nil,
//		func(rt Runtime, name string) Return {
//			return rt.OK("hello " + name)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.WriteString("world")
//			return stream
//		},
//		nil,
//	)).Equal("hello world", nil, nil)
//}
//
//func TestRpcThread_PutStream(t *testing.T) {
//	assert := base.NewAssert(t)
//
//	// Test(1)
//	thread1 := getFakeThread(true)
//	defer thread1.Close()
//	assert(thread1.PutStream(NewStream())).IsTrue()
//
//	// Test(2)
//	thread2 := getFakeThread(true)
//	thread2.Close()
//	assert(thread2.PutStream(NewStream())).IsFalse()
//
//	// Test(3)
//	thread3 := getFakeThread(true)
//	thread3.Close()
//	atomic.StorePointer(&thread3.closeCH, unsafe.Pointer(thread3))
//	assert(thread3.PutStream(NewStream())).IsFalse()
//}
//
//func TestRpcThread_Eval1(t *testing.T) {
//	assert := base.NewAssert(t)
//
//	// Test(1) read action path error
//	assert(testRunWithProcessor(true, nil,
//		func(rt Runtime, name string) Return {
//			return rt.OK("hello " + name)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			// path format error
//			stream.WriteBytes([]byte("#.test:Eval"))
//			stream.WriteString("#")
//			stream.WriteString("world")
//			return stream
//		},
//		nil,
//	)).Equal(nil, base.NewProtocolError(base.ErrStringBadStream), nil)
//}
//
//func TestRpcThread_Eval(t *testing.T) {
//	assert := base.NewAssert(t)
//
//	// Test(0) basic
//	assert(testRunWithProcessor(true, nil,
//		func(rt Runtime, name string) Return {
//			return rt.OK("hello " + name)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.WriteString("world")
//			return stream
//		},
//		nil,
//	)).Equal("hello world", nil, nil)
//
//	// Test(1) read action path error
//	assert(testRunWithProcessor(true, nil,
//		func(rt Runtime, name string) Return {
//			return rt.OK("hello " + name)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			// path format error
//			stream.WriteBytes([]byte("#.test:Eval"))
//			stream.WriteString("#")
//			stream.WriteString("world")
//			return stream
//		},
//		nil,
//	)).Equal(nil, base.NewProtocolError(base.ErrStringBadStream), nil)
//
//	// Test(2) action path is not mounted
//	assert(testRunWithProcessor(true, nil,
//		func(rt Runtime, name string) Return {
//			return rt.OK("hello " + name)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			// action path is not mounted
//			stream.WriteString("#.system:Eval")
//			stream.WriteString("#")
//			stream.WriteString("world")
//			return stream
//		},
//		nil,
//	)).Equal(nil, base.NewActionError("target #.system:Eval does not exist"), nil)
//
//	// Test(4) depth is overflow
//	ret4, error4, panic4 := testRunWithProcessor(true, nil,
//		func(rt Runtime, name string) Return {
//			return rt.OK("hello " + name)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			//  depth is overflow
//			stream.SetDepth(17)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.WriteString("world")
//			return stream
//		},
//		nil,
//	)
//	assert(ret4, panic4).IsNil()
//	assert(error4.GetKind()).Equal(base.ErrorKindAction)
//	assert(error4.GetMessage()).
//		Equal("call #.test:Eval level(17) overflows")
//	assert(strings.Contains(error4.GetDebug(), "#.test:Eval")).IsTrue()
//	assert(strings.Contains(error4.GetDebug(), "type_test.go")).IsTrue()
//
//	// Test(5) execFrom data format error
//	assert(testRunWithProcessor(true, nil,
//		func(rt Runtime, name string) Return {
//			return rt.OK("hello " + name)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			// execFrom data format error
//			stream.WriteBool(true)
//			stream.WriteString("world")
//			return stream
//		},
//		nil,
//	)).Equal(nil, base.NewProtocolError(base.ErrStringBadStream), nil)
//
//	// Test(6) ok call with all type value
//	assert(testRunWithProcessor(true, nil,
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			stream.Write(int64(3))
//			stream.Write(uint64(3))
//			stream.Write(float64(3))
//			stream.Write("hello")
//			stream.Write(([]byte)("world"))
//			stream.Write(Array{1})
//			stream.Write(Map{"name": "world"})
//			return stream
//		},
//		nil,
//	)).Equal(true, nil, nil)
//
//	// Test(7) error with 1st param
//	assert(testRunWithProcessor(false, nil,
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			// error rpc.Bool
//			stream.Write(2)
//			stream.Write(int64(3))
//			stream.Write(uint64(3))
//			stream.Write(float64(3))
//			stream.Write("hello")
//			stream.Write(([]byte)("world"))
//			stream.Write(Array{1})
//			stream.Write(Map{"name": "world"})
//			return stream
//		},
//		nil,
//	)).Equal(
//		nil,
//		base.NewActionError("#.test:Eval action arguments does not match"),
//		nil,
//	)
//
//	// Test(8) error with 1st param
//	ret8, error8, panic8 := testRunWithProcessor(true, nil,
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			// error rpc.Bool
//			stream.Write(2)
//			stream.Write(int64(3))
//			stream.Write(uint64(3))
//			stream.Write(float64(3))
//			stream.Write("hello")
//			stream.Write(([]byte)("world"))
//			stream.Write(Array{1})
//			stream.Write(Map{"name": "world"})
//			return stream
//		},
//		nil,
//	)
//	assert(ret8, panic8).IsNil()
//	assert(error8.GetKind()).Equal(base.ErrorKindAction)
//	assert(error8.GetMessage()).Equal(
//		"#.test:Eval action arguments does not match\n" +
//			"want: #.test:Eval(rpc.Runtime, rpc.Bool, rpc.Int64, rpc.Uint64, " +
//			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
//			"got: #.test:Eval(rpc.Runtime, rpc.Int64, rpc.Int64, rpc.Uint64, " +
//			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return",
//	)
//
//	assert(strings.Contains(error8.GetDebug(), "#.test:Eval")).IsTrue()
//	assert(strings.Contains(error8.GetDebug(), "type_test.go:")).IsTrue()
//
//	// Test(9) error with 2nd param
//	assert(testRunWithProcessor(false, nil,
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			// error rpc.Int64
//			stream.Write(true)
//			stream.Write(uint64(3))
//			stream.Write(float64(3))
//			stream.Write("hello")
//			stream.Write(([]byte)("world"))
//			stream.Write(Array{1})
//			stream.Write(Map{"name": "world"})
//			return stream
//		},
//		nil,
//	)).Equal(
//		nil,
//		base.NewActionError("#.test:Eval action arguments does not match"),
//		nil,
//	)
//
//	// Test(10) error with 2nd param
//	ret10, error10, panic10 := testRunWithProcessor(true, nil,
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			// error rpc.Int64
//			stream.Write(true)
//			stream.Write(uint64(3))
//			stream.Write(float64(3))
//			stream.Write("hello")
//			stream.Write(([]byte)("world"))
//			stream.Write(Array{1})
//			stream.Write(Map{"name": "world"})
//			return stream
//		},
//		nil,
//	)
//	assert(ret10, panic10).IsNil()
//	assert(error10.GetKind()).Equal(base.ErrorKindAction)
//	assert(error10.GetMessage()).Equal(
//		"#.test:Eval action arguments does not match\n" +
//			"want: #.test:Eval(rpc.Runtime, rpc.Bool, rpc.Int64, rpc.Uint64, " +
//			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
//			"got: #.test:Eval(rpc.Runtime, rpc.Bool, rpc.Bool, rpc.Uint64, " +
//			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return",
//	)
//	assert(strings.Contains(error10.GetDebug(), "#.test:Eval")).IsTrue()
//	assert(strings.Contains(error10.GetDebug(), "type_test.go:")).IsTrue()
//
//	// Test(11) error with 3rd param
//	assert(testRunWithProcessor(false, nil,
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			stream.Write(int64(3))
//			// error rpc.Uint64
//			stream.Write(true)
//			stream.Write(float64(3))
//			stream.Write("hello")
//			stream.Write(([]byte)("world"))
//			stream.Write(Array{1})
//			stream.Write(Map{"name": "world"})
//			return stream
//		},
//		nil,
//	)).Equal(
//		nil,
//		base.NewActionError("#.test:Eval action arguments does not match"),
//		nil,
//	)
//
//	// Test(12) error with 3rd param
//	ret12, error12, panic12 := testRunWithProcessor(true, nil,
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			stream.Write(int64(3))
//			// error rpc.Uint64
//			stream.Write(true)
//			stream.Write(float64(3))
//			stream.Write("hello")
//			stream.Write(([]byte)("world"))
//			stream.Write(Array{1})
//			stream.Write(Map{"name": "world"})
//			return stream
//		},
//		nil,
//	)
//	assert(ret12, panic12).IsNil()
//	assert(error12.GetKind()).Equal(base.ErrorKindAction)
//	assert(error12.GetMessage()).Equal(
//		"#.test:Eval action arguments does not match\n" +
//			"want: #.test:Eval(rpc.Runtime, rpc.Bool, rpc.Int64, rpc.Uint64, " +
//			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
//			"got: #.test:Eval(rpc.Runtime, rpc.Bool, rpc.Int64, rpc.Bool, " +
//			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return",
//	)
//	assert(strings.Contains(error12.GetDebug(), "#.test:Eval")).IsTrue()
//	assert(strings.Contains(error12.GetDebug(), "type_test.go:")).IsTrue()
//
//	// Test(13) error with 4th param
//	assert(testRunWithProcessor(false, nil,
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			stream.Write(int64(3))
//			stream.Write(uint(3))
//			// error rpc.Float64
//			stream.Write(true)
//			stream.Write("hello")
//			stream.Write(([]byte)("world"))
//			stream.Write(Array{1})
//			stream.Write(Map{"name": "world"})
//			return stream
//		},
//		nil,
//	)).Equal(
//		nil,
//		base.NewActionError("#.test:Eval action arguments does not match"),
//		nil,
//	)
//
//	// Test(14) error with 4th param
//	ret14, error14, panic14 := testRunWithProcessor(true, nil,
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			stream.Write(int64(3))
//			stream.Write(uint(3))
//			// error rpc.Float64
//			stream.Write(true)
//			stream.Write("hello")
//			stream.Write(([]byte)("world"))
//			stream.Write(Array{1})
//			stream.Write(Map{"name": "world"})
//			return stream
//		},
//		nil,
//	)
//	assert(ret14, panic14).IsNil()
//	assert(error14.GetKind()).Equal(base.ErrorKindAction)
//	assert(error14.GetMessage()).Equal(
//		"#.test:Eval action arguments does not match\n" +
//			"want: #.test:Eval(rpc.Runtime, rpc.Bool, rpc.Int64, rpc.Uint64, " +
//			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
//			"got: #.test:Eval(rpc.Runtime, rpc.Bool, rpc.Int64, rpc.Uint64, " +
//			"rpc.Bool, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return",
//	)
//	assert(strings.Contains(error14.GetDebug(), "#.test:Eval")).IsTrue()
//	assert(strings.Contains(error14.GetDebug(), "type_test.go:")).IsTrue()
//
//	// Test(15) error with 5th param
//	assert(testRunWithProcessor(false, nil,
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			stream.Write(int64(3))
//			stream.Write(uint(3))
//			stream.Write(float64(3))
//			// error rpc.String
//			stream.Write(true)
//			stream.Write(([]byte)("world"))
//			stream.Write(Array{1})
//			stream.Write(Map{"name": "world"})
//			return stream
//		},
//		nil,
//	)).Equal(
//		nil,
//		base.NewActionError("#.test:Eval action arguments does not match"),
//		nil,
//	)
//
//	// Test(16) error with 5th param
//	ret16, error16, panic16 := testRunWithProcessor(true, nil,
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			stream.Write(int64(3))
//			stream.Write(uint(3))
//			stream.Write(float64(3))
//			// error rpc.String
//			stream.Write(true)
//			stream.Write(([]byte)("world"))
//			stream.Write(Array{1})
//			stream.Write(Map{"name": "world"})
//			return stream
//		},
//		nil,
//	)
//	assert(ret16, panic16).IsNil()
//	assert(error16.GetKind()).Equal(base.ErrorKindAction)
//	assert(error16.GetMessage()).Equal(
//		"#.test:Eval action arguments does not match\n" +
//			"want: #.test:Eval(rpc.Runtime, rpc.Bool, rpc.Int64, rpc.Uint64, " +
//			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
//			"got: #.test:Eval(rpc.Runtime, rpc.Bool, rpc.Int64, rpc.Uint64, " +
//			"rpc.Float64, rpc.Bool, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return",
//	)
//	assert(strings.Contains(error16.GetDebug(), "#.test:Eval")).IsTrue()
//	assert(strings.Contains(error16.GetDebug(), "type_test.go:")).IsTrue()
//
//	// Test(17) error with 6th param
//	assert(testRunWithProcessor(false, nil,
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			stream.Write(int64(3))
//			stream.Write(uint(3))
//			stream.Write(float64(3))
//			stream.Write("hello")
//			// error rpc.Bytes
//			stream.Write(true)
//			stream.Write(Array{1})
//			stream.Write(Map{"name": "world"})
//			return stream
//		},
//		nil,
//	)).Equal(
//		nil,
//		base.NewActionError("#.test:Eval action arguments does not match"),
//		nil,
//	)
//
//	// Test(18) error with 6th param
//	ret18, error18, panic18 := testRunWithProcessor(true, nil,
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			stream.Write(int64(3))
//			stream.Write(uint(3))
//			stream.Write(float64(3))
//			stream.Write("hello")
//			// error rpc.Bytes
//			stream.Write(true)
//			stream.Write(Array{1})
//			stream.Write(Map{"name": "world"})
//			return stream
//		},
//		nil,
//	)
//	assert(ret18, panic18).IsNil()
//	assert(error18.GetKind()).Equal(base.ErrorKindAction)
//	assert(error18.GetMessage()).Equal(
//		"#.test:Eval action arguments does not match\n" +
//			"want: #.test:Eval(rpc.Runtime, rpc.Bool, rpc.Int64, rpc.Uint64, " +
//			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
//			"got: #.test:Eval(rpc.Runtime, rpc.Bool, rpc.Int64, rpc.Uint64, " +
//			"rpc.Float64, rpc.String, rpc.Bool, rpc.Array, rpc.Map) rpc.Return",
//	)
//	assert(strings.Contains(error18.GetDebug(), "#.test:Eval")).IsTrue()
//	assert(strings.Contains(error18.GetDebug(), "type_test.go:")).IsTrue()
//
//	// Test(19) error with 7th param
//	assert(testRunWithProcessor(false, nil,
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			stream.Write(int64(3))
//			stream.Write(uint(3))
//			stream.Write(float64(3))
//			stream.Write("hello")
//			stream.Write(([]byte)("world"))
//			// error rpc.Array
//			stream.Write(true)
//			stream.Write(Map{"name": "world"})
//			return stream
//		},
//		nil,
//	)).Equal(
//		nil,
//		base.NewActionError("#.test:Eval action arguments does not match"),
//		nil,
//	)
//
//	// Test(20) error with 6th param
//	ret20, error20, panic20 := testRunWithProcessor(true, nil,
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			stream.Write(int64(3))
//			stream.Write(uint(3))
//			stream.Write(float64(3))
//			stream.Write("hello")
//			stream.Write(([]byte)("world"))
//			// error rpc.Array
//			stream.Write(true)
//			stream.Write(Map{"name": "world"})
//			return stream
//		},
//		nil,
//	)
//	assert(ret20, panic20).IsNil()
//	assert(error20.GetKind()).Equal(base.ErrorKindAction)
//	assert(error20.GetMessage()).Equal(
//		"#.test:Eval action arguments does not match\n" +
//			"want: #.test:Eval(rpc.Runtime, rpc.Bool, rpc.Int64, rpc.Uint64, " +
//			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
//			"got: #.test:Eval(rpc.Runtime, rpc.Bool, rpc.Int64, rpc.Uint64, " +
//			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Bool, rpc.Map) rpc.Return",
//	)
//	assert(strings.Contains(error20.GetDebug(), "#.test:Eval")).IsTrue()
//	assert(strings.Contains(error20.GetDebug(), "type_test.go:")).IsTrue()
//
//	// Test(21) error with 8th param
//	assert(testRunWithProcessor(false, nil,
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			stream.Write(int64(3))
//			stream.Write(uint(3))
//			stream.Write(float64(3))
//			stream.Write("hello")
//			stream.Write(([]byte)("world"))
//			stream.Write(Array{1})
//			// error rpc.Map
//			stream.Write(true)
//			return stream
//		},
//		nil,
//	)).Equal(
//		nil,
//		base.NewActionError("#.test:Eval action arguments does not match"),
//		nil,
//	)
//
//	// Test(22) error with 8th param
//	ret22, error22, panic22 := testRunWithProcessor(true, nil,
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			stream.Write(int64(3))
//			stream.Write(uint(3))
//			stream.Write(float64(3))
//			stream.Write("hello")
//			stream.Write(([]byte)("world"))
//			stream.Write(Array{1})
//			// error rpc.Map
//			stream.Write(true)
//			return stream
//		},
//		nil,
//	)
//	assert(ret22, panic22).IsNil()
//	assert(error22.GetKind()).Equal(base.ErrorKindAction)
//	assert(error22.GetMessage()).Equal(
//		"#.test:Eval action arguments does not match\n" +
//			"want: #.test:Eval(rpc.Runtime, rpc.Bool, rpc.Int64, rpc.Uint64, " +
//			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
//			"got: #.test:Eval(rpc.Runtime, rpc.Bool, rpc.Int64, rpc.Uint64, " +
//			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Bool) rpc.Return",
//	)
//	assert(strings.Contains(error22.GetDebug(), "#.test:Eval")).IsTrue()
//	assert(strings.Contains(error22.GetDebug(), "type_test.go:")).IsTrue()
//
//	// Test(23) nil rpcBytes
//	assert(testRunWithProcessor(true, nil,
//		func(rt Runtime, a Bytes) Return {
//			if a != nil {
//				return rt.Error(errors.New("param is not nil"))
//			}
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(nil)
//			return stream
//		},
//		nil,
//	)).Equal(true, nil, nil)
//
//	// Test(24) nil rpcArray
//	assert(testRunWithProcessor(true, nil,
//		func(rt Runtime, a Array) Return {
//			if a != nil {
//				return rt.Error(errors.New("param is not nil"))
//			}
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(nil)
//			return stream
//		},
//		nil,
//	)).Equal(true, nil, nil)
//
//	// Test(25) nil rpcMap
//	assert(testRunWithProcessor(true, nil,
//		func(rt Runtime, a Map) Return {
//			if a != nil {
//				return rt.Error(errors.New("param is not nil"))
//			}
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(nil)
//			return stream
//		},
//		nil,
//	)).Equal(true, nil, nil)
//
//	// Test(26) unsupported type
//	assert(testRunWithProcessor(false, nil,
//		func(rt Runtime, a bool) Return {
//			return rt.OK(a)
//		},
//		func(processor *Processor) *Stream {
//			actionNode := processor.actionsMap["#.test:Eval"]
//			actionNode.argTypes[1] = reflect.ValueOf(int16(0)).Type()
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			return stream
//		},
//		nil,
//	)).Equal(
//		nil,
//		base.NewActionError("#.test:Eval action arguments does not match"),
//		nil,
//	)
//
//	// Test(27) test
//	ret27, error27, panic27 := testRunWithProcessor(true, nil,
//		func(rt Runtime, bVal bool, rpcMap Map) Return {
//			return rt.OK(bVal)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(nil)
//			stream.Write(nil)
//			stream.Write(nil)
//			return stream
//		},
//		nil,
//	)
//	assert(ret27, panic27).IsNil()
//	assert(error27.GetKind()).Equal(base.ErrorKindAction)
//	assert(error27.GetMessage()).Equal(
//		"#.test:Eval action arguments does not match\n" +
//			"want: #.test:Eval(rpc.Runtime, rpc.Bool, rpc.Map) rpc.Return\n" +
//			"got: #.test:Eval(rpc.Runtime, <nil>, rpc.Map, <nil>) rpc.Return",
//	)
//	assert(strings.Contains(error27.GetDebug(), "#.test:Eval")).IsTrue()
//	assert(strings.Contains(error27.GetDebug(), "type_test.go:")).IsTrue()
//
//	// Test(28) badStream
//	assert(testRunWithProcessor(true, nil,
//		func(rt Runtime, bVal bool, rpcMap Map) Return {
//			return rt.OK(bVal)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write("helloWorld")
//			stream.SetWritePos(stream.GetWritePos() - 1)
//			return stream
//		},
//		nil,
//	)).Equal(nil, base.NewProtocolError(base.ErrStringBadStream), nil)
//
//	// Test(29) call function error
//	ret29, error29, panic29 := testRunWithProcessor(false, nil,
//		func(rt Runtime, bVal bool) Return {
//			if bVal {
//				panic("this is a error")
//			}
//			return rt.OK(bVal)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			return stream
//		},
//		nil,
//	)
//	assert(ret29, error29).IsNil()
//	assert(panic29).IsNotNil()
//	if panic29 != nil {
//		assert(panic29.GetKind()).Equal(base.ErrorKindActionPanic)
//		assert(panic29.GetMessage()).
//			Equal("runtime error: this is a error")
//		assert(strings.Contains(panic29.GetDebug(), "#.test:Eval")).IsTrue()
//		assert(strings.Contains(panic29.GetDebug(), "thread_test.go")).IsTrue()
//	}
//
//	// Test(30) call function error
//	ret30, error30, panic30 := testRunWithProcessor(true, nil,
//		func(rt Runtime, bVal bool) Return {
//			if bVal {
//				panic("this is a error")
//			}
//			return rt.OK(bVal)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			return stream
//		},
//		nil,
//	)
//	assert(ret30, error30).IsNil()
//	assert(panic30).IsNotNil()
//	if panic30 != nil {
//		assert(panic30.GetKind()).Equal(base.ErrorKindActionPanic)
//		assert(panic30.GetMessage()).
//			Equal("runtime error: this is a error")
//		assert(strings.Contains(panic30.GetDebug(), "#.test:Eval")).IsTrue()
//		assert(strings.Contains(panic30.GetDebug(), "thread_test.go")).IsTrue()
//	}
//
//	// Test(31) return TransportError to make onEvalFinish panic
//	ret31, error31, panic31 := testRunWithProcessor(true, nil,
//		func(rt Runtime, bVal bool) Return {
//			return rt.Error(base.NewTransportError("it makes onEvalFinish panic"))
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			return stream
//		},
//		nil,
//	)
//	assert(ret31, error31).Equal(nil, nil)
//	assert(panic31).IsNotNil()
//	if panic31 != nil {
//		assert(panic31.GetKind()).Equal(base.ErrorKindKernelPanic)
//		assert(panic31.GetMessage()).
//			Equal("kernel error: it makes onEvalFinish panic")
//		assert(strings.Contains(panic31.GetDebug(), "type_test.go")).IsTrue()
//	}
//
//	// Test(32) return without rt
//	ret32, error32, panic32 := testRunWithProcessor(true, nil,
//		func(rt Runtime, bVal bool) Return {
//			return emptyReturn
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			return stream
//		},
//		nil,
//	)
//	assert(ret32, error32).Equal(nil, nil)
//	assert(panic32).IsNotNil()
//	if panic32 != nil {
//		assert(panic32.GetKind()).Equal(base.ErrorKindActionPanic)
//		assert(panic32.GetMessage()).
//			Equal("action must return through Runtime.OK or Runtime.Error")
//		assert(strings.Contains(panic32.GetDebug(), "type_test.go")).IsTrue()
//	} else {
//		assert().Fail("nil)")
//	}
//
//	// Test(33) ok call with  cache
//	assert(testRunWithProcessor(true, &testFuncCache{},
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			stream.Write(int64(3))
//			stream.Write(uint64(3))
//			stream.Write(float64(3))
//			stream.Write("hello")
//			stream.Write(([]byte)("world"))
//			stream.Write(Array{1})
//			stream.Write(Map{"name": "world"})
//			return stream
//		},
//		nil,
//	)).Equal(true, nil, nil)
//
//	// Test(34) stream is not finish
//	ret34, error34, panic34 := testRunWithProcessor(true, nil,
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			stream.Write(int64(3))
//			stream.Write(uint64(3))
//			stream.Write(float64(3))
//			stream.Write("hello")
//			stream.Write(([]byte)("world"))
//			stream.Write(Array{1})
//			stream.Write(Map{"name": "world"})
//			// error
//			stream.Write(true)
//			return stream
//		},
//		nil,
//	)
//	assert(ret34, panic34).Equal(nil, nil)
//	assert(error34).IsNotNil()
//	assert(error34.GetMessage())
//
//	assert(error34.GetKind()).Equal(base.ErrorKindAction)
//	assert(error34.GetMessage()).Equal(
//		"#.test:Eval action arguments does not match\n" +
//			"want: #.test:Eval(rpc.Runtime, rpc.Bool, rpc.Int64, rpc.Uint64, " +
//			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
//			"got: #.test:Eval(rpc.Runtime, rpc.Bool, rpc.Int64, rpc.Uint64, " +
//			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map, rpc.Bool) " +
//			"rpc.Return",
//	)
//	assert(strings.Contains(error34.GetDebug(), "#.test:Eval")).IsTrue()
//	assert(strings.Contains(error34.GetDebug(), "type_test.go:")).IsTrue()
//
//	// Test(35) bad stream
//	assert(testRunWithProcessor(true, &testFuncCache{},
//		func(rt Runtime,
//			b bool, i int64, u uint64, f float64, s string,
//			x Bytes, a Array, m Map,
//		) Return {
//			return rt.OK(true)
//		},
//		func(_ *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("#")
//			stream.Write(true)
//			stream.Write(int64(3))
//			stream.Write(uint64(3))
//			stream.Write(float64(3))
//			stream.Write("hello")
//			stream.Write(([]byte)("world"))
//			stream.Write(Array{1})
//			stream.Write(Map{"name": "world"})
//			// error
//			stream.Write(true)
//			stream.Write(Map{"name": "world"})
//			stream.SetWritePos(stream.GetWritePos() - 2)
//
//			return stream
//		},
//		nil,
//	)).Equal(nil, base.NewProtocolError(base.ErrStringBadStream), nil)
//}
