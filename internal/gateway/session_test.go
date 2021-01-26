package gateway

import (
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
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

func prepareTestSession() (*Session, adapter.IConn, *testNetConn) {
	gateway := NewGateWay(
		0,
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
