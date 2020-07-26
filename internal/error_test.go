package internal

import (
	"testing"
)

func TestReportPanic(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(testRunWithCatchPanic(
		func() {
			ReportPanic(NewReplyFatal("reply panic error"))
		},
	)).Equals(NewReplyFatal("reply panic error"))

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
	o1 := NewError(ErrorKindRuntimeFatal, "message", "fileLine")
	assert(o1.GetKind()).Equals(ErrorKindRuntimeFatal)
	assert(o1.GetMessage()).Equals("message")
	assert(o1.GetDebug()).Equals("fileLine")
}

func TestNewBaseError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewBaseError("message")).
		Equals(NewError(ErrorKindBase, "message", ""))
}

func TestNewReplyError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewReplyError("message")).
		Equals(NewError(ErrorKindReply, "message", ""))
}

func TestNewReplyPanic(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewReplyFatal("message")).
		Equals(NewError(ErrorKindReplyFatal, "message", ""))
}

func TestNewRuntimeError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewRuntimeFatal("message")).
		Equals(NewError(ErrorKindRuntimeFatal, "message", ""))
}

func TestNewProtocolError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewProtocolError("message")).
		Equals(NewError(ErrorKindProtocol, "message", ""))
}

func TestNewTransportError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewTransportError("message")).
		Equals(NewError(ErrorKindTransport, "message", ""))
}

func TestNewKernelError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewKernelFatal("message")).
		Equals(NewError(ErrorKindKernelFatal, "message", ""))
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
	o1 := NewError(ErrorKindKernelFatal, "message", "fileLine")
	assert(o1.GetKind()).Equals(ErrorKindKernelFatal)
}

func TestRpcError_GetMessage(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := NewError(ErrorKindKernelFatal, "message", "fileLine")
	assert(o1.GetMessage()).Equals("message")
}

func TestRpcError_GetDebug(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := NewError(ErrorKindKernelFatal, "message", "fileLine")
	assert(o1.GetDebug()).Equals("fileLine")
}

func TestRpcError_AddDebug(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := NewError(ErrorKindKernelFatal, "message", "")
	assert(o1.AddDebug("fileLine").GetDebug()).Equals("fileLine")

	// Test(2)
	o2 := NewError(ErrorKindKernelFatal, "message", "fileLine")
	assert(o2.AddDebug("fileLine").GetDebug()).Equals("fileLine\nfileLine")
}

func TestRpcError_Error(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewError(ErrorKindKernelFatal, "", "").Error()).Equals("")

	// Test(2)
	assert(NewError(ErrorKindKernelFatal, "message", "").Error()).Equals("message")

	// Test(3)
	assert(NewError(ErrorKindKernelFatal, "", "fileLine").Error()).Equals("fileLine")

	// Test(4)
	assert(NewError(ErrorKindKernelFatal, "message", "fileLine").Error()).
		Equals("message\nfileLine")
}
