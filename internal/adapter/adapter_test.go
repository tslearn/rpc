package adapter

import (
	"crypto/tls"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"net"
	"path"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

type testSingleReceiver struct {
	onOpenCount   int
	onCloseCount  int
	onErrorCount  int
	onStreamCount int

	streamConn *StreamConn
	errCH      chan *base.Error
	streamCH   chan *core.Stream
	sync.Mutex
}

func newTestSingleReceiver() *testSingleReceiver {
	return &testSingleReceiver{
		streamConn: nil,
		errCH:      make(chan *base.Error, 1024),
		streamCH:   make(chan *core.Stream, 1024),
	}
}

func (p *testSingleReceiver) OnConnOpen(streamConn *StreamConn) {
	p.Lock()
	defer p.Unlock()
	p.onOpenCount++
	if p.streamConn == nil {
		p.streamConn = streamConn
	} else {
		panic("error")
	}
}

func (p *testSingleReceiver) OnConnClose(streamConn *StreamConn) {
	p.Lock()
	defer p.Unlock()
	p.onCloseCount++
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
	p.onStreamCount++
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
	p.onErrorCount++
	if streamConn != nil && p.streamConn != streamConn {
		panic("error")
	}
	p.errCH <- err
}

func (p *testSingleReceiver) GetOnOpenCount() int {
	p.Lock()
	defer p.Unlock()
	return p.onOpenCount
}

func (p *testSingleReceiver) GetOnCloseCount() int {
	p.Lock()
	defer p.Unlock()
	return p.onCloseCount
}

func (p *testSingleReceiver) GetOnStreamCount() int {
	p.Lock()
	defer p.Unlock()
	return p.onStreamCount
}

func (p *testSingleReceiver) GetOnErrorCount() int {
	p.Lock()
	defer p.Unlock()
	return p.onErrorCount
}

func (p *testSingleReceiver) GetStream() *core.Stream {
	return <-p.streamCH
}

func (p *testSingleReceiver) GetError() *base.Error {
	select {
	case ret := <-p.errCH:
		return ret
	default:
		return nil
	}
}

func TestAdapter(t *testing.T) {
	type testItem struct {
		network string
		isTLS   bool
		e       error
	}

	fnTest := func(isTLS bool, network string) {
		assert := base.NewAssert(t)
		_, curFile, _, _ := runtime.Caller(0)
		curDir := path.Dir(curFile)

		tlsClientConfig := (*tls.Config)(nil)
		tlsServerConfig := (*tls.Config)(nil)

		if isTLS {
			tlsServerConfig, _ = base.GetTLSServerConfig(
				path.Join(curDir, "_cert_", "server", "server.pem"),
				path.Join(curDir, "_cert_", "server", "server-key.pem"),
			)
			tlsClientConfig, _ = base.GetTLSClientConfig(true, []string{
				path.Join(curDir, "_cert_", "ca", "ca.pem"),
			})
		}

		clientReceiver := newTestSingleReceiver()
		clientAdapter := NewClientAdapter(
			network, "localhost:65432", tlsClientConfig,
			1200, 1200, clientReceiver,
		)
		serverReceiver := newTestSingleReceiver()
		serverAdapter := NewServerAdapter(
			network, "localhost:65432", tlsServerConfig,
			1200, 1200, serverReceiver,
		)

		waitCH := make(chan bool)
		assert(serverAdapter.Open()).IsTrue()
		go func() {
			assert(serverAdapter.Run()).IsTrue()
		}()

		assert(clientAdapter.Open()).IsTrue()
		go func() {
			for clientReceiver.GetOnOpenCount() == 0 &&
				clientReceiver.GetOnErrorCount() == 0 {
				time.Sleep(50 * time.Millisecond)
			}
			assert(clientAdapter.Close()).IsTrue()
			waitCH <- true
		}()

		clientAdapter.Run()
		<-waitCH
		assert(serverAdapter.Close()).IsTrue()

		assert(clientReceiver.GetOnOpenCount()).Equal(1)
		assert(clientReceiver.GetOnCloseCount()).Equal(1)
		assert(clientReceiver.GetOnErrorCount()).Equal(0)
		assert(clientReceiver.GetOnStreamCount()).Equal(0)

		assert(serverReceiver.GetOnOpenCount()).Equal(1)
		assert(serverReceiver.GetOnCloseCount()).Equal(1)
		assert(serverReceiver.GetOnErrorCount()).Equal(0)
		assert(serverReceiver.GetOnStreamCount()).Equal(0)
	}

	t.Run("test basic", func(t *testing.T) {
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

	t.Run("test", func(t *testing.T) {
		for _, it := range []testItem{
			{network: "tcp", isTLS: false},
			{network: "tcp", isTLS: true},
			{network: "tcp4", isTLS: false},
			{network: "tcp4", isTLS: true},
			{network: "tcp6", isTLS: false},
			{network: "tcp6", isTLS: true},
			{network: "ws", isTLS: false},
			{network: "wss", isTLS: true},
		} {
			fnTest(it.isTLS, it.network)
		}
	})

	t.Run("open return false", func(t *testing.T) {
		assert := base.NewAssert(t)

		assert(NewClientAdapter(
			"err", "localhost:65432", nil,
			1200, 1200, newTestSingleReceiver(),
		).Open()).IsFalse()

		assert(NewServerAdapter(
			"err", "localhost:65432", nil,
			1200, 1200, newTestSingleReceiver(),
		).Open()).IsFalse()
	})
}