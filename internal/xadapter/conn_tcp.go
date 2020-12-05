package xadapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"net"
)

type TCPConn struct {
	channel *Channel
	fd      int
	next    XConn
}

func (p *TCPConn) NewTCPConn(fd int, next XConn) *TCPConn {
	return &TCPConn{
		fd:   fd,
		next: next,
	}
}

func (p *TCPConn) OnOpen() {
	p.next.OnOpen()
}

func (p *TCPConn) OnClose() {
	p.next.OnClose()
}

func (p *TCPConn) OnError(err *base.Error) {
	p.next.OnError(err)
}

func (p *TCPConn) OnReadBytes(b []byte) {
	p.next.OnReadBytes(b)
}

func (p *TCPConn) OnFillWrite(b []byte) int {
	return p.next.OnFillWrite(b)
}

func (p *TCPConn) TriggerWrite() {
	panic("not implement")
}

func (p *TCPConn) Close() {
	p.channel.CloseFD(p.fd)
}

func (p *TCPConn) LocalAddr() net.Addr {
	panic("not implement")
}

func (p *TCPConn) RemoteAddr() net.Addr {
	panic("not implement")
}

func (p *TCPConn) GetFD() int {
	return p.fd
}

func (p *TCPConn) GetRootConn() XConn {
	return p
}
