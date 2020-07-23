package internal

import (
	"testing"
)

func TestReportPanic(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(testRunAndCatchPanic(
		func() {
			ReportPanic(NewReplyPanic("reply panic error"))
		},
	)).Equals(NewReplyPanic("reply panic error"))

}

func TestSubscribePanic(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := SubscribePanic(nil)
	assert(o1).IsNil()

	// Test(2)
	o2 := SubscribePanic(func(e Error) {})
	defer o2.Close()
	assert(o2).IsNotNil()
	assert(o2.id > 0).IsTrue()
	assert(o2.onPanic).IsNotNil()
	assert(len(gPanicSubscriptions)).Equals(1)
	assert(gPanicSubscriptions[0]).Equals(o2)
}

func TestRpcPanicSubscription_Close(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := (*rpcPanicSubscription)(nil)
	assert(o1.Close()).IsFalse()

	// Test(2)
	o2 := SubscribePanic(func(e Error) {})
	assert(o2.Close()).IsTrue()

	// Test(3)
	o3 := SubscribePanic(func(e Error) {})
	o3.Close()
	assert(o3.Close()).IsFalse()
}

func TestNewError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := NewError(ErrKindRuntime, "message", "debug")
	assert(o1.GetKind()).Equals(ErrKindRuntime)
	assert(o1.GetMessage()).Equals("message")
	assert(o1.GetDebug()).Equals("debug")
}

func TestNewBaseError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewBaseError("message")).
		Equals(NewError(ErrKindBase, "message", ""))
}

func TestNewReplyError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewReplyError("message")).
		Equals(NewError(ErrKindReply, "message", ""))
}

func TestNewReplyPanic(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewReplyPanic("message")).
		Equals(NewError(ErrKindReplyPanic, "message", ""))
}

func TestNewRuntimeError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewRuntimeError("message")).
		Equals(NewError(ErrKindRuntime, "message", ""))
}

func TestNewProtocolError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewProtocolError("message")).
		Equals(NewError(ErrKindProtocol, "message", ""))
}

func TestNewTransportError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewTransportError("message")).
		Equals(NewError(ErrKindTransport, "message", ""))
}

func TestNewKernelError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewKernelError("message")).
		Equals(NewError(ErrKindKernel, "message", ""))
}

func TestConvertToError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(ConvertToError(0)).IsNil()

	// Test(2)
	assert(ConvertToError(make(chan bool))).IsNil()

	// Test(3)
	assert(ConvertToError(nil)).IsNil()

	// Test(4)
	assert(ConvertToError(NewBaseError("test"))).IsNotNil()
}

func TestRpcError_GetKind(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := NewError(ErrKindKernel, "message", "debug")
	assert(o1.GetKind()).Equals(ErrKindKernel)
}

func TestRpcError_GetMessage(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := NewError(ErrKindKernel, "message", "debug")
	assert(o1.GetMessage()).Equals("message")
}

func TestRpcError_GetDebug(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := NewError(ErrKindKernel, "message", "debug")
	assert(o1.GetDebug()).Equals("debug")
}

func TestRpcError_AddDebug(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := NewError(ErrKindKernel, "message", "")
	assert(o1.AddDebug("debug").GetDebug()).Equals("debug")

	// Test(2)
	o2 := NewError(ErrKindKernel, "message", "debug")
	assert(o2.AddDebug("debug").GetDebug()).Equals("debug\ndebug")
}

func TestRpcError_Error(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewError(ErrKindKernel, "", "").Error()).Equals("")

	// Test(2)
	assert(NewError(ErrKindKernel, "message", "").Error()).Equals("message")

	// Test(3)
	assert(NewError(ErrKindKernel, "", "debug").Error()).Equals("debug")

	// Test(4)
	assert(NewError(ErrKindKernel, "message", "debug").Error()).
		Equals("message\ndebug")
}
