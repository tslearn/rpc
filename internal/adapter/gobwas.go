package adapter

import (
	"github.com/gobwas/ws/wsutil"
	"net"
	"time"
)

type syncWSClientConn struct {
	conn       net.Conn
	readBuffer []byte
}

func (p *syncWSClientConn) Read(b []byte) (int, error) {
	if p.readBuffer != nil {
		n := copy(b, p.readBuffer)
		p.readBuffer = nil
		return n, nil
	}
	msg, e := wsutil.ReadServerBinary(p.conn)

	if e != nil {
		return -1, e
	}

	n := copy(b, msg)
	if n < len(msg) {
		p.readBuffer = msg[n:]
	}
	return n, nil
}

func (p *syncWSClientConn) Write(b []byte) (n int, err error) {
	err = wsutil.WriteClientBinary(p.conn, b)
	if err != nil {
		return -1, err
	}
	return len(b), nil
}

func (p *syncWSClientConn) Close() error {
	return p.conn.Close()
}

func (p *syncWSClientConn) LocalAddr() net.Addr {
	return p.conn.LocalAddr()
}

func (p *syncWSClientConn) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}

func (p *syncWSClientConn) SetDeadline(t time.Time) error {
	return p.conn.SetDeadline(t)
}

func (p *syncWSClientConn) SetReadDeadline(t time.Time) error {
	return p.conn.SetReadDeadline(t)
}

func (p *syncWSClientConn) SetWriteDeadline(t time.Time) error {
	return p.conn.SetWriteDeadline(t)
}

type syncWSServerConn struct {
	conn       net.Conn
	readBuffer []byte
}

func (p *syncWSServerConn) Read(b []byte) (int, error) {
	if p.readBuffer != nil {
		n := copy(b, p.readBuffer)
		p.readBuffer = nil
		return n, nil
	}
	msg, e := wsutil.ReadClientBinary(p.conn)

	if e != nil {
		return -1, e
	}

	n := copy(b, msg)
	if n < len(msg) {
		p.readBuffer = msg[n:]
	}
	return n, nil
}

func (p *syncWSServerConn) Write(b []byte) (n int, err error) {
	err = wsutil.WriteServerBinary(p.conn, b)
	if err != nil {
		return -1, err
	}
	return len(b), nil
}

func (p *syncWSServerConn) Close() error {
	return p.conn.Close()
}

func (p *syncWSServerConn) LocalAddr() net.Addr {
	return p.conn.LocalAddr()
}

func (p *syncWSServerConn) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}

func (p *syncWSServerConn) SetDeadline(t time.Time) error {
	return p.conn.SetDeadline(t)
}

func (p *syncWSServerConn) SetReadDeadline(t time.Time) error {
	return p.conn.SetReadDeadline(t)
}

func (p *syncWSServerConn) SetWriteDeadline(t time.Time) error {
	return p.conn.SetWriteDeadline(t)
}
