package errors

import "github.com/rpccloud/rpc/internal/base"

const gatewayErrorSeg = 3001

var (
	// ErrGatewayNoAvailableAdapters ...
	ErrGatewayNoAvailableAdapters = base.DefineConfigError(
		(gatewayErrorSeg<<16)|1,
		base.ErrorLevelFatal,
		"no listener is set on the server",
	)

	// ErrGatewayAlreadyRunning ...
	ErrGatewayAlreadyRunning = base.DefineConfigError(
		(gatewayErrorSeg<<16)|2,
		base.ErrorLevelFatal,
		"it is already running",
	)

	// ErrGatewayNotRunning ...
	ErrGatewayNotRunning = base.DefineConfigError(
		(gatewayErrorSeg<<16)|3,
		base.ErrorLevelFatal,
		"it is not running",
	)

	// ErrGatewaySequenceError ...
	ErrGatewaySequenceError = base.DefineConfigError(
		(gatewayErrorSeg<<16)|4,
		base.ErrorLevelError,
		"channel sequence error",
	)

	// ErrGateWaySessionNotFound ...
	ErrGateWaySessionNotFound = base.DefineConfigError(
		(gatewayErrorSeg<<16)|5,
		base.ErrorLevelWarn,
		"session not found",
	)

	// ErrGateWayIDOverflows ...
	ErrGateWayIDOverflows = base.DefineConfigError(
		(gatewayErrorSeg<<16)|6,
		base.ErrorLevelWarn,
		"gateway id overflows",
	)

	// ErrGateWaySeedOverflows ...
	ErrGateWaySeedOverflows = base.DefineConfigError(
		(gatewayErrorSeg<<16)|7,
		base.ErrorLevelWarn,
		"gateway seed overflows",
	)
)
