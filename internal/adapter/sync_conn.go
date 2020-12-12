package adapter

import (
	"github.com/rpccloud/rpc/internal/adapter/netpoll"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
	"sync"
)

type SyncConn struct {
	netConn net.Conn
	next    netpoll.Conn
	rBuf    []byte
	wBuf    []byte

	sync.Mutex
}

func NewSyncConn(
	netConn net.Conn,
	rBufSize int,
	wBufSize int,
) *SyncConn {
	return &SyncConn{
		netConn: netConn,
		next:    nil,
		rBuf:    make([]byte, rBufSize),
		wBuf:    make([]byte, wBufSize),
	}
}

func (p *SyncConn) SetNext(next netpoll.Conn) {
	p.next = next
}

func (p *SyncConn) TriggerRead() *base.Error {
	if n, e := p.netConn.Read(p.rBuf); e != nil {
		return errors.ErrTemp.AddDebug(e.Error())
	} else {
		p.OnReadBytes(p.rBuf[:n])
		return nil
	}
}

func (p *SyncConn) OnOpen() {
	p.next.OnOpen()
}

func (p *SyncConn) OnClose() {
	p.next.OnClose()
}

func (p *SyncConn) OnError(err *base.Error) {
	p.next.OnError(err)
}

func (p *SyncConn) OnReadBytes(b []byte) {
	p.next.OnReadBytes(b)
}

func (p *SyncConn) OnFillWrite(b []byte) int {
	return p.next.OnFillWrite(b)
}

func (p *SyncConn) TriggerWrite() {
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

func (p *SyncConn) Close() {
	if e := p.netConn.Close(); e != nil {
		p.OnError(errors.ErrTemp.AddDebug(e.Error()))
	}

	p.OnClose()
}

func (p *SyncConn) LocalAddr() net.Addr {
	return p.netConn.LocalAddr()
}

func (p *SyncConn) RemoteAddr() net.Addr {
	return p.netConn.RemoteAddr()
}

func (p *SyncConn) OnReadReady() {
	panic("kernel error, this code should not be called")
}

func (p *SyncConn) OnWriteReady() {
	panic("kernel error, this code should not be called")
}

func (p *SyncConn) GetFD() int {
	panic("kernel error, this code should not be called")
}
