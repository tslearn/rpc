package errors

import (
	"github.com/rpccloud/rpc/internal/base"
)

const adapterErrorSeg = 2000

var (
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

	ErrConnClose = base.DefineNetError(
		(adapterErrorSeg<<16)|30,
		base.ErrorLevelFatal,
		"",
	)

	ErrConnRead = base.DefineNetError(
		(adapterErrorSeg<<16)|31,
		base.ErrorLevelFatal,
		"",
	)

	ErrConnWrite = base.DefineNetError(
		(adapterErrorSeg<<16)|32,
		base.ErrorLevelFatal,
		"",
	)

	ErrOnFillWriteFatal = base.DefineKernelError(
		(adapterErrorSeg<<16)|5,
		base.ErrorLevelFatal,
		"kernel error",
	)
)
