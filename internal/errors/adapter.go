package errors

import (
	"github.com/rpccloud/rpc/internal/base"
)

const adapterErrorSeg = 2000

var (
	// ErrUnsupportedProtocol ...
	ErrUnsupportedProtocol = base.DefineNetError(
		(adapterErrorSeg<<16)|21,
		base.ErrorLevelFatal,
		"",
	)

	// ErrSyncTCPServerServiceListen ...
	ErrSyncTCPServerServiceListen = base.DefineNetError(
		(adapterErrorSeg<<16)|22,
		base.ErrorLevelFatal,
		"",
	)

	// ErrSyncTCPServerServiceAccept ...
	ErrSyncTCPServerServiceAccept = base.DefineNetError(
		(adapterErrorSeg<<16)|23,
		base.ErrorLevelFatal,
		"",
	)

	// ErrSyncTCPServerServiceClose ...
	ErrSyncTCPServerServiceClose = base.DefineNetError(
		(adapterErrorSeg<<16)|24,
		base.ErrorLevelFatal,
		"",
	)

	// ErrSyncWSServerServiceListen ...
	ErrSyncWSServerServiceListen = base.DefineNetError(
		(adapterErrorSeg<<16)|25,
		base.ErrorLevelFatal,
		"",
	)

	// ErrSyncWSServerServiceUpgrade ...
	ErrSyncWSServerServiceUpgrade = base.DefineNetError(
		(adapterErrorSeg<<16)|26,
		base.ErrorLevelFatal,
		"",
	)

	// ErrSyncWSServerServiceServe ...
	ErrSyncWSServerServiceServe = base.DefineNetError(
		(adapterErrorSeg<<16)|27,
		base.ErrorLevelFatal,
		"",
	)

	// ErrSyncWSServerServiceClose ...
	ErrSyncWSServerServiceClose = base.DefineNetError(
		(adapterErrorSeg<<16)|28,
		base.ErrorLevelFatal,
		"",
	)

	// ErrSyncClientServiceDial ...
	ErrSyncClientServiceDial = base.DefineNetError(
		(adapterErrorSeg<<16)|29,
		base.ErrorLevelFatal,
		"",
	)

	// ErrConnClose ...
	ErrConnClose = base.DefineNetError(
		(adapterErrorSeg<<16)|30,
		base.ErrorLevelFatal,
		"",
	)

	// ErrConnRead ...
	ErrConnRead = base.DefineNetError(
		(adapterErrorSeg<<16)|31,
		base.ErrorLevelFatal,
		"",
	)

	// ErrConnWrite ...
	ErrConnWrite = base.DefineNetError(
		(adapterErrorSeg<<16)|32,
		base.ErrorLevelFatal,
		"",
	)

	// ErrOnFillWriteFatal ...
	ErrOnFillWriteFatal = base.DefineKernelError(
		(adapterErrorSeg<<16)|5,
		base.ErrorLevelFatal,
		"kernel error",
	)
)
