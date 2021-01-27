package gateway

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/router"
	"testing"
)

type fakeSender struct {
	receiver *router.IRouteReceiver
}

func (p *fakeSender) SendStreamToRouter(stream *core.Stream) *base.Error {
	return (*p.receiver).ReceiveStreamFromRouter(stream)
}

type fakeRouter struct {
	isPlugged bool
	receivers [2]router.IRouteReceiver
}

func (p *fakeRouter) Plug(receiver router.IRouteReceiver) router.IRouteSender {
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
		v := NewGateWay(
			132,
			GetDefaultConfig(),
			router,
			func(sessionID uint64, err *base.Error) {

			},
		)
		assert(router.isPlugged).Equal(true)
		assert(v.id).Equal(uint32(132))
		assert(v.isRunning).Equal(false)

	})
}
