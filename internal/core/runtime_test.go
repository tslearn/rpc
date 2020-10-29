package core

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"testing"
)

var testRuntime = Runtime{id: 0, thread: getFakeThread(true)}

func TestRuntime_lock(t *testing.T) {
	t.Run("thread is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(Runtime{}.lock()).IsNil()
	})

	t.Run("thread is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		thread := getFakeThread(false)
		assert(Runtime{id: thread.top.lockStatus, thread: thread}.lock()).
			Equal(thread)
	})
}

func TestRuntime_unlock(t *testing.T) {
	t.Run("thread is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(base.RunWithCatchPanic(func() {
			Runtime{}.unlock()
		})).IsNil()
	})

	t.Run("thread is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		thread := getFakeThread(false)
		rt := Runtime{id: thread.top.lockStatus, thread: thread}
		assert(base.RunWithCatchPanic(func() {
			rt.lock()
			rt.unlock()
		})).IsNil()
		// if unlock is success, the thread can lock again
		assert(rt.lock()).Equal(thread)
	})
}

func TestRuntime_Reply(t *testing.T) {
	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(base.RunWithSubscribePanic(func() {
			_, source = Runtime{}.Reply(true), base.GetFileLine(0)
		})).Equal(
			errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(source),
		)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(true, func(_ *Processor, rt Runtime) Return {
				return rt.Reply(true)
			}),
		)).Equal(true, nil)
	})

	t.Run("test ok (Error)", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(true, func(_ *Processor, rt Runtime) Return {
				err := errors.ErrStream.AddDebug("error")
				ret, s := rt.Reply(err), base.GetFileLine(0)
				source = rt.thread.GetActionNode().path + " " + s
				assert(ret).Equal(emptyReturn)
				return ret
			}),
		)).Equal(nil, errors.ErrStream.AddDebug("error").AddDebug(source))
	})

	t.Run("test ok (error)", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(true, func(_ *Processor, rt Runtime) Return {
				err := errors.ErrActionCustom.AddDebug("error")
				ret, s := rt.Reply(err), base.GetFileLine(0)
				source = rt.thread.GetActionNode().path + " " + s
				assert(ret).Equal(emptyReturn)
				return ret
			}),
		)).Equal(nil, errors.ErrActionCustom.AddDebug("error").AddDebug(source))
	})

	t.Run("argument is (*Error)(nil)", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(true, func(_ *Processor, rt Runtime) Return {
				ret, s := rt.Reply((*base.Error)(nil)), base.GetFileLine(0)
				source = rt.thread.GetActionNode().path + " " + s
				assert(ret).Equal(emptyReturn)
				return ret
			}),
		)).Equal(
			nil,
			errors.ErrUnsupportedValue.
				AddDebug("value is nil").
				AddDebug(source),
		)
	})

	t.Run("argument is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(true, func(_ *Processor, rt Runtime) Return {
				ret, s := rt.Reply((*base.Error)(nil)), base.GetFileLine(0)
				source = rt.thread.GetActionNode().path + " " + s
				assert(ret).Equal(emptyReturn)
				return ret
			}),
		)).Equal(
			nil,
			errors.ErrUnsupportedValue.
				AddDebug("value is nil").
				AddDebug(source),
		)
	})
}

func TestRuntime_Call(t *testing.T) {
	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		ret, source := Runtime{}.Call("#"), base.GetFileLine(0)
		assert(ret).Equal(RTValue{
			err: errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(source),
		})
	})

	t.Run("make stream error", func(t *testing.T) {
		assert := base.NewAssert(t)
		source1 := ""
		source2 := ""

		assert(ParseResponseStream(
			testWithProcessorAndRuntime(true, func(_ *Processor, rt Runtime) Return {
				errArg := make(chan bool)
				rtValue, s1 := rt.Call("#.test.SayHello", errArg), base.GetFileLine(0)
				source1 = rt.thread.GetActionNode().path + " " + s1
				_, err := rtValue.ToString()
				ret, s2 := rt.Reply(err), base.GetFileLine(0)
				source2 = rt.thread.GetActionNode().path + " " + s2
				return ret
			}),
		)).Equal(nil, errors.ErrUnsupportedValue.
			AddDebug("2nd argument: value is not supported").
			AddDebug(source1).AddDebug(source2),
		)
	})

	t.Run("depth overflow", func(t *testing.T) {
		assert := base.NewAssert(t)
		source1 := ""
		source2 := ""
		source3 := ""
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(
				true,
				func(processor *Processor, rt Runtime) Return {
					// set max call depth to 1
					oldMaxCallDepth := processor.maxCallDepth
					processor.maxCallDepth = 1
					defer func() {
						processor.maxCallDepth = oldMaxCallDepth
					}()

					rtValue, s2 := rt.Call("#.test:SayHello", "ts"), base.GetFileLine(0)
					source1 = "#.test:SayHello " +
						rt.thread.processor.actionsMap["#.test:SayHello"].meta.fileLine
					source2 = rt.thread.GetActionNode().path + " " + s2
					_, err := rtValue.ToString()
					ret, s3 := rt.Reply(err), base.GetFileLine(0)
					source3 = rt.thread.GetActionNode().path + " " + s3
					return ret
				},
			),
		)).Equal(nil, errors.ErrCallOverflow.
			AddDebug("call #.test:SayHello level(1) overflows").
			AddDebug(source1).AddDebug(source2).AddDebug(source3),
		)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(
				true,
				func(processor *Processor, rt Runtime) Return {
					rtValue := rt.Call("#.test:SayHello", "ts")
					fmt.Println(rtValue.ToString())
					return rt.Reply(rtValue)
				},
			),
		)).Equal("hello ts", nil)
	})
}

func TestRuntime_ParseResponseStream(t *testing.T) {
	t.Run("errCode format error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.WriteInt64(3)
		assert(testRuntime.ParseResponseStream(v)).Equal(
			RTValue{err: errors.ErrStream},
		)
	})

	t.Run("Read ret error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.WriteUint64(0)
		assert(testRuntime.ParseResponseStream(v)).Equal(
			RTValue{err: errors.ErrStream},
		)
	})

	t.Run("Read ret ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.WriteUint64(0)
		v.WriteBool(true)
		assert(testRuntime.ParseResponseStream(v).ToBool()).Equal(true, nil)
	})

	t.Run("error message Read error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.WriteUint64(uint64(base.ErrorTypeSecurity))
		v.WriteBool(true)
		assert(testRuntime.ParseResponseStream(v)).Equal(
			RTValue{err: errors.ErrStream},
		)
	})

	t.Run("error stream is not finish", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.WriteUint64(errors.ErrStream.GetCode())
		v.WriteString(errors.ErrStream.GetMessage())
		v.WriteBool(true)
		assert(testRuntime.ParseResponseStream(v)).Equal(
			RTValue{err: errors.ErrStream},
		)
	})

	t.Run("error stream ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.WriteUint64(errors.ErrStream.GetCode())
		v.WriteString(errors.ErrStream.GetMessage())
		assert(testRuntime.ParseResponseStream(v)).Equal(
			RTValue{err: errors.ErrStream},
		)
	})
}

func TestRuntime_NewRTArray(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 0; i < 100; i++ {
			testRuntime.thread.Reset()
			v := testRuntime.NewRTArray(i)
			assert(v.rt).Equal(testRuntime)
			assert(len(*v.items), cap(*v.items)).Equal(0, i)
		}
	})
}

func TestRuntime_NewRTMap(t *testing.T) {

}

func TestRuntime_GetServiceData(t *testing.T) {

}

func TestRuntime_SetServiceData(t *testing.T) {

}
