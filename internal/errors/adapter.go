package errors

import (
	"github.com/rpccloud/rpc/internal/base"
)

const streamConnErrorSeg = 2000

var (
	// ErrStreamConnIsClosed ...
	ErrStreamConnIsClosed = base.DefineTransportError(
		(streamConnErrorSeg<<16)|1,
		base.ErrorLevelWarn,
		"stream conn is closed",
	)
)

const websocketStreamConnErrorSeg = 2001

var (
	ErrWebsocketStreamConnStreamIsNil = base.DefineKernelError(
		(websocketStreamConnErrorSeg<<16)|1,
		base.ErrorLevelFatal,
		"stream is nil",
	)

	ErrWebsocketStreamConnDataIsNotBinary = base.DefineSecurityError(
		(websocketStreamConnErrorSeg<<16)|2,
		base.ErrorLevelWarn,
		"websocket data is not binary",
	)

	ErrWebsocketStreamConnWSConnWriteMessage = base.DefineTransportError(
		(websocketStreamConnErrorSeg<<16)|3,
		base.ErrorLevelWarn,
		"",
	)

	ErrWebsocketStreamConnWSConnSetReadDeadline = base.DefineTransportError(
		(websocketStreamConnErrorSeg<<16)|4,
		base.ErrorLevelWarn,
		"",
	)

	ErrWebsocketStreamConnWSConnReadMessage = base.DefineTransportError(
		(websocketStreamConnErrorSeg<<16)|5,
		base.ErrorLevelWarn,
		"",
	)

	ErrWebsocketStreamConnWSConnClose = base.DefineTransportError(
		(websocketStreamConnErrorSeg<<16)|6,
		base.ErrorLevelWarn,
		"",
	)
)

const websocketServerAdapterErrorSeg = 2002

var (
	ErrWebsocketServerAdapterUpgrade = base.DefineSecurityError(
		(websocketServerAdapterErrorSeg<<16)|1,
		base.ErrorLevelWarn,
		"websocket upgrade error",
	)

	ErrWebsocketServerAdapterAlreadyRunning = base.DefineKernelError(
		(websocketServerAdapterErrorSeg<<16)|2,
		base.ErrorLevelFatal,
		"it is already running",
	)

	ErrWebsocketServerAdapterNotRunning = base.DefineKernelError(
		(websocketServerAdapterErrorSeg<<16)|3,
		base.ErrorLevelFatal,
		"it is not running",
	)

	ErrWebsocketServerAdapterWSServerListenAndServe = base.DefineTransportError(
		(websocketServerAdapterErrorSeg<<16)|4,
		base.ErrorLevelFatal,
		"",
	)

	ErrWebsocketServerAdapterWSServerClose = base.DefineDevelopError(
		(websocketServerAdapterErrorSeg<<16)|5,
		base.ErrorLevelError,
		"",
	)

	ErrWebsocketServerAdapterCloseTimeout = base.DefineDevelopError(
		(websocketServerAdapterErrorSeg<<16)|6,
		base.ErrorLevelError,
		"close timeout",
	)
)

const websocketClientAdapterErrorSeg = 2003

var (
	ErrWebsocketClientAdapterDial = base.DefineTransportError(
		(websocketClientAdapterErrorSeg<<16)|1,
		base.ErrorLevelFatal,
		"",
	)

	ErrWebsocketClientAdapterAlreadyRunning = base.DefineKernelError(
		(websocketClientAdapterErrorSeg<<16)|2,
		base.ErrorLevelFatal,
		"it is already running",
	)

	ErrWebsocketClientAdapterNotRunning = base.DefineKernelError(
		(websocketClientAdapterErrorSeg<<16)|3,
		base.ErrorLevelFatal,
		"it is not running",
	)

	ErrWebsocketClientAdapterCloseTimeout = base.DefineDevelopError(
		(websocketClientAdapterErrorSeg<<16)|4,
		base.ErrorLevelError,
		"close timeout",
	)
)
