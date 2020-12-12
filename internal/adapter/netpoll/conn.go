package netpoll

import (
	"github.com/rpccloud/rpc/internal/base"
	"net"
)

type Conn interface {
	OnOpen()
	OnClose()
	OnError(err *base.Error)
	OnReadReady()
	OnWriteReady()
	OnReadBytes(b []byte)
	OnFillWrite(b []byte) int

	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	GetFD() int
	TriggerWrite()
	Close()
}
