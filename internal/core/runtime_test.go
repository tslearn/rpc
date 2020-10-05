package core

import (
	sysError "errors"
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
				err := errors.ErrStreamIsBroken.AddDebug("error")
				ret, s := rt.Error(err), base.GetFileLine(0)
				source = rt.thread.GetReplyNode().path + " " + s
				assert(ret).Equal(emptyReturn)
				return ret
			}),
		)).Equal(nil, errors.ErrStreamIsBroken.AddDebug("error").AddDebug(source))
	})

	t.Run("test ok (error)", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(true, func(_ *Processor, rt Runtime) Return {
				err := sysError.New("error")
				ret, s := rt.Error(err), base.GetFileLine(0)
				source = rt.thread.GetReplyNode().path + " " + s
				assert(ret).Equal(emptyReturn)
				return ret
			}),
		)).Equal(nil, errors.ErrGeneralReplyError.
			AddDebug("error").
			AddDebug(source),
		)
	})

	t.Run("argument is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(true, func(_ *Processor, rt Runtime) Return {
				ret, s := rt.Error(nil), base.GetFileLine(0)
				source = rt.thread.GetReplyNode().path + " " + s
				assert(ret).Equal(emptyReturn)
				return ret
			}),
		)).Equal(nil, errors.ErrRuntimeErrorArgumentIsNil.AddDebug(source))
	})
}

//func TestRuntime_Call(t *testing.T) {
//  t.Run("thread lock error", func(t *testing.T) {
//    assert := base.NewAssert(t)
//    source := ""
//    assert(base.RunWithSubscribePanic(func() {
//      ret, s := Runtime{}.Call("#"), base.GetFileLine(0)
//      source = s
//      assert(ret).Equal(emptyReturn)
//    })).Equal(
//      errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(source),
//    )
//  })
//}
