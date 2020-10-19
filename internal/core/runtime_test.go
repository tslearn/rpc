package core

import (
	sysError "errors"
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"testing"
)

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

func TestContextObject_OK(t *testing.T) {
	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(base.RunWithSubscribePanic(func() {
			_, source = Runtime{}.OK(true), base.GetFileLine(0)
		})).Equal(
			errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(source),
		)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(true, func(_ *Processor, rt Runtime) Return {
				return rt.OK(true)
			}),
		)).Equal(true, nil)
	})
}

func TestContextObject_Error(t *testing.T) {
	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(base.RunWithSubscribePanic(func() {
			ret, s := Runtime{}.Error(sysError.New("error")), base.GetFileLine(0)
			source = s
			assert(ret).Equal(emptyReturn)
		})).Equal(
			errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(source),
		)
	})

	t.Run("test ok (Error)", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(true, func(_ *Processor, rt Runtime) Return {
				err := errors.ErrStream.AddDebug("error")
				ret, s := rt.Error(err), base.GetFileLine(0)
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
				err := sysError.New("error")
				ret, s := rt.Error(err), base.GetFileLine(0)
				source = rt.thread.GetActionNode().path + " " + s
				assert(ret).Equal(emptyReturn)
				return ret
			}),
		)).Equal(nil, errors.ErrActionCustom.AddDebug("error").AddDebug(source))
	})

	t.Run("argument is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(true, func(_ *Processor, rt Runtime) Return {
				ret, s := rt.Error(nil), base.GetFileLine(0)
				source = rt.thread.GetActionNode().path + " " + s
				assert(ret).Equal(emptyReturn)
				return ret
			}),
		)).Equal(
			nil,
			errors.ErrUnsupportedValue.
				AddDebug("value type(nil) is not supported").
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
				ret, s2 := rt.Error(err), base.GetFileLine(0)
				source2 = rt.thread.GetActionNode().path + " " + s2
				return ret
			}),
		)).Equal(nil, errors.ErrUnsupportedValue.
			AddDebug("2nd argument: value type(chan bool) is not supported").
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
					ret, s3 := rt.Error(err), base.GetFileLine(0)
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
					return rt.OK(rtValue)
				},
			),
		)).Equal("hello ts", nil)
	})
}
