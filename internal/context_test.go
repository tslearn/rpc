package internal

import (
	"testing"
	"unsafe"
)

func TestContext_getThread(t *testing.T) {
	assert := NewAssert(t)

	// Test(1) thread is nil
	ctx1 := &ContextObject{thread: nil}
	assert(ctx1.getThread()).IsNil()

	// Test(2) thread is not nil\
	fakeThread2 := getFakeThread()
	defer fakeThread2.Stop()
	ctx2 := &ContextObject{thread: unsafe.Pointer(fakeThread2)}
	assert(ctx2.getThread()).Equals(fakeThread2)
}

func TestContext_stop(t *testing.T) {
	assert := NewAssert(t)

	// Test(1) thread is nil
	ctx1 := &ContextObject{thread: nil}
	assert(ctx1.getThread()).IsNil()
	ctx1.stop()
	assert(ctx1.getThread()).IsNil()

	// Test(2) thread is not nil
	fakeThread2 := getFakeThread()
	defer fakeThread2.Stop()
	ctx2 := &ContextObject{thread: unsafe.Pointer(fakeThread2)}
	assert(ctx2.getThread()).IsNotNil()
	ctx2.stop()
	assert(ctx2.getThread()).IsNil()
}

func TestContext_OK(t *testing.T) {
	// prepare
	//assert := NewAssert(t)

	//// Test(1) ctx is ok
	//fakeThread1 := getFakeThread()
	//defer fakeThread1.Stop()
	//ctx1 := &ContextObject{thread: unsafe.Pointer(fakeThread1)}
	//ctx1.OK("Hello World")
	//assert(fakeThread1.execSuccessful).IsTrue()
	//fakeThread1.outStream.SetReadPosToBodyStart()
	//assert(fakeThread1.outStream.ReadBool()).Equals(true, true)
	//assert(fakeThread1.outStream.ReadString()).Equals("Hello World", true)
	//assert(fakeThread1.outStream.CanRead()).IsFalse()
	//
	//// Test(2) ctx is stop
	//fakeThread2 := getFakeThread()
	//defer fakeThread2.Stop()
	//ctx2 := &ContextObject{thread: unsafe.Pointer(fakeThread2)}
	//ctx2.stop()
	//ctx2.OK("Hello World")
	//fakeThread2.outStream.SetReadPosToBodyStart()
	//assert(fakeThread2.outStream.GetWritePos()).Equals(streamBodyPos)
	//assert(fakeThread2.outStream.GetReadPos()).Equals(streamBodyPos)

	//// Test(3) value is illegal
	//fakeThread3 := getFakeThread()
	//defer fakeThread3.Stop()
	//ctx3 := ContextObject{thread: unsafe.Pointer(fakeThread3)}
	//ctx3.OK(make(chan bool))
	//assert(fakeThread3.execSuccessful).IsFalse()
	//fakeThread3.outStream.SetReadPosToBodyStart()
	//assert(fakeThread3.outStream.ReadBool()).Equals(false, true)
	//assert(fakeThread3.outStream.ReadString()).Equals("return type is error", true)
	//dbgMessage, ok := thread2.outStream.ReadString()
	//assert(ok).IsTrue()
	//assert(strings.Contains(dbgMessage, "TestRpcContext_OK")).IsTrue()
	//assert(thread2.outStream.CanRead()).IsFalse()
}

func TestContext_writeError(t *testing.T) {
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
