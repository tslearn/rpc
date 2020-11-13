package errors

import (
	"github.com/rpccloud/rpc/internal/base"
)

// Kind 0 (Kernel error)
var (
	// ErrFnCacheIllegalKindString ... *
	ErrFnCacheIllegalKindString = base.DefineKernelError(
		1,
		base.ErrorLevelFatal,
		"",
	)

	// ErrFnCacheDuplicateKindString ... *
	ErrFnCacheDuplicateKindString = base.DefineKernelError(
		2,
		base.ErrorLevelFatal,
		"",
	)

	// ErrProcessorOnReturnStreamIsNil ... *
	ErrProcessorOnReturnStreamIsNil = base.DefineKernelError(
		3,
		base.ErrorLevelFatal,
		"onReturnStream is nil",
	)

	// ErrProcessorMaxCallDepthIsWrong ... *
	ErrProcessorMaxCallDepthIsWrong = base.DefineKernelError(
		4,
		base.ErrorLevelFatal,
		"maxCallDepth is wrong",
	)

	// ErrProcessorNodeMetaIsNil ... *
	ErrProcessorNodeMetaIsNil = base.DefineKernelError(
		5,
		base.ErrorLevelFatal,
		"node meta is nil",
	)

	// ErrProcessorActionMetaIsNil ... *
	ErrProcessorActionMetaIsNil = base.DefineKernelError(
		6,
		base.ErrorLevelFatal,
		"action meta is nil",
	)

	// ErrThreadEvalFatal ...
	ErrThreadEvalFatal = base.DefineKernelError(
		7,
		base.ErrorLevelFatal,
		"",
	)

	// ErrProcessorIsNotRunning ... *
	ErrProcessorIsNotRunning = base.DefineKernelError(
		8,
		base.ErrorLevelFatal,
		"processor is not running",
	)
)

// Kind 1 (Caused by remote abuse or error)
var (
	// ErrStream ...
	ErrStream = base.DefineSecurityError(
		10001,
		base.ErrorLevelWarn,
		"stream error",
	)

	// ErrUnsupportedValue ...
	ErrUnsupportedValue = base.DefineDevelopError(
		10002,
		base.ErrorLevelError,
		"",
	)

	// ErrTargetNotExist ...
	ErrTargetNotExist = base.DefineSecurityError(
		10011,
		base.ErrorLevelError,
		"",
	)

	// ErrCallOverflow ...
	ErrCallOverflow = base.DefineSecurityError(
		10012,
		base.ErrorLevelError,
		"",
	)

	// ErrArgumentsNotMatch ...
	ErrArgumentsNotMatch = base.DefineSecurityError(
		10013,
		base.ErrorLevelError,
		"",
	)
)

// Kind 2 (Caused by incorrect use of api)
var (
	// ErrServiceName ... *
	ErrServiceName = base.DefineDevelopError(
		10003,
		base.ErrorLevelFatal,
		"",
	)

	// ErrServiceIsNil ... *
	ErrServiceIsNil = base.DefineDevelopError(
		10004,
		base.ErrorLevelFatal,
		"service is nil",
	)

	// ErrServiceOverflow ... *
	ErrServiceOverflow = base.DefineDevelopError(
		10005,
		base.ErrorLevelFatal,
		"",
	)

	// ErrActionName ... *
	ErrActionName = base.DefineDevelopError(
		10006,
		base.ErrorLevelFatal,
		"",
	)

	// ErrActionHandler ... *
	ErrActionHandler = base.DefineDevelopError(
		10007,
		base.ErrorLevelFatal,
		"",
	)

	// ErrNumOfThreadsIsWrong ... *
	ErrNumOfThreadsIsWrong = base.DefineConfigError(
		10019,
		base.ErrorLevelFatal,
		"numOfThreads is wrong",
	)

	// ErrMaxNodeDepthIsWrong ... *
	ErrMaxNodeDepthIsWrong = base.DefineConfigError(
		10020,
		base.ErrorLevelFatal,
		"maxNodeDepth is wrong",
	)

	// ErrCacheMkdirAll ... *
	ErrCacheMkdirAll = base.DefineConfigError(
		10021,
		base.ErrorLevelFatal,
		"create directory error",
	)

	// ErrCacheWriteFile ... *
	ErrCacheWriteFile = base.DefineConfigError(
		10022,
		base.ErrorLevelFatal,
		"write to file error",
	)
)

// Kind 3 (Caused by Actions)
var (
	// ErrActionCloseTimeout ... *
	ErrActionCloseTimeout = base.DefineDevelopError(
		10008,
		base.ErrorLevelError,
		"",
	)

	// ErrActionCustom ...
	ErrActionCustom = base.DefineActionError(
		10009,
		base.ErrorLevelError,
		"",
	)

	// ErrActionPanic ...
	ErrActionPanic = base.DefineDevelopError(
		10010,
		base.ErrorLevelFatal,
		"",
	)

	// ErrRuntimeIllegalInCurrentGoroutine ...
	ErrRuntimeIllegalInCurrentGoroutine = base.DefineDevelopError(
		10014,
		base.ErrorLevelFatal,
		"Runtime is illegal in current goroutine",
	)

	// ErrRuntimeReplyHasBeenCalled ...
	ErrRuntimeReplyHasBeenCalled = base.DefineDevelopError(
		10015,
		base.ErrorLevelError,
		"Runtime.Reply has been called before",
	)

	// ErrRuntimeExternalReturn ...
	ErrRuntimeExternalReturn = base.DefineDevelopError(
		10016,
		base.ErrorLevelError,
		"action must be return through Runtime.OK or Runtime.Error",
	)

	// ErrRTArrayIndexOverflow ...
	ErrRTArrayIndexOverflow = base.DefineDevelopError(
		10017,
		base.ErrorLevelFatal,
		"",
	)

	// ErrRTMapNameNotFound ...
	ErrRTMapNameNotFound = base.DefineDevelopError(
		10018,
		base.ErrorLevelFatal,
		"",
	)
)
