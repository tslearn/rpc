package adapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
)

type XReceiver interface {
	OnEventConnOpen(eventConn *EventConn)
	OnEventConnClose(eventConn *EventConn)
	OnEventConnStream(eventConn *EventConn, stream *core.Stream)
	OnEventConnError(eventConn *EventConn, err *base.Error)
}

type ReceiverHook struct {
	receiver XReceiver

	onEventConnOpen   func(eventConn *EventConn)
	onEventConnClose  func(eventConn *EventConn)
	onEventConnStream func(eventConn *EventConn, stream *core.Stream)
	onEventConnError  func(eventConn *EventConn, err *base.Error)
}

func NewReceiverHook(
	receiver XReceiver,
	onEventConnOpen func(eventConn *EventConn),
	onEventConnClose func(eventConn *EventConn),
	onEventConnStream func(eventConn *EventConn, stream *core.Stream),
	onEventConnError func(eventConn *EventConn, err *base.Error),
) *ReceiverHook {
	return &ReceiverHook{
		receiver:          receiver,
		onEventConnOpen:   onEventConnOpen,
		onEventConnStream: onEventConnStream,
		onEventConnClose:  onEventConnClose,
		onEventConnError:  onEventConnError,
	}
}

func (p *ReceiverHook) OnEventConnOpen(eventConn *EventConn) {
	if fn := p.onEventConnOpen; fn != nil {
		fn(eventConn)
	}

	p.receiver.OnEventConnOpen(eventConn)
}

func (p *ReceiverHook) OnEventConnClose(eventConn *EventConn) {
	if fn := p.onEventConnClose; fn != nil {
		fn(eventConn)
	}

	p.receiver.OnEventConnClose(eventConn)
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

func (p *ReceiverHook) OnEventConnError(
	eventConn *EventConn,
	err *base.Error,
) {
	if fn := p.onEventConnError; fn != nil {
		fn(eventConn, err)
	}

	p.receiver.OnEventConnError(eventConn, err)
}
