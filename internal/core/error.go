package core

import (
	"github.com/rpccloud/rpc/internal/base"
)

const fnCacheErrorSeg = 1001

var (
	ErrFnCacheMkdirAll = base.DefineRuntimeError(
		(fnCacheErrorSeg<<16)|1,
		base.ErrorLevelFatal,
		"",
	)

	ErrFnCacheWriteFile = base.DefineRuntimeError(
		(fnCacheErrorSeg<<16)|2,
		base.ErrorLevelFatal,
		"",
	)

	ErrFnCacheIllegalKindString = base.DefineKernelError(
		(fnCacheErrorSeg<<16)|3,
		base.ErrorLevelFatal,
		"",
	)

	ErrFnCacheDuplicateKindString = base.DefineKernelError(
		(fnCacheErrorSeg<<16)|4,
		base.ErrorLevelFatal,
		"",
	)
)

const processorErrorSeg = 1002

var (
	ErrProcessorNumOfThreadsIsWrong = base.DefineKernelError(
		(processorErrorSeg<<16)|1,
		base.ErrorLevelFatal,
		"numOfThreads is wrong",
	)

	ErrProcessorMaxNodeDepthIsWrong = base.DefineKernelError(
		(processorErrorSeg<<16)|2,
		base.ErrorLevelFatal,
		"maxNodeDepth is wrong",
	)

	ErrProcessorMaxCallDepthIsWrong = base.DefineKernelError(
		(processorErrorSeg<<16)|3,
		base.ErrorLevelFatal,
		"maxCallDepth is wrong",
	)

	ErrProcessorCloseTimeout = base.DefineRuntimeError(
		(processorErrorSeg<<16)|4,
		base.ErrorLevelError,
		"",
	)

	ErrProcessorNodeMetaIsNil = base.DefineKernelError(
		(processorErrorSeg<<16)|5,
		base.ErrorLevelFatal,
		"nodeMeta is nil",
	)

	ErrProcessorServiceNameIsIllegal = base.DefineKernelError(
		(processorErrorSeg<<16)|6,
		base.ErrorLevelFatal,
		"",
	)

	ErrProcessorNodeMetaServiceIsNil = base.DefineRuntimeError(
		(processorErrorSeg<<16)|7,
		base.ErrorLevelFatal,
		"",
	)

	ErrProcessorServicePathOverflow = base.DefineRuntimeError(
		(processorErrorSeg<<16)|8,
		base.ErrorLevelFatal,
		"",
	)

	ErrProcessorDuplicatedServiceName = base.DefineRuntimeError(
		(processorErrorSeg<<16)|9,
		base.ErrorLevelFatal,
		"",
	)

	ErrProcessorOnMount = base.DefineRuntimeError(
		(processorErrorSeg<<16)|10,
		base.ErrorLevelFatal,
		"",
	)

	ErrProcessorMetaIsNil = base.DefineKernelError(
		(processorErrorSeg<<16)|11,
		base.ErrorLevelFatal,
		"meta is nil",
	)

	ErrProcessReplyNameIsIllegal = base.DefineRuntimeError(
		(processorErrorSeg<<16)|12,
		base.ErrorLevelFatal,
		"",
	)

	ErrProcessHandlerIsNil = base.DefineRuntimeError(
		(processorErrorSeg<<16)|13,
		base.ErrorLevelFatal,
		"handler is nil",
	)

	ErrProcessIllegalHandler = base.DefineRuntimeError(
		(processorErrorSeg<<16)|14,
		base.ErrorLevelFatal,
		"",
	)

	ErrProcessDuplicatedReplyName = base.DefineRuntimeError(
		(processorErrorSeg<<16)|15,
		base.ErrorLevelFatal,
		"",
	)
)

const replyRuntimeSeg = 1003

var (
	ErrRuntimeIllegalInCurrentGoroutine = base.DefineReplyError(
		(replyRuntimeSeg<<16)|1,
		base.ErrorLevelFatal,
		"Runtime is illegal in current goroutine",
	)

	ErrRuntimeErrorArgumentIsNil = base.DefineReplyError(
		(replyRuntimeSeg<<16)|2,
		base.ErrorLevelError,
		"Runtime.Error argument is nil",
	)

	ErrRuntimeArgumentNotSupported = base.DefineReplyError(
		(replyRuntimeSeg<<16)|3,
		base.ErrorLevelError,
		"",
	)

	ErrRuntimeOKHasBeenCalled = base.DefineReplyError(
		(replyRuntimeSeg<<16)|4,
		base.ErrorLevelError,
		"Runtime.OK has been called before",
	)

	ErrRuntimeErrorHasBeenCalled = base.DefineReplyError(
		(replyRuntimeSeg<<16)|5,
		base.ErrorLevelError,
		"Runtime.Error has been called before",
	)

	ErrRuntimeExternelReturn = base.DefineReplyError(
		(replyRuntimeSeg<<16)|6,
		base.ErrorLevelError,
		"reply must return through Runtime.OK or Runtime.Error",
	)
)

const threadErrorSeg = 1004

var (
	ErrThreadEvalFatal = base.DefineKernelError(
		(threadErrorSeg<<16)|1,
		base.ErrorLevelFatal,
		"",
	)

	ErrThreadReplyFatal = base.DefineReplyError(
		(threadErrorSeg<<16)|2,
		base.ErrorLevelFatal,
		"",
	)

	ErrThreadTargetNotExist = base.DefineProtocolError(
		(threadErrorSeg<<16)|3,
		base.ErrorLevelError,
		"",
	)

	ErrThreadCallDepthOverflow = base.DefineProtocolError(
		(threadErrorSeg<<16)|4,
		base.ErrorLevelError,
		"",
	)

	ErrThreadArgumentsNotMatch = base.DefineProtocolError(
		(threadErrorSeg<<16)|5,
		base.ErrorLevelError,
		"",
	)
)

const customErrorSeg = 1005

var (
	ErrGeneralCustomError = base.DefineCustomError(
		(customErrorSeg<<16)|1,
		base.ErrorLevelError,
		"",
	)
)
