package internal

import (
	"testing"
)

func TestRpcContext_getThread(t *testing.T) {
	//assert := NewAssert(t)
	//
	//ctx1 := RPCContext{}
	//assert(ctx1.thread).IsNil()
	//
	//thread := newThread(nil)
	//thread.stop()
	//
	//ctx2 := RPCContext{thread: unsafe.Pointer(thread)}
	//assert(ctx2.thread).Equals(unsafe.Pointer(thread))

}

func TestRpcContext_stop(t *testing.T) {
	//assert := NewAssert(t)
	//
	//thread := newThread(nil)
	//thread.stop()
	//
	//ctx := RPCContext{thread: unsafe.Pointer(thread)}
	//assert(ctx.thread).IsNotNil()
	//ctx.stop()
	//assert(ctx.thread).IsNil()
}

func TestRpcContext_OK(t *testing.T) {
	//assert := NewAssert(t)
	//
	//// ctx is ok
	//thread := newThread(nil)
	//thread.stop()
	//assert(thread.execSuccessful).IsFalse()
	//ctx := RPCContext{thread: unsafe.Pointer(thread)}
	//ctx.OK(uint(215))
	//assert(thread.execSuccessful).IsTrue()
	//thread.outStream.SetReadPos(StreamBodyPos)
	//assert(thread.outStream.ReadBool()).Equals(true, true)
	//assert(thread.outStream.ReadUint64()).Equals(RPCUint(215), true)
	//assert(thread.outStream.CanRead()).IsFalse()
	//
	//// ctx is stop
	//thread1 := newThread(nil)
	//thread1.stop()
	//assert(thread1.execSuccessful).IsFalse()
	//ctx1 := RPCContext{thread: unsafe.Pointer(thread1)}
	//ctx1.stop()
	//ctx1.OK(uint(215))
	//thread1.outStream.SetReadPos(StreamBodyPos)
	//assert(thread1.outStream.GetWritePos()).Equals(StreamBodyPos)
	//assert(thread1.outStream.GetReadPos()).Equals(StreamBodyPos)
	//
	//// value is illegal
	//thread2 := newThread(nil)
	//thread2.stop()
	//assert(thread2.execSuccessful).IsFalse()
	//ctx2 := RPCContext{thread: unsafe.Pointer(thread2)}
	//ctx2.OK(make(chan bool))
	//assert(thread2.execSuccessful).IsFalse()
	//thread2.outStream.SetReadPos(StreamBodyPos)
	//assert(thread2.outStream.ReadBool()).Equals(false, true)
	//assert(thread2.outStream.ReadString()).Equals("return type is error", true)
	//dbgMessage, ok := thread2.outStream.ReadString()
	//assert(ok).IsTrue()
	//assert(strings.Contains(dbgMessage, "TestRpcContext_OK")).IsTrue()
	//assert(thread2.outStream.CanRead()).IsFalse()
}

func TestRpcContext_writeError(t *testing.T) {
	//assert := NewAssert(t)
	//
	//// ctx is ok
	//processor := NewRPCProcessor(16, 16, nil, nil)
	//thread := newThread(newThreadPool(processor))
	//thread.stop()
	//thread.execSuccessful = true
	//ctx := RPCContext{thread: unsafe.Pointer(thread)}
	//ctx.writeError("errorMessage", "errorDebug")
	//assert(thread.execSuccessful).IsFalse()
	//thread.outStream.SetReadPos(StreamBodyPos)
	//assert(thread.outStream.ReadBool()).Equals(false, true)
	//assert(thread.outStream.ReadString()).Equals("errorMessage", true)
	//assert(thread.outStream.ReadString()).Equals("errorDebug", true)
	//assert(thread.outStream.CanRead()).IsFalse()
	//
	//// ctx is stop
	//thread1 := newThread(nil)
	//thread1.stop()
	//thread1.execSuccessful = true
	//ctx1 := RPCContext{thread: unsafe.Pointer(thread1)}
	//ctx1.stop()
	//ctx1.writeError("errorMessage", "errorDebug")
	//thread1.outStream.SetReadPos(StreamBodyPos)
	//assert(thread1.execSuccessful).IsTrue()
	//assert(thread1.outStream.GetWritePos()).Equals(StreamBodyPos)
	//assert(thread1.outStream.GetReadPos()).Equals(StreamBodyPos)
}

func TestRpcContext_Error(t *testing.T) {
	//assert := NewAssert(t)
	//
	//// ctx is ok
	//thread := newThread(nil)
	//thread.stop()
	//ctx := RPCContext{thread: unsafe.Pointer(thread)}
	//assert(ctx.Error(nil)).IsNil()
	//assert(thread.outStream.GetWritePos()).Equals(StreamBodyPos)
	//assert(thread.outStream.GetReadPos()).Equals(StreamBodyPos)
	//
	//assert(
	//	ctx.Error(NewRPCErrorByDebug("errorMessage", "errorDebug")),
	//).IsNil()
	//thread.outStream.SetReadPos(StreamBodyPos)
	//assert(thread.outStream.ReadBool()).Equals(false, true)
	//assert(thread.outStream.ReadString()).Equals("errorMessage", true)
	//assert(thread.outStream.ReadString()).Equals("errorDebug", true)
	//
	//// ctx have execReplyNode
	//thread1 := newThread(nil)
	//thread1.stop()
	//thread1.execReplyNode = &rpcReplyNode{debugString: "nodeDebug"}
	//ctx1 := RPCContext{thread: unsafe.Pointer(thread1)}
	//assert(
	//	ctx1.Error(NewRPCErrorByDebug("errorMessage", "errorDebug")),
	//).IsNil()
	//thread1.outStream.SetReadPos(StreamBodyPos)
	//assert(thread1.outStream.ReadBool()).Equals(false, true)
	//assert(thread1.outStream.ReadString()).Equals("errorMessage", true)
	//assert(thread1.outStream.ReadString()).Equals("errorDebug\nnodeDebug", true)
}

func TestRpcContext_Errorf(t *testing.T) {
	//assert := NewAssert(t)
	//
	//// ctx is ok
	//thread := newThread(nil)
	//thread.stop()
	//ctx := RPCContext{thread: unsafe.Pointer(thread)}
	//assert(ctx.Errorf("error%s", "Message")).IsNil()
	//thread.outStream.SetReadPos(StreamBodyPos)
	//assert(thread.outStream.ReadBool()).Equals(false, true)
	//assert(thread.outStream.ReadString()).Equals("errorMessage", true)
	//dbgMessage, ok := thread.outStream.ReadString()
	//assert(ok).IsTrue()
	//
	//assert(strings.Contains(dbgMessage, "TestRpcContext_Errorf")).IsTrue()
}
