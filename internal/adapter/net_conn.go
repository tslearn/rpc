package adapter

import (
	"io"
	"net"
	"strings"
	"sync"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
)

// NetConn ...
type NetConn struct {
	isServer  bool
	isRunning bool
	conn      net.Conn
	next      IConn
	rBuf      []byte
	wBuf      []byte
	sync.Mutex
}

// NewNetConn ...
func NewNetConn(
	isServer bool,
	netConn net.Conn,
	rBufSize int,
	wBufSize int,
) *NetConn {
	return &NetConn{
		isServer:  isServer,
		isRunning: true,
		conn:      netConn,
		next:      nil,
		rBuf:      make([]byte, rBufSize),
		wBuf:      make([]byte, wBufSize),
	}
}

// SetNext ...
func (p *NetConn) SetNext(next IConn) {
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
		if e := p.conn.Close(); e != nil {
			p.OnError(errors.ErrTemp.AddDebug(e.Error()))
		}
	}
}

// LocalAddr ...
func (p *NetConn) LocalAddr() net.Addr {
	return p.conn.LocalAddr()
}

// RemoteAddr ...
func (p *NetConn) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}

// OnReadReady ...
func (p *NetConn) OnReadReady() bool {
	n, e := p.conn.Read(p.rBuf)
	if e != nil {
		if p.isServer {
			if e != io.EOF {
				p.OnError(errors.ErrTemp.AddDebug(e.Error()))
			}
		} else {
			p.Lock()
			ignoreReport := (!p.isRunning) &&
				strings.HasSuffix(e.Error(), ErrNetClosingSuffix)
			p.Unlock()

			if !ignoreReport {
				p.OnError(errors.ErrTemp.AddDebug(e.Error()))
			}
		}

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
			if n, e := p.conn.Write(p.wBuf[start:bufLen]); e != nil {
				p.OnError(errors.ErrTemp.AddDebug(e.Error()))
			} else {
				start += n
			}
		}
	}
}

// OnReadBytes ...
func (p *NetConn) OnReadBytes(_ []byte) {
	panic("kernel error, this code should not be called")
}

// OnFillWrite ...
func (p *NetConn) OnFillWrite(_ []byte) int {
	panic("kernel error, this code should not be called")
}

// GetFD ...
func (p *NetConn) GetFD() int {
	panic("kernel error, this code should not be called")
}

// SetFD ...
func (p *NetConn) SetFD(_ int) {
	panic("kernel error, this code should not be called")
}
