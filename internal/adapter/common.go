package adapter

import (
	"net"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
)

// ErrNetClosingSuffix ...
const ErrNetClosingSuffix = "use of closed network connection"

// IConn ...
type IConn interface {
	OnOpen()
	OnClose()
	OnError(err *base.Error)
	OnReadReady() bool
	OnWriteReady() bool
	OnReadBytes(b []byte)
	OnFillWrite(b []byte) int

	SetNext(conn IConn)
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	GetFD() int
	Close()
}

// IReceiver ...
type IReceiver interface {
	OnConnOpen(streamConn *StreamConn)
	OnConnClose(streamConn *StreamConn)
	OnConnReadStream(streamConn *StreamConn, stream *core.Stream)
	OnConnError(streamConn *StreamConn, err *base.Error)
}

//
//// ReceiverHook ...
//type ReceiverHook struct {
//	receiver         IReceiver
//	onConnOpen       func(streamConn *StreamConn)
//	onConnClose      func(streamConn *StreamConn)
//	onConnReadStream func(streamConn *StreamConn, stream *core.Stream)
//	onConnError      func(streamConn *StreamConn, err *base.Error)
//}
//
//// NewReceiverHook ...
//func NewReceiverHook(
//	receiver IReceiver,
//	onConnOpen func(streamConn *StreamConn),
//	onConnClose func(streamConn *StreamConn),
//	onConnReadStream func(streamConn *StreamConn, stream *core.Stream),
//	onConnError func(streamConn *StreamConn, err *base.Error),
//) *ReceiverHook {
//	return &ReceiverHook{
//		receiver:         receiver,
//		onConnOpen:       onConnOpen,
//		onConnReadStream: onConnReadStream,
//		onConnClose:      onConnClose,
//		onConnError:      onConnError,
//	}
//}
//
//// OnConnOpen ...
//func (p *ReceiverHook) OnConnOpen(streamConn *StreamConn) {
//	if fn := p.onConnOpen; fn != nil {
//		fn(streamConn)
//	}
//
//	p.receiver.OnConnOpen(streamConn)
//}
//
//// OnConnClose ...
//func (p *ReceiverHook) OnConnClose(streamConn *StreamConn) {
//	if fn := p.onConnClose; fn != nil {
//		fn(streamConn)
//	}
//
//	p.receiver.OnConnClose(streamConn)
//}
//
//// OnConnReadStream ...
//func (p *ReceiverHook) OnConnReadStream(
//	streamConn *StreamConn,
//	stream *core.Stream,
//) {
//	if fn := p.onConnReadStream; fn != nil {
//		fn(streamConn, stream)
//	}
//
//	p.receiver.OnConnReadStream(streamConn, stream)
//}
//
//// OnConnError ...
//func (p *ReceiverHook) OnConnError(
//	streamConn *StreamConn,
//	err *base.Error,
//) {
//	if fn := p.onConnError; fn != nil {
//		fn(streamConn, err)
//	}
//
//	p.receiver.OnConnError(streamConn, err)
//}
