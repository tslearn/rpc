package internal

import (
	"errors"
	"testing"
)

func TestNewRPCError(t *testing.T) {
	assert := NewAssert(t)

	assert(NewRPCError("hello").GetMessage()).Equals("hello")
	assert(NewRPCError("hello").GetDebug()).Equals("")
}

func TestNewRPCErrorByDebug(t *testing.T) {
	assert := NewAssert(t)

	var testCollection = [][2]interface{}{
		{
			NewRPCErrorByDebug("", ""),
			"",
		}, {
			NewRPCErrorByDebug("message", ""),
			"message\n",
		}, {
			NewRPCErrorByDebug("", "debug"),
			"Debug:\n\tdebug\n",
		}, {
			NewRPCErrorByDebug("message", "debug"),
			"message\nDebug:\n\tdebug\n",
		},
	}
	for _, item := range testCollection {
		assert(item[0].(*rpcError).Error()).Equals(item[1])
	}
}

func TestNewRPCErrorByError(t *testing.T) {
	assert := NewAssert(t)

	// wrap nil error
	err := NewRPCErrorByError(nil)
	assert(err == nil).IsTrue()

	// wrap error
	err = NewRPCErrorByError(errors.New("custom error"))
	assert(err.GetMessage()).Equals("custom error")
	assert(err.GetDebug()).Equals("")
}

func TestConvertToRPCError(t *testing.T) {
	assert := NewAssert(t)

	assert(ConvertToRPCError(0)).IsNil()
	assert(ConvertToRPCError(make(chan bool))).IsNil()
	assert(ConvertToRPCError(nil)).IsNil()
	assert(ConvertToRPCError(NewRPCError("test"))).IsNotNil()
}

func TestRpcError_GetMessage(t *testing.T) {
	assert := NewAssert(t)
	assert(NewRPCErrorByDebug("message", "debug").GetMessage()).Equals("message")
}

func TestRpcError_GetDebug(t *testing.T) {
	assert := NewAssert(t)
	assert(NewRPCErrorByDebug("message", "debug").GetDebug()).Equals("debug")
}

func TestRpcError_AddDebug(t *testing.T) {
	assert := NewAssert(t)

	err := NewRPCError("message")
	err.AddDebug("m1")
	assert(err.GetDebug()).Equals("m1")
	err.AddDebug("m2")
	assert(err.GetDebug()).Equals("m1\nm2")
}

func TestRpcError_GetExtra(t *testing.T) {
	assert := NewAssert(t)

	err := NewRPCError("message")
	err.SetExtra("extra")
	assert(err.GetExtra()).Equals("extra")
}

func TestRpcError_SetExtra(t *testing.T) {
	assert := NewAssert(t)

	err := NewRPCError("message")
	err.SetExtra("extra")
	assert(err.(*rpcError).extra).Equals("extra")
}
