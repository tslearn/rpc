package router

import (
	"github.com/rpccloud/rpc/internal"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
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

func (p *DirectRouter) Plug(
	receiver internal.IStreamReceiver,
) (internal.IStreamRouterSlot, *base.Error) {
	if p.receivers[0] == nil {
		p.receivers[0] = receiver
		return &DirectRouterSlot{receiverPtr: &p.receivers[1]}, nil
	} else if p.receivers[1] == nil {
		p.receivers[1] = receiver
		return &DirectRouterSlot{receiverPtr: &p.receivers[0]}, nil
	} else {
		return nil, errors.ErrDirectRouterConfigError
	}
}
