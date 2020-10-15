package errors

import (
	"github.com/rpccloud/rpc/internal/base"
)

const generalErrorSeg = 0

var (
	ErrStream = base.DefineSecurityError(
		(generalErrorSeg<<16)|1,
		base.ErrorLevelWarn,
		"stream is broken",
	)

	ErrUnsupportedValue = base.DefineDevelopError(
		(generalErrorSeg<<16)|2,
		base.ErrorLevelError,
		"",
	)

	ErrServiceName = base.DefineDevelopError(
		(generalErrorSeg<<16)|3,
		base.ErrorLevelFatal,
		"",
	)

	ErrReplyName = base.DefineDevelopError(
		(generalErrorSeg<<16)|4,
		base.ErrorLevelFatal,
		"",
	)

	ErrReplyHandler = base.DefineDevelopError(
		(generalErrorSeg<<16)|5,
		base.ErrorLevelFatal,
		"",
	)

	ErrServiceIsNil = base.DefineDevelopError(
		(generalErrorSeg<<16)|6,
		base.ErrorLevelFatal,
		"service is nil",
	)

	ErrServiceOverflow = base.DefineDevelopError(
		(generalErrorSeg<<16)|7,
		base.ErrorLevelFatal,
		"",
	)

	ErrReplyCloseTimeout = base.DefineDevelopError(
		(generalErrorSeg<<16)|8,
		base.ErrorLevelError,
		"",
	)

	ErrReplyPanic = base.DefineDevelopError(
		(generalErrorSeg<<16)|9,
		base.ErrorLevelFatal,
		"",
	)

	ErrReplyReturn = base.DefineReplyError(
		(generalErrorSeg<<16)|10,
		base.ErrorLevelError,
		"",
	)

	ErrRPCTargetNotExist = base.DefineSecurityError(
		(generalErrorSeg<<16)|11,
		base.ErrorLevelError,
		"",
	)

	ErrRPCCallOverflow = base.DefineSecurityError(
		(generalErrorSeg<<16)|12,
		base.ErrorLevelError,
		"",
	)

	ErrRPCArgumentsNotMatch = base.DefineSecurityError(
		(generalErrorSeg<<16)|13,
		base.ErrorLevelError,
		"",
	)

	ErrRuntimeIllegalInCurrentGoroutine = base.DefineDevelopError(
		(generalErrorSeg<<16)|14,
		base.ErrorLevelFatal,
		"Runtime is illegal in current goroutine",
	)

	ErrRuntimeErrorArgumentIsNil = base.DefineDevelopError(
		(generalErrorSeg<<16)|15,
		base.ErrorLevelError,
		"Runtime.Error argument is nil",
	)

	ErrRuntimeOKHasBeenCalled = base.DefineDevelopError(
		(generalErrorSeg<<16)|16,
		base.ErrorLevelError,
		"Runtime.OK has been called before",
	)

	ErrRuntimeErrorHasBeenCalled = base.DefineDevelopError(
		(generalErrorSeg<<16)|17,
		base.ErrorLevelError,
		"Runtime.Error has been called before",
	)

	ErrRuntimeExternalReturn = base.DefineDevelopError(
		(generalErrorSeg<<16)|18,
		base.ErrorLevelError,
		"reply must be return through Runtime.OK or Runtime.Error",
	)

	ErrNumOfThreadsIsWrong = base.DefineConfigError(
		(generalErrorSeg<<16)|19,
		base.ErrorLevelFatal,
		"numOfThreads is wrong",
	)

	ErrMaxNodeDepthIsWrong = base.DefineConfigError(
		(generalErrorSeg<<16)|20,
		base.ErrorLevelFatal,
		"maxNodeDepth is wrong",
	)

	// ErrFnCacheMkdirAll ... *
	ErrCacheMkdirAll = base.DefineConfigError(
		(generalErrorSeg<<16)|21,
		base.ErrorLevelFatal,
		"create directory error",
	)

	// ErrFnCacheWriteFile ... *
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

	ErrProcessorMaxCallDepthIsWrong = base.DefineKernelError(
		(kernelErrorSeg<<16)|3,
		base.ErrorLevelFatal,
		"maxCallDepth is wrong",
	)

	ErrProcessorNodeMetaIsNil = base.DefineKernelError(
		(kernelErrorSeg<<16)|4,
		base.ErrorLevelFatal,
		"nodeMeta is nil",
	)

	ErrProcessorMetaIsNil = base.DefineKernelError(
		(kernelErrorSeg<<16)|5,
		base.ErrorLevelFatal,
		"meta is nil",
	)

	ErrGetServiceDataServiceNodeIsNil = base.DefineKernelError(
		(kernelErrorSeg<<16)|6,
		base.ErrorLevelFatal,
		"serviceNode is nil",
	)

	ErrSetServiceDataServiceNodeIsNil = base.DefineKernelError(
		(kernelErrorSeg<<16)|7,
		base.ErrorLevelFatal,
		"serviceNode is nil",
	)

	ErrThreadEvalFatal = base.DefineKernelError(
		(kernelErrorSeg<<16)|8,
		base.ErrorLevelFatal,
		"",
	)
)
