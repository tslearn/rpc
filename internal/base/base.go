package base

var (
	// ErrRuntimeGeneral ...
	ErrRuntimeGeneral = DefineRuntimeError(
		1, ErrorLevelError, "",
	)

	// ErrBadStream ...
	ErrBadStream = DefineProtocolError(
		2, ErrorLevelWarn, "bad stream",
	)

	// ErrStreamConnIsClosed ...
	ErrStreamConnIsClosed = DefineTransportError(
		3, ErrorLevelWarn, "stream conn is closed",
	)

	ErrReadFile = DefineKernelError(
		4, ErrorLevelFatal, "",
	)
)
