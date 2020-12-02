package adapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
)

type XReceiver interface {
	OnEventConnStream(eventConn *EventConn, stream *core.Stream)
	OnEventConnClose(eventConn *EventConn)
	OnEventConnError(eventConn *EventConn, err *base.Error)
}
