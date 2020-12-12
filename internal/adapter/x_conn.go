package adapter

import (
	"github.com/rpccloud/rpc/internal/adapter/netpoll"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"golang.org/x/sys/unix"
	"net"
	"sync"
)

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

func (p *XConn) SetNext(next netpoll.Conn) {
	p.next = next
}

func (p *XConn) OnReadReady() {
	for {
		if n, e := netpoll.ReadFD(p.fd, p.rBuf); e != nil {
			if e != unix.EWOULDBLOCK && e != unix.EAGAIN {
				p.OnError(errors.ErrTemp.AddDebug(e.Error()))
			}
			return
		} else {
			p.OnReadBytes(p.rBuf[:n])
		}
	}
}

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
			if n, e := netpoll.WriteFD(p.fd, p.wBuf[p.wStartPos:p.wEndPos]); e != nil {
				if e != unix.EWOULDBLOCK && e != unix.EAGAIN {
					p.OnError(errors.ErrTemp.AddDebug(e.Error()))
					return true
				} else {
					return false
				}
			} else {
				p.wStartPos += n
			}
		}

		p.wStartPos = 0
		p.wEndPos = 0

		if isFinish {
			return true
		}
	}
}

func (p *XConn) OnWriteReady() {
	p.Lock()
	defer p.Unlock()

	if p.canWriteReady {
		p.canWriteReady = !p.DoWrite()
	}
}

func (p *XConn) OnOpen() {
	p.next.OnOpen()
}

func (p *XConn) OnClose() {
	p.next.OnClose()
}

func (p *XConn) OnError(err *base.Error) {
	p.next.OnError(err)
}

func (p *XConn) OnReadBytes(b []byte) {
	p.next.OnReadBytes(b)
}

func (p *XConn) OnFillWrite(b []byte) int {
	return p.next.OnFillWrite(b)
}

func (p *XConn) TriggerWrite() {
	p.Lock()
	defer p.Unlock()

	p.canWriteReady = !p.DoWrite()
}

func (p *XConn) Close() {
	if e := netpoll.CloseFD(p.fd); e != nil {
		p.OnError(errors.ErrTemp.AddDebug(e.Error()))
	}
}

func (p *XConn) LocalAddr() net.Addr {
	return p.lAddr
}

func (p *XConn) RemoteAddr() net.Addr {
	return p.rAddr
}

func (p *XConn) GetFD() int {
	return p.fd
}
