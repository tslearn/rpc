package errors

import "github.com/rpccloud/rpc/internal/base"

const directRouterErrorSeg = 4001

var (
	// ErrDirectRouterConfigError ...
	ErrDirectRouterConfigError = base.DefineConfigError(
		(directRouterErrorSeg<<16)|1,
		base.ErrorLevelFatal,
		"DirectRouter config error",
	)
)
