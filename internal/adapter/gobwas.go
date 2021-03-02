package adapter

import (
	"github.com/gobwas/ws/wsutil"
	"io"
	"net"
	"time"
)

type syncWSConn struct {
	conn        net.Conn
	readBuffer  []byte
	readBinary  func(rw io.ReadWriter) ([]byte, error)
	writeBinary func(w io.Writer, p []byte) error
}

func newSyncWSServerConn(conn net.Conn) *syncWSConn {
	return &syncWSConn{
		conn:        conn,
		readBuffer:  nil,
		readBinary:  wsutil.ReadClientBinary,
		writeBinary: wsutil.WriteServerBinary,
	}
}

func newSyncWSClientConn(conn net.Conn) *syncWSConn {
	return &syncWSConn{
		conn:        conn,
		readBuffer:  nil,
		readBinary:  wsutil.ReadServerBinary,
		writeBinary: wsutil.WriteClientBinary,
	}
}

func (p *syncWSConn) Read(b []byte) (int, error) {
	if p.readBuffer != nil {
		n := copy(b, p.readBuffer)

		if n < len(p.readBuffer) {
			p.readBuffer = p.readBuffer[n:]
		} else {
			p.readBuffer = nil
		}

		return n, nil
	}
	msg, e := p.readBinary(p.conn)

	if e != nil {
		return -1, e
	}

	n := copy(b, msg)
	if n < len(msg) {
		p.readBuffer = msg[n:]
	}
	return n, nil
}

func (p *syncWSConn) Write(b []byte) (n int, err error) {
	err = p.writeBinary(p.conn, b)
	if err != nil {
		return -1, err
	}
	return len(b), nil
}

func (p *syncWSConn) Close() error {
	return p.conn.Close()
}

func (p *syncWSConn) LocalAddr() net.Addr {
	return p.conn.LocalAddr()
}

func (p *syncWSConn) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}

func (p *syncWSConn) SetDeadline(t time.Time) error {
	return p.conn.SetDeadline(t)
}

func (p *syncWSConn) SetReadDeadline(t time.Time) error {
	return p.conn.SetReadDeadline(t)
}

func (p *syncWSConn) SetWriteDeadline(t time.Time) error {
	return p.conn.SetWriteDeadline(t)
}
