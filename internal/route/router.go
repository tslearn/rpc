package route

import (
	"github.com/rpccloud/rpc/internal/rpc"
)

// IRouteReceiver ...
type IRouteReceiver interface {
	ReceiveStreamFromRouter(stream *rpc.Stream)
}

// IRouteSender ...
type IRouteSender interface {
	SendStreamToRouter(stream *rpc.Stream)
}

// IRouter ...
type IRouter interface {
	Plug(receiver IRouteReceiver) IRouteSender
}
