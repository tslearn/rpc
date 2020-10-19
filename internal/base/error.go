package base

import (
	"fmt"
	"strconv"
	"sync"
)

// ErrorType ...
type ErrorType uint8

const (
	// ErrorTypeConfig ...
	ErrorTypeConfig = ErrorType(1)
	// ErrorTypeNet ...
	ErrorTypeNet = ErrorType(2)
	// ErrorTypeAction ...
	ErrorTypeAction = ErrorType(3)
	// ErrorTypeDevelop ...
	ErrorTypeDevelop = ErrorType(4)
	// ErrorTypeKernel ...
	ErrorTypeKernel = ErrorType(5)
	// ErrorTypeSecurity ..
	ErrorTypeSecurity = ErrorType(6)
)

// ErrorLevel ...
type ErrorLevel uint8

const (
	// ErrorLevelWarn ...
	ErrorLevelWarn = ErrorLevel(1)
	// ErrorLevelError ...
	ErrorLevelError = ErrorLevel(2)
	// ErrorLevelFatal ...
	ErrorLevelFatal = ErrorLevel(3)
)

// ErrorNumber ...
type ErrorNumber uint32

var (
	errorDefineMutex = &sync.Mutex{}
	errorDefineMap   = map[ErrorNumber]string{}
)

// Error ...
type Error struct {
	code    uint64
	message string
}

// NewError ...
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

// DefineConfigError ...
func DefineConfigError(num ErrorNumber, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeConfig, num, level, msg, GetFileLine(1))
}

// DefineNetError ...
func DefineNetError(num ErrorNumber, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeNet, num, level, msg, GetFileLine(1))
}

// DefineActionError ...
func DefineActionError(num ErrorNumber, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeAction, num, level, msg, GetFileLine(1))
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

// GetCode ...
func (p *Error) GetCode() uint64 {
	return p.code
}

// GetType ...
func (p *Error) GetType() ErrorType {
	return ErrorType(p.code >> 42)
}

// GetLevel ...
func (p *Error) GetLevel() ErrorLevel {
	return ErrorLevel((p.code >> 34) & 0xFF)
}

// GetNumber ...
func (p *Error) GetNumber() ErrorNumber {
	return ErrorNumber((p.code >> 2) & 0xFFFFFFFF)
}

// GetMessage ...
func (p *Error) GetMessage() string {
	return p.message
}

// AddDebug ...
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
	case ErrorTypeConfig:
		return "Config"
	case ErrorTypeNet:
		return "Transport"
	case ErrorTypeAction:
		return "Reply"
	case ErrorTypeDevelop:
		return "Develop"
	case ErrorTypeKernel:
		return "Kernel"
	case ErrorTypeSecurity:
		return "Security"
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
