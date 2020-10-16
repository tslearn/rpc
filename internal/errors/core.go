package errors

import (
	"github.com/rpccloud/rpc/internal/base"
)

const generalErrorSeg = 0

var (
	// ErrStream ...
	ErrStream = base.DefineSecurityError(
		(generalErrorSeg<<16)|1,
		base.ErrorLevelWarn,
		"stream is broken",
	)

	// ErrUnsupportedValue ...
	ErrUnsupportedValue = base.DefineDevelopError(
		(generalErrorSeg<<16)|2,
		base.ErrorLevelError,
		"",
	)

	// ErrServiceName ... *
	ErrServiceName = base.DefineDevelopError(
		(generalErrorSeg<<16)|3,
		base.ErrorLevelFatal,
		"",
	)

	// ErrServiceIsNil ... *
	ErrServiceIsNil = base.DefineDevelopError(
		(generalErrorSeg<<16)|6,
		base.ErrorLevelFatal,
		"service is nil",
	)

	// ErrServiceOverflow ... *
	ErrServiceOverflow = base.DefineDevelopError(
		(generalErrorSeg<<16)|7,
		base.ErrorLevelFatal,
		"",
	)

	// ErrReplyName ... *
	ErrReplyName = base.DefineDevelopError(
		(generalErrorSeg<<16)|4,
		base.ErrorLevelFatal,
		"",
	)

	// ErrReplyHandler ... *
	ErrReplyHandler = base.DefineDevelopError(
		(generalErrorSeg<<16)|5,
		base.ErrorLevelFatal,
		"",
	)

	// ErrReplyCloseTimeout ... *
	ErrReplyCloseTimeout = base.DefineDevelopError(
		(generalErrorSeg<<16)|8,
		base.ErrorLevelError,
		"",
	)

	// ErrReplyCustom ...
	ErrReplyCustom = base.DefineReplyError(
		(generalErrorSeg<<16)|10,
		base.ErrorLevelError,
		"",
	)

	// ErrReplyPanic ...
	ErrReplyPanic = base.DefineDevelopError(
		(generalErrorSeg<<16)|9,
		base.ErrorLevelFatal,
		"",
	)

	// ErrRPCTargetNotExist ...
	ErrRPCTargetNotExist = base.DefineSecurityError(
		(generalErrorSeg<<16)|11,
		base.ErrorLevelError,
		"",
	)

	// ErrRPCCallOverflow ...
	ErrRPCCallOverflow = base.DefineSecurityError(
		(generalErrorSeg<<16)|12,
		base.ErrorLevelError,
		"",
	)

	// ErrRPCArgumentsNotMatch ...
	ErrRPCArgumentsNotMatch = base.DefineSecurityError(
		(generalErrorSeg<<16)|13,
		base.ErrorLevelError,
		"",
	)

	// ErrRuntimeIllegalInCurrentGoroutine ...
	ErrRuntimeIllegalInCurrentGoroutine = base.DefineDevelopError(
		(generalErrorSeg<<16)|14,
		base.ErrorLevelFatal,
		"Runtime is illegal in current goroutine",
	)

	// ErrRuntimeErrorArgumentIsNil ...
	ErrRuntimeErrorArgumentIsNil = base.DefineDevelopError(
		(generalErrorSeg<<16)|15,
		base.ErrorLevelError,
		"Runtime.Error argument is nil",
	)

	// ErrRuntimeOKHasBeenCalled ...
	ErrRuntimeOKHasBeenCalled = base.DefineDevelopError(
		(generalErrorSeg<<16)|16,
		base.ErrorLevelError,
		"Runtime.OK has been called before",
	)

	// ErrRuntimeErrorHasBeenCalled ...
	ErrRuntimeErrorHasBeenCalled = base.DefineDevelopError(
		(generalErrorSeg<<16)|17,
		base.ErrorLevelError,
		"Runtime.Error has been called before",
	)

	// ErrRuntimeExternalReturn ...
	ErrRuntimeExternalReturn = base.DefineDevelopError(
		(generalErrorSeg<<16)|18,
		base.ErrorLevelError,
		"reply must be return through Runtime.OK or Runtime.Error",
	)

	// ErrNumOfThreadsIsWrong ... *
	ErrNumOfThreadsIsWrong = base.DefineConfigError(
		(generalErrorSeg<<16)|19,
		base.ErrorLevelFatal,
		"numOfThreads is wrong",
	)

	// ErrMaxNodeDepthIsWrong ... *
	ErrMaxNodeDepthIsWrong = base.DefineConfigError(
		(generalErrorSeg<<16)|20,
		base.ErrorLevelFatal,
		"maxNodeDepth is wrong",
	)

	// ErrCacheMkdirAll ... *
	ErrCacheMkdirAll = base.DefineConfigError(
		(generalErrorSeg<<16)|21,
		base.ErrorLevelFatal,
		"create directory error",
	)

	// ErrCacheWriteFile ... *
	ErrCacheWriteFile = base.DefineConfigError(
		(generalErrorSeg<<16)|22,
		base.ErrorLevelFatal,
		"write to file error",
	)
)

const kernelErrorSeg = 1000

var (
	// ErrFnCacheIllegalKindString ... *
	ErrFnCacheIllegalKindString = base.DefineKernelError(
		(kernelErrorSeg<<16)|1,
		base.ErrorLevelFatal,
		"",
	)

	// ErrFnCacheDuplicateKindString ... *
	ErrFnCacheDuplicateKindString = base.DefineKernelError(
		(kernelErrorSeg<<16)|2,
		base.ErrorLevelFatal,
		"",
	)

	// ErrProcessorOnReturnStreamIsNil ... *
	ErrProcessorOnReturnStreamIsNil = base.DefineKernelError(
		(kernelErrorSeg<<16)|3,
		base.ErrorLevelFatal,
		"onReturnStream is nil",
	)

	// ErrProcessorMaxCallDepthIsWrong ... *
	ErrProcessorMaxCallDepthIsWrong = base.DefineKernelError(
		(kernelErrorSeg<<16)|4,
		base.ErrorLevelFatal,
		"maxCallDepth is wrong",
	)

	// ErrProcessorNodeMetaIsNil ... *
	ErrProcessorNodeMetaIsNil = base.DefineKernelError(
		(kernelErrorSeg<<16)|5,
		base.ErrorLevelFatal,
		"node meta is nil",
	)

	// ErrProcessorReplyMetaIsNil ... *
	ErrProcessorReplyMetaIsNil = base.DefineKernelError(
		(kernelErrorSeg<<16)|6,
		base.ErrorLevelFatal,
		"reply meta is nil",
	)

	// ErrGetServiceDataServiceNodeIsNil ...
	ErrGetServiceDataServiceNodeIsNil = base.DefineKernelError(
		(kernelErrorSeg<<16)|7,
		base.ErrorLevelFatal,
		"serviceNode is nil",
	)

	// ErrSetServiceDataServiceNodeIsNil ...
	ErrSetServiceDataServiceNodeIsNil = base.DefineKernelError(
		(kernelErrorSeg<<16)|8,
		base.ErrorLevelFatal,
		"serviceNode is nil",
	)

	// ErrThreadEvalFatal ...
	ErrThreadEvalFatal = base.DefineKernelError(
		(kernelErrorSeg<<16)|9,
		base.ErrorLevelFatal,
		"",
	)
)
