package gateway

import (
	"crypto/tls"
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/route"
	"net"
	"testing"
	"time"
)

type fakeSender struct {
	receiver *route.IRouteReceiver
}

func (p *fakeSender) SendStreamToRouter(stream *core.Stream) {
	(*p.receiver).ReceiveStreamFromRouter(stream)
}

type fakeRouter struct {
	isPlugged bool
	receivers [2]route.IRouteReceiver
}

func (p *fakeRouter) Plug(receiver route.IRouteReceiver) route.IRouteSender {
	p.isPlugged = true
	if p.receivers[0] == nil {
		p.receivers[0] = receiver
		return &fakeSender{receiver: &p.receivers[1]}
	} else if p.receivers[1] == nil {
		p.receivers[1] = receiver
		return &fakeSender{receiver: &p.receivers[0]}
	} else {
		panic("DirectRouter can only be plugged twice")
	}
}

func TestGateWayBasic(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(sessionManagerVectorSize).Equal(1024)
	})
}

func TestNewGateWay(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		router := &fakeRouter{}
		onError := func(sessionID uint64, err *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), router, onError)
		assert(router.isPlugged).Equal(true)
		assert(v.id).Equal(uint32(132))
		assert(v.isRunning).Equal(false)
		assert(v.sessionSeed).Equal(uint64(0))
		assert(v.totalSessions).Equal(int64(0))
		assert(len(v.sessionMapList)).Equal(sessionManagerVectorSize)
		assert(cap(v.sessionMapList)).Equal(sessionManagerVectorSize)
		assert(v.routeSender).Equal(&fakeSender{receiver: &router.receivers[1]})
		assert(len(v.closeCH)).Equal(0)
		assert(cap(v.closeCH)).Equal(1)
		assert(v.config).Equal(GetDefaultConfig())
		assert(v.onError).IsNotNil()
		assert(len(v.adapters)).Equal(0)
		assert(cap(v.adapters)).Equal(0)
		assert(v.orcManager).IsNotNil()

		for i := 0; i < sessionManagerVectorSize; i++ {
			assert(v.sessionMapList[i]).IsNotNil()
		}
	})
}

func TestGateWay_TotalSessions(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		onError := func(sessionID uint64, err *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		v.totalSessions = 54321
		assert(v.TotalSessions()).Equal(int64(54321))
	})
}

func TestGateWay_AddSession(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		onError := func(sessionID uint64, err *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)

		for i := uint64(1); i < 100; i++ {
			session := newSession(i, v)
			assert(v.AddSession(session)).IsTrue()
		}

		for i := uint64(1); i < 100; i++ {
			session := newSession(i, v)
			assert(v.AddSession(session)).IsFalse()
		}
	})
}

func TestGateWay_GetSession(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		onError := func(sessionID uint64, err *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)

		for i := uint64(1); i < 100; i++ {
			session := newSession(i, v)
			assert(v.AddSession(session)).IsTrue()
		}

		for i := uint64(1); i < 100; i++ {
			s, ok := v.GetSession(i)
			assert(s).IsNotNil()
			assert(ok).IsTrue()
		}

		for i := uint64(100); i < 200; i++ {
			s, ok := v.GetSession(i)
			assert(s).IsNil()
			assert(ok).IsFalse()
		}
	})
}

func TestGateWay_CreateSessionID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		onError := func(sessionID uint64, err *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		assert(v.CreateSessionID()).Equal(uint64(1))
		assert(v.CreateSessionID()).Equal(uint64(2))
	})
}

func TestGateWay_TimeCheck(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		onError := func(sessionID uint64, err *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		for i := uint64(1); i <= sessionManagerVectorSize; i++ {
			session := newSession(i, v)
			session.activeTimeNS = 0
			assert(v.AddSession(session)).IsTrue()
		}

		assert(v.TotalSessions()).Equal(int64(sessionManagerVectorSize))
		v.TimeCheck(base.TimeNow().UnixNano())
		assert(v.TotalSessions()).Equal(int64(0))
	})
}

func TestGateWay_Listen(t *testing.T) {
	t.Run("gateway is running", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		onError := func(sessionID uint64, e *base.Error) {
			err = e
		}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		v.isRunning = true
		assert(v.Listen("tcp", "0.0.0.0:8080", nil)).Equal(v)
		assert(err).Equal(base.ErrGatewayAlreadyRunning)
	})

	t.Run("gateway is not running", func(t *testing.T) {
		assert := base.NewAssert(t)
		tlsConfig := &tls.Config{}
		onError := func(sessionID uint64, e *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		assert(v.Listen("tcp", "0.0.0.0:8080", tlsConfig)).Equal(v)
		assert(len(v.adapters)).Equal(1)
		assert(v.adapters[0]).Equal(adapter.NewServerAdapter(
			false,
			"tcp",
			"0.0.0.0:8080",
			tlsConfig,
			v.config.serverReadBufferSize,
			v.config.serverWriteBufferSize,
			v,
		))
	})
}

func TestGateWay_ListenWithDebug(t *testing.T) {
	t.Run("gateway is running", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		onError := func(sessionID uint64, e *base.Error) {
			err = e
		}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		v.isRunning = true
		assert(v.ListenWithDebug("tcp", "0.0.0.0:8080", nil)).Equal(v)
		assert(err).Equal(base.ErrGatewayAlreadyRunning)
	})

	t.Run("gateway is not running", func(t *testing.T) {
		assert := base.NewAssert(t)
		tlsConfig := &tls.Config{}
		onError := func(sessionID uint64, e *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		assert(v.ListenWithDebug("tcp", "0.0.0.0:8080", tlsConfig)).Equal(v)
		assert(len(v.adapters)).Equal(1)
		assert(v.adapters[0]).Equal(adapter.NewServerAdapter(
			true,
			"tcp",
			"0.0.0.0:8080",
			tlsConfig,
			v.config.serverReadBufferSize,
			v.config.serverWriteBufferSize,
			v,
		))
	})
}

func TestGateWay_Open(t *testing.T) {
	t.Run("it is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		onError := func(sessionID uint64, e *base.Error) { err = e }
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		v.isRunning = true
		v.Open()
		assert(err).Equal(base.ErrGatewayAlreadyRunning)
	})

	t.Run("no valid adapter", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		onError := func(sessionID uint64, e *base.Error) { err = e }
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		v.Open()
		assert(err).Equal(base.ErrGatewayNoAvailableAdapter)
	})

	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		waitCH := make(chan bool)
		onError := func(sessionID uint64, e *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		v.AddSession(&Session{id: 10})
		v.Listen("tcp", "127.0.0.1:8000", nil)
		v.Listen("tcp", "127.0.0.1:8001", nil)

		go func() {
			for v.TotalSessions() == 1 {
				time.Sleep(10 * time.Millisecond)
			}
			assert(v.isRunning).IsTrue()
			_, err1 := net.Listen("tcp", "127.0.0.1:8000")
			_, err2 := net.Listen("tcp", "127.0.0.1:8001")
			assert(err1).IsNotNil()
			assert(err2).IsNotNil()
			v.Close()
			waitCH <- true
		}()
		assert(v.TotalSessions()).Equal(int64(1))
		v.Open()
		<-waitCH
	})
}

func TestGateWay_Close(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		waitCH := make(chan bool)
		onError := func(sessionID uint64, e *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		v.AddSession(&Session{id: 10})
		v.Listen("tcp", "127.0.0.1:8000", nil)

		go func() {
			for v.TotalSessions() == 1 {
				time.Sleep(10 * time.Millisecond)
			}
			assert(v.isRunning).IsTrue()
			v.Close()
			assert(v.isRunning).IsFalse()
			ln1, err1 := net.Listen("tcp", "127.0.0.1:8000")
			ln2, err2 := net.Listen("tcp", "127.0.0.1:8001")
			assert(err1).IsNil()
			assert(err2).IsNil()
			_ = ln1.Close()
			_ = ln2.Close()
			waitCH <- true
		}()

		assert(v.TotalSessions()).Equal(int64(1))
		v.Open()
		<-waitCH
	})
}

func TestGateWay_ReceiveStreamFromRouter(t *testing.T) {
	t.Run("session is exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		onError := func(sessionID uint64, e *base.Error) { err = e }
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		v.AddSession(newSession(10, v))
		stream := core.NewStream()
		stream.SetSessionID(10)
		v.ReceiveStreamFromRouter(stream)
		assert(err).IsNil()
	})

	t.Run("session is not exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		onError := func(sessionID uint64, e *base.Error) { err = e }
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		v.AddSession(newSession(10, v))
		stream := core.NewStream()
		stream.SetSessionID(11)
		v.ReceiveStreamFromRouter(stream)
		assert(err).Equal(base.ErrGateWaySessionNotFound)
	})
}

func TestGateWay_OnConnOpen(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		onError := func(sessionID uint64, e *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		v.AddSession(newSession(10, v))
		assert(base.RunWithCatchPanic(func() {
			v.OnConnOpen(nil)
		})).IsNil()
	})
}

func TestGateWay_OnConnReadStream(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		onError := func(sessionID uint64, e *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		syncConn := adapter.NewServerSyncConn(newTestNetConn(), 1200, 1200)
		streamConn := adapter.NewStreamConn(false, syncConn, v)
		syncConn.SetNext(streamConn)

		stream := core.NewStream()
		stream.SetKind(core.ControlStreamConnectRequest)
		stream.WriteString("")
		stream.BuildStreamCheck()
		streamConn.OnReadBytes(stream.GetBuffer())
		assert(v.totalSessions).Equal(int64(1))
	})
}

func TestGateWay_OnConnError(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		onError := func(sessionID uint64, e *base.Error) { err = e }
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		v.AddSession(newSession(10, v))
		netConn := newTestNetConn()
		syncConn := adapter.NewServerSyncConn(netConn, 1200, 1200)
		streamConn := adapter.NewStreamConn(false, syncConn, v)
		assert(netConn.isRunning).IsTrue()
		v.OnConnError(streamConn, base.ErrStream)
		assert(netConn.isRunning).IsFalse()
		assert(err).Equal(base.ErrStream)
	})
}

func TestGateWay_OnConnClose(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		onError := func(sessionID uint64, e *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		v.AddSession(newSession(10, v))
		assert(base.RunWithCatchPanic(func() {
			v.OnConnClose(nil)
		})).IsNil()
	})
}
