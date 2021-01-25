package gateway

import (
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/router"
	"net"
	"testing"
	"time"
)

type testNetConn struct {
	writeCH   chan []byte
	isRunning bool
}

func newTestNetConn() *testNetConn {
	return &testNetConn{
		isRunning: true,
		writeCH:   make(chan []byte, 1024),
	}
}

func (p *testNetConn) Read(_ []byte) (n int, err error) {
	panic("not implemented")
}

func (p *testNetConn) Write(b []byte) (n int, err error) {
	buf := make([]byte, len(b))
	copy(buf, b)
	p.writeCH <- buf
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
	//
	//syncConn.OnOpen()
	//for {
	//    if !syncConn.OnReadReady() {
	//        break
	//    }
	//}
	//syncConn.OnClose()
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
		session.TimeCheck(base.TimeNow().UnixNano())
		assert(session.gateway.TotalSessions()).Equal(int64(1))
	})

	t.Run("p.conn is nil, session is timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, _ := prepareTestSession()
		gw := session.gateway
		gw.config.serverSessionTimeout = 10 * time.Millisecond
		time.Sleep(20 * time.Millisecond)
		session.TimeCheck(base.TimeNow().UnixNano())
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
