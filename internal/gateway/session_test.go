package gateway

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
	"net"
	"testing"
	"time"
)

func newSession(id uint64, gateway *GateWay) *Session {
	return &Session{
		id:           id,
		gateway:      gateway,
		security:     base.GetRandString(32),
		conn:         nil,
		channels:     make([]Channel, gateway.config.numOfChannels),
		activeTimeNS: base.TimeNow().UnixNano(),
		prev:         nil,
		next:         nil,
	}
}

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
		3,
		GetDefaultConfig(),
		rpc.NewTestStreamHub(),
	)
	session := newSession(11, gateway)
	netConn := newTestNetConn()
	syncConn := adapter.NewServerSyncConn(netConn, 1200, 1200)
	streamConn := adapter.NewStreamConn(false, syncConn, session)
	syncConn.SetNext(streamConn)
	gateway.AddSession(session)
	return session, syncConn, netConn
}

func TestInitSession(t *testing.T) {
	t.Run("stream callbackID != 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn()
		streamHub := rpc.NewTestStreamHub()
		gw := NewGateWay(132, GetDefaultConfig(), streamHub)

		streamConn := adapter.NewStreamConn(
			false,
			adapter.NewServerSyncConn(netConn, 1200, 1200),
			gw,
		)

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindConnectRequest)
		stream.SetCallbackID(1)
		stream.BuildStreamCheck()
		streamConn.OnReadBytes(stream.GetBuffer())
		assert(netConn.isRunning).IsFalse()
		assert(rpc.ParseResponseStream(streamHub.GetStream())).
			Equal(nil, base.ErrStream)
	})

	t.Run("kind is not StreamKindConnectRequest", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn()
		streamHub := rpc.NewTestStreamHub()
		gw := NewGateWay(132, GetDefaultConfig(), streamHub)

		streamConn := adapter.NewStreamConn(
			false,
			adapter.NewServerSyncConn(netConn, 1200, 1200),
			gw,
		)

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindConnectResponse)
		stream.BuildStreamCheck()
		streamConn.OnReadBytes(stream.GetBuffer())
		assert(netConn.isRunning).IsFalse()
		assert(rpc.ParseResponseStream(streamHub.GetStream())).
			Equal(nil, base.ErrStream)
	})

	t.Run("read session string error", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn()
		streamHub := rpc.NewTestStreamHub()
		gw := NewGateWay(132, GetDefaultConfig(), streamHub)

		streamConn := adapter.NewStreamConn(
			false,
			adapter.NewServerSyncConn(netConn, 1200, 1200),
			gw,
		)

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindConnectRequest)
		stream.WriteBool(true)
		stream.BuildStreamCheck()
		streamConn.OnReadBytes(stream.GetBuffer())
		assert(netConn.isRunning).IsFalse()
		assert(rpc.ParseResponseStream(streamHub.GetStream())).
			Equal(nil, base.ErrStream)
	})

	t.Run("read stream is not finish", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn()
		streamHub := rpc.NewTestStreamHub()
		gw := NewGateWay(132, GetDefaultConfig(), streamHub)

		streamConn := adapter.NewStreamConn(
			false,
			adapter.NewServerSyncConn(netConn, 1200, 1200),
			gw,
		)

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindConnectRequest)
		stream.WriteString("")
		stream.WriteBool(false)
		stream.BuildStreamCheck()
		streamConn.OnReadBytes(stream.GetBuffer())
		assert(netConn.isRunning).IsFalse()
		assert(rpc.ParseResponseStream(streamHub.GetStream())).
			Equal(nil, base.ErrStream)
	})

	t.Run("max sessions limit", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamHub := rpc.NewTestStreamHub()
		gw := NewGateWay(132, GetDefaultConfig(), streamHub)
		gw.config.serverMaxSessions = 1
		gw.AddSession(&Session{
			id:       234,
			security: "12345678123456781234567812345678",
		})

		syncConn := adapter.NewServerSyncConn(newTestNetConn(), 1200, 1200)
		streamConn := adapter.NewStreamConn(false, syncConn, gw)
		syncConn.SetNext(streamConn)

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindConnectRequest)
		stream.WriteString("")
		stream.BuildStreamCheck()
		streamConn.OnReadBytes(stream.GetBuffer())

		assert(rpc.ParseResponseStream(streamHub.GetStream())).
			Equal(nil, base.ErrGateWaySeedOverflows)
	})

	t.Run("stream is ok, create new session", func(t *testing.T) {
		assert := base.NewAssert(t)
		id := uint64(234)
		security := "12345678123456781234567812345678"
		testCollection := map[string]bool{
			"234-12345678123456781234567812345678":   true,
			"0234-12345678123456781234567812345678":  true,
			"":                                       false,
			"-":                                      false,
			"-S":                                     false,
			"-SecurityPasswordSecurityPass":          false,
			"-SecurityPasswordSecurityPassword":      false,
			"-SecurityPasswordSecurityPasswordEx":    false,
			"*-S":                                    false,
			"*-SecurityPasswordSecurityPassword":     false,
			"*-SecurityPasswordSecurityPasswordEx":   false,
			"ABC-S":                                  false,
			"ABC-SecurityPasswordSecurityPassword":   false,
			"ABC-SecurityPasswordSecurityPasswordEx": false,
			"1-S":                                    false,
			"1-SecurityPasswordSecurityPassword":     false,
			"1-SecurityPasswordSecurityPasswordEx":   false,
			"-234-SecurityPasswordSecurityPassword":  false,
			"234-":                                   false,
			"234-S":                                  false,
			"234-SecurityPasswordSecurityPassword":   false,
			"234-SecurityPasswordSecurityPasswordEx": false,
			"-234-":                                  false,
			"-234-234-":                              false,
			"------":                                 false,
		}

		for connStr, exist := range testCollection {
			gw := NewGateWay(
				132, GetDefaultConfig(), rpc.NewTestStreamHub(),
			)
			gw.AddSession(&Session{id: id, security: security, gateway: gw})
			netConn := newTestNetConn()
			syncConn := adapter.NewServerSyncConn(netConn, 1200, 1200)
			streamConn := adapter.NewStreamConn(false, syncConn, gw)
			syncConn.SetNext(streamConn)

			stream := rpc.NewStream()
			stream.SetKind(rpc.StreamKindConnectRequest)
			stream.WriteString(connStr)
			stream.BuildStreamCheck()
			streamConn.OnReadBytes(stream.GetBuffer())

			if exist {
				assert(gw.totalSessions).Equal(int64(1))
			} else {
				assert(gw.totalSessions).Equal(int64(2))
				v, _ := gw.GetSession(1)
				assert(v.id).Equal(uint64(1))
				assert(v.gateway).Equal(gw)
				assert(len(v.security)).Equal(32)
				assert(v.conn).IsNotNil()
				assert(len(v.channels)).Equal(GetDefaultConfig().numOfChannels)
				assert(cap(v.channels)).Equal(GetDefaultConfig().numOfChannels)
				nowNS := base.TimeNow().UnixNano()
				assert(nowNS-v.activeTimeNS < int64(time.Second)).IsTrue()
				assert(nowNS-v.activeTimeNS >= 0).IsTrue()
				assert(v.prev).IsNil()
				assert(v.next).IsNil()

				rs := rpc.NewStream()
				rs.PutBytesTo(netConn.writeBuffer, 0)

				cfg := gw.config

				assert(rs.GetKind()).
					Equal(uint8(rpc.StreamKindConnectResponse))
				assert(rs.ReadString()).
					Equal(fmt.Sprintf("%d-%s", v.id, v.security), nil)
				assert(rs.ReadInt64()).Equal(int64(cfg.numOfChannels), nil)
				assert(rs.ReadInt64()).Equal(int64(cfg.transLimit), nil)
				assert(rs.ReadInt64()).
					Equal(int64(cfg.heartbeat/time.Millisecond), nil)
				assert(rs.ReadInt64()).
					Equal(int64(cfg.heartbeatTimeout/time.Millisecond), nil)
				assert(rs.IsReadFinish()).IsTrue()
				assert(rs.CheckStream()).IsTrue()
			}
		}
	})
}

func TestNewSession(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		gateway := NewGateWay(
			43,
			GetDefaultConfig(),
			rpc.NewTestStreamHub(),
		)
		v := newSession(3, gateway)
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
		session.gateway.config.heartbeatTimeout = 1 * time.Millisecond
		syncConn.OnOpen()
		time.Sleep(30 * time.Millisecond)
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
		gw.config.serverSessionTimeout = 1 * time.Millisecond
		time.Sleep(30 * time.Millisecond)
		gw.TimeCheck(base.TimeNow().UnixNano())
		assert(gw.TotalSessions()).Equal(int64(0))
	})

	t.Run("p.channels is not timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, _ := prepareTestSession()

		// fill the channels
		for i := 0; i < session.gateway.config.numOfChannels; i++ {
			stream := rpc.NewStream()
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
			stream := rpc.NewStream()
			stream.SetCallbackID(uint64(i) + 1)
			session.channels[i].In(stream.GetCallbackID())
			session.channels[i].Out(stream)
		}

		session.gateway.config.serverCacheTimeout = 1 * time.Millisecond
		time.Sleep(30 * time.Millisecond)
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
		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindRPCResponseOK)
		session.OutStream(stream)
		assert(len(netConn.writeBuffer)).Equal(0)
		assert(session.channels[0].backStream).Equal(stream)
	})

	t.Run("stream kind error", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()
		syncConn.OnOpen()
		// ignore the init stream
		netConn.writeBuffer = make([]byte, 0)

		for i := 1; i <= len(session.channels); i++ {
			(&session.channels[i%len(session.channels)]).In(uint64(i))
		}

		for i := 1; i <= len(session.channels); i++ {
			stream := rpc.NewStream()
			stream.SetKind(rpc.StreamKindConnectResponse)
			stream.SetCallbackID(uint64(i))
			session.OutStream(stream)
		}

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
			stream := rpc.NewStream()
			stream.SetKind(rpc.StreamKindRPCResponseOK)
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
			stream := rpc.NewStream()
			stream.SetCallbackID(uint64(i))
			stream.SetKind(rpc.StreamKindRPCResponseOK)
			stream.BuildStreamCheck()
			exceptBuffer = append(exceptBuffer, stream.GetBuffer()...)
			session.OutStream(stream)
		}

		assert(netConn.writeBuffer).Equal(exceptBuffer)
	})

	t.Run("stream is StreamKindRPCBoardCast", func(t *testing.T) {
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
			stream := rpc.NewStream()
			stream.SetGatewayID(1234)
			stream.SetSessionID(5678)
			stream.SetKind(rpc.StreamKindRPCBoardCast)
			stream.WriteString("#.test%Msg")
			stream.WriteString("HI")
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
		session, syncConn, _ := prepareTestSession()
		streamConn := adapter.NewStreamConn(false, syncConn, session)
		syncConn.SetNext(streamConn)
		session.OnConnOpen(streamConn)
		assert(session.conn).Equal(streamConn)
	})
}

func TestSession_OnConnReadStream(t *testing.T) {
	t.Run("cbID == 0, kind == StreamKindPing ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()
		syncConn.OnOpen()
		netConn.writeBuffer = make([]byte, 0)
		streamConn := adapter.NewStreamConn(false, syncConn, session)
		syncConn.SetNext(streamConn)
		sendStream := rpc.NewStream()
		sendStream.SetKind(rpc.StreamKindPing)
		session.OnConnReadStream(streamConn, sendStream)
		backStream := rpc.NewStream()
		backStream.PutBytesTo(netConn.writeBuffer, 0)
		assert(backStream.GetKind()).Equal(uint8(rpc.StreamKindPong))
		assert(backStream.IsReadFinish()).IsTrue()
	})

	t.Run("cbID == 0, kind == StreamKindPing error", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()
		syncConn.OnOpen()
		netConn.writeBuffer = make([]byte, 0)
		streamConn := adapter.NewStreamConn(false, syncConn, session)
		syncConn.SetNext(streamConn)
		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindPing)
		stream.WriteBool(true)

		streamHub := rpc.NewTestStreamHub()
		session.gateway.streamHub = streamHub
		session.OnConnReadStream(streamConn, stream)
		assert(rpc.ParseResponseStream(streamHub.GetStream())).
			Equal(nil, base.ErrStream)
	})

	t.Run("cbID > 0, accept = true, backStream = nil", func(t *testing.T) {
		assert := base.NewAssert(t)

		streamHub := rpc.NewTestStreamHub()
		session, syncConn, _ := prepareTestSession()
		session.gateway.streamHub = streamHub

		streamConn := adapter.NewStreamConn(false, syncConn, session)
		stream := rpc.NewStream()
		stream.SetCallbackID(10)
		stream.SetKind(rpc.StreamKindRPCRequest)
		session.OnConnReadStream(streamConn, stream)

		backStream := streamHub.GetStream()
		assert(backStream.GetGatewayID()).Equal(uint32(3))
		assert(backStream.GetSessionID()).Equal(uint64(11))
	})

	t.Run("cbID > 0, accept = false, backStream != nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()

		cacheStream := rpc.NewStream()
		cacheStream.SetCallbackID(10)
		cacheStream.BuildStreamCheck()

		streamConn := adapter.NewStreamConn(false, syncConn, session)
		syncConn.SetNext(streamConn)

		(&session.channels[10%len(session.channels)]).In(10)
		(&session.channels[10%len(session.channels)]).Out(cacheStream)

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindRPCRequest)
		stream.SetCallbackID(10)
		session.OnConnOpen(streamConn)
		netConn.writeBuffer = make([]byte, 0)
		session.OnConnReadStream(streamConn, stream)
		assert(netConn.writeBuffer).Equal(cacheStream.GetBuffer())
	})

	t.Run("cbID > 0, accept = false, backStream == nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()

		streamConn := adapter.NewStreamConn(false, syncConn, session)
		syncConn.SetNext(streamConn)
		(&session.channels[10%len(session.channels)]).In(10)

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindRPCRequest)
		stream.SetCallbackID(10)
		session.OnConnOpen(streamConn)
		netConn.writeBuffer = make([]byte, 0)
		session.OnConnReadStream(streamConn, stream)
		assert(len(netConn.writeBuffer)).Equal(0)
	})

	t.Run("cbID == 0, accept = true, backStream = nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, _ := prepareTestSession()
		streamHub := rpc.NewTestStreamHub()
		session.gateway.streamHub = streamHub

		streamConn := adapter.NewStreamConn(false, syncConn, session)
		stream := rpc.NewStream()
		stream.SetCallbackID(0)
		stream.SetKind(rpc.StreamKindRPCRequest)

		session.OnConnReadStream(streamConn, stream)
		assert(rpc.ParseResponseStream(streamHub.GetStream())).
			Equal(nil, base.ErrStream)
	})

	t.Run("cbID == 0, kind err", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, _ := prepareTestSession()
		streamConn := adapter.NewStreamConn(false, syncConn, session)
		streamHub := rpc.NewTestStreamHub()
		session.gateway.streamHub = streamHub
		session.OnConnReadStream(streamConn, rpc.NewStream())
		assert(rpc.ParseResponseStream(streamHub.GetStream())).
			Equal(nil, base.ErrStream)
	})
}

func TestSession_OnConnError(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, _ := prepareTestSession()
		streamConn := adapter.NewStreamConn(false, syncConn, session)
		streamHub := rpc.NewTestStreamHub()
		session.gateway.streamHub = streamHub
		session.OnConnError(streamConn, base.ErrStream)
		assert(rpc.ParseResponseStream(streamHub.GetStream())).
			Equal(nil, base.ErrStream)
	})
}

func TestSession_OnConnClose(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, _ := prepareTestSession()
		streamConn := adapter.NewStreamConn(false, syncConn, session)
		session.OnConnOpen(streamConn)
		assert(session.conn).IsNotNil()
		session.OnConnClose(streamConn)
		assert(session.conn).IsNil()
	})
}

func TestNewSessionPool(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		gateway := &GateWay{}
		v := NewSessionPool(gateway)
		assert(v.gateway).Equal(gateway)
		assert(len(v.idMap)).Equal(0)
		assert(v.head).IsNil()
	})
}

func TestSessionPool_Add(t *testing.T) {
	t.Run("session is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionPool(&GateWay{})
		assert(v.Add(nil)).IsFalse()
	})

	t.Run("session has already exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionPool(&GateWay{})
		session := &Session{id: 10}
		assert(v.Add(session)).IsTrue()
		assert(v.Add(session)).IsFalse()
		assert(v.gateway.totalSessions).Equal(int64(1))
	})

	t.Run("add one session", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionPool(&GateWay{})
		session := &Session{id: 10}
		assert(v.Add(session)).IsTrue()
		assert(v.gateway.totalSessions).Equal(int64(1))
		assert(v.head).Equal(session)
		assert(session.prev).Equal(nil)
		assert(session.next).Equal(nil)
	})

	t.Run("add two sessions", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionPool(&GateWay{})
		session1 := &Session{id: 11}
		session2 := &Session{id: 12}
		assert(v.Add(session1)).IsTrue()
		assert(v.Add(session2)).IsTrue()
		assert(v.gateway.totalSessions).Equal(int64(2))
		assert(v.head).Equal(session2)
		assert(session2.prev).Equal(nil)
		assert(session2.next).Equal(session1)
		assert(session1.prev).Equal(session2)
		assert(session1.next).Equal(nil)
	})
}

func TestSessionPool_Get(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionPool(&GateWay{})
		session1 := &Session{id: 11}
		session2 := &Session{id: 12}
		assert(v.Add(session1)).IsTrue()
		assert(v.Add(session2)).IsTrue()
		assert(v.gateway.totalSessions).Equal(int64(2))
		assert(v.Get(11)).Equal(session1, true)
		assert(v.Get(12)).Equal(session2, true)
		assert(v.Get(10)).Equal(nil, false)
	})
}

func TestSessionPool_TimeCheck(t *testing.T) {
	fnTest := func(pos int) bool {
		nowNS := base.TimeNow().UnixNano()
		v := NewSessionPool(&GateWay{config: GetDefaultConfig()})

		s1 := &Session{id: 1, activeTimeNS: nowNS}
		s2 := &Session{id: 2, activeTimeNS: nowNS}
		s3 := &Session{id: 3, activeTimeNS: nowNS}
		v.Add(s3)
		v.Add(s2)
		v.Add(s1)

		firstSession := (*Session)(nil)
		lastSession := (*Session)(nil)

		if pos == 1 {
			s1.activeTimeNS = 0
			firstSession = s2
			lastSession = s3
		} else if pos == 2 {
			s2.activeTimeNS = 0
			firstSession = s1
			lastSession = s3
		} else if pos == 3 {
			s3.activeTimeNS = 0
			firstSession = s1
			lastSession = s2
		} else {
			v.TimeCheck(nowNS)
			return v.gateway.totalSessions == 3 && v.head == s1 &&
				s1.next == s2 && s2.next == s3 && s3.next == nil &&
				s3.prev == s2 && s2.prev == s1 && s1.prev == nil
		}

		v.TimeCheck(nowNS)

		return v.gateway.totalSessions == 2 && v.head == firstSession &&
			firstSession.next == lastSession && lastSession.next == nil &&
			lastSession.prev == firstSession && firstSession.prev == nil
	}

	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(fnTest(0)).IsTrue()
		assert(fnTest(1)).IsTrue()
		assert(fnTest(2)).IsTrue()
		assert(fnTest(3)).IsTrue()
	})
}
