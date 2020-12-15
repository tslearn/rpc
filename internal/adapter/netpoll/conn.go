package netpoll

import (
	"net"

	"github.com/rpccloud/rpc/internal/base"
)

// Conn ...
type Conn interface {
	OnOpen()
	OnClose()
	OnError(err *base.Error)
	OnReadReady() bool
	OnWriteReady()
	OnReadBytes(b []byte)
	OnFillWrite(b []byte) int

	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	GetFD() int
	Close()
}
