package xserver

import (
	"github.com/rpccloud/rpc/internal/base"
)

type IStreamAsyncConn interface {
	Close() *base.Error
}
