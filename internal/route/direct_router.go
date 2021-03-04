package route

import (
	"github.com/rpccloud/rpc/internal/rpc"
)

// DirectRouterSender ...
type DirectRouterSender struct {
	receiver *IRouteReceiver
}

// SendStreamToRouter ...
func (p *DirectRouterSender) SendStreamToRouter(stream *rpc.Stream) {
	(*p.receiver).ReceiveStreamFromRouter(stream)
}

// DirectRouter ...
type DirectRouter struct {
	receivers [2]IRouteReceiver
}

// NewDirectRouter ...
func NewDirectRouter() *DirectRouter {
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
