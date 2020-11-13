package errors

import "github.com/rpccloud/rpc/internal/base"

const clientErrorSeg = 6001

var (
	ErrClientAlreadyRunning = base.DefineConfigError(
		(clientErrorSeg<<16)|2,
		base.ErrorLevelFatal,
		"it is already running",
	)

	ErrClientNotRunning = base.DefineConfigError(
		(clientErrorSeg<<16)|3,
		base.ErrorLevelFatal,
		"it is not running",
	)
)
