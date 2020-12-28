package errors

import "github.com/rpccloud/rpc/internal/base"

const clientErrorSeg = 6001

var (
	// ErrClientAlreadyRunning ...
	ErrClientAlreadyRunning = base.DefineConfigError(
		(clientErrorSeg<<16)|1,
		base.ErrorLevelFatal,
		"it is already running",
	)

	// ErrClientNotRunning ...
	ErrClientNotRunning = base.DefineConfigError(
		(clientErrorSeg<<16)|2,
		base.ErrorLevelFatal,
		"it is not running",
	)

	// ErrClientTimeout ...
	ErrClientTimeout = base.DefineNetError(
		(clientErrorSeg<<16)|3,
		base.ErrorLevelWarn,
		"timeout",
	)

	// ErrClientConnectString ...
	ErrClientConnectString = base.DefineNetError(
		(clientErrorSeg<<16)|4,
		base.ErrorLevelWarn,
		"",
	)

	// ErrClientConfigChanges ...
	ErrClientConfigChanges = base.DefineKernelError(
		(clientErrorSeg<<16)|5,
		base.ErrorLevelWarn,
		"client config changes",
	)
)
