package internal

import (
	"net"
	"time"
)

// IStreamConn ...
type IStreamConn interface {
	ReadStream(timeout time.Duration, readLimit int64) (*Stream, Error)
	WriteStream(stream *Stream, timeout time.Duration) Error
	Close() Error
}

// IServerAdapter ...
type IServerAdapter interface {
	Open(onConnRun func(IStreamConn, net.Addr), onError func(uint64, Error))
	Close(onError func(uint64, Error))
}

// IClientAdapter ...
type IClientAdapter interface {
	Open(onConnRun func(IStreamConn), onError func(Error))
	Close(onError func(Error))
}

// ReplyCache ...
type ReplyCache interface {
	Get(fnString string) ReplyCacheFunc
}

// ReplyCacheFunc ...
type ReplyCacheFunc = func(
	rt Runtime,
	stream *Stream,
	fn interface{},
) bool

// Bool ...
type Bool = bool

// Int64 ...
type Int64 = int64

// Uint64 ...
type Uint64 = uint64

// Float64 ...
type Float64 = float64

// String ...
type String = string

// Bytes ...
type Bytes = []byte

// Array ...
type Array = []interface{}

// Map common Map type
type Map = map[string]interface{}

// Any ...
type Any = interface{}

// Return ...
type returnObject struct{}
type Return = *returnObject

var emptyReturn = &returnObject{}
