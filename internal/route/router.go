package route

import (
	"github.com/rpccloud/rpc/internal/core"
)

// IRouteReceiver ...
type IRouteReceiver interface {
	ReceiveStreamFromRouter(stream *core.Stream)
}

// IRouteSender ...
type IRouteSender interface {
	SendStreamToRouter(stream *core.Stream)
}

// IRouter ...
type IRouter interface {
	Plug(receiver IRouteReceiver) IRouteSender
}
