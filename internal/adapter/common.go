package adapter

import (
	"net"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
)

// ErrNetClosingSuffix ...
const ErrNetClosingSuffix = "use of closed network connection"

// IConn ...
type IConn interface {
	OnOpen()
	OnClose()
	OnError(err *base.Error)
	OnReadReady() bool
	OnWriteReady() bool
	OnReadBytes(b []byte)
	OnFillWrite(b []byte) int

	SetNext(conn IConn)
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	GetFD() int
	Close()
}

// IReceiver ...
type IReceiver interface {
	OnConnOpen(streamConn *StreamConn)
	OnConnClose(streamConn *StreamConn)
	OnConnReadStream(streamConn *StreamConn, stream *core.Stream)
	OnConnError(streamConn *StreamConn, err *base.Error)
}
