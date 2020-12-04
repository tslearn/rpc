package adapter

import (
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
)

type XConn interface {
	// Read reads data from the connection.
	Read(b []byte) (n int, err error)

	// Write writes data to the connection.
	Write(b []byte) (n int, err error)

	// Close closes the connection.
	// Any blocked Read or Write operations will be unblocked and return errors.
	Close() error

	// LocalAddr returns the local network address.
	LocalAddr() net.Addr

	// RemoteAddr returns the remote network address.
	RemoteAddr() net.Addr
}

type EventConn struct {
	receiver   XReceiver
	conn       XConn
	fd         int
	transLimit int
	rBuf       []byte
	rPos       int
	rStream    *core.Stream
}

func NewEventConn(
	receiver XReceiver,
	conn XConn,
	fd int,
	rBufSize int,
) *EventConn {
	return &EventConn{
		receiver:   receiver,
		conn:       conn,
		fd:         fd,
		transLimit: 4 * 1024 * 1024,
		rBuf:       make([]byte, rBufSize),
		rPos:       0,
	}
}

func (p *EventConn) GetTransLimit() int {
	return p.transLimit
}

func (p *EventConn) SetTransLimit(transLimit int) {
	p.transLimit = transLimit
}

func (p *EventConn) GetReceiver() XReceiver {
	return p.receiver
}

func (p *EventConn) SetReceiver(receiver XReceiver) {
	p.receiver = receiver
}

func (p *EventConn) GetFD() int {
	return p.fd
}

func (p *EventConn) GetConn() XConn {
	return p.conn
}

func (p *EventConn) OnReadReady() {
	if len(p.rBuf) < core.StreamHeadSize {
		p.receiver.OnEventConnError(p, errors.ErrEventConnReadBufferIsTooSmall)
	}

	if n, e := p.conn.Read(p.rBuf[p.rPos:]); e != nil {
		p.receiver.OnEventConnError(p, errors.ErrEventConnRead.AddDebug(e.Error()))
	} else {
		p.rPos += n
		start := 0

		for start < p.rPos {
			if p.rStream != nil {
				remains := int(p.rStream.GetLength()) - p.rStream.GetWritePos()
				if p.rPos < start+remains {
					p.rStream.PutBytes(p.rBuf[start:p.rPos])
					start = p.rPos
				} else {
					p.rStream.PutBytes(p.rBuf[start : start+remains])
					start += remains
					p.receiver.OnEventConnStream(p, p.rStream)
					p.rStream = nil
				}
			} else {
				if p.rPos-start >= core.StreamHeadSize {
					p.rStream = core.NewStream()
					p.rStream.PutBytesTo(p.rBuf[start:start+core.StreamHeadSize], 0)
					start += core.StreamHeadSize
				} else {
					break
				}
			}
		}

		if start < p.rPos {
			p.rPos = copy(p.rBuf, p.rBuf[start:p.rPos])
		} else {
			p.rPos = 0
		}
	}
}
