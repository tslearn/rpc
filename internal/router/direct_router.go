package router

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
)

// DirectRouterSender ...
type DirectRouterSender struct {
	receiver *IRouteReceiver
}

// SendStreamToRouter ...
func (p *DirectRouterSender) SendStreamToRouter(
	stream *core.Stream,
) *base.Error {
	return (*p.receiver).ReceiveStreamFromRouter(stream)
}

// DirectRouter ...
type DirectRouter struct {
	receivers [2]IRouteReceiver
}

// NewDirectRouter ...
func NewDirectRouter() IRouter {
	return &DirectRouter{}
}

// Plug ...
func (p *DirectRouter) Plug(receiver IRouteReceiver) IRouteSender {
	if p.receivers[0] == nil {
		p.receivers[0] = receiver
		return &DirectRouterSender{receiver: &p.receivers[1]}
	} else if p.receivers[1] == nil {
		p.receivers[1] = receiver
		return &DirectRouterSender{receiver: &p.receivers[0]}
	} else {
		panic("DirectRouter can only be plugged twice")
	}
}
