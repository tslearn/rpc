package gateway

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
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
		router := &fakeRouter{}
		onError := func(sessionID uint64, err *base.Error) {}
		v := NewGateWay(132, GetDefaultConfig(), router, onError)
		v.totalSessions = 54321
		assert(v.TotalSessions()).Equal(int64(54321))
	})
}

func TestGateWay_Get(t *testing.T) {

}
