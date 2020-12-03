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

type ReceiverHook struct {
	receiver XReceiver

	onEventConnStream func(eventConn *EventConn, stream *core.Stream)
	onEventConnClose  func(eventConn *EventConn)
	onEventConnError  func(eventConn *EventConn, err *base.Error)
}

func NewReceiverHook(
	receiver XReceiver,
	onEventConnStream func(eventConn *EventConn, stream *core.Stream),
	onEventConnClose func(eventConn *EventConn),
	onEventConnError func(eventConn *EventConn, err *base.Error),
) *ReceiverHook {
	return &ReceiverHook{
		receiver:          receiver,
		onEventConnStream: onEventConnStream,
		onEventConnClose:  onEventConnClose,
		onEventConnError:  onEventConnError,
	}
}

func (p *ReceiverHook) OnEventConnStream(
	eventConn *EventConn,
	stream *core.Stream,
) {
	if fn := p.onEventConnStream; fn != nil {
		fn(eventConn, stream)
	}

	p.receiver.OnEventConnStream(eventConn, stream)
}

func (p *ReceiverHook) OnEventConnClose(eventConn *EventConn) {
	if fn := p.onEventConnClose; fn != nil {
		fn(eventConn)
	}

	p.receiver.OnEventConnClose(eventConn)
}

func (p *ReceiverHook) OnEventConnError(
	eventConn *EventConn,
	err *base.Error,
) {
	if fn := p.onEventConnError; fn != nil {
		fn(eventConn, err)
	}

	p.receiver.OnEventConnError(eventConn, err)
}
