package core

import (
	"errors"
	"testing"
)

func TestNewRPCError(t *testing.T) {
	assert := newAssert(t)

	var testCollection = [][4]interface{}{
		{
			NewRPCError("hello"),
			"hello",
			"error_test.go",
		},
	}

	for _, item := range testCollection {
		assert(item[0].(*rpcError).GetMessage()).Equals(item[1])
		assert(item[0].(*rpcError).GetDebug()).Contains(item[2])
	}
}

func TestNewRPCErrorWithDebug(t *testing.T) {
	assert := newAssert(t)

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

func TestWrapSystemError(t *testing.T) {
	assert := newAssert(t)

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
	assert := newAssert(t)

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
	assert := newAssert(t)

	err := &rpcError{
		message: "message",
		debug:   "debug",
	}
	assert(err.GetMessage()).Equals("message")
}

func TestRpcError_GetDebug(t *testing.T) {
	assert := newAssert(t)

	err := &rpcError{
		message: "message",
		debug:   "debug",
	}
	assert(err.GetDebug()).Equals("debug")
}

func TestRpcError_AddDebug(t *testing.T) {
	assert := newAssert(t)

	err := &rpcError{
		message: "message",
		debug:   "",
	}
	err.AddDebug("m1")
	assert(err.GetDebug()).Equals("m1")
	err.AddDebug("m2")
	assert(err.GetDebug()).Equals("m1\nm2")
}
