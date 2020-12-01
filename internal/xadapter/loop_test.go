package xadapter

import (
	"fmt"
	"net"
	"reflect"
	"testing"
)

type TCPConn struct {
	fd   int
	conn net.Conn
}

func NewTCPConn(conn net.Conn) *TCPConn {
	return &TCPConn{
		fd:   tcpFD(conn.(*net.TCPConn)),
		conn: conn,
	}
}

func (p *TCPConn) FD() int {
	return p.fd
}

func (p *TCPConn) OnRead() error {
	buf := make([]byte, 1024)
	n, err := p.conn.Read(buf)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Read:", string(buf[:n]))

		go func() {
			_, _ = p.conn.Write(buf[:n])
		}()
	}

	return nil
}

func (p *TCPConn) OnClose() error {
	fmt.Println("OnClose")
	return nil
}

func tcpFD(conn *net.TCPConn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}

func Test_Debug(t *testing.T) {
	server, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		panic(err)
	}

	manager, err := NewLoopManager(4)
	if err != nil {
		panic(err)
	}

	manager.Open()

	for {
		// Listen for an incoming connection.
		conn, err := server.Accept()
		fmt.Println("AAA")
		if err == nil {
			channel := manager.AllocChannel()
			channel.AddConn(NewTCPConn(conn))
		} else {
			panic(err)
		}
	}
}
