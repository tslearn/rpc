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

	ErrUnsupportedProtocol = base.DefineNetError(
		(adapterErrorSeg<<16)|21,
		base.ErrorLevelFatal,
		"",
	)

	ErrSyncTCPServerServiceListen = base.DefineNetError(
		(adapterErrorSeg<<16)|22,
		base.ErrorLevelFatal,
		"",
	)

	ErrSyncTCPServerServiceAccept = base.DefineNetError(
		(adapterErrorSeg<<16)|23,
		base.ErrorLevelFatal,
		"",
	)

	ErrSyncTCPServerServiceClose = base.DefineNetError(
		(adapterErrorSeg<<16)|24,
		base.ErrorLevelFatal,
		"",
	)

	ErrSyncWSServerServiceListen = base.DefineNetError(
		(adapterErrorSeg<<16)|25,
		base.ErrorLevelFatal,
		"",
	)

	ErrSyncWSServerServiceUpgrade = base.DefineNetError(
		(adapterErrorSeg<<16)|26,
		base.ErrorLevelFatal,
		"",
	)

	ErrSyncWSServerServiceServe = base.DefineNetError(
		(adapterErrorSeg<<16)|27,
		base.ErrorLevelFatal,
		"",
	)

	ErrSyncWSServerServiceClose = base.DefineNetError(
		(adapterErrorSeg<<16)|28,
		base.ErrorLevelFatal,
		"",
	)

	ErrSyncClientServiceDial = base.DefineNetError(
		(adapterErrorSeg<<16)|29,
		base.ErrorLevelFatal,
		"",
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
