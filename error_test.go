package common

import (
	"errors"
	"testing"
)

func TestNewRPCError(t *testing.T) {
	assert := NewAssert(t)

	var testCollection = [][4]interface{}{
		{
			NewRPCError("hello"),
			"hello",
			"error_test.go",
		},
	}

	for _, item := range testCollection {
		assert(item[0].(RPCError).GetMessage()).Equals(item[1])
		assert(item[0].(RPCError).GetDebug()).Contains(item[2])
	}
}

func TestNewRPCErrorWithDebug(t *testing.T) {
	assert := NewAssert(t)

	var testCollection = [][2]interface{}{
		{
			NewRPCErrorWithDebug("", ""),
			"[RPCError]\n",
		}, {
			NewRPCErrorWithDebug("message", ""),
			"[RPCError message]\n",
		}, {
			NewRPCErrorWithDebug("", "debug"),
			"[RPCError]\nDebug:\n\tdebug\n",
		}, {
			NewRPCErrorWithDebug("message", "debug"),
			"[RPCError message]\nDebug:\n\tdebug\n",
		},
	}
	for _, item := range testCollection {
		assert(item[0].(RPCError).Error()).Equals(item[1])
	}
}

func TestWrapSystemError(t *testing.T) {
	assert := NewAssert(t)

	// wrap nil error
	err := WrapSystemError(nil)
	assert(err).IsNil()

	// wrap error
	err = WrapSystemError(
		errors.New("custom error"),
	)
	assert(err.GetMessage()).Equals("custom error")
	assert(err.GetDebug()).Equals("")
}

func TestWrapSystemErrorWithDebug(t *testing.T) {
	assert := NewAssert(t)

	// wrap nil error
	err := WrapSystemErrorWithDebug(nil)
	assert(err).IsNil()

	// wrap error
	err = WrapSystemErrorWithDebug(
		errors.New("custom error"),
	)
	assert(err.GetMessage()).Equals("custom error")
	assert(err.GetDebug()).Contains("error_test.go")
}

func TestRpcError_GetMessage(t *testing.T) {
	assert := NewAssert(t)

	err := &rpcError{
		message: "message",
		debug:   "debug",
	}
	assert(err.GetMessage()).Equals("message")
}

func TestRpcError_GetDebug(t *testing.T) {
	assert := NewAssert(t)

	err := &rpcError{
		message: "message",
		debug:   "debug",
	}
	assert(err.GetDebug()).Equals("debug")
}

func TestRpcError_AddDebug(t *testing.T) {
	assert := NewAssert(t)

	err := &rpcError{
		message: "message",
		debug:   "",
	}
	err.AddDebug("m1")
	assert(err.GetDebug()).Equals("m1")
	err.AddDebug("m2")
	assert(err.GetDebug()).Equals("m1\nm2")
}
