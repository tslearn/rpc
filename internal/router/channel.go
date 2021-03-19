package router

import (
	"errors"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
	"net"
	"sync"
	"time"
)

type Channel struct {
	isRunning bool
	streamCH  chan *rpc.Stream
	conn      net.Conn
	streamHub rpc.IStreamHub
	sync.Mutex
}

func NewChannel(streamCH chan *rpc.Stream, streamHub rpc.IStreamHub) *Channel {
	return &Channel{
		isRunning: true,
		streamCH:  streamCH,
		conn:      nil,
		streamHub: streamHub,
	}
}

func (p *Channel) IsNeedConnected() bool {
	p.Lock()
	defer p.Unlock()

	return p.isRunning && p.conn == nil
}

func (p *Channel) getConn() net.Conn {
	p.Lock()
	defer p.Unlock()
	return p.conn
}

func (p *Channel) setConn(conn net.Conn) bool {
	p.Lock()
	defer p.Unlock()

	if !p.isRunning {
		return false
	}

	if p.conn != nil {
		_ = p.conn.Close()
	}

	p.conn = conn

	return true
}

func (p *Channel) waitUntilConnIsNil() {
	for p.getConn() != nil {
		time.Sleep(50 * time.Millisecond)
	}
}

func (p *Channel) RunWithConn(conn net.Conn) bool {
	p.waitUntilConnIsNil()

	if !p.setConn(conn) {
		return false
	}

	waitCH := make(chan bool)

	go func() {
		p.runRead(conn)
		_ = conn.Close()
		waitCH <- true
	}()

	go func() {
		p.runWrite(conn)
		_ = conn.Close()
		waitCH <- true
	}()

	<-waitCH
	<-waitCH

	p.setConn(nil)
	return true
}

func (p *Channel) runRead(conn net.Conn) error {
	buffer := make([]byte, 8192)
	bufferStart := 0
	bufferPos := 0

	stream := (*rpc.Stream)(nil)

	for {
		if stream == nil {
			// read the head
			for bufferPos-bufferStart < rpc.StreamHeadSize {
				if n, e := conn.Read(buffer[bufferPos:]); e != nil {
					return e
				} else {
					bufferPos += n
				}
			}

			// create the stream
			stream = rpc.NewStream()

			// put header to stream
			stream.PutBytesTo(
				buffer[bufferStart:bufferStart+rpc.StreamHeadSize],
				0,
			)
			bufferStart += rpc.StreamHeadSize
		} else if remainBytes := bufferPos - bufferStart; remainBytes > 0 {
			streamLength := int(stream.GetLength())
			readMaxBytes := base.MinInt(
				remainBytes,
				streamLength-stream.GetWritePos(),
			)

			if readMaxBytes >= 0 {
				stream.PutBytes(buffer[bufferStart : bufferStart+readMaxBytes])
				if stream.GetWritePos() == streamLength {
					if stream.CheckStream() {
						p.streamHub.OnReceiveStream(stream)
					} else {
						return errors.New("stream error")
					}
				}
			} else {
				return errors.New("stream error")
			}
		}

		// move the buffer
		if len(buffer)-bufferStart <= rpc.StreamHeadSize {
			copy(buffer[0:], buffer[bufferStart:])
			bufferPos -= bufferStart
			bufferStart = 0
		}
	}
}

func (p *Channel) runWrite(conn net.Conn) error {
	for {
		stream := <-p.streamCH
		writePos := 0

		for {
			peekBuf, finish := stream.PeekBufferSlice(writePos, 1024)

			peekLen := len(peekBuf)

			if peekLen <= 0 {
				return errors.New("error stream")
			}

			start := 0
			for start < peekLen {
				if n, e := conn.Write(peekBuf[start:]); e != nil {
					return e
				} else {
					start += n
				}
			}

			writePos += peekLen

			if finish {
				break
			}
		}
	}
}

func (p *Channel) Close() {
	func() {
		p.Lock()
		defer p.Unlock()
		p.isRunning = false
		if p.conn != nil {
			_ = p.conn.Close()
		}
	}()

	p.waitUntilConnIsNil()
}
