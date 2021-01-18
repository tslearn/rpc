package adapter

import (
	"github.com/rpccloud/rpc/internal/core"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

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

// NewServerNetConn ...
func NewServerNetConn(netConn net.Conn, rBufSize int, wBufSize int) *NetConn {
	return &NetConn{
		isServer:  true,
		isRunning: true,
		conn:      netConn,
		next:      nil,
		rBuf:      make([]byte, rBufSize),
		wBuf:      make([]byte, wBufSize),
	}
}

// NewClientNetConn ...
func NewClientNetConn(netConn net.Conn, rBufSize int, wBufSize int) *NetConn {
	return &NetConn{
		isServer:  false,
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
			p.OnError(errors.ErrConnClose.AddDebug(e.Error()))
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
				p.OnError(errors.ErrConnRead.AddDebug(e.Error()))
			}
		} else {
			p.Lock()
			ignoreReport := (!p.isRunning) &&
				strings.HasSuffix(e.Error(), ErrNetClosingSuffix)
			p.Unlock()

			if !ignoreReport {
				p.OnError(errors.ErrConnRead.AddDebug(e.Error()))
			}
		}

		return false
	}
	p.next.OnReadBytes(p.rBuf[:n])
	return true
}

// OnWriteReady ...
func (p *NetConn) OnWriteReady() bool {
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
				p.OnError(errors.ErrConnWrite.AddDebug(e.Error()))
				return false
			} else if n == 0 {
				return false
			} else {
				start += n
			}
		}
	}

	return true
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

const streamConnStatusRunning = int32(1)
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

	activeTimeNS int64
}

// NewStreamConn ...
func NewStreamConn(prev IConn, receiver IReceiver) *StreamConn {
	return &StreamConn{
		status:      streamConnStatusRunning,
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

// SetReceiver ...
func (p *StreamConn) SetReceiver(receiver IReceiver) {
	p.receiver = receiver
}

// OnOpen ...
func (p *StreamConn) OnOpen() {
	p.receiver.OnConnOpen(p)
}

// OnClose ...
func (p *StreamConn) OnClose() {
	p.receiver.OnConnClose(p)
}

// OnError ...
func (p *StreamConn) OnError(err *base.Error) {
	p.receiver.OnConnError(p, err)
}

// OnReadBytes ...
func (p *StreamConn) OnReadBytes(b []byte) {
	if p.readStream == nil {
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
				atomic.StoreInt64(&p.activeTimeNS, base.TimeNow().UnixNano())
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
		p.OnError(
			errors.ErrOnFillWriteFatal.AddDebug("OnFillWrite internal error"),
		)
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
		streamConnStatusClosed,
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
			recover()
		}()

		stream.BuildStreamCheck()
		p.writeCH <- stream
	}()

	p.prev.OnWriteReady()
}

func (p *StreamConn) IsActive(nowNS int64, timeout time.Duration) bool {
	return nowNS-atomic.LoadInt64(&p.activeTimeNS) < int64(timeout)
}

// SetNext ...
func (p *StreamConn) SetNext(_ IConn) {
	panic("kernel error, this code should not be called")
}

// OnReadReady ...
func (p *StreamConn) OnReadReady() bool {
	panic("kernel error, this code should not be called")
}

// OnWriteReady ...
func (p *StreamConn) OnWriteReady() bool {
	panic("kernel error, this code should not be called")
}

// GetFD ...
func (p *StreamConn) GetFD() int {
	panic("kernel error, this code should not be called")
}
