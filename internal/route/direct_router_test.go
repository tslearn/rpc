package route

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
	"testing"
)

type testReceiver struct {
	streamCH chan *rpc.Stream
}

func newTestReceiver() *testReceiver {
	return &testReceiver{
		streamCH: make(chan *rpc.Stream, 1024),
	}
}

func (p *testReceiver) ReceiveStreamFromRouter(s *rpc.Stream) {
	p.streamCH <- s
}

func TestDirectRouterSender_SendStreamToRouter(t *testing.T) {
	t.Run("ReceiveStreamFromRouter return ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := IRouteReceiver(newTestReceiver())
		v := &DirectRouterSender{receiver: &receiver}
		sendStream := rpc.NewStream()
		v.SendStreamToRouter(sendStream)
		assert(len(receiver.(*testReceiver).streamCH)).Equal(1)
		assert(<-receiver.(*testReceiver).streamCH).Equal(sendStream)
	})
}

func TestNewDirectRouter(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(len(NewDirectRouter().receivers)).Equal(2)
	})
}

func TestDirectRouter_Plug(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewDirectRouter()
		receiver1 := IRouteReceiver(newTestReceiver())
		receiver2 := IRouteReceiver(newTestReceiver())
		receiver3 := IRouteReceiver(newTestReceiver())

		sender1 := v.Plug(receiver1)
		sender2 := v.Plug(receiver2)
		assert(sender1).Equal(&DirectRouterSender{receiver: &receiver2})
		assert(sender2).Equal(&DirectRouterSender{receiver: &receiver1})

		assert(base.RunWithCatchPanic(func() {
			v.Plug(receiver3)
		})).Equal("DirectRouter can only be plugged twice")
	})
}
