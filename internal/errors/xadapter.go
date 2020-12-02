package errors

import "github.com/rpccloud/rpc/internal/base"

const xAdapterErrorSeg = 2500

var (
	// ErrKqueueSystem  ...
	ErrKqueueSystem = base.DefineNetError(
		(xAdapterErrorSeg<<16)|1,
		base.ErrorLevelFatal,
		"",
	)
)