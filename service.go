package rpc

import "github.com/rpccloud/rpc/internal/core"

// Bool ...
type Bool = core.Bool

// Int64 ...
type Int64 = core.Int64

// Uint64 ...
type Uint64 = core.Uint64

// Float64 ...
type Float64 = core.Float64

// String ...
type String = core.String

// Bytes ...
type Bytes = core.Bytes

// Array common Array type
type Array = core.Array

// Map common Map type
type Map = core.Map

// Any common Any type
type Any = core.Any

// RTValue ...
type RTValue = core.RTValue

// RTArray ...
type RTArray = core.RTArray

// RTMap ...
type RTMap = core.RTMap

// Return ...
type Return = core.Return

// Runtime ...
type Runtime = core.Runtime

// Service ...
type Service = core.Service

// NewService ...
var NewService = core.NewService

// ActionCache ...
type ActionCache = core.ActionCache

// ActionCacheFunc ...
type ActionCacheFunc = core.ActionCacheFunc
