package async

import (
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
)

type AsyncConn struct {
	status   uint32
	fd       int
	next     adapter.XConn
	lAddr    net.Addr
	rAddr    net.Addr
	rBufSize int
	wBufSize int
}

func NewAsyncConn(
	fd int,
	lAddr net.Addr,
	rAddr net.Addr,
	rBufSize int,
	wBufSize int,
) *AsyncConn {
	return &AsyncConn{
		status:   0,
		fd:       fd,
		next:     nil,
		lAddr:    lAddr,
		rAddr:    rAddr,
		rBufSize: rBufSize,
		wBufSize: wBufSize,
	}
}

func (p *AsyncConn) SetNext(next adapter.XConn) {
	p.next = next
}

func (p *AsyncConn) OnReadOpen() {
	panic("not implement")
}

func (p *AsyncConn) OnWriteOpen() {
	panic("not implement")
}

func (p *AsyncConn) OnReadReady() {

}

func (p *AsyncConn) OnWriteReady() {

}

func (p *AsyncConn) OnReadClose() {

}

func (p *AsyncConn) OnWriteClose() {

}

func (p *AsyncConn) OnOpen() {
	p.next.OnOpen()
}

func (p *AsyncConn) OnClose() {
	p.next.OnClose()
}

func (p *AsyncConn) OnError(err *base.Error) {
	p.next.OnError(err)
}

func (p *AsyncConn) OnReadBytes(b []byte) {
	p.next.OnReadBytes(b)
}

func (p *AsyncConn) OnFillWrite(b []byte) int {
	return p.next.OnFillWrite(b)
}

func (p *AsyncConn) TriggerWrite() {
	panic("not implement")
}

func (p *AsyncConn) Close() {
	if e := closeFD(p.fd); e != nil {
		p.OnError(errors.ErrTemp.AddDebug(e.Error()))
	}
}

func (p *AsyncConn) LocalAddr() net.Addr {
	return p.lAddr
}

func (p *AsyncConn) RemoteAddr() net.Addr {
	return p.rAddr
}

func (p *AsyncConn) GetFD() int {
	return p.fd
}
