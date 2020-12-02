package adapter

import (
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
	receiver XReceiver
	conn     XConn
	fd       int
}

func NewEventConn(receiver XReceiver, conn XConn, fd int) *EventConn {
	return &EventConn{
		receiver: receiver,
		conn:     conn,
		fd:       fd,
	}
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
