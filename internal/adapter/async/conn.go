package async

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
	"sync/atomic"
)

type Conn struct {
	openStatus  int32
	closeStatus int32
	fd          int
	next        adapter.XConn
	lAddr       net.Addr
	rAddr       net.Addr
	rBuf        []byte
	wBuf        []byte
	wStartPos   int
	wEndPos     int
	canWrite    bool
}

func NewConn(
	fd int,
	lAddr net.Addr,
	rAddr net.Addr,
	rBufSize int,
	wBufSize int,
) *Conn {
	return &Conn{
		openStatus:  0,
		closeStatus: 0,
		fd:          fd,
		next:        nil,
		lAddr:       lAddr,
		rAddr:       rAddr,
		rBuf:        make([]byte, rBufSize),
		wBuf:        make([]byte, wBufSize),
		wStartPos:   0,
		wEndPos:     0,
		canWrite:    false,
	}
}

func (p *Conn) SetNext(next adapter.XConn) {
	p.next = next
}

func (p *Conn) OnReadOpen() {
	if atomic.AddInt32(&p.openStatus, 1) == 2 {
		p.OnOpen()
	}
}

func (p *Conn) OnWriteOpen() {
	if atomic.AddInt32(&p.openStatus, 1) == 2 {
		p.OnOpen()
	}
}

func (p *Conn) OnReadReady() {
	if n, e := readFD(p.fd, p.rBuf); e != nil {
		p.OnError(errors.ErrTemp.AddDebug(e.Error()))
	} else {
		p.OnReadBytes(p.rBuf[:n])
	}
}

func (p *Conn) OnWriteReady() {
	if p.wEndPos < len(p.wBuf) {
		if n := p.OnFillWrite(p.wBuf[p.wEndPos:]); n > 0 {
			p.wEndPos += n
		}
	}

	if p.wEndPos == 0 {
		p.canWrite = true
	} else {
		p.canWrite = false
		if n, e := writeFD(p.fd, p.wBuf[p.wStartPos:p.wEndPos]); e != nil {
			p.OnError(errors.ErrTemp.AddDebug(e.Error()))
		} else {
			p.wStartPos += n
		}

		if p.wStartPos == p.wEndPos {
			p.wStartPos = 0
			p.wEndPos = 0
		}
	}
}

func (p *Conn) OnReadClose() {
	if atomic.AddInt32(&p.closeStatus, 1) == 2 {
		p.OnClose()
	}
}

func (p *Conn) OnWriteClose() {
	if atomic.AddInt32(&p.closeStatus, 1) == 2 {
		p.OnClose()
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
	fmt.Println("TriggerWrite A")
	if p.canWrite {
		fmt.Println("TriggerWrite B")
		p.OnWriteReady()
	}
}

func (p *Conn) Close() {
	if e := closeFD(p.fd); e != nil {
		p.OnError(errors.ErrTemp.AddDebug(e.Error()))
	}
}

func (p *Conn) LocalAddr() net.Addr {
	return p.lAddr
}

func (p *Conn) RemoteAddr() net.Addr {
	return p.rAddr
}

func (p *Conn) GetFD() int {
	return p.fd
}
