package xadapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"net"
)

type ChannelConn struct {
	status   uint32
	channel  *Channel
	fd       int
	next     XConn
	lAddr    net.Addr
	rAddr    net.Addr
	rBufSize int
	wBufSize int
}

func NewChannelConn(
	channel *Channel,
	fd int,
	lAddr net.Addr,
	rAddr net.Addr,
	rBufSize int,
	wBufSize int,
) *ChannelConn {
	return &ChannelConn{
		status:   0,
		channel:  channel,
		fd:       fd,
		next:     nil,
		lAddr:    lAddr,
		rAddr:    rAddr,
		rBufSize: rBufSize,
		wBufSize: wBufSize,
	}
}

func (p *ChannelConn) SetNext(next XConn) {
	p.next = next
}

func (p *ChannelConn) OnReadOpen() {
	panic("not implement")
}

func (p *ChannelConn) OnWriteOpen() {
	panic("not implement")
}

func (p *ChannelConn) OnReadReady() {

}

func (p *ChannelConn) OnWriteReady() {

}

func (p *ChannelConn) OnReadClose() {

}

func (p *ChannelConn) OnWriteClose() {

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
	return p.lAddr
}

func (p *ChannelConn) RemoteAddr() net.Addr {
	return p.rAddr
}

func (p *ChannelConn) GetFD() int {
	return p.fd
}
