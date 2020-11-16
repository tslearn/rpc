package router

import (
	"github.com/rpccloud/rpc/internal"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
)

type DirectRouterSlot struct {
	receiverPtr *internal.IStreamReceiver
}

func (p *DirectRouterSlot) SendStream(stream *core.Stream) *base.Error {
	return (*p.receiverPtr).OnStream(stream)
}

type DirectRouter struct {
	receivers []internal.IStreamReceiver
}

func NewDirectRouter() internal.IStreamRouter {
	return &DirectRouter{
		receivers: make([]internal.IStreamReceiver, 2),
	}
}

func (p *DirectRouter) Plug(
	receiver internal.IStreamReceiver,
) internal.IStreamRouterSlot {
	if p.receivers[0] == nil {
		p.receivers[0] = receiver
		return &DirectRouterSlot{receiverPtr: &p.receivers[1]}
	} else if p.receivers[1] == nil {
		p.receivers[1] = receiver
		return &DirectRouterSlot{receiverPtr: &p.receivers[0]}
	} else {
		panic("DirectRouter can only be plugged twice")
	}
}
