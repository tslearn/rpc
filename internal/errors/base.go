package errors

import "github.com/rpccloud/rpc/internal/base"

var (
	// ErrRuntimeGeneral ...
	ErrRuntimeGeneral = base.DefineRuntimeError(
		1, base.ErrorLevelError, "",
	)

	// ErrBadStream ...
	ErrBadStream = base.DefineProtocolError(
		2, base.ErrorLevelWarn, "bad stream",
	)

	// ErrStreamConnIsClosed ...
	ErrStreamConnIsClosed = base.DefineTransportError(
		3, base.ErrorLevelWarn, "stream conn is closed",
	)

	ErrReadFile = base.DefineKernelError(
		4, base.ErrorLevelFatal, "",
	)
)
