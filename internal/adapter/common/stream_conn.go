package common

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
)

const streamConnStatusRunning = int32(1)
const streamConnStatusClosing = int32(2)
const streamConnStatusClosed = int32(0)

// StreamConn ...
type StreamConn struct {
	status   int32
	prev     IConn
	receiver IReceiver
	writeCH  chan *core.Stream

	readHeadPos int
	readHeadBuf []byte
	readStream  *core.Stream
	writeStream *core.Stream
	writePos    int
}

// NewStreamConn ...
func NewStreamConn(prev IConn, receiver IReceiver) *StreamConn {
	return &StreamConn{
		status:      streamConnStatusClosed,
		prev:        prev,
		receiver:    receiver,
		writeCH:     make(chan *core.Stream, 16),
		readHeadPos: 0,
		readHeadBuf: make([]byte, core.StreamHeadSize),
		readStream:  nil,
		writeStream: nil,
		writePos:    0,
	}
}

// OnOpen ...
func (p *StreamConn) OnOpen() {
	atomic.StoreInt32(&p.status, streamConnStatusRunning)
	p.receiver.OnConnOpen(p)
}

// OnClose ...
func (p *StreamConn) OnClose() {
	p.receiver.OnConnClose(p)
	atomic.StoreInt32(&p.status, streamConnStatusClosed)
}

// OnError ...
func (p *StreamConn) OnError(err *base.Error) {
	p.receiver.OnConnError(p, err)
}

// OnReadBytes ...
func (p *StreamConn) OnReadBytes(b []byte) {
	if p.readStream == nil {
		if p.readHeadPos == 0 { // fast cache
			if bytesLen := len(b); bytesLen >= core.StreamHeadSize {
				streamLength := int(core.GetStreamLengthByHeadBuffer(b))
				if bytesLen == streamLength {
					stream := core.NewStream()
					stream.PutBytesTo(b, 0)
					p.receiver.OnConnReadStream(p, stream)
					return
				} else if bytesLen > streamLength {
					stream := core.NewStream()
					stream.PutBytesTo(b, 0)
					p.receiver.OnConnReadStream(p, stream)
					p.OnReadBytes(b[streamLength:])
					return
				} else {
					p.readStream = core.NewStream()
					p.readStream.PutBytesTo(b, 0)
					return
				}
			}
		}

		if p.readHeadPos < core.StreamHeadSize {
			copyBytes := copy(p.readHeadBuf[p.readHeadPos:], b)
			p.readHeadPos += copyBytes
			b = b[copyBytes:]
		}

		if p.readHeadPos < core.StreamHeadSize {
			return
		}

		p.readStream = core.NewStream()
		p.readStream.PutBytesTo(p.readHeadBuf, 0)
		p.readHeadPos = 0
	}

	if byteLen := len(b); byteLen > 0 {
		streamLength := int(p.readStream.GetLength())
		remains := streamLength - p.readStream.GetWritePos()
		writeBuf := b[:base.MinInt(byteLen, remains)]
		p.readStream.PutBytes(writeBuf)
		if p.readStream.GetWritePos() == streamLength {
			if p.readStream.CheckStream() {
				p.receiver.OnConnReadStream(p, p.readStream)
				p.readStream = nil
			} else {
				p.receiver.OnConnError(p, errors.ErrStream)
				return
			}
		}

		if byteLen > len(writeBuf) {
			p.OnReadBytes(b[:len(writeBuf)])
		}
	}
}

// OnFillWrite ...
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

	if len(peekBuf) <= 0 {
		p.OnError(errors.ErrTemp.AddDebug("OnFillWrite internal error"))
		return 0
	}

	copyBytes := copy(b, peekBuf)
	p.writePos += copyBytes

	if finish {
		p.writeStream.Release()
		p.writeStream = nil
		p.writePos = 0
	}

	return copyBytes
}

// Close ...
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

// LocalAddr ...
func (p *StreamConn) LocalAddr() net.Addr {
	return p.prev.LocalAddr()
}

// RemoteAddr ...
func (p *StreamConn) RemoteAddr() net.Addr {
	return p.prev.RemoteAddr()
}

// WriteStreamAndRelease ...
func (p *StreamConn) WriteStreamAndRelease(stream *core.Stream) {
	func() {
		defer func() {
			if v := recover(); v != nil {
				p.OnError(errors.ErrTemp.AddDebug(fmt.Sprintf("%v", v)))
			}
		}()

		p.writeCH <- stream
	}()

	p.prev.OnWriteReady()
}

// OnReadReady ...
func (p *StreamConn) OnReadReady() bool {
	panic("kernel error, this code should not be called")
}

// OnWriteReady ...
func (p *StreamConn) OnWriteReady() {
	panic("kernel error, this code should not be called")
}

// GetFD ...
func (p *StreamConn) GetFD() int {
	panic("kernel error, this code should not be called")
}
