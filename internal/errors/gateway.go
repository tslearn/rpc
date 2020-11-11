package errors

import "github.com/rpccloud/rpc/internal/base"

const gatewayErrorSeg = 3000

var (
	ErrGatewayNoAvailableAdapters = base.DefineConfigError(
		(gatewayErrorSeg<<16)|1,
		base.ErrorLevelFatal,
		"no listener is set on the server",
	)

	ErrGatewayAlreadyRunning = base.DefineConfigError(
		(gatewayErrorSeg<<16)|2,
		base.ErrorLevelFatal,
		"it is already running",
	)

	ErrGatewayNotRunning = base.DefineConfigError(
		(gatewayErrorSeg<<16)|3,
		base.ErrorLevelFatal,
		"it is not running",
	)

	ErrGatewaySequenceError = base.DefineConfigError(
		(gatewayErrorSeg<<16)|4,
		base.ErrorLevelError,
		"channel sequence error",
	)

	ErrGateWaySessionNotFound = base.DefineConfigError(
		(gatewayErrorSeg<<16)|5,
		base.ErrorLevelWarn,
		"session not found",
	)
)
