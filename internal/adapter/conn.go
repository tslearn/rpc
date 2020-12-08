package adapter

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"net"
)

type XConn interface {
	OnOpen()
	OnClose()
	OnError(err *base.Error)
	OnReadBytes(b []byte)
	OnFillWrite(b []byte) int

	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	TriggerWrite()
	Close()
}

type StreamConn struct {
	prev     XConn
	receiver IReceiver
}

func NewStreamConn(prev XConn, receiver IReceiver) *StreamConn {
	return &StreamConn{
		prev:     prev,
		receiver: receiver,
	}
}

func (p *StreamConn) OnOpen() {
	p.receiver.OnConnOpen(p)
}

func (p *StreamConn) OnClose() {
	p.receiver.OnConnClose(p)
}

func (p *StreamConn) OnError(err *base.Error) {
	p.receiver.OnConnError(p, err)
}

func (p *StreamConn) OnReadBytes(b []byte) {
	panic("not implement")
}

func (p *StreamConn) OnFillWrite(b []byte) int {
	panic("not implement")
}

func (p *StreamConn) TriggerWrite() {
	p.prev.TriggerWrite()
}

func (p *StreamConn) Close() {
	fmt.Println(p.prev)
	p.prev.Close()
}

func (p *StreamConn) LocalAddr() net.Addr {
	return p.prev.LocalAddr()
}

func (p *StreamConn) RemoteAddr() net.Addr {
	return p.prev.RemoteAddr()
}

func (p *StreamConn) WriteStream(stream *core.Stream) {
	panic("not implement")
}
