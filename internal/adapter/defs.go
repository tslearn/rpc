package adapter

import "github.com/rpccloud/rpc/internal/base"

type IAdapter interface {
	Open(receiver XReceiver) *base.Error
	Close() *base.Error
}
