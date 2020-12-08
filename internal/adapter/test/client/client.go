package main

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/adapter/sync"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"net"
	"time"
)

type receiver struct {
	streamCH   chan *core.Stream
	streamConn *adapter.StreamConn
}

func (p *receiver) OnConnOpen(streamConn *adapter.StreamConn) {
	fmt.Println("Client: OnConnOpen")
	p.streamConn = streamConn
}

func (p *receiver) OnConnClose(streamConn *adapter.StreamConn) {
	fmt.Println("Client: OnConnClose")

	p.streamConn = nil
}

func (p *receiver) OnConnReadStream(
	streamConn *adapter.StreamConn,
	stream *core.Stream,
) {
	p.streamCH <- stream
}

func (p *receiver) OnConnError(
	streamConn *adapter.StreamConn,
	err *base.Error,
) {
	if streamConn != nil {
		streamConn.Close()
	}

	fmt.Println("Client: OnConnError", err)
}

func SimpleTestRead() {
	conn, err := net.Dial("tcp", "0.0.0.0:8080")
	if err != nil {
		panic(err)
	}
	buf := make([]byte, 1024)
	fmt.Println(conn.Read(buf))
}

func TestReceiver() {

	clientReceiver := &receiver{streamCH: make(chan *core.Stream)}

	clientAdapter := sync.NewSyncClientAdapter(
		"tcp", "0.0.0.0:8080", 1200, 1200, clientReceiver,
	)

	go func() {
		clientAdapter.Open()
	}()

	time.Sleep(time.Second)
	if clientReceiver.streamConn == nil {
		panic("not connect")
	}

	fmt.Println("Start Test")

	stream := core.NewStream()
	stream.WriteInt64(12)
	stream.BuildStreamCheck()

	for i := 0; i < 10; i++ {
		clientReceiver.streamConn.WriteStream(stream)
		s := <-clientReceiver.streamCH
		s.SetReadPosToBodyStart()
		if v, _ := s.ReadInt64(); v != int64(12) {
			panic("error")
		} else {
			fmt.Println(v)
		}
		s.Release()
	}

	fmt.Println("End Test")
	time.Sleep(10 * time.Second)
}

func main() {
	TestReceiver()
	return
}
