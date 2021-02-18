package base

const generalErrorSeg = 0 << 8

var (
	// ErrStream ...
	ErrStream = DefineSecurityError(
		generalErrorSeg|1,
		ErrorLevelWarn,
		"stream error",
	)

	// ErrUnsupportedValue ...
	ErrUnsupportedValue = DefineDevelopError(
		generalErrorSeg|2,
		ErrorLevelError,
		"",
	)

	// ErrServiceName ... *
	ErrServiceName = DefineDevelopError(
		generalErrorSeg|3,
		ErrorLevelFatal,
		"",
	)

	// ErrServiceIsNil ... *
	ErrServiceIsNil = DefineDevelopError(
		generalErrorSeg|4,
		ErrorLevelFatal,
		"service is nil",
	)

	// ErrServiceOverflow ... *
	ErrServiceOverflow = DefineDevelopError(
		generalErrorSeg|5,
		ErrorLevelFatal,
		"",
	)

	// ErrActionName ... *
	ErrActionName = DefineDevelopError(
		generalErrorSeg|6,
		ErrorLevelFatal,
		"",
	)

	// ErrActionHandler ... *
	ErrActionHandler = DefineDevelopError(
		generalErrorSeg|7,
		ErrorLevelFatal,
		"",
	)

	// ErrActionCloseTimeout ... *
	ErrActionCloseTimeout = DefineDevelopError(
		generalErrorSeg|8,
		ErrorLevelError,
		"",
	)

	// ErrAction ...
	ErrAction = DefineActionError(
		generalErrorSeg|9,
		ErrorLevelError,
		"",
	)

	// ErrActionPanic ...
	ErrActionPanic = DefineDevelopError(
		generalErrorSeg|10,
		ErrorLevelFatal,
		"",
	)

	// ErrTargetNotExist ...
	ErrTargetNotExist = DefineSecurityError(
		generalErrorSeg|11,
		ErrorLevelError,
		"",
	)

	// ErrCallOverflow ...
	ErrCallOverflow = DefineSecurityError(
		generalErrorSeg|12,
		ErrorLevelError,
		"",
	)

	// ErrArgumentsNotMatch ...
	ErrArgumentsNotMatch = DefineSecurityError(
		generalErrorSeg|13,
		ErrorLevelError,
		"",
	)

	// ErrRuntimeIllegalInCurrentGoroutine ...
	ErrRuntimeIllegalInCurrentGoroutine = DefineDevelopError(
		generalErrorSeg|14,
		ErrorLevelFatal,
		"Runtime is illegal in current goroutine",
	)

	// ErrRuntimeReplyHasBeenCalled ...
	ErrRuntimeReplyHasBeenCalled = DefineDevelopError(
		generalErrorSeg|15,
		ErrorLevelError,
		"Runtime.Reply has been called before",
	)

	// ErrRuntimeExternalReturn ...
	ErrRuntimeExternalReturn = DefineDevelopError(
		generalErrorSeg|16,
		ErrorLevelError,
		"action must be return through Runtime.OK or Runtime.Error",
	)

	// ErrRTArrayIndexOverflow ...
	ErrRTArrayIndexOverflow = DefineDevelopError(
		generalErrorSeg|17,
		ErrorLevelFatal,
		"",
	)

	// ErrRTMapNameNotFound ...
	ErrRTMapNameNotFound = DefineDevelopError(
		generalErrorSeg|18,
		ErrorLevelFatal,
		"",
	)

	// ErrNumOfThreadsIsWrong ... *
	ErrNumOfThreadsIsWrong = DefineConfigError(
		generalErrorSeg|19,
		ErrorLevelFatal,
		"numOfThreads is wrong",
	)

	// ErrThreadBufferSizeIsWrong ... *
	ErrThreadBufferSizeIsWrong = DefineConfigError(
		generalErrorSeg|20,
		ErrorLevelFatal,
		"threadBufferSize is wrong",
	)

	// ErrMaxNodeDepthIsWrong ... *
	ErrMaxNodeDepthIsWrong = DefineConfigError(
		generalErrorSeg|21,
		ErrorLevelFatal,
		"maxNodeDepth is wrong",
	)

	// ErrCacheMkdirAll ... *
	ErrCacheMkdirAll = DefineConfigError(
		generalErrorSeg|22,
		ErrorLevelFatal,
		"create directory error",
	)

	// ErrCacheWriteFile ... *
	ErrCacheWriteFile = DefineConfigError(
		generalErrorSeg|23,
		ErrorLevelFatal,
		"write to file error",
	)
)

const coreErrorSeg = 1 << 8

var (
	// ErrFnCacheIllegalKindString ... *
	ErrFnCacheIllegalKindString = DefineKernelError(
		coreErrorSeg|1,
		ErrorLevelFatal,
		"",
	)

	// ErrFnCacheDuplicateKindString ... *
	ErrFnCacheDuplicateKindString = DefineKernelError(
		coreErrorSeg|2,
		ErrorLevelFatal,
		"",
	)

	// ErrProcessorMaxCallDepthIsWrong ... *
	ErrProcessorMaxCallDepthIsWrong = DefineKernelError(
		coreErrorSeg|4,
		ErrorLevelFatal,
		"maxCallDepth is wrong",
	)

	// ErrProcessorNodeMetaIsNil ... *
	ErrProcessorNodeMetaIsNil = DefineKernelError(
		coreErrorSeg|5,
		ErrorLevelFatal,
		"node meta is nil",
	)

	// ErrProcessorActionMetaIsNil ... *
	ErrProcessorActionMetaIsNil = DefineKernelError(
		coreErrorSeg|6,
		ErrorLevelFatal,
		"action meta is nil",
	)

	// ErrThreadEvalFatal ...
	ErrThreadEvalFatal = DefineKernelError(
		coreErrorSeg|7,
		ErrorLevelFatal,
		"",
	)

	// ErrProcessorIsNotRunning ... *
	ErrProcessorIsNotRunning = DefineKernelError(
		coreErrorSeg|8,
		ErrorLevelFatal,
		"processor is not running",
	)
)

const gatewayErrorSeg = 2 << 8

var (
	// ErrGatewayNoAvailableAdapter ...
	ErrGatewayNoAvailableAdapter = DefineConfigError(
		gatewayErrorSeg|1,
		ErrorLevelFatal,
		"no listener is set on the server",
	)

	// ErrGatewayAlreadyRunning ...
	ErrGatewayAlreadyRunning = DefineConfigError(
		gatewayErrorSeg|2,
		ErrorLevelFatal,
		"it is already running",
	)

	// ErrGateWaySessionNotFound ...
	ErrGateWaySessionNotFound = DefineConfigError(
		gatewayErrorSeg|3,
		ErrorLevelWarn,
		"session not found",
	)

	// ErrGateWaySeedOverflows ...
	ErrGateWaySeedOverflows = DefineConfigError(
		gatewayErrorSeg|4,
		ErrorLevelWarn,
		"gateway seed overflows",
	)
)

const serverErrorSeg = 3 << 8

var (
	// ErrServerAlreadyRunning ...
	ErrServerAlreadyRunning = DefineConfigError(
		serverErrorSeg|1,
		ErrorLevelFatal,
		"it is already running",
	)

	// ErrServerNotRunning ...
	ErrServerNotRunning = DefineConfigError(
		serverErrorSeg|2,
		ErrorLevelFatal,
		"it is not running",
	)
)

const clientErrorSeg = 4 << 8

var (
	// ErrClientTimeout ...
	ErrClientTimeout = DefineNetError(
		clientErrorSeg|1,
		ErrorLevelWarn,
		"timeout",
	)

	// ErrClientConfig ...
	ErrClientConfig = DefineNetError(
		clientErrorSeg|2,
		ErrorLevelWarn,
		"client config error",
	)
)

const adapterErrorSeg = 5 << 8

var (
	// ErrUnsupportedProtocol ...
	ErrUnsupportedProtocol = DefineNetError(
		adapterErrorSeg|1,
		ErrorLevelFatal,
		"",
	)

	// ErrSyncTCPServerServiceListen ...
	ErrSyncTCPServerServiceListen = DefineNetError(
		adapterErrorSeg|2,
		ErrorLevelFatal,
		"",
	)

	// ErrSyncTCPServerServiceAccept ...
	ErrSyncTCPServerServiceAccept = DefineNetError(
		adapterErrorSeg|3,
		ErrorLevelFatal,
		"",
	)

	// ErrSyncTCPServerServiceClose ...
	ErrSyncTCPServerServiceClose = DefineNetError(
		adapterErrorSeg|4,
		ErrorLevelFatal,
		"",
	)

	// ErrSyncWSServerServiceListen ...
	ErrSyncWSServerServiceListen = DefineNetError(
		adapterErrorSeg|5,
		ErrorLevelFatal,
		"",
	)

	// ErrSyncWSServerServiceUpgrade ...
	ErrSyncWSServerServiceUpgrade = DefineNetError(
		adapterErrorSeg|6,
		ErrorLevelFatal,
		"",
	)

	// ErrSyncWSServerServiceServe ...
	ErrSyncWSServerServiceServe = DefineNetError(
		adapterErrorSeg|7,
		ErrorLevelFatal,
		"",
	)

	// ErrSyncWSServerServiceClose ...
	ErrSyncWSServerServiceClose = DefineNetError(
		adapterErrorSeg|8,
		ErrorLevelFatal,
		"",
	)

	// ErrSyncClientServiceDial ...
	ErrSyncClientServiceDial = DefineNetError(
		adapterErrorSeg|9,
		ErrorLevelFatal,
		"",
	)

	// ErrConnClose ...
	ErrConnClose = DefineNetError(
		adapterErrorSeg|10,
		ErrorLevelFatal,
		"",
	)

	// ErrConnRead ...
	ErrConnRead = DefineNetError(
		adapterErrorSeg|11,
		ErrorLevelFatal,
		"",
	)

	// ErrConnWrite ...
	ErrConnWrite = DefineNetError(
		adapterErrorSeg|12,
		ErrorLevelFatal,
		"",
	)

	// ErrOnFillWriteFatal ...
	ErrOnFillWriteFatal = DefineKernelError(
		adapterErrorSeg|13,
		ErrorLevelFatal,
		"kernel error",
	)
)
