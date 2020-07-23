package internal

import (
	"testing"
	"time"
	"unsafe"
)

func TestContextObject_nilContext(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(nilContext).Equals(Context(nil))
}

func TestContextObject_getThread(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	ret1, error1, panic1, source1 :=
		testRunOnContext(false, func(ctx Context) Return {
			o1 := ContextObject{thread: nil}
			assert(o1.getThread()).IsNil()
			return ctx.OK(false)
		})
	assert(ret1, error1, panic1).Equals(
		false,
		nil,
		NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(source1),
	)

	// Test(2)
	ret2, error2, panic2, source2 :=
		testRunOnContext(false, func(ctx Context) Return {
			thread := getFakeThread(false)
			thread.execReplyNode = nil
			o1 := ContextObject{thread: unsafe.Pointer(thread)}
			assert(o1.getThread()).IsNil()
			return ctx.OK(false)
		})
	assert(ret2, error2, panic2).Equals(
		false,
		nil,
		NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(source2),
	)

	// Test(3)
	ret3, error3, panic3, _ :=
		testRunOnContext(false, func(ctx Context) Return {
			assert(ctx.getThread()).IsNotNil()
			return ctx.OK(true)
		})
	assert(ret3, error3, panic3).Equals(true, nil, nil)

	// Test(4)
	source4 := ""
	ret4, error4, panic4, _ :=
		testRunOnContext(true, func(ctx Context) Return {
			go func() {
				func(source string) {
					source4 = source
					assert(ctx.getThread()).IsNil()
				}(GetFileLine(0))
			}()
			time.Sleep(200 * time.Millisecond)
			return ctx.OK(false)
		})
	assert(ret4, error4, panic4).Equals(
		false,
		nil,
		NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(source4),
	)

	// Test(5)
	ret5, error5, panic5, _ :=
		testRunOnContext(true, func(ctx Context) Return {
			assert(ctx.getThread()).IsNotNil()
			return ctx.OK(true)
		})
	assert(ret5, error5, panic5).Equals(true, nil, nil)
}

func TestContextObject_stop(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	source1 := ""
	assert(testRunAndCatchPanic(func() {
		ret, source := Context(nil).stop(), GetFileLine(0)
		source1 = source
		assert(ret).IsFalse()
	})).Equals(NewKernelError(ErrStringUnexpectedNil).AddDebug(source1))

	// Test(2)
	ctx2 := getFakeContext(true)
	assert(ctx2.thread).IsNotNil()
	assert(ctx2.stop()).IsTrue()
	assert(ctx2.thread).IsNil()
}

func TestContextObject_OK(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	source1 := ""
	assert(testRunAndCatchPanic(func() {
		ret, source := Context(nil).OK(true), GetFileLine(0)
		source1 = source
		assert(ret).Equals(nilReturn)
	})).Equals(NewReplyPanic(ErrStringUnexpectedNil).AddDebug(source1))

	// Test(2)
	source2 := ""
	assert(testRunAndCatchPanic(func() {
		ret, source := (&ContextObject{thread: nil}).OK(true), GetFileLine(0)
		source2 = source
		assert(ret).Equals(nilReturn)
	})).Equals(NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(source2))

	// Test(3)
	ret3, error3, panic3, _ :=
		testRunOnContext(true, func(ctx Context) Return {
			assert(ctx.getThread()).IsNotNil()
			return ctx.OK(true)
		})
	assert(ret3, error3, panic3).Equals(true, nil, nil)
}

func TestContextObject_Error(t *testing.T) {
	//assert := NewAssert(t)
	//
	//// ctx is ok
	//processor := NewProcessor(16, 16, nil, nil)
	//rpcThread := newThread(newThreadPool(processor))
	//rpcThread.stop()
	//rpcThread.execSuccessful = true
	//ctx := ContextObject{rpcThread: unsafe.Pointer(rpcThread)}
	//ctx.writeError("errorMessage", "errorDebug")
	//assert(rpcThread.execSuccessful).IsFalse()
	//rpcThread.outStream.SetReadPos(streamBodyPos)
	//assert(rpcThread.outStream.ReadBool()).Equals(false, true)
	//assert(rpcThread.outStream.ReadString()).Equals("errorMessage", true)
	//assert(rpcThread.outStream.ReadString()).Equals("errorDebug", true)
	//assert(rpcThread.outStream.CanRead()).IsFalse()
	//
	//// ctx is stop
	//thread1 := newThread(nil)
	//thread1.stop()
	//thread1.execSuccessful = true
	//ctx1 := ContextObject{rpcThread: unsafe.Pointer(thread1)}
	//ctx1.stop()
	//ctx1.writeError("errorMessage", "errorDebug")
	//thread1.outStream.SetReadPos(streamBodyPos)
	//assert(thread1.execSuccessful).IsTrue()
	//assert(thread1.outStream.GetWritePos()).Equals(streamBodyPos)
	//assert(thread1.outStream.GetReadPos()).Equals(streamBodyPos)
}

func TestContext_Error(t *testing.T) {
	//assert := NewAssert(t)
	//
	//// ctx is ok
	//rpcThread := newThread(nil)
	//rpcThread.stop()
	//ctx := ContextObject{rpcThread: unsafe.Pointer(rpcThread)}
	//assert(ctx.Error(nil)).IsNil()
	//assert(rpcThread.outStream.GetWritePos()).Equals(streamBodyPos)
	//assert(rpcThread.outStream.GetReadPos()).Equals(streamBodyPos)
	//
	//assert(
	//	ctx.Error(NewErrorByDebug("errorMessage", "errorDebug")),
	//).IsNil()
	//rpcThread.outStream.SetReadPos(streamBodyPos)
	//assert(rpcThread.outStream.ReadBool()).Equals(false, true)
	//assert(rpcThread.outStream.ReadString()).Equals("errorMessage", true)
	//assert(rpcThread.outStream.ReadString()).Equals("errorDebug", true)
	//
	//// ctx have execReplyNode
	//thread1 := newThread(nil)
	//thread1.stop()
	//thread1.execReplyNode = &rpcReplyNode{debugString: "nodeDebug"}
	//ctx1 := ContextObject{rpcThread: unsafe.Pointer(thread1)}
	//assert(
	//	ctx1.Error(NewErrorByDebug("errorMessage", "errorDebug")),
	//).IsNil()
	//thread1.outStream.SetReadPos(streamBodyPos)
	//assert(thread1.outStream.ReadBool()).Equals(false, true)
	//assert(thread1.outStream.ReadString()).Equals("errorMessage", true)
	//assert(thread1.outStream.ReadString()).Equals("errorDebug\nnodeDebug", true)
}

func TestContext_Errorf(t *testing.T) {
	//assert := NewAssert(t)
	//
	//// ctx is ok
	//rpcThread := newThread(nil)
	//rpcThread.stop()
	//ctx := ContextObject{rpcThread: unsafe.Pointer(rpcThread)}
	//assert(ctx.Errorf("error%s", "Message")).IsNil()
	//rpcThread.outStream.SetReadPos(streamBodyPos)
	//assert(rpcThread.outStream.ReadBool()).Equals(false, true)
	//assert(rpcThread.outStream.ReadString()).Equals("errorMessage", true)
	//dbgMessage, ok := rpcThread.outStream.ReadString()
	//assert(ok).IsTrue()
	//
	//assert(strings.Contains(dbgMessage, "TestRpcContext_Errorf")).IsTrue()
}
