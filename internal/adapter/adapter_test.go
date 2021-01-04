package adapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"net"
	"strings"
	"sync"
	"testing"
)

type testSingleReceiver struct {
	streamConn *StreamConn
	errCH      chan *base.Error
	streamCH   chan *core.Stream
	sync.Mutex
}

func newTestSingleReceiver() *testSingleReceiver {
	return &testSingleReceiver{
		streamConn: nil,
		errCH:      make(chan *base.Error),
		streamCH:   make(chan *core.Stream),
	}
}

func (p *testSingleReceiver) OnConnOpen(streamConn *StreamConn) {
	p.Lock()
	defer p.Unlock()
	if p.streamConn == nil {
		p.streamConn = streamConn
	} else {
		panic("error")
	}
}

func (p *testSingleReceiver) OnConnClose(streamConn *StreamConn) {
	p.Lock()
	defer p.Unlock()
	if p.streamConn != nil && p.streamConn == streamConn {
		p.streamConn = nil
	} else {
		panic("error")
	}
}

func (p *testSingleReceiver) OnConnReadStream(
	streamConn *StreamConn,
	stream *core.Stream,
) {
	p.Lock()
	defer p.Unlock()
	if p.streamConn != nil && p.streamConn == streamConn {
		p.streamCH <- stream
	} else {
		panic("error")
	}
}

func (p *testSingleReceiver) OnConnError(
	streamConn *StreamConn,
	err *base.Error,
) {
	p.Lock()
	defer p.Unlock()
	if p.streamConn != nil && p.streamConn == streamConn {
		p.errCH <- err
	} else {
		panic("error")
	}
}

func (p *testSingleReceiver) GetStream() *core.Stream {
	return <-p.streamCH
}

func (p *testSingleReceiver) GetError() *base.Error {
	return <-p.errCH
}

func TestAdapter(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(ErrNetClosingSuffix).Equal("use of closed network connection")

		// check ErrNetClosingSuffix on all platform
		waitCH := make(chan bool)
		go func() {
			ln, e := net.Listen("tcp", "0.0.0.0:65432")
			if e != nil {
				panic(e)
			}

			waitCH <- true
			_, _ = ln.Accept()
			_ = ln.Close()
		}()

		<-waitCH
		conn, e := net.Dial("tcp", "0.0.0.0:65432")
		if e != nil {
			panic(e)
		}
		_ = conn.Close()
		e = conn.Close()
		assert(e).IsNotNil()
		assert(strings.HasSuffix(e.Error(), ErrNetClosingSuffix)).IsTrue()
	})
}
