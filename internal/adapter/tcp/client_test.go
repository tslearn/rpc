package tcp

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"net"
	"testing"
	"time"
)

type testReceiver struct {
}

func (p *testReceiver) OnEventConnOpen(eventConn *adapter.EventConn) {
	fmt.Println("OnEventConnOpen")

	_, _ = eventConn.GetConn().Write([]byte("hello"))
}

func (p *testReceiver) OnEventConnClose(eventConn *adapter.EventConn) {
	fmt.Println("OnEventConnClose")
}

func (p *testReceiver) OnEventConnStream(
	eventConn *adapter.EventConn,
	stream *core.Stream,
) {
	fmt.Println("OnEventConnStream")
}

func (p *testReceiver) OnEventConnError(
	eventConn *adapter.EventConn,
	err *base.Error,
) {
	fmt.Println("OnEventConnError", err)
}

func TestDebug(t *testing.T) {
	go func() {
		ln, e := net.Listen("tcp", "0.0.0.0:8080")
		if e != nil {
			panic(e)
		}
		for {
			tcpConn, e := ln.Accept()
			if e != nil {
				return
			}

			go func(conn net.Conn) {
				buf := make([]byte, 1024)
				n, _ := conn.Read(buf)
				fmt.Println(conn.Write(buf[:n]))
				//  _ = conn.Close()
			}(tcpConn)
		}
	}()

	time.Sleep(time.Second)

	tcpAdapter := NewTCPClientAdapter("0.0.0.0:8080")
	tcpAdapter.Open(&testReceiver{})

	fmt.Println("End")
}
