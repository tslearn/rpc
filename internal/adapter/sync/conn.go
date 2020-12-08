package sync

import (
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
)

type Conn struct {
	netConn net.Conn
	next    adapter.XConn
	rBuf    []byte
	wBuf    []byte
}

func NewConn(
	netConn net.Conn,
	rBufSize int,
	wBufSize int,
) *Conn {
	return &Conn{
		netConn: netConn,
		next:    nil,
		rBuf:    make([]byte, rBufSize),
		wBuf:    make([]byte, wBufSize),
	}
}

func (p *Conn) SetNext(next adapter.XConn) {
	p.next = next
}

func (p *Conn) TriggerRead() *base.Error {
	if n, e := p.netConn.Read(p.rBuf); e != nil {
		return errors.ErrTemp.AddDebug(e.Error())
	} else {
		p.OnReadBytes(p.rBuf[:n])
		return nil
	}
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
	if e := p.netConn.Close(); e != nil {
		p.OnError(errors.ErrTemp.AddDebug(e.Error()))
	}
}

func (p *Conn) LocalAddr() net.Addr {
	return p.netConn.LocalAddr()
}

func (p *Conn) RemoteAddr() net.Addr {
	return p.netConn.RemoteAddr()
}
