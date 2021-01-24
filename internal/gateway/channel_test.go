package gateway

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"testing"
	"time"
)

func TestChannel_In(t *testing.T) {
	t.Run("old id without back stream", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10, backStream: core.NewStream()}
		assert(v.In(0)).Equal(false, nil)
		assert(v.In(9)).Equal(false, nil)
	})

	t.Run("old id with back stream", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10, backStream: core.NewStream()}
		assert(v.In(10)).Equal(false, core.NewStream())
	})

	t.Run("new id", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10, backStream: core.NewStream(), backTimeNS: 1}
		assert(v.In(11)).Equal(true, nil)
		assert(v.backTimeNS, v.backStream).Equal(int64(0), nil)
	})
}

func TestChannel_Out(t *testing.T) {
	t.Run("id is zero", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10}
		stream := core.NewStream()
		stream.SetCallbackID(0)
		assert(v.Out(stream)).Equal(true)
		assert(v.backTimeNS, v.backStream).Equal(int64(0), nil)
	})

	t.Run("id equals sequence", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10}
		stream := core.NewStream()
		stream.SetCallbackID(10)
		assert(v.Out(stream)).Equal(true)
		assert(v.backTimeNS > 0, v.backStream != nil).Equal(true, true)
	})

	t.Run("id equals sequence, but not in", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10, backTimeNS: 10}
		stream := core.NewStream()
		stream.SetCallbackID(10)
		assert(v.Out(stream)).Equal(false)
		assert(v.backTimeNS, v.backStream).Equal(int64(10), nil)
	})

	t.Run("id is wrong 01", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10}
		stream := core.NewStream()
		stream.SetCallbackID(9)
		assert(v.Out(stream)).Equal(false)
		assert(v.backTimeNS, v.backStream).Equal(int64(0), nil)
	})

	t.Run("id is wrong 02", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10}
		stream := core.NewStream()
		stream.SetCallbackID(11)
		assert(v.Out(stream)).Equal(false)
		assert(v.backTimeNS, v.backStream).Equal(int64(0), nil)
	})
}

func TestChannel_TimeCheck(t *testing.T) {
	t.Run("backTimeNS is zero", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10, backTimeNS: 0, backStream: core.NewStream()}
		v.TimeCheck(base.TimeNow().UnixNano(), int64(time.Second))
		assert(v.backStream).IsNotNil()
	})

	t.Run("not timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		nowNS := base.TimeNow().UnixNano()
		v := &Channel{
			sequence:   10,
			backTimeNS: nowNS,
			backStream: core.NewStream(),
		}
		v.TimeCheck(nowNS, int64(100*time.Second))
		assert(v.backStream).IsNotNil()
	})

	t.Run("timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		nowNS := base.TimeNow().UnixNano()
		v := &Channel{
			sequence:   10,
			backTimeNS: nowNS - 101,
			backStream: core.NewStream(),
		}
		v.TimeCheck(nowNS, 100)
		assert(v.backStream).IsNil()
	})
}

func TestChannel_Clean(t *testing.T) {
	t.Run("backStream is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10, backTimeNS: 1, backStream: nil}
		v.Clean()
		assert(v.backTimeNS, v.backStream).Equal(int64(0), nil)
	})

	t.Run("backStream is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10, backTimeNS: 1, backStream: core.NewStream()}
		v.Clean()
		assert(v.backTimeNS, v.backStream).Equal(int64(0), nil)
	})
}
