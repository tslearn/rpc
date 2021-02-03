// Package errors ...
package base

const generalErrorSeg = 1000

var (
	// ErrStream ...
	ErrStream = DefineSecurityError(
		(generalErrorSeg<<16)|1,
		ErrorLevelWarn,
		"stream error",
	)

	// ErrUnsupportedValue ...
	ErrUnsupportedValue = DefineDevelopError(
		(generalErrorSeg<<16)|2,
		ErrorLevelError,
		"",
	)

	// ErrServiceName ... *
	ErrServiceName = DefineDevelopError(
		(generalErrorSeg<<16)|3,
		ErrorLevelFatal,
		"",
	)

	// ErrServiceIsNil ... *
	ErrServiceIsNil = DefineDevelopError(
		(generalErrorSeg<<16)|4,
		ErrorLevelFatal,
		"service is nil",
	)

	// ErrServiceOverflow ... *
	ErrServiceOverflow = DefineDevelopError(
		(generalErrorSeg<<16)|5,
		ErrorLevelFatal,
		"",
	)

	// ErrActionName ... *
	ErrActionName = DefineDevelopError(
		(generalErrorSeg<<16)|6,
		ErrorLevelFatal,
		"",
	)

	// ErrActionHandler ... *
	ErrActionHandler = DefineDevelopError(
		(generalErrorSeg<<16)|7,
		ErrorLevelFatal,
		"",
	)

	// ErrActionCloseTimeout ... *
	ErrActionCloseTimeout = DefineDevelopError(
		(generalErrorSeg<<16)|8,
		ErrorLevelError,
		"",
	)

	// ErrAction ...
	ErrAction = DefineActionError(
		(generalErrorSeg<<16)|9,
		ErrorLevelError,
		"",
	)

	// ErrActionPanic ...
	ErrActionPanic = DefineDevelopError(
		(generalErrorSeg<<16)|10,
		ErrorLevelFatal,
		"",
	)

	// ErrTargetNotExist ...
	ErrTargetNotExist = DefineSecurityError(
		(generalErrorSeg<<16)|11,
		ErrorLevelError,
		"",
	)

	// ErrCallOverflow ...
	ErrCallOverflow = DefineSecurityError(
		(generalErrorSeg<<16)|12,
		ErrorLevelError,
		"",
	)

	// ErrArgumentsNotMatch ...
	ErrArgumentsNotMatch = DefineSecurityError(
		(generalErrorSeg<<16)|13,
		ErrorLevelError,
		"",
	)

	// ErrRuntimeIllegalInCurrentGoroutine ...
	ErrRuntimeIllegalInCurrentGoroutine = DefineDevelopError(
		(generalErrorSeg<<16)|14,
		ErrorLevelFatal,
		"Runtime is illegal in current goroutine",
	)

	// ErrRuntimeReplyHasBeenCalled ...
	ErrRuntimeReplyHasBeenCalled = DefineDevelopError(
		(generalErrorSeg<<16)|15,
		ErrorLevelError,
		"Runtime.Reply has been called before",
	)

	// ErrRuntimeExternalReturn ...
	ErrRuntimeExternalReturn = DefineDevelopError(
		(generalErrorSeg<<16)|16,
		ErrorLevelError,
		"action must be return through Runtime.OK or Runtime.Error",
	)

	// ErrRTArrayIndexOverflow ...
	ErrRTArrayIndexOverflow = DefineDevelopError(
		(generalErrorSeg<<16)|17,
		ErrorLevelFatal,
		"",
	)

	// ErrRTMapNameNotFound ...
	ErrRTMapNameNotFound = DefineDevelopError(
		(generalErrorSeg<<16)|18,
		ErrorLevelFatal,
		"",
	)

	// ErrNumOfThreadsIsWrong ... *
	ErrNumOfThreadsIsWrong = DefineConfigError(
		(generalErrorSeg<<16)|19,
		ErrorLevelFatal,
		"numOfThreads is wrong",
	)

	// ErrThreadBufferSizeIsWrong ... *
	ErrThreadBufferSizeIsWrong = DefineConfigError(
		(generalErrorSeg<<16)|20,
		ErrorLevelFatal,
		"threadBufferSize is wrong",
	)

	// ErrMaxNodeDepthIsWrong ... *
	ErrMaxNodeDepthIsWrong = DefineConfigError(
		(generalErrorSeg<<16)|21,
		ErrorLevelFatal,
		"maxNodeDepth is wrong",
	)

	// ErrCacheMkdirAll ... *
	ErrCacheMkdirAll = DefineConfigError(
		(generalErrorSeg<<16)|22,
		ErrorLevelFatal,
		"create directory error",
	)

	// ErrCacheWriteFile ... *
	ErrCacheWriteFile = DefineConfigError(
		(generalErrorSeg<<16)|23,
		ErrorLevelFatal,
		"write to file error",
	)
)

const coreErrorSeg = 1001

var (
	// ErrFnCacheIllegalKindString ... *
	ErrFnCacheIllegalKindString = DefineKernelError(
		(coreErrorSeg<<16)|1,
		ErrorLevelFatal,
		"",
	)

	// ErrFnCacheDuplicateKindString ... *
	ErrFnCacheDuplicateKindString = DefineKernelError(
		(coreErrorSeg<<16)|2,
		ErrorLevelFatal,
		"",
	)

	// ErrProcessorOnReturnStreamIsNil ... *
	ErrProcessorOnReturnStreamIsNil = DefineKernelError(
		(coreErrorSeg<<16)|3,
		ErrorLevelFatal,
		"onReturnStream is nil",
	)

	// ErrProcessorMaxCallDepthIsWrong ... *
	ErrProcessorMaxCallDepthIsWrong = DefineKernelError(
		(coreErrorSeg<<16)|4,
		ErrorLevelFatal,
		"maxCallDepth is wrong",
	)

	// ErrProcessorNodeMetaIsNil ... *
	ErrProcessorNodeMetaIsNil = DefineKernelError(
		(coreErrorSeg<<16)|5,
		ErrorLevelFatal,
		"node meta is nil",
	)

	// ErrProcessorActionMetaIsNil ... *
	ErrProcessorActionMetaIsNil = DefineKernelError(
		(coreErrorSeg<<16)|6,
		ErrorLevelFatal,
		"action meta is nil",
	)

	// ErrThreadEvalFatal ...
	ErrThreadEvalFatal = DefineKernelError(
		(coreErrorSeg<<16)|7,
		ErrorLevelFatal,
		"",
	)

	// ErrProcessorIsNotRunning ... *
	ErrProcessorIsNotRunning = DefineKernelError(
		(coreErrorSeg<<16)|8,
		ErrorLevelFatal,
		"processor is not running",
	)
)

const gatewayErrorSeg = 1002

var (
	// ErrGatewayNoAvailableAdapter ...
	ErrGatewayNoAvailableAdapter = DefineConfigError(
		(gatewayErrorSeg<<16)|1,
		ErrorLevelFatal,
		"no listener is set on the server",
	)

	// ErrGatewayAlreadyRunning ...
	ErrGatewayAlreadyRunning = DefineConfigError(
		(gatewayErrorSeg<<16)|2,
		ErrorLevelFatal,
		"it is already running",
	)

	// ErrGateWaySessionNotFound ...
	ErrGateWaySessionNotFound = DefineConfigError(
		(gatewayErrorSeg<<16)|3,
		ErrorLevelWarn,
		"session not found",
	)

	// ErrGateWaySeedOverflows ...
	ErrGateWaySeedOverflows = DefineConfigError(
		(gatewayErrorSeg<<16)|4,
		ErrorLevelWarn,
		"gateway seed overflows",
	)
)

const serverErrorSeg = 1003

var (
	// ErrServerAlreadyRunning ...
	ErrServerAlreadyRunning = DefineConfigError(
		(serverErrorSeg<<16)|1,
		ErrorLevelFatal,
		"it is already running",
	)

	// ErrServerNotRunning ...
	ErrServerNotRunning = DefineConfigError(
		(serverErrorSeg<<16)|2,
		ErrorLevelFatal,
		"it is not running",
	)
)

const clientErrorSeg = 1004

var (
	// ErrClientTimeout ...
	ErrClientTimeout = DefineNetError(
		(clientErrorSeg<<16)|1,
		ErrorLevelWarn,
		"timeout",
	)

	// ErrClientConfig ...
	ErrClientConfig = DefineNetError(
		(clientErrorSeg<<16)|2,
		ErrorLevelWarn,
		"client config error",
	)
)

const adapterErrorSeg = 2000

var (
	// ErrUnsupportedProtocol ...
	ErrUnsupportedProtocol = DefineNetError(
		(adapterErrorSeg<<16)|1,
		ErrorLevelFatal,
		"",
	)

	// ErrSyncTCPServerServiceListen ...
	ErrSyncTCPServerServiceListen = DefineNetError(
		(adapterErrorSeg<<16)|2,
		ErrorLevelFatal,
		"",
	)

	// ErrSyncTCPServerServiceAccept ...
	ErrSyncTCPServerServiceAccept = DefineNetError(
		(adapterErrorSeg<<16)|3,
		ErrorLevelFatal,
		"",
	)

	// ErrSyncTCPServerServiceClose ...
	ErrSyncTCPServerServiceClose = DefineNetError(
		(adapterErrorSeg<<16)|4,
		ErrorLevelFatal,
		"",
	)

	// ErrSyncWSServerServiceListen ...
	ErrSyncWSServerServiceListen = DefineNetError(
		(adapterErrorSeg<<16)|5,
		ErrorLevelFatal,
		"",
	)

	// ErrSyncWSServerServiceUpgrade ...
	ErrSyncWSServerServiceUpgrade = DefineNetError(
		(adapterErrorSeg<<16)|6,
		ErrorLevelFatal,
		"",
	)

	// ErrSyncWSServerServiceServe ...
	ErrSyncWSServerServiceServe = DefineNetError(
		(adapterErrorSeg<<16)|7,
		ErrorLevelFatal,
		"",
	)

	// ErrSyncWSServerServiceClose ...
	ErrSyncWSServerServiceClose = DefineNetError(
		(adapterErrorSeg<<16)|8,
		ErrorLevelFatal,
		"",
	)

	// ErrSyncClientServiceDial ...
	ErrSyncClientServiceDial = DefineNetError(
		(adapterErrorSeg<<16)|9,
		ErrorLevelFatal,
		"",
	)

	// ErrConnClose ...
	ErrConnClose = DefineNetError(
		(adapterErrorSeg<<16)|10,
		ErrorLevelFatal,
		"",
	)

	// ErrConnRead ...
	ErrConnRead = DefineNetError(
		(adapterErrorSeg<<16)|11,
		ErrorLevelFatal,
		"",
	)

	// ErrConnWrite ...
	ErrConnWrite = DefineNetError(
		(adapterErrorSeg<<16)|12,
		ErrorLevelFatal,
		"",
	)

	// ErrOnFillWriteFatal ...
	ErrOnFillWriteFatal = DefineKernelError(
		(adapterErrorSeg<<16)|13,
		ErrorLevelFatal,
		"kernel error",
	)
)
