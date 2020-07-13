package rpcc

import (
	"github.com/tslearn/rpcc/core"
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
	ReadStream(
		timeout time.Duration,
		readLimit int64,
	) (*core.RPCStream, core.RPCError)
	WriteStream(
		stream *core.RPCStream,
		timeout time.Duration,
		writeLimit int64,
	) core.RPCError
	Close() core.RPCError
}

type IEndPoint interface {
	ConnectString() string
	IsRunning() bool
	Open(onConnRun func(IStreamConn), onError func(core.RPCError)) bool
	Close(onError func(core.RPCError)) bool
}

// Bool ...
type Bool = core.RPCBool

// Int ...
type Int = core.RPCInt

// Uint ...
type Uint = core.RPCUint

// Float ...
type Float = core.RPCFloat

// String ...
type String = core.RPCString

// Bytes ...
type Bytes = core.RPCBytes

// Any common Any type
type Any = core.RPCAny

// Array common Array type
type Array = core.RPCArray

// Map common Map type
type Map = core.RPCMap

// Error ...
type Error = core.RPCError

// Return ...
type Return = core.RPCReturn

// Context ...
type Context = core.RPCContext

// Service ...
type Service = core.Service

// NewService ...
var NewService = core.NewService
