package errors

import (
	"github.com/rpccloud/rpc/internal/base"
)

const generalErrorSeg = 1000

var (
	ErrStream = base.DefineSecurityError(
		(generalErrorSeg<<16)|1,
		base.ErrorLevelWarn,
		"stream is broken",
	)

	ErrServiceName = base.DefineDevelopError(
		(generalErrorSeg<<16)|2,
		base.ErrorLevelFatal,
		"",
	)

	ErrReplyName = base.DefineDevelopError(
		(generalErrorSeg<<16)|3,
		base.ErrorLevelFatal,
		"",
	)

	ErrReplyHandler = base.DefineDevelopError(
		(generalErrorSeg<<16)|4,
		base.ErrorLevelFatal,
		"",
	)

	ErrServiceIsNil = base.DefineDevelopError(
		(generalErrorSeg<<16)|5,
		base.ErrorLevelFatal,
		"service is nil",
	)

	ErrServiceOverflow = base.DefineDevelopError(
		(generalErrorSeg<<16)|6,
		base.ErrorLevelFatal,
		"",
	)

	ErrReplyCloseTimeout = base.DefineDevelopError(
		(generalErrorSeg<<16)|7,
		base.ErrorLevelError,
		"",
	)

	ErrReplyPanic = base.DefineDevelopError(
		(generalErrorSeg<<16)|8,
		base.ErrorLevelFatal,
		"",
	)

	ErrReplyReturn = base.DefineReplyError(
		(generalErrorSeg<<16)|9,
		base.ErrorLevelError,
		"",
	)
)

const fnCacheErrorSeg = 1001

var (
	// ErrFnCacheMkdirAll ... *
	ErrFnCacheMkdirAll = base.DefineConfigError(
		(fnCacheErrorSeg<<16)|1,
		base.ErrorLevelFatal,
		"create directory error",
	)

	// ErrFnCacheWriteFile ... *
	ErrFnCacheWriteFile = base.DefineConfigError(
		(fnCacheErrorSeg<<16)|2,
		base.ErrorLevelFatal,
		"write to file error",
	)

	// ErrFnCacheIllegalKindString ... *
	ErrFnCacheIllegalKindString = base.DefineKernelError(
		(fnCacheErrorSeg<<16)|3,
		base.ErrorLevelFatal,
		"",
	)

	// ErrFnCacheDuplicateKindString ... *
	ErrFnCacheDuplicateKindString = base.DefineKernelError(
		(fnCacheErrorSeg<<16)|4,
		base.ErrorLevelFatal,
		"",
	)
)

const processorErrorSeg = 1002

var (
	ErrProcessorNumOfThreadsIsWrong = base.DefineConfigError(
		(processorErrorSeg<<16)|1,
		base.ErrorLevelFatal,
		"numOfThreads is wrong",
	)

	ErrProcessorMaxNodeDepthIsWrong = base.DefineConfigError(
		(processorErrorSeg<<16)|2,
		base.ErrorLevelFatal,
		"maxNodeDepth is wrong",
	)

	ErrProcessorMaxCallDepthIsWrong = base.DefineKernelError(
		(processorErrorSeg<<16)|3,
		base.ErrorLevelFatal,
		"maxCallDepth is wrong",
	)

	ErrProcessorNodeMetaIsNil = base.DefineKernelError(
		(processorErrorSeg<<16)|5,
		base.ErrorLevelFatal,
		"nodeMeta is nil",
	)

	ErrProcessorMetaIsNil = base.DefineKernelError(
		(processorErrorSeg<<16)|11,
		base.ErrorLevelFatal,
		"meta is nil",
	)
)

const replyRuntimeSeg = 1003

var (
	ErrRuntimeIllegalInCurrentGoroutine = base.DefineDevelopError(
		(replyRuntimeSeg<<16)|1,
		base.ErrorLevelFatal,
		"Runtime is illegal in current goroutine",
	)

	ErrRuntimeErrorArgumentIsNil = base.DefineDevelopError(
		(replyRuntimeSeg<<16)|2,
		base.ErrorLevelError,
		"Runtime.Error argument is nil",
	)

	ErrRuntimeArgumentNotSupported = base.DefineDevelopError(
		(replyRuntimeSeg<<16)|3,
		base.ErrorLevelError,
		"",
	)

	ErrRuntimeOKHasBeenCalled = base.DefineDevelopError(
		(replyRuntimeSeg<<16)|4,
		base.ErrorLevelError,
		"Runtime.OK has been called before",
	)

	ErrRuntimeErrorHasBeenCalled = base.DefineDevelopError(
		(replyRuntimeSeg<<16)|5,
		base.ErrorLevelError,
		"Runtime.Error has been called before",
	)

	ErrRuntimeExternalReturn = base.DefineDevelopError(
		(replyRuntimeSeg<<16)|6,
		base.ErrorLevelError,
		"reply must be return through Runtime.OK or Runtime.Error",
	)

	ErrRuntimeServiceNodeIsNil = base.DefineKernelError(
		(replyRuntimeSeg<<16)|7,
		base.ErrorLevelFatal,
		"serviceNode is nil",
	)
)

const threadErrorSeg = 1004

var (
	ErrThreadEvalFatal = base.DefineKernelError(
		(threadErrorSeg<<16)|1,
		base.ErrorLevelFatal,
		"",
	)

	ErrThreadTargetNotExist = base.DefineSecurityError(
		(threadErrorSeg<<16)|3,
		base.ErrorLevelError,
		"",
	)

	ErrThreadCallDepthOverflow = base.DefineSecurityError(
		(threadErrorSeg<<16)|4,
		base.ErrorLevelError,
		"",
	)

	ErrThreadArgumentsNotMatch = base.DefineSecurityError(
		(threadErrorSeg<<16)|5,
		base.ErrorLevelError,
		"",
	)
)
