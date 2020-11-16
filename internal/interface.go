package internal

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"net"
	"time"
)

// IStreamConn ...
type IStreamConn interface {
	ReadStream(timeout time.Duration, readLimit int64) (*core.Stream, *base.Error)
	WriteStream(stream *core.Stream, timeout time.Duration) *base.Error
	Close() *base.Error
}

// IServerAdapter ...
type IServerAdapter interface {
	Open(onConnRun func(IStreamConn, net.Addr), onError func(uint64, *base.Error))
	Close(onError func(uint64, *base.Error))
}

// IClientAdapter ...
type IClientAdapter interface {
	Open(onConnRun func(IStreamConn), onError func(*base.Error))
	Close(onError func(*base.Error))
}

type IStreamReceiver interface {
	OnStream(stream *core.Stream) *base.Error
}

type IStreamRouter interface {
	Plug(receiver IStreamReceiver) IStreamRouterSlot
}

type IStreamRouterSlot interface {
	SendStream(stream *core.Stream) *base.Error
}
