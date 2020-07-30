package rpc

import (
	"github.com/rpccloud/rpc/internal"
	"time"
)

const SystemStreamKindInit = int64(1)
const SystemStreamKindInitBack = int64(2)
const SystemStreamKindRequestIds = int64(3)
const SystemStreamKindRequestIdsBack = int64(4)

type IStreamConn interface {
	ReadStream(timeout time.Duration, readLimit int64) (*Stream, Error)
	WriteStream(stream *Stream, timeout time.Duration) Error
	Close() Error
}

type IAdapter interface {
	Open(onConnRun func(IStreamConn), onConnError func(Error)) Error
	Close() Error
}

// Bool ...
type Bool = internal.Bool

// Int64 ...
type Int64 = internal.Int64

// Uint64 ...
type Uint64 = internal.Uint64

// Float64 ...
type Float64 = internal.Float64

// String ...
type String = internal.String

// Bytes ...
type Bytes = internal.Bytes

// Any common Any type
type Any = internal.Any

// Array common Array type
type Array = internal.Array

// Map common Map type
type Map = internal.Map

// ReturnObject ...
type Return = internal.Return

// ContextObject ...
type Context = internal.Context

// Service ...
type Service = internal.Service

// NewService ...
var NewService = internal.NewService

// Error ...
type Error = internal.Error

// Stream ...
type Stream = internal.Stream

// ReplyCache ...
type ReplyCache = internal.ReplyCache

// ReplyCacheFunc ...
type ReplyCacheFunc = internal.ReplyCacheFunc
