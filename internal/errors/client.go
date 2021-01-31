package errors

import "github.com/rpccloud/rpc/internal/base"

const clientErrorSeg = 6001

var (
	// ErrClientTimeout ...
	ErrClientTimeout = base.DefineNetError(
		(clientErrorSeg<<16)|1,
		base.ErrorLevelWarn,
		"timeout",
	)

	// ErrClientConfig ...
	ErrClientConfig = base.DefineNetError(
		(clientErrorSeg<<16)|2,
		base.ErrorLevelWarn,
		"client config error",
	)
)
