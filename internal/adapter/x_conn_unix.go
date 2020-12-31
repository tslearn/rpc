// +build linux darwin

package adapter

import (
	"net"
	"sync"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"golang.org/x/sys/unix"
)

// XConn ...
type XConn struct {
	isRunning  bool
	channel    *Channel
	fd         int
	next       IConn
	lAddr      net.Addr
	rAddr      net.Addr
	rBuf       []byte
	wBuf       []byte
	wStartPos  int
	wEndPos    int
	watchWrite bool
	sync.Mutex
}

// NewXConn ...
func NewXConn(
	channel *Channel,
	fd int,
	lAddr net.Addr,
	rAddr net.Addr,
	rBufSize int,
	wBufSize int,
) *XConn {
	return &XConn{
		isRunning:  true,
		channel:    channel,
		fd:         fd,
		next:       nil,
		lAddr:      lAddr,
		rAddr:      rAddr,
		rBuf:       make([]byte, rBufSize),
		wBuf:       make([]byte, wBufSize),
		wStartPos:  0,
		wEndPos:    0,
		watchWrite: false,
	}
}

// SetNext ...
func (p *XConn) SetNext(next IConn) {
	p.next = next
}

func (p *XConn) setWatchWrite(isWatch bool) *base.Error {
	if p.watchWrite != isWatch {
		if e := p.channel.SetWriteFD(p.fd, isWatch); e != nil {
			return errors.ErrTemp.AddDebug(e.Error())
		}
		p.watchWrite = isWatch
	}

	return nil
}

// OnReadReady ...
func (p *XConn) OnReadReady() bool {
	if n, e := unix.Read(p.fd, p.rBuf); e != nil {
		if e != unix.EWOULDBLOCK && e != unix.EAGAIN {
			p.OnError(errors.ErrTemp.AddDebug(e.Error()))
			return false
		}
	} else {
		p.next.OnReadBytes(p.rBuf[:n])
	}

	return true
}

func (p *XConn) doWrite() bool {
	isFillFinish := false

	// fill buffer
	for p.wEndPos < len(p.wBuf) {
		if n := p.next.OnFillWrite(p.wBuf[p.wEndPos:]); n > 0 {
			p.wEndPos += n
		} else {
			isFillFinish = true
			break
		}
	}

	// write buffer
	if p.wStartPos < p.wEndPos {
		if n, e := unix.Write(p.fd, p.wBuf[p.wStartPos:p.wEndPos]); e == nil {
			p.wStartPos += n
		} else {
			if e != unix.EWOULDBLOCK && e != unix.EAGAIN {
				p.OnError(errors.ErrTemp.AddDebug(e.Error()))
				return true
			}
			return false
		}
	}

	if p.wStartPos == p.wEndPos {
		p.wStartPos = 0
		p.wEndPos = 0
		return isFillFinish
	}

	return false
}

// OnWriteReady ...
func (p *XConn) OnWriteReady() {
	p.Lock()
	defer p.Unlock()

	if err := p.setWatchWrite(!p.doWrite()); err != nil {
		p.OnError(err)
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

// Close ...
func (p *XConn) Close() {
	p.Lock()
	defer p.Unlock()

	if p.isRunning {
		if e := p.channel.CloseFD(p.fd); e != nil {
			p.OnError(errors.ErrTemp.AddDebug(e.Error()))
		}
		p.isRunning = false
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

// OnReadBytes ...
func (p *XConn) OnReadBytes(b []byte) {
	panic("kernel error, this code should not be called")
}

// OnFillWrite ...
func (p *XConn) OnFillWrite(b []byte) int {
	panic("kernel error, this code should not be called")
}
