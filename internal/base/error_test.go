package base

import "testing"

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

//
//func TestNewError(t *testing.T) {
//	assert := NewAssert(t)
//
//	// Test(1)
//	o1 := NewError(ErrorKindRuntimePanic, "message", "fileLine")
//	assert(o1.GetKind()).Equal(ErrorKindRuntimePanic)
//	assert(o1.GetMessage()).Equal("message")
//	assert(o1.GetDebug()).Equal("fileLine")
//}
//
//func TestNewReplyError(t *testing.T) {
//	assert := NewAssert(t)
//
//	// Test(1)
//	assert(NewReplyError("message")).
//		Equal(NewError(ErrorKindReply, "message", ""))
//}
//
//func TestNewReplyPanic(t *testing.T) {
//	assert := NewAssert(t)
//
//	// Test(1)
//	assert(NewReplyPanic("message")).
//		Equal(NewError(ErrorKindReplyPanic, "message", ""))
//}
//
//func TestNewRuntimeError(t *testing.T) {
//	assert := NewAssert(t)
//
//	// Test(1)
//	assert(NewRuntimePanic("message")).
//		Equal(NewError(ErrorKindRuntimePanic, "message", ""))
//}
//
//func TestNewProtocolError(t *testing.T) {
//	assert := NewAssert(t)
//
//	// Test(1)
//	assert(NewProtocolError("message")).
//		Equal(NewError(ErrorKindProtocol, "message", ""))
//}
//
//func TestNewTransportError(t *testing.T) {
//	assert := NewAssert(t)
//
//	// Test(1)
//	assert(NewTransportError("message")).
//		Equal(NewError(ErrorKindTransport, "message", ""))
//}
//
//func TestNewKernelError(t *testing.T) {
//	assert := NewAssert(t)
//
//	// Test(1)
//	assert(NewKernelPanic("message")).
//		Equal(NewError(ErrorKindKernelPanic, "message", ""))
//}
//
//func TestNewSecurityLimitError(t *testing.T) {
//	assert := NewAssert(t)
//
//	// Test(1)
//	assert(NewSecurityLimitError("message")).
//		Equal(NewError(ErrorKindSecurityLimit, "message", ""))
//}
//
//func TestConvertToError(t *testing.T) {
//	assert := NewAssert(t)
//
//	// Test(1)
//	assert(ConvertToError(0)).IsNil()
//
//	// Test(2)
//	assert(ConvertToError(make(chan bool))).IsNil()
//
//	// Test(3)
//	assert(ConvertToError(nil)).IsNil()
//
//	// Test(4)
//	assert(ConvertToError(NewKernelPanic("test"))).IsNotNil()
//}
//
//func TestRpcError_GetKind(t *testing.T) {
//	assert := NewAssert(t)
//
//	// Test(1)
//	o1 := NewError(ErrorKindKernelPanic, "message", "fileLine")
//	assert(o1.GetKind()).Equal(ErrorKindKernelPanic)
//}
//
//func TestRpcError_GetMessage(t *testing.T) {
//	assert := NewAssert(t)
//
//	// Test(1)
//	o1 := NewError(ErrorKindKernelPanic, "message", "fileLine")
//	assert(o1.GetMessage()).Equal("message")
//}
//
//func TestRpcError_GetDebug(t *testing.T) {
//	assert := NewAssert(t)
//
//	// Test(1)
//	o1 := NewError(ErrorKindKernelPanic, "message", "fileLine")
//	assert(o1.GetDebug()).Equal("fileLine")
//}
//
//func TestRpcError_AddDebug(t *testing.T) {
//	assert := NewAssert(t)
//
//	// Test(1)
//	o1 := NewError(ErrorKindKernelPanic, "message", "")
//	assert(o1.AddDebug("fileLine").GetDebug()).Equal("fileLine")
//
//	// Test(2)
//	o2 := NewError(ErrorKindKernelPanic, "message", "fileLine")
//	assert(o2.AddDebug("fileLine").GetDebug()).Equal("fileLine\nfileLine")
//}
//
//func TestRpcError_Error(t *testing.T) {
//	assert := NewAssert(t)
//
//	// Test(1)
//	assert(NewError(ErrorKindKernelPanic, "", "").Error()).Equal("")
//
//	// Test(2)
//	assert(NewError(ErrorKindKernelPanic, "message", "").Error()).Equal("message")
//
//	// Test(3)
//	assert(NewError(ErrorKindKernelPanic, "", "fileLine").Error()).Equal("fileLine")
//
//	// Test(4)
//	assert(NewError(ErrorKindKernelPanic, "message", "fileLine").Error()).
//		Equal("message\nfileLine")
//}
