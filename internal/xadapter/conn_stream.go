package xadapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"net"
)

type StreamConn struct {
	prev XConn
}

func NewStreamConn(prev XConn) *StreamConn {
	return &StreamConn{
		prev: prev,
	}
}

func (p *StreamConn) OnOpen() {
	panic("not implement")
}

func (p *StreamConn) OnClose() {
	panic("not implement")
}

func (p *StreamConn) OnError(err *base.Error) {
	panic("not implement")
}

func (p *StreamConn) OnReadBytes(b []byte) {
	panic("not implement")
}

func (p *StreamConn) OnFillWrite(b []byte) int {
	panic("not implement")
}

func (p *StreamConn) TriggerWrite() {
	p.GetRootConn().TriggerWrite()
}

func (p *StreamConn) Close() {
	p.GetRootConn().Close()
}

func (p *StreamConn) LocalAddr() net.Addr {
	return p.GetRootConn().LocalAddr()
}

func (p *StreamConn) RemoteAddr() net.Addr {
	return p.GetRootConn().RemoteAddr()
}

func (p *StreamConn) GetFD() int {
	return p.GetRootConn().GetFD()
}

func (p *StreamConn) GetRootConn() XConn {
	return p.prev.GetRootConn()
}
