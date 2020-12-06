package tcp

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"testing"
	"time"
)

type testReceiver struct {
	streamCH  chan *core.Stream
	eventConn *adapter.EventConn
	isClient  bool
}

func (p *testReceiver) OnEventConnOpen(eventConn *adapter.EventConn) {
	if p.isClient {
		fmt.Println("Client: OnEventConnOpen")
	} else {
		fmt.Println("Server: OnEventConnOpen")
	}

	p.eventConn = eventConn
}

func (p *testReceiver) OnEventConnClose(eventConn *adapter.EventConn) {
	if p.isClient {
		fmt.Println("Client: OnEventConnClose")
	} else {
		fmt.Println("Server: OnEventConnClose")
	}

	p.eventConn = nil
}

func (p *testReceiver) OnEventConnStream(
	eventConn *adapter.EventConn,
	stream *core.Stream,
) {
	if p.isClient {
		p.streamCH <- stream
	} else {
		if p.eventConn != nil {
			if err := p.eventConn.WriteStream(stream); err != nil {
				p.OnEventConnError(p.eventConn, err)
			}
		}
		stream.Release()
	}
}

func (p *testReceiver) OnEventConnError(
	eventConn *adapter.EventConn,
	err *base.Error,
) {
	if p.isClient {
		fmt.Println("Client: OnEventConnError")
	} else {
		fmt.Println("Server: OnEventConnError")
	}
}

func BenchmarkDebug(b *testing.B) {
	go func() {
		serverAdapter := NewTCPServerAdapter("tcp", "0.0.0.0:8080", 1024, 1024)
		serverAdapter.Open(&testReceiver{isClient: false})
	}()

	time.Sleep(time.Second)

	clientReceiver := &testReceiver{isClient: true, streamCH: make(chan *core.Stream)}
	tcpAdapter := NewTCPClientAdapter("0.0.0.0:8080", 1024, 1024)
	go func() {
		tcpAdapter.Open(clientReceiver)
	}()
	time.Sleep(time.Second)

	stream := core.NewStream()
	stream.WriteInt64(12)
	stream.BuildStreamCheck()

	b.ResetTimer()
	b.ReportAllocs()
	b.N = 1000000

	for i := 0; i < b.N; i++ {
		clientReceiver.eventConn.WriteStream(stream)
		s := <-clientReceiver.streamCH
		s.SetReadPosToBodyStart()
		if v, _ := s.ReadInt64(); v != int64(12) {
			panic("error")
		}
		s.Release()
	}
}
