package base

import (
	"fmt"
	"math"
	"testing"
)

func TestErrorType(t *testing.T) {
	t.Run("check constant", func(t *testing.T) {
		assert := NewAssert(t)

		assert(ErrorTypeProtocol).Equal(ErrorType(1))
		assert(ErrorTypeTransport).Equal(ErrorType(2))
		assert(ErrorTypeReply).Equal(ErrorType(3))
		assert(ErrorTypeRuntime).Equal(ErrorType(4))
		assert(ErrorTypeKernel).Equal(ErrorType(5))
		assert(ErrorTypeSecurity).Equal(ErrorType(6))
	})
}

func TestErrorLevel(t *testing.T) {
	t.Run("check constant", func(t *testing.T) {
		assert := NewAssert(t)

		assert(ErrorLevelWarn).Equal(ErrorLevel(1))
		assert(ErrorLevelError).Equal(ErrorLevel(2))
		assert(ErrorLevelFatal).Equal(ErrorLevel(3))
	})
}

func TestDefineError(t *testing.T) {
	t.Run("error redefined", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := defineError(ErrorTypeReply, 321, ErrorLevelWarn, "msg", "source")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(RunWithCatchPanic(func() {
			_ = defineError(ErrorTypeReply, 321, ErrorLevelWarn, "msg", "source")
		})).Equal("Error redefined :\n>>> source\n>>> source\n")
	})

	t.Run("ok", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := defineError(ErrorTypeReply, 321, ErrorLevelWarn, "msg", "source")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetCode(), v1.GetLevel(), v1.GetMessage()).
			Equal(ErrorTypeReply, ErrorCode(321), ErrorLevelWarn, "msg")
		assert(errorDefineMap[v1.code]).Equal("source")
	})
}

func TestDefineProtocolError(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		assert := NewAssert(t)
		v1, s1 := DefineProtocolError(321, ErrorLevelWarn, "msg"), GetFileLine(0)
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetCode(), v1.GetLevel(), v1.GetMessage()).
			Equal(ErrorTypeProtocol, ErrorCode(321), ErrorLevelWarn, "msg")
		assert(errorDefineMap[v1.code]).Equal(s1)
	})
}

func TestDefineTransportError(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		assert := NewAssert(t)
		v1, s1 := DefineTransportError(321, ErrorLevelWarn, "msg"), GetFileLine(0)
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetCode(), v1.GetLevel(), v1.GetMessage()).
			Equal(ErrorTypeTransport, ErrorCode(321), ErrorLevelWarn, "msg")
		assert(errorDefineMap[v1.code]).Equal(s1)
	})
}

func TestDefineReplyError(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		assert := NewAssert(t)
		v1, s1 := DefineReplyError(321, ErrorLevelWarn, "msg"), GetFileLine(0)
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetCode(), v1.GetLevel(), v1.GetMessage()).
			Equal(ErrorTypeReply, ErrorCode(321), ErrorLevelWarn, "msg")
		assert(errorDefineMap[v1.code]).Equal(s1)
	})
}

func TestDefineRuntimeError(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		assert := NewAssert(t)
		v1, s1 := DefineRuntimeError(321, ErrorLevelWarn, "msg"), GetFileLine(0)
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetCode(), v1.GetLevel(), v1.GetMessage()).
			Equal(ErrorTypeRuntime, ErrorCode(321), ErrorLevelWarn, "msg")
		assert(errorDefineMap[v1.code]).Equal(s1)
	})
}

func TestDefineKernelError(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		assert := NewAssert(t)
		v1, s1 := DefineKernelError(321, ErrorLevelWarn, "msg"), GetFileLine(0)
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetCode(), v1.GetLevel(), v1.GetMessage()).
			Equal(ErrorTypeKernel, ErrorCode(321), ErrorLevelWarn, "msg")
		assert(errorDefineMap[v1.code]).Equal(s1)
	})
}

func TestDefineSecurityError(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		assert := NewAssert(t)
		v1, s1 := DefineSecurityError(321, ErrorLevelWarn, "msg"), GetFileLine(0)
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetCode(), v1.GetLevel(), v1.GetMessage()).
			Equal(ErrorTypeSecurity, ErrorCode(321), ErrorLevelWarn, "msg")
		assert(errorDefineMap[v1.code]).Equal(s1)
	})
}

func TestNewReplyError(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewReplyError(ErrorLevelWarn, "msg")
		assert(v1.GetType(), v1.GetCode(), v1.GetLevel(), v1.GetMessage()).
			Equal(ErrorTypeReply, ErrorCode(0), ErrorLevelWarn, "msg")
		s1, ok1 := errorDefineMap[v1.code]
		assert(s1, ok1).Equal("", false)
	})
}

func TestError_GetType(t *testing.T) {
	t.Run("ok ErrorTypeProtocol", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineProtocolError(321, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType()).Equal(ErrorTypeProtocol)
	})

	t.Run("ok ErrorTypeTransport", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineTransportError(321, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType()).Equal(ErrorTypeTransport)
	})

	t.Run("ok ErrorTypeReply", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineReplyError(321, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType()).Equal(ErrorTypeReply)
	})

	t.Run("ok ErrorTypeRuntime", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineRuntimeError(321, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType()).Equal(ErrorTypeRuntime)
	})

	t.Run("ok ErrorTypeKernel", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineKernelError(321, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType()).Equal(ErrorTypeKernel)
	})

	t.Run("ok ErrorTypeSecurity", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(321, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType()).Equal(ErrorTypeSecurity)
	})

	t.Run("ok ErrorTypeSecurity", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewReplyError(ErrorLevelWarn, "msg")
		assert(v1.GetType()).Equal(ErrorTypeReply)
	})
}

func TestError_GetLevel(t *testing.T) {
	t.Run("ok ErrorLevelWarn", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(321, ErrorLevelWarn, "msg")
		v2 := NewReplyError(ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetLevel()).Equal(ErrorLevelWarn)
		assert(v2.GetLevel()).Equal(ErrorLevelWarn)
	})

	t.Run("ok ErrorLevelError", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(321, ErrorLevelError, "msg")
		v2 := NewReplyError(ErrorLevelError, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetLevel()).Equal(ErrorLevelError)
		assert(v2.GetLevel()).Equal(ErrorLevelError)
	})

	t.Run("ok ErrorLevelFatal", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(321, ErrorLevelFatal, "msg")
		v2 := NewReplyError(ErrorLevelFatal, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetLevel()).Equal(ErrorLevelFatal)
		assert(v2.GetLevel()).Equal(ErrorLevelFatal)
	})
}

func TestError_GetCode(t *testing.T) {
	t.Run("ok with uint32 min", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(0, ErrorLevelFatal, "msg")
		v2 := NewReplyError(ErrorLevelFatal, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetCode()).Equal(ErrorCode(0))
		assert(v2.GetCode()).Equal(ErrorCode(0))
	})

	t.Run("ok with uint32 max", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(math.MaxUint32, ErrorLevelFatal, "msg")
		v2 := NewReplyError(ErrorLevelFatal, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetCode()).Equal(ErrorCode(math.MaxUint32))
		assert(v2.GetCode()).Equal(ErrorCode(0))
	})
}

func TestError_GetMessage(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(math.MaxUint32, ErrorLevelFatal, "msg1")
		v2 := NewReplyError(ErrorLevelFatal, "msg2")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetMessage()).Equal("msg1")
		assert(v2.GetMessage()).Equal("msg2")
	})
}

func TestError_AddDebug(t *testing.T) {
	t.Run("ok from origin error", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(1, ErrorLevelFatal, "")
		v2 := DefineSecurityError(2, ErrorLevelFatal, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			delete(errorDefineMap, v2.code)
			errorDefineMutex.Unlock()
		}()
		v3 := v1.AddDebug("dbg")
		v4 := v2.AddDebug("dbg")
		assert(fmt.Sprintf("%p", v3) == fmt.Sprintf("%p", v1)).IsFalse()
		assert(fmt.Sprintf("%p", v4) == fmt.Sprintf("%p", v2)).IsFalse()
		assert(v3.GetMessage()).Equal(">>> dbg")
		assert(v4.GetMessage()).Equal("msg\n>>> dbg")
	})

	t.Run("ok from derived error", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewReplyError(ErrorLevelFatal, "")
		v2 := NewReplyError(ErrorLevelFatal, "msg")
		v3 := v1.AddDebug("dbg")
		v4 := v2.AddDebug("dbg")
		assert(fmt.Sprintf("%p", v3) == fmt.Sprintf("%p", v1)).IsTrue()
		assert(fmt.Sprintf("%p", v4) == fmt.Sprintf("%p", v2)).IsTrue()
		assert(v3.GetMessage()).Equal(">>> dbg")
		assert(v4.GetMessage()).Equal("msg\n>>> dbg")
	})
}

func TestError_getErrorTypeString(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineProtocolError(2, ErrorLevelFatal, "msg")
		v2 := DefineTransportError(2, ErrorLevelFatal, "msg")
		v3 := DefineReplyError(2, ErrorLevelFatal, "msg")
		v4 := DefineRuntimeError(2, ErrorLevelFatal, "msg")
		v5 := DefineKernelError(2, ErrorLevelFatal, "msg")
		v6 := DefineSecurityError(2, ErrorLevelFatal, "msg")
		v7 := &Error{}
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			delete(errorDefineMap, v2.code)
			delete(errorDefineMap, v3.code)
			delete(errorDefineMap, v4.code)
			delete(errorDefineMap, v5.code)
			delete(errorDefineMap, v6.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.getErrorTypeString()).Equal("Protocol")
		assert(v2.getErrorTypeString()).Equal("Transport")
		assert(v3.getErrorTypeString()).Equal("Reply")
		assert(v4.getErrorTypeString()).Equal("Runtime")
		assert(v5.getErrorTypeString()).Equal("Kernel")
		assert(v6.getErrorTypeString()).Equal("Security")
		assert(v7.getErrorTypeString()).Equal("")
	})
}

func TestError_getErrorLevelString(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineProtocolError(2, ErrorLevelWarn, "msg")
		v2 := DefineProtocolError(2, ErrorLevelError, "msg")
		v3 := DefineProtocolError(2, ErrorLevelFatal, "msg")
		v4 := DefineRuntimeError(2, 0, "msg")
		v5 := DefineRuntimeError(2, 255, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			delete(errorDefineMap, v2.code)
			delete(errorDefineMap, v3.code)
			delete(errorDefineMap, v4.code)
			delete(errorDefineMap, v5.code)
			errorDefineMutex.Unlock()
		}()
		assert(v1.getErrorLevelString()).Equal("Warn")
		assert(v2.getErrorLevelString()).Equal("Error")
		assert(v3.getErrorLevelString()).Equal("Fatal")
		assert(v4.getErrorLevelString()).Equal("")
		assert(v5.getErrorLevelString()).Equal("")
	})
}

func TestError_Error(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineProtocolError(2, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, v1.code)
			errorDefineMutex.Unlock()
		}()
		v2 := v1.AddDebug("dbg")
		assert(v2.Error()).Equal("ProtocolWarn: msg\n>>> dbg")
	})
}
