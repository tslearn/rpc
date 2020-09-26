package base

import (
	"fmt"
	"math"
	"testing"
)

func TestError(t *testing.T) {
	t.Run("check constant", func(t *testing.T) {
		assert := NewAssert(t)

		assert(ErrorTypeProtocol).Equal(ErrorType(1))
		assert(ErrorTypeTransport).Equal(ErrorType(2))
		assert(ErrorTypeReply).Equal(ErrorType(3))
		assert(ErrorTypeRuntime).Equal(ErrorType(4))
		assert(ErrorTypeKernel).Equal(ErrorType(5))
		assert(ErrorTypeSecurity).Equal(ErrorType(6))
		assert(ErrorTypeCustom).Equal(ErrorType(7))

		assert(ErrorLevelWarn).Equal(ErrorLevel(1))
		assert(ErrorLevelError).Equal(ErrorLevel(2))
		assert(ErrorLevelFatal).Equal(ErrorLevel(3))
	})
}

func TestNewError(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(NewError(123, "msg")).Equal(&Error{code: 123, message: "msg"})
	})
}

func TestDefineError(t *testing.T) {
	num := ErrorNumber(math.MaxUint32)

	t.Run("error redefined", func(t *testing.T) {
		assert := NewAssert(t)
		_ = defineError(ErrorTypeReply, num, ErrorLevelWarn, "msg", "source")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(RunWithCatchPanic(func() {
			_ = defineError(ErrorTypeReply, num, ErrorLevelWarn, "msg", "source")
		})).Equal("Error redefined :\n>>> source\n>>> source\n")
	})

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := defineError(ErrorTypeReply, num, ErrorLevelWarn, "msg", "source")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetNumber(), v1.GetLevel(), v1.GetMessage()).
			Equal(ErrorTypeReply, num, ErrorLevelWarn, "msg")
		assert(errorDefineMap[num]).Equal("source")
	})
}

func TestDefineProtocolError(t *testing.T) {
	num := ErrorNumber(math.MaxUint32)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1, s1 := DefineProtocolError(num, ErrorLevelWarn, "msg"), GetFileLine(0)
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetNumber(), v1.GetLevel(), v1.GetMessage()).
			Equal(ErrorTypeProtocol, num, ErrorLevelWarn, "msg")
		assert(errorDefineMap[num]).Equal(s1)
	})
}

func TestDefineTransportError(t *testing.T) {
	num := ErrorNumber(math.MaxUint32)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1, s1 := DefineTransportError(num, ErrorLevelWarn, "msg"), GetFileLine(0)
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetNumber(), v1.GetLevel(), v1.GetMessage()).
			Equal(ErrorTypeTransport, num, ErrorLevelWarn, "msg")
		assert(errorDefineMap[num]).Equal(s1)
	})
}

func TestDefineReplyError(t *testing.T) {
	num := ErrorNumber(math.MaxUint32)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1, s1 := DefineReplyError(num, ErrorLevelWarn, "msg"), GetFileLine(0)
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetNumber(), v1.GetLevel(), v1.GetMessage()).
			Equal(ErrorTypeReply, num, ErrorLevelWarn, "msg")
		assert(errorDefineMap[num]).Equal(s1)
	})
}

func TestDefineRuntimeError(t *testing.T) {
	num := ErrorNumber(math.MaxUint32)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1, s1 := DefineRuntimeError(num, ErrorLevelWarn, "msg"), GetFileLine(0)
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetNumber(), v1.GetLevel(), v1.GetMessage()).
			Equal(ErrorTypeRuntime, num, ErrorLevelWarn, "msg")
		assert(errorDefineMap[num]).Equal(s1)
	})
}

func TestDefineKernelError(t *testing.T) {
	num := ErrorNumber(math.MaxUint32)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1, s1 := DefineKernelError(num, ErrorLevelWarn, "msg"), GetFileLine(0)
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetNumber(), v1.GetLevel(), v1.GetMessage()).
			Equal(ErrorTypeKernel, num, ErrorLevelWarn, "msg")
		assert(errorDefineMap[num]).Equal(s1)
	})
}

func TestDefineSecurityError(t *testing.T) {
	num := ErrorNumber(math.MaxUint32)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1, s1 := DefineSecurityError(num, ErrorLevelWarn, "msg"), GetFileLine(0)
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetNumber(), v1.GetLevel(), v1.GetMessage()).
			Equal(ErrorTypeSecurity, num, ErrorLevelWarn, "msg")
		assert(errorDefineMap[num]).Equal(s1)
	})
}

func TestDefineCustomError(t *testing.T) {
	num := ErrorNumber(math.MaxUint32)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1, s1 := DefineCustomError(num, ErrorLevelWarn, "msg"), GetFileLine(0)
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetNumber(), v1.GetLevel(), v1.GetMessage()).
			Equal(ErrorTypeCustom, num, ErrorLevelWarn, "msg")
		assert(errorDefineMap[num]).Equal(s1)
	})
}

func TestError_GetCode(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert((&Error{code: 123}).GetCode()).Equal(uint64(123))
	})
}

func TestError_GetType(t *testing.T) {
	num := ErrorNumber(math.MaxUint32)

	t.Run("test ErrorTypeProtocol", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineProtocolError(num, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType()).Equal(ErrorTypeProtocol)
	})

	t.Run("test ErrorTypeTransport", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineTransportError(num, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType()).Equal(ErrorTypeTransport)
	})

	t.Run("test ErrorTypeReply", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineReplyError(num, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType()).Equal(ErrorTypeReply)
	})

	t.Run("test ErrorTypeRuntime", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineRuntimeError(num, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType()).Equal(ErrorTypeRuntime)
	})

	t.Run("test ErrorTypeKernel", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineKernelError(num, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType()).Equal(ErrorTypeKernel)
	})

	t.Run("test ErrorTypeSecurity", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(num, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType()).Equal(ErrorTypeSecurity)
	})
}

func TestError_GetLevel(t *testing.T) {
	num := ErrorNumber(math.MaxUint32)

	t.Run("test ErrorLevelWarn", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(num, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetLevel()).Equal(ErrorLevelWarn)
	})

	t.Run("test ErrorLevelError", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(num, ErrorLevelError, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetLevel()).Equal(ErrorLevelError)
	})

	t.Run("test ErrorLevelFatal", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(num, ErrorLevelFatal, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetLevel()).Equal(ErrorLevelFatal)
	})
}

func TestError_GetNumber(t *testing.T) {
	num := ErrorNumber(math.MaxUint32)

	t.Run("test with uint32 min", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(num, ErrorLevelFatal, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetNumber()).Equal(num)
	})

	t.Run("test with uint32 max", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(num, ErrorLevelFatal, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetNumber()).Equal(num)
	})
}

func TestError_GetMessage(t *testing.T) {
	num := ErrorNumber(math.MaxUint32)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(num, ErrorLevelFatal, "msg1")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetMessage()).Equal("msg1")
	})
}

func TestError_AddDebug(t *testing.T) {
	num := ErrorNumber(math.MaxUint32)

	t.Run("test from origin error", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(num, ErrorLevelWarn, "")
		v2 := DefineSecurityError(num-1, ErrorLevelFatal, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			delete(errorDefineMap, num-1)
			errorDefineMutex.Unlock()
		}()
		v3 := v1.AddDebug("dbg")
		v4 := v2.AddDebug("dbg")
		assert(fmt.Sprintf("%p", v3) == fmt.Sprintf("%p", v1)).IsFalse()
		assert(fmt.Sprintf("%p", v4) == fmt.Sprintf("%p", v2)).IsFalse()
		assert(v3.GetMessage()).Equal("dbg")
		assert(v4.GetMessage()).Equal("msg\ndbg")
	})

	t.Run("test from derived error", func(t *testing.T) {
		assert := NewAssert(t)
		v := DefineSecurityError(num, ErrorLevelWarn, "")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()

		v1 := v.AddDebug("")
		v2 := v.AddDebug("msg")
		v3 := v1.AddDebug("dbg")
		v4 := v2.AddDebug("dbg")
		assert(fmt.Sprintf("%p", v3) == fmt.Sprintf("%p", v1)).IsTrue()
		assert(fmt.Sprintf("%p", v4) == fmt.Sprintf("%p", v2)).IsTrue()
		assert(v3.GetMessage()).Equal("dbg")
		assert(v4.GetMessage()).Equal("msg\ndbg")
	})
}

func TestError_getErrorTypeString(t *testing.T) {
	num := ErrorNumber(math.MaxUint32)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineProtocolError(num, ErrorLevelFatal, "msg")
		v2 := DefineTransportError(num-1, ErrorLevelFatal, "msg")
		v3 := DefineReplyError(num-2, ErrorLevelFatal, "msg")
		v4 := DefineRuntimeError(num-3, ErrorLevelFatal, "msg")
		v5 := DefineKernelError(num-4, ErrorLevelFatal, "msg")
		v6 := DefineSecurityError(num-5, ErrorLevelFatal, "msg")
		v7 := &Error{}
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			delete(errorDefineMap, num-1)
			delete(errorDefineMap, num-2)
			delete(errorDefineMap, num-3)
			delete(errorDefineMap, num-4)
			delete(errorDefineMap, num-5)
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
	num := ErrorNumber(math.MaxUint32)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineProtocolError(num, ErrorLevelWarn, "msg")
		v2 := DefineProtocolError(num-1, ErrorLevelError, "msg")
		v3 := DefineProtocolError(num-2, ErrorLevelFatal, "msg")
		v4 := DefineRuntimeError(num-3, 0, "msg")
		v5 := DefineRuntimeError(num-4, 255, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			delete(errorDefineMap, num-1)
			delete(errorDefineMap, num-2)
			delete(errorDefineMap, num-3)
			delete(errorDefineMap, num-4)
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
	num := ErrorNumber(math.MaxUint32)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineProtocolError(num, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		v2 := v1.AddDebug("dbg")
		assert(v2.Error()).Equal("ProtocolWarn: msg\ndbg")
	})
}
