package adapter

import (
	"fmt"
	"testing"
	"time"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
)

type serverReceiver struct {
}

func (p *serverReceiver) OnConnOpen(_ *StreamConn) {
	fmt.Println("Server: OnConnOpen")
}

func (p *serverReceiver) OnConnClose(_ *StreamConn) {
	fmt.Println("Server: OnConnClose")
}

func (p *serverReceiver) OnConnReadStream(
	streamConn *StreamConn,
	stream *core.Stream,
) {
	streamConn.WriteStreamAndRelease(stream)
}

func (p *serverReceiver) OnConnError(
	streamConn *StreamConn,
	err *base.Error,
) {
	if streamConn != nil {
		streamConn.Close()
	}

	fmt.Println("Server: OnConnError", err)
}

type clientReceiver struct {
	streamCH   chan *core.Stream
	streamConn *StreamConn
}

func (p *clientReceiver) OnConnOpen(streamConn *StreamConn) {
	fmt.Println("Client: OnConnOpen")
	p.streamConn = streamConn
}

func (p *clientReceiver) OnConnClose(_ *StreamConn) {
	fmt.Println("Client: OnConnClose")

	p.streamConn = nil
}

func (p *clientReceiver) OnConnReadStream(
	_ *StreamConn,
	stream *core.Stream,
) {
	p.streamCH <- stream
}

func (p *clientReceiver) OnConnError(
	streamConn *StreamConn,
	err *base.Error,
) {
	if streamConn != nil {
		streamConn.Close()
	}

	fmt.Println("Client: OnConnError", err)
}

func BenchmarkDebug(b *testing.B) {
	go func() {
		serverAdapter := NewServerAdapter(
			"tcp", "0.0.0.0:8080", nil, 1200, 1200, &serverReceiver{},
		)
		serverAdapter.Open()
		serverAdapter.Run()
	}()

	time.Sleep(time.Second)

	cReceiver := &clientReceiver{streamCH: make(chan *core.Stream)}
	clientAdapter := NewClientAdapter(
		"tcp", "0.0.0.0:8080", nil, 1200, 1200, cReceiver,
	)

	go func() {
		clientAdapter.Open()
		clientAdapter.Run()
	}()

	time.Sleep(time.Second)

	if cReceiver.streamConn == nil {
		panic("not connect")
	}

	ptrBN := &b.N
	*ptrBN = 100000

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
