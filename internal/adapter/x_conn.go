package adapter

import (
	"net"
	"sync"

	"github.com/rpccloud/rpc/internal/adapter/netpoll"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"golang.org/x/sys/unix"
)

// XConn ...
type XConn struct {
	channel       *netpoll.Channel
	fd            int
	next          netpoll.Conn
	lAddr         net.Addr
	rAddr         net.Addr
	rBuf          []byte
	wBuf          []byte
	wStartPos     int
	wEndPos       int
	canWriteReady bool
	sync.Mutex
}

// NewXConn ...
func NewXConn(
	channel *netpoll.Channel,
	fd int,
	lAddr net.Addr,
	rAddr net.Addr,
	rBufSize int,
	wBufSize int,
) *XConn {
	return &XConn{
		channel:       channel,
		fd:            fd,
		next:          nil,
		lAddr:         lAddr,
		rAddr:         rAddr,
		rBuf:          make([]byte, rBufSize),
		wBuf:          make([]byte, wBufSize),
		wStartPos:     0,
		wEndPos:       0,
		canWriteReady: false,
	}
}

// SetNext ...
func (p *XConn) SetNext(next netpoll.Conn) {
	p.next = next
}

// OnReadReady ...
func (p *XConn) OnReadReady() {
	for {
		if n, e := netpoll.ReadFD(p.fd, p.rBuf); e == nil {
			p.OnReadBytes(p.rBuf[:n])
		} else {
			if e != unix.EWOULDBLOCK && e != unix.EAGAIN {
				p.OnError(errors.ErrTemp.AddDebug(e.Error()))
			}
			return
		}
	}
}

// DoWrite ...
func (p *XConn) DoWrite() bool {
	for {
		isFinish := false

		// fill buffer
		if p.wEndPos == 0 {
			for p.wEndPos < len(p.wBuf) {
				if n := p.OnFillWrite(p.wBuf[p.wEndPos:]); n > 0 {
					p.wEndPos += n
				} else {
					isFinish = true
					break
				}
			}
		}

		// write buffer
		for p.wStartPos < p.wEndPos {
			if n, e := netpoll.WriteFD(p.fd, p.wBuf[p.wStartPos:p.wEndPos]); e == nil {
				p.wStartPos += n
			} else {
				if e != unix.EWOULDBLOCK && e != unix.EAGAIN {
					p.OnError(errors.ErrTemp.AddDebug(e.Error()))
					return true
				}
				return false
			}
		}

		p.wStartPos = 0
		p.wEndPos = 0

		if isFinish {
			return true
		}
	}
}

// OnWriteReady ...
func (p *XConn) OnWriteReady() {
	p.Lock()
	defer p.Unlock()

	if p.canWriteReady {
		p.canWriteReady = !p.DoWrite()
	}
}

// OnOpen ...
func (p *XConn) OnOpen() {
	p.next.OnOpen()
}

// OnClose ...
func (p *XConn) OnClose() {
	p.next.OnClose()
}

// OnError ...
func (p *XConn) OnError(err *base.Error) {
	p.next.OnError(err)
}

// OnReadBytes ...
func (p *XConn) OnReadBytes(b []byte) {
	p.next.OnReadBytes(b)
}

// OnFillWrite ...
func (p *XConn) OnFillWrite(b []byte) int {
	return p.next.OnFillWrite(b)
}

// TriggerWrite ...
func (p *XConn) TriggerWrite() {
	p.Lock()
	defer p.Unlock()

	p.canWriteReady = !p.DoWrite()
}

// Close ...
func (p *XConn) Close() {
	if e := netpoll.CloseFD(p.fd); e != nil {
		p.OnError(errors.ErrTemp.AddDebug(e.Error()))
	}
}

// LocalAddr ...
func (p *XConn) LocalAddr() net.Addr {
	return p.lAddr
}

// RemoteAddr ...
func (p *XConn) RemoteAddr() net.Addr {
	return p.rAddr
}

// GetFD ...
func (p *XConn) GetFD() int {
	return p.fd
}
