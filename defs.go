package rpcc

import (
	"github.com/tslearn/rpcc/internal"
	"time"
)

const configServerReadLimit = int64(1024 * 1024)
const configServerWriteLimit = int64(1024 * 1024)
const configReadTimeout = 10 * time.Second
const configWriteTimeout = 1 * time.Second

const SystemStreamKindInit = int64(1)
const SystemStreamKindInitBack = int64(2)
const SystemStreamKindRequestIds = int64(3)
const SystemStreamKindRequestIdsBack = int64(4)

type IStreamConn interface {
	ReadStream(timeout time.Duration, readLimit int64) (Stream, Error)
	WriteStream(stream Stream, timeout time.Duration, writeLimit int64) Error
	Close() Error
}

type IEndPoint interface {
	ConnectString() string
	IsRunning() bool
	Open(onConnRun func(IStreamConn), onError func(Error)) bool
	Close(onError func(Error)) bool
}

// Stream ...
type Stream = *internal.RPCStream

// Bool ...
type Bool = internal.RPCBool

// Int ...
type Int = internal.RPCInt

// Uint ...
type Uint = internal.RPCUint

// Float ...
type Float = internal.RPCFloat

// String ...
type String = internal.RPCString

// Bytes ...
type Bytes = internal.RPCBytes

// Any common Any type
type Any = internal.RPCAny

// Array common Array type
type Array = internal.RPCArray

// Map common Map type
type Map = internal.RPCMap

// Error ...
type Error = internal.RPCError

// Return ...
type Return = *internal.RPCReturn

// Context ...
type Context = *internal.RPCContext

// Service ...
type Service = internal.Service

// NewService ...
var NewService = internal.NewService
