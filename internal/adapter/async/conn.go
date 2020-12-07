package async

import (
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
)

type Conn struct {
	status   uint32
	fd       int
	next     adapter.XConn
	lAddr    net.Addr
	rAddr    net.Addr
	rBufSize int
	wBufSize int
}

func NewConn(
	fd int,
	lAddr net.Addr,
	rAddr net.Addr,
	rBufSize int,
	wBufSize int,
) *Conn {
	return &Conn{
		status:   0,
		fd:       fd,
		next:     nil,
		lAddr:    lAddr,
		rAddr:    rAddr,
		rBufSize: rBufSize,
		wBufSize: wBufSize,
	}
}

func (p *Conn) SetNext(next adapter.XConn) {
	p.next = next
}

func (p *Conn) OnReadOpen() {
	panic("not implement")
}

func (p *Conn) OnWriteOpen() {
	panic("not implement")
}

func (p *Conn) OnReadReady() {

}

func (p *Conn) OnWriteReady() {

}

func (p *Conn) OnReadClose() {

}

func (p *Conn) OnWriteClose() {

}

func (p *Conn) OnOpen() {
	p.next.OnOpen()
}

func (p *Conn) OnClose() {
	p.next.OnClose()
}

func (p *Conn) OnError(err *base.Error) {
	p.next.OnError(err)
}

func (p *Conn) OnReadBytes(b []byte) {
	p.next.OnReadBytes(b)
}

func (p *Conn) OnFillWrite(b []byte) int {
	return p.next.OnFillWrite(b)
}

func (p *Conn) TriggerWrite() {
	panic("not implement")
}

func (p *Conn) Close() {
	if e := closeFD(p.fd); e != nil {
		p.OnError(errors.ErrTemp.AddDebug(e.Error()))
	}
}

func (p *Conn) LocalAddr() net.Addr {
	return p.lAddr
}

func (p *Conn) RemoteAddr() net.Addr {
	return p.rAddr
}

func (p *Conn) GetFD() int {
	return p.fd
}
