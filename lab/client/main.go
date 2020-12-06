package main

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/xadapter"
	"net"
	"time"
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
	xadapter.NewTCPListener(
		"tcp", "0.0.0.0:8080",
		func(fd int, localAddr net.Addr, remoteAddr net.Addr) {
			fmt.Println(fd, localAddr, remoteAddr)
		}, func(err *base.Error) {

		})

	time.Sleep(1 * time.Second)

	fmt.Println("Start")
	for i := 0; i < 200; i++ {
		net.Dial("tcp", "0.0.0.0:8080")
		fmt.Println(i)
	}
	fmt.Println("End")
	time.Sleep(10 * time.Second)

	//net.Listen("tcp", "0.0.0.0:8888")
	//
	//serverAdapter := tcp.NewTCPServerAdapter("0.0.0.0:8080", 1024, 1024)
	//serverAdapter.Open(&testReceiver{})
}
