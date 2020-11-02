package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
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
	t.Run("test regex", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(actionNameRegex.MatchString("$onMount")).IsTrue()
		assert(actionNameRegex.MatchString("$onUnmount")).IsTrue()
		assert(actionNameRegex.MatchString("$onUpdateConfig")).IsTrue()
		assert(actionNameRegex.MatchString("onMount")).IsTrue()
		assert(actionNameRegex.MatchString("sayHello")).IsTrue()
		assert(actionNameRegex.MatchString("$sayHello")).IsFalse()
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
				go func() {
					time.Sleep(200 * time.Millisecond)
					processor.Close()
				}()

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
		assert(<-wait).Equal("$onUnmount called")
	})
}

//func TestNewProcessor(t *testing.T) {
//	assert := base.NewAssert(t)
//	// Test(6) OK
//	helper6 := newTestProcessorReturnHelper()
//	processor6 := NewProcessor(
//		true,
//		2048,
//		2,
//		3,
//		nil,
//		5*time.Second,
//		[]*ServiceMeta{{
//			name: "test",
//			service: NewService().On("Eval", func(rt Runtime) Return {
//				time.Sleep(time.Second)
//				return rt.OK(true)
//			}),
//			fileLine: "",
//		}},
//		helper6.GetFunction(),
//	)
//	for i := 0; i < 2048; i++ {
//		stream := NewStream()
//		stream.SetDepth(3)
//		stream.WriteString("#.test:Eval")
//		stream.WriteString("")
//		processor6.PutStream(stream)
//	}
//	assert(processor6).IsNotNil()
//	assert(processor6.isDebug).IsTrue()
//	assert(len(processor6.actionsMap)).Equal(1)
//	assert(len(processor6.servicesMap)).Equal(2)
//	assert(processor6.maxNodeDepth).Equal(uint16(2))
//	assert(processor6.maxCallDepth).Equal(uint16(3))
//	assert(len(processor6.threads)).Equal(2048)
//	assert(len(processor6.freeCHArray)).Equal(freeGroups)
//	assert(processor6.readThreadPos).Equal(uint64(2048))
//	assert(processor6.fnError).IsNotNil()
//	processor6.Close()
//	assert(processor6.writeThreadPos).Equal(uint64(2048))
//	sumFrees := 0
//	for _, freeCH := range processor6.freeCHArray {
//		sumFrees += len(freeCH)
//	}
//	assert(sumFrees).Equal(2048)
//}
//
//func TestProcessor_Close(t *testing.T) {
//	assert := base.NewAssert(t)
//
//	// Test(1) p.panicSubscription == nil
//	processor1 := getFakeProcessor(true)
//	assert(processor1.Close()).IsFalse()
//
//	// Test(2)
//	lock2 := sync.Mutex{}
//	actionFileLine2 := ""
//	helper2 := newTestProcessorReturnHelper()
//	processor2 := NewProcessor(
//		true,
//		1024,
//		2,
//		3,
//		nil,
//		time.Second,
//		[]*ServiceMeta{{
//			name: "test",
//			service: NewService().On("Eval", func(rt Runtime) Return {
//				lock2.Lock()
//				actionFileLine2 = rt.thread.GetExecActionDebug()
//				lock2.Unlock()
//				time.Sleep(4 * time.Second)
//				return rt.OK(true)
//			}),
//			fileLine: "",
//		}},
//		helper2.GetFunction(),
//	)
//	for i := 0; i < 1; i++ {
//		stream := NewStream()
//		stream.SetDepth(3)
//		stream.WriteString("#.test:Eval")
//		stream.WriteString("")
//		processor2.PutStream(stream)
//	}
//	assert(processor2.Close()).IsFalse()
//	lock2.Lock()
//	assert(helper2.GetReturn()).Equal([]Any{}, []base.Error{}, []base.Error{
//		base.NewActionPanic(
//			"the following actions can not close: \n\t" +
//				actionFileLine2 + " (1 goroutine)",
//		),
//	})
//	lock2.Unlock()
//	time.Sleep(2 * time.Second)
//
//	// Test(3)
//	actionFileLine3 := ""
//	lock3 := sync.Mutex{}
//	helper3 := newTestProcessorReturnHelper()
//	processor3 := NewProcessor(
//		true,
//		1024,
//		2,
//		3,
//		nil,
//		time.Second,
//		[]*ServiceMeta{{
//			name: "test",
//			service: NewService().On("Eval", func(rt Runtime) Return {
//				lock3.Lock()
//				actionFileLine3 = rt.thread.GetExecActionDebug()
//				lock3.Unlock()
//				time.Sleep(4 * time.Second)
//				return rt.OK(true)
//			}),
//			fileLine: "",
//		}},
//		helper3.GetFunction(),
//	)
//	for i := 0; i < 2; i++ {
//		stream := NewStream()
//		stream.SetDepth(3)
//		stream.WriteString("#.test:Eval")
//		stream.WriteString("")
//		processor3.PutStream(stream)
//	}
//	assert(processor3.Close()).IsFalse()
//	lock3.Lock()
//	assert(helper3.GetReturn()).Equal([]Any{}, []base.Error{}, []base.Error{
//		base.NewActionPanic(
//			"the following actions can not close: \n\t" +
//				actionFileLine3 + " (2 goroutines)",
//		),
//	})
//	lock3.Unlock()
//	time.Sleep(2 * time.Second)
//}
//
//func TestProcessor_PutStream(t *testing.T) {
//	assert := base.NewAssert(t)
//
//	// Test(1)
//	processor1 := NewProcessor(
//		true,
//		1024,
//		32,
//		32,
//		nil,
//		5*time.Second,
//		nil,
//		func(stream *Stream) {},
//	)
//	defer processor1.Close()
//	assert(processor1.PutStream(NewStream())).IsTrue()
//
//	// Test(2)
//	processor2 := getFakeProcessor(true)
//	for i := 0; i < 2048; i++ {
//		assert(processor2.PutStream(NewStream())).IsFalse()
//	}
//}
//
//func TestProcessor_fnError(t *testing.T) {
//	assert := base.NewAssert(t)
//
//	// Test(1)
//	err1 := base.NewRuntimePanic("message").AddDebug("debug")
//	helper1 := newTestProcessorReturnHelper()
//	processor1 := NewProcessor(
//		true,
//		1,
//		1,
//		1,
//		nil,
//		5*time.Second,
//		nil,
//		helper1.GetFunction(),
//	)
//	defer processor1.Close()
//	processor1.fnError(err1)
//	assert(helper1.GetReturn()).Equal([]Any{}, []base.Error{}, []base.Error{err1})
//}
//
//func TestProcessor_BuildCache(t *testing.T) {
//	assert := base.NewAssert(t)
//	_, file, _, _ := runtime.Caller(0)
//	currDir := path.Dir(file)
//	defer func() {
//		_ = os.RemoveAll(path.Join(path.Dir(file), "_tmp_"))
//	}()
//
//	// Test(1)
//	tmpFile1 := path.Join(currDir, "_tmp_/test-processor-01.go")
//	snapshotFile3 := path.Join(currDir, "_snapshot_/test-processor-01.snapshot")
//	helper1 := newTestProcessorReturnHelper()
//	processor1 := NewProcessor(
//		true,
//		1024,
//		2,
//		3,
//		nil,
//		5*time.Second,
//		[]*ServiceMeta{},
//		helper1.GetFunction(),
//	)
//	defer processor1.Close()
//	assert(processor1.BuildCache("pkgName", tmpFile1)).IsNil()
//	assert(helper1.GetReturn()).Equal([]Any{}, []base.Error{}, []base.Error{})
//	assert(testReadFromFile(tmpFile1)).Equal(testReadFromFile(snapshotFile3))
//
//	// Test(2)
//	tmpFile2 := path.Join(currDir, "_tmp_/test-processor-02.go")
//	snapshotFile2 := path.Join(currDir, "_snapshot_/test-processor-02.snapshot")
//	helper2 := newTestProcessorReturnHelper()
//	processor2 := NewProcessor(
//		true,
//		1024,
//		2,
//		3,
//		nil,
//		5*time.Second,
//		[]*ServiceMeta{{
//			name: "test",
//			service: NewService().On("Eval", func(rt Runtime) Return {
//				return rt.OK(true)
//			}),
//			fileLine: "",
//		}},
//		helper2.GetFunction(),
//	)
//	defer processor2.Close()
//	assert(processor2.BuildCache("pkgName", tmpFile2)).IsNil()
//	assert(helper2.GetReturn()).Equal([]Any{}, []base.Error{}, []base.Error{})
//	assert(testReadFromFile(tmpFile2)).Equal(testReadFromFile(snapshotFile2))
//
//	// Test(3)
//	helper3 := newTestProcessorReturnHelper()
//	processor3 := NewProcessor(
//		true,
//		1024,
//		2,
//		3,
//		nil,
//		5*time.Second,
//		nil,
//		helper3.GetFunction(),
//	)
//	defer processor3.Close()
//	err1 := processor3.BuildCache(
//		"pkgName",
//		path.Join(currDir, "processor_test.go/err_dir.go"),
//	)
//	assert(err1.GetKind()).Equal(base.ErrorKindRuntimePanic)
//	assert(strings.Contains(err1.GetMessage(), "processor_test.go")).
//		IsTrue()
//}
//
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
