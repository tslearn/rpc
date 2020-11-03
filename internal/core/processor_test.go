package core

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"os"
	"path"
	"runtime"
	"testing"
	"time"
)

func TestRpcActionNode_GetConfig(t *testing.T) {
	t.Run("data is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &rpcServiceNode{}
		assert(v.GetConfig("name")).Equal(nil, false)
	})

	t.Run("key does not exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &rpcServiceNode{data: Map{"age": 18}}
		assert(v.GetConfig("name")).Equal(nil, false)
	})

	t.Run("key exists", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &rpcServiceNode{data: Map{"age": 18}}
		assert(v.GetConfig("age")).Equal(18, true)
	})
}

func TestRpcActionNode_SetConfig(t *testing.T) {
	t.Run("data is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &rpcServiceNode{}
		v.SetConfig("age", 3)
		assert(v.data).Equal(nil)
	})

	t.Run("key does not exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &rpcServiceNode{data: Map{}}
		v.SetConfig("age", 3)
		assert(v.data).Equal(Map{"age": 3})
	})

	t.Run("key exists", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &rpcServiceNode{data: Map{"age": 5}}
		v.SetConfig("age", 3)
		assert(v.data).Equal(Map{"age": 3})
	})
}

func TestProcessor(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		emptyEvalBack(nil)
		emptyEvalFinish(nil)
		assert(actionNameRegex.MatchString("$onMount")).IsTrue()
		assert(actionNameRegex.MatchString("$onUnmount")).IsTrue()
		assert(actionNameRegex.MatchString("$onUpdateConfig")).IsTrue()
		assert(actionNameRegex.MatchString("onMount")).IsTrue()
		assert(actionNameRegex.MatchString("sayHello")).IsTrue()
		assert(actionNameRegex.MatchString("$sayHello")).IsFalse()
		assert(rootName).Equal("#")
		assert(freeGroups).Equal(1024)
		assert(processorStatusClosed).Equal(0)
		assert(processorStatusRunning).Equal(1)
	})
}

func TestNewProcessor(t *testing.T) {
	t.Run("onReturnStream is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(NewProcessor(
			1024, 16, 16, 2048, nil, 5*time.Second, nil, nil,
		)).Equal(nil, errors.ErrProcessorOnReturnStreamIsNil)
	})

	t.Run("numOfThreads <= 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(NewProcessor(
			0, 16, 16, 2048, nil, 5*time.Second, nil, func(stream *Stream) {},
		)).Equal(nil, errors.ErrNumOfThreadsIsWrong)
	})

	t.Run("maxNodeDepth <= 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(NewProcessor(
			1024, 0, 16, 2048, nil, 5*time.Second, nil, func(stream *Stream) {},
		)).Equal(nil, errors.ErrMaxNodeDepthIsWrong)
	})

	t.Run("maxCallDepth <= 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(NewProcessor(
			1024, 16, 0, 2048, nil, 5*time.Second, nil, func(stream *Stream) {},
		)).Equal(nil, errors.ErrProcessorMaxCallDepthIsWrong)
	})

	t.Run("mount service error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(NewProcessor(
			1024, 16, 16, 2048, nil, 5*time.Second,
			[]*ServiceMeta{nil}, func(stream *Stream) {},
		)).Equal(nil, errors.ErrProcessorNodeMetaIsNil)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		processor, err := NewProcessor(
			1, 16, 16, 2048, nil, 5*time.Second,
			[]*ServiceMeta{{
				name: "test",
				service: NewService().On("Eval", func(rt Runtime) Return {
					return rt.Reply(true)
				}),
				fileLine: "",
			}},
			func(stream *Stream) {},
		)
		assert(err).IsNil()
		assert(processor).IsNotNil()
		assert(len(processor.threads)).Equal(freeGroups)
		_ = processor.Close()
	})

	t.Run("test ok (subscribe error)", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamCH := make(chan *Stream, 1)
		processor, err := NewProcessor(
			1, 16, 16, 2048, nil, 5*time.Second,
			[]*ServiceMeta{{
				name: "test",
				service: NewService().On("Eval", func(rt Runtime) Return {
					return rt.Reply(true)
				}),
				fileLine: "",
			}},
			func(stream *Stream) {
				streamCH <- stream
			},
		)
		assert(err).IsNil()
		assert(processor).IsNotNil()
		base.PublishPanic(errors.ErrStream)
		assert(ParseResponseStream(<-streamCH)).Equal(nil, errors.ErrStream)
		_ = processor.Close()
	})

	t.Run("test ok (system action)", func(t *testing.T) {
		assert := base.NewAssert(t)
		processor := (*Processor)(nil)
		wait := make(chan string, 3)
		service := NewService().
			On("$onMount", func(rt Runtime) Return {
				wait <- "$onMount called"
				return rt.Reply(true)
			}).
			On("$onUpdateConfig", func(rt Runtime) Return {
				wait <- "$onUpdateConfig called"
				return rt.Reply(true)
			}).
			On("$onUnmount", func(rt Runtime) Return {
				wait <- "$onUnmount called"
				return rt.Reply(true)
			})
		processor, _ = NewProcessor(
			1, 16, 16, 2048, nil, 5*time.Second,
			[]*ServiceMeta{{
				name:     "test",
				service:  service,
				fileLine: "",
			}},
			func(stream *Stream) {
				stream.Release()
			},
		)
		assert(processor).IsNotNil()
		assert(<-wait).Equal("$onMount called")
		assert(<-wait).Equal("$onUpdateConfig called")
		processor.Close()
		assert(<-wait).Equal("$onUnmount called")
	})

	t.Run("test ok (1M calls)", func(t *testing.T) {
		assert := base.NewAssert(t)
		processor := (*Processor)(nil)
		wait := make(chan bool)
		service := NewService().
			On("Eval", func(rt Runtime) Return {
				return rt.Reply(true)
			})
		processor, _ = NewProcessor(
			1, 16, 16, 2048, nil, 5*time.Second,
			[]*ServiceMeta{{
				name:     "test",
				service:  service,
				fileLine: "",
			}},
			func(stream *Stream) {
				v, _ := ParseResponseStream(stream)
				if ret, ok := v.(bool); ok {
					wait <- ret
				} else {
					wait <- false
				}
				stream.Release()
			},
		)

		go func() {
			for i := 0; i < 1000000; i++ {
				stream, _ := MakeRequestStream(true, 0, "#.test:Eval", "")
				processor.PutStream(stream)
			}
		}()

		for i := 0; i < 1000000; i++ {
			assert(<-wait).IsTrue()
		}

		processor.Close()
	})
}

func TestProcessor_Close(t *testing.T) {
	t.Run("processor is not running", func(t *testing.T) {
		assert := base.NewAssert(t)
		processor := getFakeProcessor()
		assert(processor.Close()).Equal(false)
	})

	t.Run("close timeout", func(t *testing.T) {
		assert := base.NewAssert(t)

		fnTest := func(count int) {
			source := ""
			waitCH := make(chan bool)
			streamCH := make(chan *Stream, 1)
			processor, _ := NewProcessor(
				1024,
				2,
				3,
				2048,
				nil,
				time.Second,
				[]*ServiceMeta{{
					name: "test",
					service: NewService().On("Eval", func(rt Runtime) Return {
						waitCH <- true
						source = rt.thread.GetExecActionDebug()
						time.Sleep(4 * time.Second)
						return rt.Reply(true)
					}),
					fileLine: "",
				}},
				func(stream *Stream) {
					streamCH <- stream
				},
			)

			for i := 0; i < count; i++ {
				stream, _ := MakeRequestStream(true, 0, "#.test:Eval", "")
				processor.PutStream(stream)
				<-waitCH
			}

			assert(processor.Close()).IsFalse()

			if count == 1 {
				assert(ParseResponseStream(<-streamCH)).Equal(
					nil,
					errors.ErrActionCloseTimeout.AddDebug(fmt.Sprintf(
						"the following actions can not close: \n\t%s (1 goroutine)",
						source,
					)),
				)
			} else {
				assert(ParseResponseStream(<-streamCH)).Equal(
					nil,
					errors.ErrActionCloseTimeout.AddDebug(fmt.Sprintf(
						"the following actions can not close: \n\t%s (%d goroutines)",
						source, count,
					)),
				)
			}
		}

		fnTest(1)
		fnTest(100)
	})
}

func TestProcessor_PutStream(t *testing.T) {
	t.Run("processor is closed", func(t *testing.T) {
		assert := base.NewAssert(t)
		processor, _ := NewProcessor(
			1024,
			2,
			3,
			2048,
			nil,
			time.Second,
			nil,
			func(_ *Stream) {},
		)
		processor.Close()

		for i := 0; i < 2048; i++ {
			assert(processor.PutStream(NewStream())).IsFalse()
		}
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		waitCH := make(chan bool)
		processor, _ := NewProcessor(
			freeGroups*16,
			2,
			3,
			2048,
			nil,
			time.Second,
			nil,
			func(_ *Stream) {
				waitCH <- true
			},
		)

		defer processor.Close()

		for i := 0; i < 204800; i++ {
			assert(processor.PutStream(NewStream())).IsTrue()
			<-waitCH
		}

		freeSum := 0
		for i := 0; i < len(processor.freeCHArray); i++ {
			freeSum += len(processor.freeCHArray[i])
		}
		assert(freeSum).Equal(freeGroups * 16)
	})
}

func TestProcessor_BuildCache(t *testing.T) {
	assert := base.NewAssert(t)
	_, file, _, _ := runtime.Caller(0)
	currDir := path.Dir(file)
	defer func() {
		_ = os.RemoveAll(path.Join(path.Dir(file), "_tmp_"))
	}()

	t.Run("services is empty", func(t *testing.T) {
		tmpFile := path.Join(currDir, "_tmp_/test-processor-01.go")
		snapshotFile := path.Join(currDir, "_snapshot_/test-processor-01.snapshot")
		processor, _ := NewProcessor(
			freeGroups*16,
			2,
			3,
			2048,
			nil,
			time.Second,
			nil,
			func(_ *Stream) {},
		)
		defer processor.Close()
		assert(processor.BuildCache("pkgName", tmpFile)).IsNil()
		assert(base.ReadFromFile(tmpFile)).Equal(base.ReadFromFile(snapshotFile))
	})

	t.Run("service is not empty", func(t *testing.T) {
		tmpFile := path.Join(currDir, "_tmp_/test-processor-02.go")
		snapshotFile := path.Join(currDir, "_snapshot_/test-processor-02.snapshot")
		processor, _ := NewProcessor(
			freeGroups*16,
			2,
			3,
			2048,
			nil,
			time.Second,
			[]*ServiceMeta{{
				name: "test",
				service: NewService().On("Eval", func(rt Runtime) Return {
					return rt.Reply(true)
				}),
				fileLine: "",
			}},
			func(_ *Stream) {},
		)
		defer processor.Close()
		assert(processor.BuildCache("pkgName", tmpFile)).IsNil()
		assert(base.ReadFromFile(tmpFile)).Equal(base.ReadFromFile(snapshotFile))
	})

	t.Run("processor is closed", func(t *testing.T) {
		processor, _ := NewProcessor(
			freeGroups*16,
			2,
			3,
			2048,
			nil,
			time.Second,
			nil,
			func(_ *Stream) {},
		)
		processor.Close()
		assert(processor.BuildCache("pkgName", "")).
			Equal(errors.ErrProcessorIsNotRunning)
	})
}

//func TestProcessor_mountNode(t *testing.T) {
//	assert := base.NewAssert(t)
//
//	fnTestMount := func(
//		services []*ServiceMeta,
//		wantPanicKind base.ErrorKind,
//		wantPanicMessage string,
//		wantPanicDebug string,
//	) {
//		helper := newTestProcessorReturnHelper()
//		assert(NewProcessor(
//			true,
//			1024,
//			2,
//			3,
//			&testFuncCache{},
//			5*time.Second,
//			services,
//			helper.GetFunction(),
//		)).IsNil()
//		retArray, errArray, panicArray := helper.GetReturn()
//		assert(retArray, errArray).Equal([]Any{}, []base.Error{})
//		assert(len(panicArray)).Equal(1)
//
//		assert(panicArray[0].GetKind()).Equal(wantPanicKind)
//		assert(panicArray[0].GetMessage()).Equal(wantPanicMessage)
//
//		if wantPanicKind == base.ErrorKindKernelPanic {
//			assert(strings.Contains(panicArray[0].GetDebug(), "goroutine")).IsTrue()
//			assert(strings.Contains(panicArray[0].GetDebug(), "[running]")).IsTrue()
//			assert(strings.Contains(panicArray[0].GetDebug(), "mountNode")).IsTrue()
//			assert(strings.Contains(panicArray[0].GetDebug(), "NewProcessor")).
//				IsTrue()
//		} else {
//			assert(panicArray[0].GetDebug()).Equal(wantPanicDebug)
//		}
//	}
//
//	// Test(1)
//	fnTestMount([]*ServiceMeta{
//		nil,
//	}, base.ErrorKindKernelPanic, "nodeMeta is nil", "")
//
//	// Test(2)
//	fnTestMount([]*ServiceMeta{{
//		name:     "+",
//		service:  NewService(),
//		fileLine: "DebugMessage",
//	}}, base.ErrorKindRuntimePanic, "service name + is illegal", "DebugMessage")
//
//	// Test(3)
//	fnTestMount([]*ServiceMeta{{
//		name:     "abc",
//		service:  nil,
//		fileLine: "DebugMessage",
//	}}, base.ErrorKindRuntimePanic, "service is nil", "DebugMessage")
//
//	// Test(4)
//	s4, source1 := NewService().
//		AddChildService("s", NewService(), nil), base.GetFileLine(0)
//	fnTestMount(
//		[]*ServiceMeta{{
//			name:     "s",
//			service:  NewService().AddChildService("s", s4, nil),
//			fileLine: "DebugMessage",
//		}},
//		base.ErrorKindRuntimePanic,
//		"service path #.s.s.s is too long",
//		source1,
//	)
//
//	// Test(5)
//	fnTestMount(
//		[]*ServiceMeta{{
//			name:     "user",
//			service:  NewService(),
//			fileLine: "Debug1",
//		}, {
//			name:     "user",
//			service:  NewService(),
//			fileLine: "Debug2",
//		}},
//		base.ErrorKindRuntimePanic,
//		"duplicated service name user",
//		"current:\n\tDebug2\nconflict:\n\tDebug1",
//	)
//
//	// Test(6)
//	fnTestMount(
//		[]*ServiceMeta{{
//			name: "user",
//			service: &Service{
//				children: []*ServiceMeta{},
//				actions:  []*rpcActionMeta{nil},
//				fileLine: "DebugAction",
//			},
//			fileLine: "DebugService",
//		}},
//		base.ErrorKindKernelPanic,
//		"meta is nil",
//		"DebugAction",
//	)
//
//	// Test(7)
//	fnTestMount(
//		[]*ServiceMeta{{
//			name: "test",
//			service: &Service{
//				children: []*ServiceMeta{},
//				actions: []*rpcActionMeta{{
//					name:     "-",
//					handler:  nil,
//					fileLine: "DebugAction",
//				}},
//				fileLine: "DebugService",
//			},
//			fileLine: "Debug1",
//		}},
//		base.ErrorKindRuntimePanic,
//		"action name - is illegal",
//		"DebugAction",
//	)
//
//	// Test(8)
//	fnTestMount(
//		[]*ServiceMeta{{
//			name: "test",
//			service: &Service{
//				children: []*ServiceMeta{},
//				actions: []*rpcActionMeta{{
//					name:     "Eval",
//					handler:  nil,
//					fileLine: "DebugAction",
//				}},
//				fileLine: "DebugService",
//			},
//			fileLine: "Debug1",
//		}},
//		base.ErrorKindRuntimePanic,
//		"handler is nil",
//		"DebugAction",
//	)
//
//	// Test(9)
//	fnTestMount(
//		[]*ServiceMeta{{
//			name: "test",
//			service: &Service{
//				children: []*ServiceMeta{},
//				actions: []*rpcActionMeta{{
//					name:     "Eval",
//					handler:  3,
//					fileLine: "DebugAction",
//				}},
//				fileLine: "DebugService",
//			},
//			fileLine: "Debug1",
//		}},
//		base.ErrorKindRuntimePanic,
//		"handler must be func(rt rpc.Runtime, ...) rpc.Return",
//		"DebugAction",
//	)
//
//	// Test(10)
//	fnTestMount(
//		[]*ServiceMeta{{
//			name: "test",
//			service: &Service{
//				children: []*ServiceMeta{},
//				actions: []*rpcActionMeta{{
//					name:     "Eval",
//					handler:  func() {},
//					fileLine: "DebugAction",
//				}},
//				fileLine: "DebugService",
//			},
//			fileLine: "Debug1",
//		}},
//		base.ErrorKindRuntimePanic,
//		"handler 1st argument type must be rpc.Runtime",
//		"DebugAction",
//	)
//
//	// Test(11)
//	fnTestMount(
//		[]*ServiceMeta{{
//			name: "test",
//			service: &Service{
//				children: []*ServiceMeta{},
//				actions: []*rpcActionMeta{{
//					name:     "Eval",
//					handler:  func(rt Runtime) Return { return rt.OK(true) },
//					fileLine: "DebugAction1",
//				}, {
//					name:     "Eval",
//					handler:  func(rt Runtime) Return { return rt.OK(true) },
//					fileLine: "DebugAction2",
//				}},
//				fileLine: "DebugService",
//			},
//			fileLine: "Debug1",
//		}},
//		base.ErrorKindRuntimePanic,
//		"action name Eval is duplicated",
//		"current:\n\tDebugAction2\nconflict:\n\tDebugAction1",
//	)
//
//	// Test(12)
//	actionMeta12Eval1 := &rpcActionMeta{
//		name: "Eval1",
//		handler: func(rt Runtime, _a Array) Return {
//			return rt.OK(true)
//		},
//		fileLine: "DebugEval1",
//	}
//	actionMeta12Eval2 := &rpcActionMeta{
//		name: "Eval2",
//		handler: func(rt Runtime,
//			_ bool, _ int64, _ uint64, _ float64,
//			_ string, _ Bytes, _ Array, _ Map,
//		) Return {
//			return rt.OK(true)
//		},
//		fileLine: "DebugEval2",
//	}
//	addMeta12 := &ServiceMeta{
//		name: "test",
//		service: &Service{
//			children: []*ServiceMeta{},
//			actions:  []*rpcActionMeta{actionMeta12Eval1, actionMeta12Eval2},
//			fileLine: base.GetFileLine(1),
//		},
//		fileLine: "serviceDebug",
//	}
//	processor12 := NewProcessor(
//		true,
//		1024,
//		2,
//		3,
//		&testFuncCache{},
//		5*time.Second,
//		[]*ServiceMeta{addMeta12},
//		getFakeOnEvalBack(),
//	)
//	assert(processor12).IsNotNil()
//	defer processor12.Close()
//	assert(*processor12.servicesMap["#"]).Equal(rpcServiceNode{
//		path:    rootName,
//		addMeta: nil,
//		depth:   0,
//	})
//	assert(*processor12.servicesMap["#.test"]).Equal(rpcServiceNode{
//		path:    "#.test",
//		addMeta: addMeta12,
//		depth:   1,
//	})
//	assert(processor12.actionsMap["#.test:Eval1"].path).Equal("#.test:Eval1")
//	assert(processor12.actionsMap["#.test:Eval1"].meta).Equal(actionMeta12Eval1)
//	assert(processor12.actionsMap["#.test:Eval1"].cacheFN).IsNil()
//	assert(getFuncKind(processor12.actionsMap["#.test:Eval1"].reflectFn)).
//		Equal("A", nil)
//	assert(processor12.actionsMap["#.test:Eval1"].callString).
//		Equal("#.test:Eval1(rpc.Runtime, rpc.Array) rpc.Return")
//	assert(processor12.actionsMap["#.test:Eval1"].argTypes).
//		Equal([]reflect.Type{runtimeType, arrayType})
//	assert(processor12.actionsMap["#.test:Eval1"].indicator).IsNotNil()
//
//	assert(processor12.actionsMap["#.test:Eval2"].path).Equal("#.test:Eval2")
//	assert(processor12.actionsMap["#.test:Eval2"].meta).Equal(actionMeta12Eval2)
//	assert(processor12.actionsMap["#.test:Eval2"].cacheFN).IsNotNil()
//	assert(getFuncKind(processor12.actionsMap["#.test:Eval2"].reflectFn)).
//		Equal("BIUFSXAM", nil)
//	assert(processor12.actionsMap["#.test:Eval2"].callString).Equal(
//		"#.test:Eval2(rpc.Runtime, rpc.Bool, rpc.Int64, rpc.Uint64, " +
//			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return",
//	)
//	assert(processor12.actionsMap["#.test:Eval2"].argTypes).
//		Equal([]reflect.Type{
//			runtimeType, boolType, int64Type, uint64Type,
//			float64Type, stringType, bytesType, arrayType, mapType,
//		})
//	assert(processor12.actionsMap["#.test:Eval2"].indicator).IsNotNil()
//}
