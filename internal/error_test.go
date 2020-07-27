package internal

import (
	"testing"
)

func TestReportPanic(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(testRunWithCatchPanic(
		func() {
			reportPanic(NewReplyPanic("reply panic error"))
		},
	)).Equals(NewReplyPanic("reply panic error"))

}

func TestSubscribePanic(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := subscribePanic(nil)
	assert(o1).IsNil()

	// Test(2)
	o2 := subscribePanic(func(e Error) {})
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
	o2 := subscribePanic(func(e Error) {})
	assert(o2.Close()).IsTrue()

	// Test(3)
	o3 := subscribePanic(func(e Error) {})
	o3.Close()
	assert(o3.Close()).IsFalse()
}

func TestNewError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := NewError(ErrorKindRuntimePanic, "message", "fileLine")
	assert(o1.GetKind()).Equals(ErrorKindRuntimePanic)
	assert(o1.GetMessage()).Equals("message")
	assert(o1.GetDebug()).Equals("fileLine")
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
	assert(NewReplyPanic("message")).
		Equals(NewError(ErrorKindReplyPanic, "message", ""))
}

func TestNewRuntimeError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewRuntimePanic("message")).
		Equals(NewError(ErrorKindRuntimePanic, "message", ""))
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
	assert(NewKernelPanic("message")).
		Equals(NewError(ErrorKindKernelPanic, "message", ""))
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
	assert(ConvertToError(NewKernelPanic("test"))).IsNotNil()
}

func TestRpcError_GetKind(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := NewError(ErrorKindKernelPanic, "message", "fileLine")
	assert(o1.GetKind()).Equals(ErrorKindKernelPanic)
}

func TestRpcError_GetMessage(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := NewError(ErrorKindKernelPanic, "message", "fileLine")
	assert(o1.GetMessage()).Equals("message")
}

func TestRpcError_GetDebug(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := NewError(ErrorKindKernelPanic, "message", "fileLine")
	assert(o1.GetDebug()).Equals("fileLine")
}

func TestRpcError_AddDebug(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := NewError(ErrorKindKernelPanic, "message", "")
	assert(o1.AddDebug("fileLine").GetDebug()).Equals("fileLine")

	// Test(2)
	o2 := NewError(ErrorKindKernelPanic, "message", "fileLine")
	assert(o2.AddDebug("fileLine").GetDebug()).Equals("fileLine\nfileLine")
}

func TestRpcError_Error(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewError(ErrorKindKernelPanic, "", "").Error()).Equals("")

	// Test(2)
	assert(NewError(ErrorKindKernelPanic, "message", "").Error()).Equals("message")

	// Test(3)
	assert(NewError(ErrorKindKernelPanic, "", "fileLine").Error()).Equals("fileLine")

	// Test(4)
	assert(NewError(ErrorKindKernelPanic, "message", "fileLine").Error()).
		Equals("message\nfileLine")
}
