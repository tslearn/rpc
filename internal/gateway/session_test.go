package gateway

import (
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"github.com/rpccloud/rpc/internal/router"
	"net"
	"strings"
	"testing"
	"time"
)

type testNetConn struct {
	writeBuffer []byte
	isRunning   bool
}

func newTestNetConn() *testNetConn {
	return &testNetConn{
		isRunning:   true,
		writeBuffer: make([]byte, 0),
	}
}

func (p *testNetConn) Read(_ []byte) (n int, err error) {
	panic("not implemented")
}

func (p *testNetConn) Write(b []byte) (n int, err error) {
	p.writeBuffer = append(p.writeBuffer, b...)
	return len(b), nil
}

func (p *testNetConn) Close() error {
	p.isRunning = false
	return nil
}

func (p *testNetConn) LocalAddr() net.Addr {
	panic("not implemented")
}

func (p *testNetConn) RemoteAddr() net.Addr {
	panic("not implemented")
}

func (p *testNetConn) SetDeadline(_ time.Time) error {
	panic("not implemented")
}

func (p *testNetConn) SetReadDeadline(_ time.Time) error {
	panic("not implemented")
}

func (p *testNetConn) SetWriteDeadline(_ time.Time) error {
	panic("not implemented")
}

type fakeRouteSender struct {
	emulateError bool
	streamCH     chan *core.Stream
}

func newFakeRouteSender(emulateError bool) *fakeRouteSender {
	return &fakeRouteSender{
		emulateError: emulateError,
		streamCH:     make(chan *core.Stream, 1024),
	}
}

func (p *fakeRouteSender) SendStreamToRouter(stream *core.Stream) *base.Error {
	if p.emulateError {
		return errors.ErrStream
	} else {
		p.streamCH <- stream
		return nil
	}
}

func prepareTestSession() (*Session, adapter.IConn, *testNetConn) {
	gateway := NewGateWay(
		3,
		GetDefaultConfig(),
		router.NewDirectRouter(),
		func(sessionID uint64, err *base.Error) {

		},
	)
	session := NewSession(11, gateway)
	netConn := newTestNetConn()
	syncConn := adapter.NewServerSyncConn(netConn, 1200, 1200)
	streamConn := adapter.NewStreamConn(syncConn, session)
	syncConn.SetNext(streamConn)
	gateway.Add(session)
	return session, syncConn, netConn
}

func TestNewSession(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		gateway := NewGateWay(
			43,
			GetDefaultConfig(),
			router.NewDirectRouter(),
			func(sessionID uint64, err *base.Error) {},
		)
		v := NewSession(3, gateway)
		assert(v.id).Equal(uint64(3))
		assert(v.gateway).Equal(gateway)
		assert(len(v.security)).Equal(32)
		assert(v.conn).IsNil()
		assert(len(v.channels)).Equal(GetDefaultConfig().numOfChannels)
		assert(cap(v.channels)).Equal(GetDefaultConfig().numOfChannels)
		assert(base.TimeNow().UnixNano()-v.activeTimeNS < int64(time.Second)).
			IsTrue()
		assert(v.prev).IsNil()
		assert(v.next).IsNil()
	})
}

func TestSession_TimeCheck(t *testing.T) {
	t.Run("p.conn is active", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()
		session.gateway.config.heartbeatTimeout = 100 * time.Millisecond
		syncConn.OnOpen()
		session.TimeCheck(base.TimeNow().UnixNano())
		assert(netConn.isRunning).IsTrue()
		assert(session.conn).IsNotNil()
	})

	t.Run("p.conn is not active", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()
		session.gateway.config.heartbeatTimeout = 10 * time.Millisecond
		syncConn.OnOpen()
		time.Sleep(20 * time.Millisecond)
		session.TimeCheck(base.TimeNow().UnixNano())
		assert(netConn.isRunning).IsFalse()
	})

	t.Run("p.conn is nil, session is not timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, _ := prepareTestSession()
		session.gateway.config.serverSessionTimeout = time.Second
		session.gateway.TimeCheck(base.TimeNow().UnixNano())
		assert(session.gateway.TotalSessions()).Equal(int64(1))
	})

	t.Run("p.conn is nil, session is timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, _ := prepareTestSession()
		gw := session.gateway
		gw.config.serverSessionTimeout = 10 * time.Millisecond
		time.Sleep(20 * time.Millisecond)
		gw.TimeCheck(base.TimeNow().UnixNano())
		assert(gw.TotalSessions()).Equal(int64(0))
	})

	t.Run("p.channels is not timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, _ := prepareTestSession()

		// fill the channels
		for i := 0; i < session.gateway.config.numOfChannels; i++ {
			stream := core.NewStream()
			stream.SetCallbackID(uint64(i) + 1)
			session.channels[i].In(stream.GetCallbackID())
			session.channels[i].Out(stream)
			assert(session.channels[i].backTimeNS > 0).IsTrue()
			assert(session.channels[i].backStream).IsNotNil()
		}

		session.gateway.config.serverCacheTimeout = 10 * time.Millisecond
		session.TimeCheck(base.TimeNow().UnixNano())

		for i := 0; i < session.gateway.config.numOfChannels; i++ {
			assert(session.channels[i].backTimeNS > 0).IsTrue()
			assert(session.channels[i].backStream).IsNotNil()
		}
	})

	t.Run("p.channels is timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, _ := prepareTestSession()

		// fill the channels
		for i := 0; i < session.gateway.config.numOfChannels; i++ {
			stream := core.NewStream()
			stream.SetCallbackID(uint64(i) + 1)
			session.channels[i].In(stream.GetCallbackID())
			session.channels[i].Out(stream)
		}

		session.gateway.config.serverCacheTimeout = 10 * time.Millisecond
		time.Sleep(20 * time.Millisecond)
		session.TimeCheck(base.TimeNow().UnixNano())

		for i := 0; i < session.gateway.config.numOfChannels; i++ {
			assert(session.channels[i].backTimeNS).Equal(int64(0))
			assert(session.channels[i].backStream).IsNil()
		}
	})
}

func TestSession_OutStream(t *testing.T) {
	t.Run("stream is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, netConn := prepareTestSession()
		session.OutStream(nil)
		assert(len(netConn.writeBuffer)).Equal(0)
	})

	t.Run("p.conn is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, netConn := prepareTestSession()
		session.OutStream(core.NewStream())
		assert(len(netConn.writeBuffer)).Equal(0)
	})

	t.Run("stream can not out", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()
		syncConn.OnOpen()
		// ignore the init stream
		netConn.writeBuffer = make([]byte, 0)

		for i := 1; i <= len(session.channels); i++ {
			(&session.channels[i%len(session.channels)]).In(uint64(i))
		}

		for i := len(session.channels) + 1; i <= 2*len(session.channels); i++ {
			stream := core.NewStream()
			stream.SetCallbackID(uint64(i))
			session.OutStream(stream)
		}

		assert(len(netConn.writeBuffer)).Equal(0)
	})

	t.Run("stream can out", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()
		syncConn.OnOpen()
		// ignore the init stream
		netConn.writeBuffer = make([]byte, 0)

		exceptBuffer := make([]byte, 0)
		for i := 1; i <= len(session.channels); i++ {
			(&session.channels[i%len(session.channels)]).In(uint64(i))
		}

		for i := 1; i <= len(session.channels); i++ {
			stream := core.NewStream()
			stream.SetCallbackID(uint64(i))
			stream.BuildStreamCheck()
			exceptBuffer = append(exceptBuffer, stream.GetBuffer()...)
			session.OutStream(stream)
		}

		assert(netConn.writeBuffer).Equal(exceptBuffer)
	})
}

func TestSession_OnConnOpen(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()
		streamConn := adapter.NewStreamConn(syncConn, session)
		syncConn.SetNext(streamConn)
		session.OnConnOpen(streamConn)

		stream := core.NewStream()
		stream.PutBytesTo(netConn.writeBuffer, 0)

		cfg := session.gateway.config

		assert(stream.ReadInt64()).
			Equal(int64(core.ControlStreamConnectResponse), nil)
		connStr, err := stream.ReadString()
		assert(err).IsNil()
		connStrArr := strings.Split(connStr, "-")
		assert(len(connStrArr)).Equal(2)
		assert(connStrArr[0]).Equal("11")
		assert(len(connStrArr[1])).Equal(32)
		assert(stream.ReadInt64()).Equal(int64(cfg.numOfChannels), nil)
		assert(stream.ReadInt64()).Equal(int64(cfg.transLimit), nil)
		assert(stream.ReadInt64()).Equal(int64(cfg.heartbeat), nil)
		assert(stream.ReadInt64()).Equal(int64(cfg.heartbeatTimeout), nil)
		assert(stream.ReadInt64()).Equal(int64(cfg.clientRequestInterval), nil)
		assert(stream.IsReadFinish()).IsTrue()
		assert(stream.CheckStream()).IsTrue()
	})
}

func TestSession_OnConnReadStream(t *testing.T) {
	t.Run("cbID > 0, accept = true, backStream = nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		fakeSender := newFakeRouteSender(false)
		session, syncConn, _ := prepareTestSession()
		session.gateway.routeSender = fakeSender

		streamConn := adapter.NewStreamConn(syncConn, session)
		stream := core.NewStream()
		stream.SetCallbackID(10)
		session.OnConnReadStream(streamConn, stream)

		backStream := <-fakeSender.streamCH
		assert(backStream.GetGatewayID()).Equal(uint32(3))
		assert(backStream.GetSessionID()).Equal(uint64(11))
	})

	t.Run("cbID > 0, accept = true, backStream = nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		fakeSender := newFakeRouteSender(true)
		session, syncConn, _ := prepareTestSession()
		session.gateway.routeSender = fakeSender

		streamConn := adapter.NewStreamConn(syncConn, session)
		stream := core.NewStream()
		stream.SetCallbackID(10)
		session.OnConnReadStream(streamConn, stream)

		assert(len(fakeSender.streamCH)).Equal(0)
	})

	t.Run("cbID > 0, accept = false, backStream != nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()

		cacheStream := core.NewStream()
		cacheStream.SetCallbackID(10)
		cacheStream.BuildStreamCheck()

		streamConn := adapter.NewStreamConn(syncConn, session)
		syncConn.SetNext(streamConn)

		(&session.channels[10%len(session.channels)]).In(10)
		(&session.channels[10%len(session.channels)]).Out(cacheStream)

		stream := core.NewStream()
		stream.SetCallbackID(10)
		session.OnConnOpen(streamConn)
		netConn.writeBuffer = make([]byte, 0)
		session.OnConnReadStream(streamConn, stream)
		assert(netConn.writeBuffer).Equal(cacheStream.GetBuffer())
	})

	t.Run("cbID > 0, accept = false, backStream == nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()

		streamConn := adapter.NewStreamConn(syncConn, session)
		syncConn.SetNext(streamConn)
		(&session.channels[10%len(session.channels)]).In(10)

		stream := core.NewStream()
		stream.SetCallbackID(10)
		session.OnConnOpen(streamConn)
		netConn.writeBuffer = make([]byte, 0)
		session.OnConnReadStream(streamConn, stream)
		assert(len(netConn.writeBuffer)).Equal(0)
	})

	t.Run("cbID == 0, kind err", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, _ := prepareTestSession()
		err := (*base.Error)(nil)
		streamConn := adapter.NewStreamConn(syncConn, session)
		session.gateway.onError = func(sessionID uint64, e *base.Error) {
			err = e
		}
		session.OnConnReadStream(streamConn, core.NewStream())
		assert(err).Equal(errors.ErrStream)
	})

	t.Run("cbID == 0, kind == ControlStreamPing", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()
		syncConn.OnOpen()
		netConn.writeBuffer = make([]byte, 0)
		streamConn := adapter.NewStreamConn(syncConn, session)
		syncConn.SetNext(streamConn)
		sendStream := core.NewStream()
		sendStream.WriteInt64(core.ControlStreamPing)
		session.OnConnReadStream(streamConn, sendStream)
		backStream := core.NewStream()
		backStream.PutBytesTo(netConn.writeBuffer, 0)
		assert(backStream.ReadInt64()).Equal(int64(core.ControlStreamPong), nil)
		assert(backStream.IsReadFinish()).IsTrue()
	})

	t.Run("cbID == 0, kind == ControlStreamPing with err", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, _ := prepareTestSession()
		streamConn := adapter.NewStreamConn(syncConn, session)

		stream := core.NewStream()
		stream.WriteInt64(core.ControlStreamPing)
		stream.WriteBool(true)

		err := (*base.Error)(nil)
		session.gateway.onError = func(sessionID uint64, e *base.Error) {
			err = e
		}
		session.OnConnReadStream(streamConn, stream)
		assert(err).Equal(errors.ErrStream)
	})
}
