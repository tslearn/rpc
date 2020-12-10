package test

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/adapter/async"
	"github.com/rpccloud/rpc/internal/adapter/sync"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"testing"
	"time"
)

type serverReceiver struct {
}

func (p *serverReceiver) OnConnOpen(streamConn *adapter.StreamConn) {
	fmt.Println("Server: OnConnOpen")
}

func (p *serverReceiver) OnConnClose(streamConn *adapter.StreamConn) {
	fmt.Println("Server: OnConnClose")
}

func (p *serverReceiver) OnConnReadStream(
	streamConn *adapter.StreamConn,
	stream *core.Stream,
) {
	streamConn.WriteStreamAndRelease(stream)
}

func (p *serverReceiver) OnConnError(
	streamConn *adapter.StreamConn,
	err *base.Error,
) {
	if streamConn != nil {
		streamConn.Close()
	}

	fmt.Println("Server: OnConnError", err)
}

type clientReceiver struct {
	streamCH   chan *core.Stream
	streamConn *adapter.StreamConn
}

func (p *clientReceiver) OnConnOpen(streamConn *adapter.StreamConn) {
	fmt.Println("Client: OnConnOpen")
	p.streamConn = streamConn
}

func (p *clientReceiver) OnConnClose(streamConn *adapter.StreamConn) {
	fmt.Println("Client: OnConnClose")

	p.streamConn = nil
}

func (p *clientReceiver) OnConnReadStream(
	streamConn *adapter.StreamConn,
	stream *core.Stream,
) {
	p.streamCH <- stream
}

func (p *clientReceiver) OnConnError(
	streamConn *adapter.StreamConn,
	err *base.Error,
) {
	if streamConn != nil {
		streamConn.Close()
	}

	fmt.Println("Client: OnConnError", err)
}

func BenchmarkDebug(b *testing.B) {
	go func() {
		serverAdapter := async.NewAsyncServerAdapter(
			"tcp", "0.0.0.0:8080", 1200, 1200, &serverReceiver{},
		)
		serverAdapter.Open()
	}()

	time.Sleep(time.Second)

	cReceiver := &clientReceiver{streamCH: make(chan *core.Stream)}
	clientAdapter := sync.NewSyncClientAdapter(
		"tcp", "0.0.0.0:8080", 1200, 1200, cReceiver,
	)

	go func() {
		clientAdapter.Open()
	}()

	time.Sleep(time.Second)

	if cReceiver.streamConn == nil {
		panic("not connect")
	}

	b.N = 100000

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		stream := core.NewStream()
		stream.WriteInt64(12)
		stream.BuildStreamCheck()
		cReceiver.streamConn.WriteStreamAndRelease(stream)

		s := <-cReceiver.streamCH
		s.SetReadPosToBodyStart()
		if v, _ := s.ReadInt64(); v != int64(12) {
			panic("error")
		}
		if i%10000 == 0 {
			fmt.Println(i)
		}
		s.Release()
	}
}