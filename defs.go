package rpc

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
)

const controlStreamKindInit = int64(1)
const controlStreamKindInitBack = int64(2)
const controlStreamKindRequestIds = int64(3)
const controlStreamKindRequestIdsBack = int64(4)

type streamHub interface {
	PutStream(*Stream) bool
	Close() bool
}

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

// Any common Any type
type Any = core.Any

// Array common Array type
type Array = core.Array

// Map common Map type
type Map = core.Map

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

// NewServiceWithOnMount ...
var NewServiceWithOnMount = core.NewServiceWithOnMount

// Error ...
type Error = base.Error

// Stream ...
type Stream = core.Stream

// ReplyCache ...
type ReplyCache = core.ReplyCache

// ReplyCacheFunc ...
type ReplyCacheFunc = core.ReplyCacheFunc

// GetRandString ...
var GetRandString = base.GetRandString

// TimeNow
var TimeNow = base.TimeNow
