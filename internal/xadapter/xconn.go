package xadapter

import "github.com/rpccloud/rpc/internal/base"

type XConn interface {
	FD() int
	OnRead() *base.Error
	OnClose()
	OnError(err *base.Error)
}
