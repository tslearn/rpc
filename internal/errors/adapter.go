package errors

import (
	"github.com/rpccloud/rpc/internal/base"
)

const adapterErrorSeg = 2000

var (
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
)

const tcpServerAdapterErrorSeg = 2101

var (
	ErrTCPServerAdapterAlreadyRunning = base.DefineNetError(
		(tcpServerAdapterErrorSeg<<16)|1,
		base.ErrorLevelWarn,
		"it is already running",
	)

	ErrTCPServerAdapterNotRunning = base.DefineNetError(
		(tcpServerAdapterErrorSeg<<16)|2,
		base.ErrorLevelWarn,
		"it is not running",
	)

	ErrTCPServerAdapterListen = base.DefineNetError(
		(tcpServerAdapterErrorSeg<<16)|3,
		base.ErrorLevelWarn,
		"",
	)

	ErrTCPServerAdapterAccept = base.DefineNetError(
		(tcpServerAdapterErrorSeg<<16)|4,
		base.ErrorLevelWarn,
		"",
	)

	ErrTCPServerAdapterClose = base.DefineNetError(
		(tcpServerAdapterErrorSeg<<16)|5,
		base.ErrorLevelWarn,
		"",
	)
)

const tcpClientAdapterErrorSeg = 2102

var (
	ErrTCPClientAdapterAlreadyRunning = base.DefineNetError(
		(tcpClientAdapterErrorSeg<<16)|1,
		base.ErrorLevelWarn,
		"it is already running",
	)

	ErrTCPClientAdapterNotRunning = base.DefineNetError(
		(tcpClientAdapterErrorSeg<<16)|2,
		base.ErrorLevelWarn,
		"it is not running",
	)

	ErrTCPClientAdapterDail = base.DefineNetError(
		(tcpClientAdapterErrorSeg<<16)|3,
		base.ErrorLevelWarn,
		"",
	)

	ErrTCPClientAdapterClose = base.DefineNetError(
		(tcpClientAdapterErrorSeg<<16)|4,
		base.ErrorLevelWarn,
		"",
	)
)
