package errors

import "github.com/rpccloud/rpc/internal/base"

const clientErrorSeg = 6001

var (
	// ErrClientTimeout ...
	ErrClientTimeout = base.DefineNetError(
		(clientErrorSeg<<16)|3,
		base.ErrorLevelWarn,
		"timeout",
	)
)
