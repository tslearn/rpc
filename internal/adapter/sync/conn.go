package sync

import (
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
	"sync"
)

type Conn struct {
	netConn net.Conn
	next    adapter.XConn
	rBuf    []byte
	wBuf    []byte

	sync.Mutex
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
	p.Lock()
	defer p.Unlock()

	isTriggerFinish := false

	for !isTriggerFinish {
		bufLen := 0

		for !isTriggerFinish && bufLen < len(p.wBuf) {
			if n := p.OnFillWrite(p.wBuf[bufLen:]); n > 0 {
				bufLen += n
			} else {
				isTriggerFinish = true
			}
		}

		start := 0
		for start < bufLen {
			if n, e := p.netConn.Write(p.wBuf[start:bufLen]); e != nil {
				p.OnError(errors.ErrTemp.AddDebug(e.Error()))
			} else {
				start += n
			}
		}
	}
}

func (p *Conn) Close() {
	if e := p.netConn.Close(); e != nil {
		p.OnError(errors.ErrTemp.AddDebug(e.Error()))
	}

	p.OnClose()
}

func (p *Conn) LocalAddr() net.Addr {
	return p.netConn.LocalAddr()
}

func (p *Conn) RemoteAddr() net.Addr {
	return p.netConn.RemoteAddr()
}
