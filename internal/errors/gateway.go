package errors

import "github.com/rpccloud/rpc/internal/base"

const gatewayErrorSeg = 3001

var (
	// ErrGatewayNoAvailableAdapter ...
	ErrGatewayNoAvailableAdapter = base.DefineConfigError(
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

	// ErrGateWaySessionNotFound ...
	ErrGateWaySessionNotFound = base.DefineConfigError(
		(gatewayErrorSeg<<16)|5,
		base.ErrorLevelWarn,
		"session not found",
	)

	// ErrGateWaySeedOverflows ...
	ErrGateWaySeedOverflows = base.DefineConfigError(
		(gatewayErrorSeg<<16)|7,
		base.ErrorLevelWarn,
		"gateway seed overflows",
	)
)
