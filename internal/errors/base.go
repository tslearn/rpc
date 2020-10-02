package errors

import "github.com/rpccloud/rpc/internal/base"

var (
	// ErrRuntimeGeneral ...
	ErrRuntimeGeneral = base.DefineDevelopError(
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
)
