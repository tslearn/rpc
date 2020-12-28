package errors

import (
	"github.com/rpccloud/rpc/internal/base"
)

const adapterErrorSeg = 2000

var (
	// ErrTemp ... *
	ErrTemp = base.DefineNetError(
		(adapterErrorSeg<<16)|1000,
		base.ErrorLevelWarn,
		"",
	)

	// ErrStreamConnIsClosed ... *
	ErrStreamConnIsClosed = base.DefineNetError(
		(adapterErrorSeg<<16)|1,
		base.ErrorLevelWarn,
		"stream conn is closed",
	)

	// ErrKqueueSystem  ...
	ErrKqueueSystem = base.DefineNetError(
		(adapterErrorSeg<<16)|2,
		base.ErrorLevelFatal,
		"",
	)

	// ErrKqueueNotRunning ...
	ErrKqueueNotRunning = base.DefineNetError(
		(adapterErrorSeg<<16)|3,
		base.ErrorLevelFatal,
		"it is not running",
	)

	// ErrTCPListener  ...
	ErrTCPListener = base.DefineNetError(
		(adapterErrorSeg<<16)|4,
		base.ErrorLevelFatal,
		"",
	)
)

const eventConnErrorSeg = 2101

var (
	// ErrEventConnReadBufferIsTooSmall ...
	ErrEventConnReadBufferIsTooSmall = base.DefineConfigError(
		(eventConnErrorSeg<<16)|1,
		base.ErrorLevelFatal,
		"event conn read buffer is too small",
	)

	// ErrEventConnRead ...
	ErrEventConnRead = base.DefineConfigError(
		(eventConnErrorSeg<<16)|2,
		base.ErrorLevelWarn,
		"",
	)

	// ErrEventConnReadLimit ...
	ErrEventConnReadLimit = base.DefineConfigError(
		(eventConnErrorSeg<<16)|3,
		base.ErrorLevelWarn,
		"",
	)

	// ErrEventConnWriteStream ...
	ErrEventConnWriteStream = base.DefineNetError(
		(eventConnErrorSeg<<16)|4,
		base.ErrorLevelWarn,
		"",
	)

	// ErrEventConnClose ...
	ErrEventConnClose = base.DefineNetError(
		(eventConnErrorSeg<<16)|5,
		base.ErrorLevelWarn,
		"",
	)
)

const tcpServerAdapterErrorSeg = 2102

var (
	// ErrTCPServerAdapterAlreadyRunning ...
	ErrTCPServerAdapterAlreadyRunning = base.DefineNetError(
		(tcpServerAdapterErrorSeg<<16)|1,
		base.ErrorLevelWarn,
		"it is already running",
	)

	// ErrTCPServerAdapterNotRunning ...
	ErrTCPServerAdapterNotRunning = base.DefineNetError(
		(tcpServerAdapterErrorSeg<<16)|2,
		base.ErrorLevelWarn,
		"it is not running",
	)

	// ErrTCPServerAdapterListen ...
	ErrTCPServerAdapterListen = base.DefineNetError(
		(tcpServerAdapterErrorSeg<<16)|3,
		base.ErrorLevelWarn,
		"",
	)

	// ErrTCPServerAdapterAccept ...
	ErrTCPServerAdapterAccept = base.DefineNetError(
		(tcpServerAdapterErrorSeg<<16)|4,
		base.ErrorLevelWarn,
		"",
	)

	// ErrTCPServerAdapterClose ...
	ErrTCPServerAdapterClose = base.DefineNetError(
		(tcpServerAdapterErrorSeg<<16)|5,
		base.ErrorLevelWarn,
		"",
	)
)

const tcpClientAdapterErrorSeg = 2103

var (
	// ErrTCPClientAdapterAlreadyRunning ...
	ErrTCPClientAdapterAlreadyRunning = base.DefineNetError(
		(tcpClientAdapterErrorSeg<<16)|1,
		base.ErrorLevelWarn,
		"it is already running",
	)

	// ErrTCPClientAdapterNotRunning ...
	ErrTCPClientAdapterNotRunning = base.DefineNetError(
		(tcpClientAdapterErrorSeg<<16)|2,
		base.ErrorLevelWarn,
		"it is not running",
	)

	// ErrTCPClientAdapterDial ...
	ErrTCPClientAdapterDial = base.DefineNetError(
		(tcpClientAdapterErrorSeg<<16)|3,
		base.ErrorLevelWarn,
		"",
	)

	// ErrTCPClientAdapterClose ...
	ErrTCPClientAdapterClose = base.DefineNetError(
		(tcpClientAdapterErrorSeg<<16)|4,
		base.ErrorLevelWarn,
		"",
	)
)
