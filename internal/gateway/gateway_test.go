package gateway

import (
	"crypto/tls"
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"github.com/rpccloud/rpc/internal/route"
	"testing"
)

type fakeSender struct {
	receiver *route.IRouteReceiver
}

func (p *fakeSender) SendStreamToRouter(stream *core.Stream) *base.Error {
	return (*p.receiver).ReceiveStreamFromRouter(stream)
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
		assert(v.sessionSeed).Equal(uint64(1))
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

func TestGateWay_addSession(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		onError := func(sessionID uint64, err *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)

		for i := uint64(1); i < 100; i++ {
			session := NewSession(i, v)
			assert(v.addSession(session)).IsTrue()
		}

		for i := uint64(1); i < 100; i++ {
			session := NewSession(i, v)
			assert(v.addSession(session)).IsFalse()
		}
	})
}

func TestGateWay_getSession(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		onError := func(sessionID uint64, err *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)

		for i := uint64(1); i < 100; i++ {
			session := NewSession(i, v)
			assert(v.addSession(session)).IsTrue()
		}

		for i := uint64(1); i < 100; i++ {
			s, ok := v.getSession(i)
			assert(s).IsNotNil()
			assert(ok).IsTrue()
		}

		for i := uint64(100); i < 200; i++ {
			s, ok := v.getSession(i)
			assert(s).IsNil()
			assert(ok).IsFalse()
		}
	})
}

func TestGateWay_TimeCheck(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		onError := func(sessionID uint64, err *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		for i := uint64(1); i <= sessionManagerVectorSize; i++ {
			session := NewSession(i, v)
			session.activeTimeNS = 0
			assert(v.addSession(session)).IsTrue()
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
		assert(err).Equal(errors.ErrGatewayAlreadyRunning)
	})

	t.Run("gateway is not running", func(t *testing.T) {
		assert := base.NewAssert(t)
		tlsConfig := &tls.Config{}
		onError := func(sessionID uint64, e *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		assert(v.Listen("tcp", "0.0.0.0:8080", tlsConfig)).Equal(v)
		assert(len(v.adapters)).Equal(1)
		assert(v.adapters[0]).Equal(adapter.NewServerAdapter(
			"tcp",
			"0.0.0.0:8080",
			tlsConfig,
			v.config.serverReadBufferSize,
			v.config.serverWriteBufferSize,
			v,
		))
	})
}

func TestGateWay_ReceiveStreamFromRouter(t *testing.T) {
	t.Run("session is exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		onError := func(sessionID uint64, e *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		v.addSession(NewSession(10, v))
		stream := core.NewStream()
		stream.SetSessionID(10)
		assert(v.ReceiveStreamFromRouter(stream)).IsNil()
	})

	t.Run("session is not exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		onError := func(sessionID uint64, e *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		v.addSession(NewSession(10, v))
		stream := core.NewStream()
		stream.SetSessionID(11)
		assert(v.ReceiveStreamFromRouter(stream)).
			Equal(errors.ErrGateWaySessionNotFound)
	})
}

func TestGateWay_OnConnOpen(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		onError := func(sessionID uint64, e *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		v.addSession(NewSession(10, v))
		assert(base.RunWithCatchPanic(func() {
			v.OnConnOpen(nil)
		})).IsNil()
	})
}

func TestGateWay_OnConnReadStream(t *testing.T) {

}

func TestGateWay_OnConnError(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		onError := func(sessionID uint64, e *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		v.addSession(NewSession(10, v))
		assert(base.RunWithCatchPanic(func() {
			v.OnConnError(nil, errors.ErrStream)
		})).Equal("kernel error: it should not be called")
	})
}

func TestGateWay_OnConnClose(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		onError := func(sessionID uint64, e *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), &fakeRouter{}, onError)
		v.addSession(NewSession(10, v))
		assert(base.RunWithCatchPanic(func() {
			v.OnConnClose(nil)
		})).Equal("kernel error: it should not be called")
	})
}
