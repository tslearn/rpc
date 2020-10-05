package base

import (
	"fmt"
	"strconv"
	"sync"
)

type ErrorType uint8

const (
	ErrorTypeProtocol  = ErrorType(1)
	ErrorTypeTransport = ErrorType(2)
	ErrorTypeReply     = ErrorType(3)
	ErrorTypeDevelop   = ErrorType(4)
	ErrorTypeKernel    = ErrorType(5)
	ErrorTypeSecurity  = ErrorType(6)
	ErrorTypeCustom    = ErrorType(7)
)

type ErrorLevel uint8

const (
	ErrorLevelWarn  = ErrorLevel(1)
	ErrorLevelError = ErrorLevel(2)
	ErrorLevelFatal = ErrorLevel(3)
)

type ErrorNumber uint32

var (
	errorDefineMutex = &sync.Mutex{}
	errorDefineMap   = map[ErrorNumber]string{}
)

type Error struct {
	code    uint64
	message string
}

func NewError(code uint64, message string) *Error {
	return &Error{
		code:    code,
		message: message,
	}
}

func defineError(
	kind ErrorType,
	num ErrorNumber,
	level ErrorLevel,
	message string,
	source string,
) *Error {
	errorDefineMutex.Lock()
	defer errorDefineMutex.Unlock()

	code := (uint64(kind) << 42) | (uint64(level) << 34) | (uint64(num) << 2)

	if value, ok := errorDefineMap[num]; ok {
		panic(fmt.Sprintf("Error redefined :\n>>> %s\n>>> %s\n", value, source))
	} else {
		errorDefineMap[num] = source
	}

	return &Error{
		code:    code,
		message: message,
	}
}

// DefineProtocolError ...
func DefineProtocolError(num ErrorNumber, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeProtocol, num, level, msg, GetFileLine(1))
}

// DefineTransportError ...
func DefineTransportError(num ErrorNumber, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeTransport, num, level, msg, GetFileLine(1))
}

// DefineReplyError ...
func DefineReplyError(num ErrorNumber, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeReply, num, level, msg, GetFileLine(1))
}

// DefineDevelopError ...
func DefineDevelopError(num ErrorNumber, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeDevelop, num, level, msg, GetFileLine(1))
}

// DefineKernelError ...
func DefineKernelError(num ErrorNumber, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeKernel, num, level, msg, GetFileLine(1))
}

// DefineSecurityError ...
func DefineSecurityError(num ErrorNumber, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeSecurity, num, level, msg, GetFileLine(1))
}

func (p *Error) GetCode() uint64 {
	return p.code
}

func (p *Error) GetType() ErrorType {
	return ErrorType(p.code >> 42)
}

func (p *Error) GetLevel() ErrorLevel {
	return ErrorLevel((p.code >> 34) & 0xFF)
}

func (p *Error) GetNumber() ErrorNumber {
	return ErrorNumber((p.code >> 2) & 0xFFFFFFFF)
}

func (p *Error) GetMessage() string {
	return p.message
}

func (p *Error) AddDebug(debug string) *Error {
	if p.code%2 == 0 {
		ret := &Error{code: p.code + 1}
		if p.message == "" {
			ret.message = ConcatString(debug)
		} else {
			ret.message = ConcatString(p.message, "\n", debug)
		}
		return ret
	}

	if p.message == "" {
		p.message = ConcatString(debug)
	} else {
		p.message = ConcatString(p.message, "\n", debug)
	}
	return p
}

func (p *Error) getErrorTypeString() string {
	switch p.GetType() {
	case ErrorTypeProtocol:
		return "Protocol"
	case ErrorTypeTransport:
		return "Transport"
	case ErrorTypeReply:
		return "Reply"
	case ErrorTypeDevelop:
		return "Develop"
	case ErrorTypeKernel:
		return "Kernel"
	case ErrorTypeSecurity:
		return "Security"
	case ErrorTypeCustom:
		return "Custom"
	default:
		return ""
	}
}

func (p *Error) getErrorLevelString() string {
	switch p.GetLevel() {
	case ErrorLevelWarn:
		return "Warn"
	case ErrorLevelError:
		return "Error"
	case ErrorLevelFatal:
		return "Fatal"
	default:
		return ""
	}
}

func (p *Error) Error() string {
	return ConcatString(
		p.getErrorTypeString(),
		p.getErrorLevelString(),
		"[",
		strconv.FormatUint(uint64(p.GetNumber()), 10),
		"]: ",
		p.message,
	)
}
