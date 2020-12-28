package errors

import "github.com/rpccloud/rpc/internal/base"

const serverErrorSeg = 5001

var (
	// ErrServerAlreadyRunning ...
	ErrServerAlreadyRunning = base.DefineConfigError(
		(serverErrorSeg<<16)|2,
		base.ErrorLevelFatal,
		"it is already running",
	)

	// ErrServerNotRunning ...
	ErrServerNotRunning = base.DefineConfigError(
		(serverErrorSeg<<16)|3,
		base.ErrorLevelFatal,
		"it is not running",
	)
)
