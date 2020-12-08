package adapter

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
	"sync/atomic"
)

type XConn interface {
	OnOpen()
	OnClose()
	OnError(err *base.Error)
	OnReadBytes(b []byte)
	OnFillWrite(b []byte) int

	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	TriggerWrite()
	Close()
}

const streamConnStatusRunning = int32(1)
const streamConnStatusClosing = int32(2)
const streamConnStatusClosed = int32(0)

type StreamConn struct {
	status   int32
	prev     XConn
	receiver IReceiver
	writeCH  chan *core.Stream

	writeStream *core.Stream
	writePos    int
}

func NewStreamConn(prev XConn, receiver IReceiver) *StreamConn {
	return &StreamConn{
		status:   streamConnStatusClosed,
		prev:     prev,
		receiver: receiver,
		writeCH:  make(chan *core.Stream, 16),
	}
}

func (p *StreamConn) OnOpen() {
	atomic.StoreInt32(&p.status, streamConnStatusRunning)
	p.receiver.OnConnOpen(p)
}

func (p *StreamConn) OnClose() {
	p.receiver.OnConnClose(p)
	atomic.StoreInt32(&p.status, streamConnStatusClosed)
}

func (p *StreamConn) OnError(err *base.Error) {
	p.receiver.OnConnError(p, err)
}

func (p *StreamConn) OnReadBytes(b []byte) {
	panic("not implement")
}

func (p *StreamConn) OnFillWrite(b []byte) int {
	if p.writeStream == nil {
		select {
		case stream := <-p.writeCH:
			p.writeStream = stream
			p.writePos = 0
		default:
			return 0
		}
	}

	peekBuf, finish := p.writeStream.PeekBufferSlice(p.writePos, len(b))
	if len(peekBuf) > 0 {
		copyBytes := copy(b, peekBuf)
		p.writePos += copyBytes

		if finish {
			p.writeStream = nil
			p.writePos = 0
		}

		return copyBytes
	} else {
		p.OnError(errors.ErrTemp.AddDebug("OnFillWrite internal error"))
		return 0
	}
}

func (p *StreamConn) TriggerWrite() {
	p.prev.TriggerWrite()
}

func (p *StreamConn) Close() {
	if atomic.CompareAndSwapInt32(
		&p.status,
		streamConnStatusRunning,
		streamConnStatusClosing,
	) {
		close(p.writeCH)
		p.prev.Close()
	}
}

func (p *StreamConn) LocalAddr() net.Addr {
	return p.prev.LocalAddr()
}

func (p *StreamConn) RemoteAddr() net.Addr {
	return p.prev.RemoteAddr()
}

func (p *StreamConn) WriteStream(stream *core.Stream) {
	func() {
		defer func() {
			if v := recover(); v != nil {
				p.OnError(errors.ErrTemp.AddDebug(fmt.Sprintf("%v", v)))
			}
		}()

		p.writeCH <- stream
	}()

	p.TriggerWrite()
}
