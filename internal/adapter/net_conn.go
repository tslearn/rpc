package adapter

import (
	"net"
	"sync"

	"github.com/rpccloud/rpc/internal/adapter/netpoll"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
)

// NetConn ...
type NetConn struct {
	isRunning bool
	fd        int
	netConn   net.Conn
	next      netpoll.Conn
	rBuf      []byte
	wBuf      []byte
	sync.Mutex
}

// NewNetConn ...
func NewNetConn(
	netConn net.Conn,
	rBufSize int,
	wBufSize int,
) *NetConn {
	return &NetConn{
		isRunning: true,
		fd:        0,
		netConn:   netConn,
		next:      nil,
		rBuf:      make([]byte, rBufSize),
		wBuf:      make([]byte, wBufSize),
	}
}

// SetNext ...
func (p *NetConn) SetNext(next netpoll.Conn) {
	p.next = next
}

// OnOpen ...
func (p *NetConn) OnOpen() {
	p.next.OnOpen()
}

// OnClose ...
func (p *NetConn) OnClose() {
	p.next.OnClose()
}

// OnError ...
func (p *NetConn) OnError(err *base.Error) {
	p.next.OnError(err)
}

// Close ...
func (p *NetConn) Close() {
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
func (p *NetConn) LocalAddr() net.Addr {
	return p.netConn.LocalAddr()
}

// RemoteAddr ...
func (p *NetConn) RemoteAddr() net.Addr {
	return p.netConn.RemoteAddr()
}

// OnReadReady ...
func (p *NetConn) OnReadReady() bool {
	n, e := p.netConn.Read(p.rBuf)
	if e != nil {
		p.OnError(errors.ErrTemp.AddDebug(e.Error()))
		return false
	}
	p.next.OnReadBytes(p.rBuf[:n])
	return true
}

// OnWriteReady ...
func (p *NetConn) OnWriteReady() {
	p.Lock()
	defer p.Unlock()

	isTriggerFinish := false

	for !isTriggerFinish {
		bufLen := 0

		for !isTriggerFinish && bufLen < len(p.wBuf) {
			if n := p.next.OnFillWrite(p.wBuf[bufLen:]); n > 0 {
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

// OnReadBytes ...
func (p *NetConn) OnReadBytes(b []byte) {
	panic("kernel error, this code should not be called")
}

// OnFillWrite ...
func (p *NetConn) OnFillWrite(b []byte) int {
	panic("kernel error, this code should not be called")
}

// GetFD ...
func (p *NetConn) GetFD() int {
	return p.fd
}

// SetFD ...
func (p *NetConn) SetFD(fd int) {
	p.fd = fd
}
