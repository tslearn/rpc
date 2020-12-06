package main

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/adapter/tcp"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"net"
)

type testReceiver struct {
}

func (p *testReceiver) OnEventConnOpen(eventConn *adapter.EventConn) {
	fmt.Println("Server: OnEventConnOpen")
}

func (p *testReceiver) OnEventConnClose(eventConn *adapter.EventConn) {
	fmt.Println("Server: OnEventConnClose")
}

func (p *testReceiver) OnEventConnStream(
	eventConn *adapter.EventConn,
	stream *core.Stream,
) {
	if err := eventConn.WriteStream(stream); err != nil {
		p.OnEventConnError(eventConn, err)
	}
	stream.Release()
}

func (p *testReceiver) OnEventConnError(
	eventConn *adapter.EventConn,
	err *base.Error,
) {
	fmt.Println("Server: OnEventConnError")
}

func main() {
	net.Listen("tcp", "0.0.0.0:8888")

	serverAdapter := tcp.NewTCPServerAdapter("0.0.0.0:8080", 1024, 1024)
	serverAdapter.Open(&testReceiver{})
}
