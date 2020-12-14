package adapter

import (
	"net"
	"sync"

	"github.com/rpccloud/rpc/internal/adapter/netpoll"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
)

// SyncConn ...
type SyncConn struct {
	isRunning bool
	netConn   net.Conn
	next      netpoll.Conn
	rBuf      []byte
	wBuf      []byte
	sync.Mutex
}

// NewSyncConn ...
func NewSyncConn(
	netConn net.Conn,
	rBufSize int,
	wBufSize int,
) *SyncConn {
	return &SyncConn{
		isRunning: true,
		netConn:   netConn,
		next:      nil,
		rBuf:      make([]byte, rBufSize),
		wBuf:      make([]byte, wBufSize),
	}
}

// SetNext ...
func (p *SyncConn) SetNext(next netpoll.Conn) {
	p.next = next
}

// OnOpen ...
func (p *SyncConn) OnOpen() {
	p.next.OnOpen()
}

// OnClose ...
func (p *SyncConn) OnClose() {
	p.next.OnClose()
}

// OnError ...
func (p *SyncConn) OnError(err *base.Error) {
	p.next.OnError(err)
}

// OnReadBytes ...
func (p *SyncConn) OnReadBytes(b []byte) {
	p.next.OnReadBytes(b)
}

// OnFillWrite ...
func (p *SyncConn) OnFillWrite(b []byte) int {
	return p.next.OnFillWrite(b)
}

// TriggerWrite ...
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

// Close ...
func (p *SyncConn) Close() {
	p.Lock()
	defer p.Unlock()

	if p.isRunning {
		p.isRunning = false
		if e := p.netConn.Close(); e != nil {
			p.OnError(errors.ErrTemp.AddDebug(e.Error()))
		}
	}
}

// LocalAddr ...
func (p *SyncConn) LocalAddr() net.Addr {
	return p.netConn.LocalAddr()
}

// RemoteAddr ...
func (p *SyncConn) RemoteAddr() net.Addr {
	return p.netConn.RemoteAddr()
}

// OnReadReady ...
func (p *SyncConn) OnReadReady() bool {
	n, e := p.netConn.Read(p.rBuf)
	if e != nil {
		p.OnError(errors.ErrTemp.AddDebug(e.Error()))
		return false
	}
	p.OnReadBytes(p.rBuf[:n])
	return true
}

// OnWriteReady ...
func (p *SyncConn) OnWriteReady() bool {
	panic("kernel error, this code should not be called")
}

// GetFD ...
func (p *SyncConn) GetFD() int {
	panic("kernel error, this code should not be called")
}
