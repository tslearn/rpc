package xadapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"net"
)

type ChannelConn struct {
	status  uint32
	channel *Channel
	fd      int
	next    XConn
	la      net.Addr
	ra      net.Addr
}

func (p *ChannelConn) NewChannelConn(
	channel *Channel,
	fd int,
	la net.Addr,
	ra net.Addr,
) *ChannelConn {
	return &ChannelConn{
		status:  0,
		channel: channel,
		fd:      fd,
		next:    nil,
		la:      la,
		ra:      ra,
	}
}

func (p *ChannelConn) SetNext(next XConn) {
	p.next = next
}

func (p *ChannelConn) OnOpen() {
	p.next.OnOpen()
}

func (p *ChannelConn) OnClose() {
	p.next.OnClose()
}

func (p *ChannelConn) OnError(err *base.Error) {
	p.next.OnError(err)
}

func (p *ChannelConn) OnReadBytes(b []byte) {
	p.next.OnReadBytes(b)
}

func (p *ChannelConn) OnFillWrite(b []byte) int {
	return p.next.OnFillWrite(b)
}

func (p *ChannelConn) TriggerWrite() {
	panic("not implement")
}

func (p *ChannelConn) Close() {
	p.channel.CloseFD(p.fd)
}

func (p *ChannelConn) LocalAddr() net.Addr {
	return p.la
}

func (p *ChannelConn) RemoteAddr() net.Addr {
	return p.ra
}

func (p *ChannelConn) GetFD() int {
	return p.fd
}

func (p *ChannelConn) GetRootConn() XConn {
	return p
}
