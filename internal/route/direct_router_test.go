package route

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"testing"
)

type testReceiver struct {
	emulateError bool
	streamCH     chan *core.Stream
}

func newTestReceiver(emulateError bool) *testReceiver {
	return &testReceiver{
		emulateError: emulateError,
		streamCH:     make(chan *core.Stream, 1024),
	}
}

func (p *testReceiver) ReceiveStreamFromRouter(s *core.Stream) *base.Error {
	if p.emulateError {
		return errors.ErrStream
	}

	p.streamCH <- s
	return nil
}

func TestDirectRouterSender_SendStreamToRouter(t *testing.T) {
	t.Run("ReceiveStreamFromRouter return error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := IRouteReceiver(newTestReceiver(true))
		v := &DirectRouterSender{receiver: &receiver}
		sendStream := core.NewStream()
		assert(v.SendStreamToRouter(sendStream)).Equal(errors.ErrStream)
	})

	t.Run("ReceiveStreamFromRouter return ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := IRouteReceiver(newTestReceiver(false))
		v := &DirectRouterSender{receiver: &receiver}
		sendStream := core.NewStream()
		assert(v.SendStreamToRouter(sendStream)).Equal(nil)
		assert(len(receiver.(*testReceiver).streamCH)).Equal(1)
		assert(<-receiver.(*testReceiver).streamCH).Equal(sendStream)
	})
}

func TestNewDirectRouter(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewDirectRouter()
		assert(len(v.receivers)).Equal(2)
	})
}

func TestDirectRouter_Plug(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewDirectRouter()
		receiver1 := IRouteReceiver(newTestReceiver(false))
		receiver2 := IRouteReceiver(newTestReceiver(false))
		receiver3 := IRouteReceiver(newTestReceiver(false))

		sender1 := v.Plug(receiver1)
		sender2 := v.Plug(receiver2)
		assert(sender1).Equal(&DirectRouterSender{receiver: &receiver2})
		assert(sender2).Equal(&DirectRouterSender{receiver: &receiver1})

		assert(base.RunWithCatchPanic(func() {
			v.Plug(receiver3)
		})).Equal("DirectRouter can only be plugged twice")
	})
}
