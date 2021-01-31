package server

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"github.com/rpccloud/rpc/internal/route"
	"testing"
	"time"
)

type testReceiver struct {
	streamCH chan *core.Stream
}

func newTestReceiver() *testReceiver {
	return &testReceiver{
		streamCH: make(chan *core.Stream, 1024),
	}
}

func (p *testReceiver) ReceiveStreamFromRouter(s *core.Stream) {
	p.streamCH <- s
}

func TestNewRPCProcessor(t *testing.T) {
	t.Run("numOfThreads <= 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(NewRPCProcessor(
			route.NewDirectRouter(),
			0,
			128,
			128,
			4096,
			nil,
			3*time.Second,
			[]*core.ServiceMeta{},
		)).Equal(nil, errors.ErrNumOfThreadsIsWrong)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestReceiver()
		router := route.NewDirectRouter()
		router.Plug(receiver)
		v, err := NewRPCProcessor(
			router,
			1024,
			128,
			128,
			4096,
			nil,
			3*time.Second,
			[]*core.ServiceMeta{},
		)
		assert(err).IsNil()
		assert(v).IsNotNil()
		assert(v.processor).IsNotNil()
		v.processor.PutStream(core.NewStream())
		assert(<-receiver.streamCH).IsNotNil()
		v.Close()
	})
}

func TestRPCProcessor_ReceiveStreamFromRouter(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestReceiver()
		router := route.NewDirectRouter()
		router.Plug(receiver)
		v, err := NewRPCProcessor(
			router,
			1024,
			128,
			128,
			4096,
			nil,
			3*time.Second,
			[]*core.ServiceMeta{},
		)
		assert(err).IsNil()
		assert(v).IsNotNil()
		assert(v.processor).IsNotNil()
		v.ReceiveStreamFromRouter(core.NewStream())
		assert(<-receiver.streamCH).IsNotNil()
		v.Close()
	})
}

func TestRPCProcessor_Close(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		router := route.NewDirectRouter()
		v, err := NewRPCProcessor(
			router,
			1024,
			128,
			128,
			4096,
			nil,
			3*time.Second,
			[]*core.ServiceMeta{},
		)
		assert(err).IsNil()
		assert(v).IsNotNil()
		assert(base.RunWithCatchPanic(func() {
			v.Close()
		})).IsNil()
	})
}
