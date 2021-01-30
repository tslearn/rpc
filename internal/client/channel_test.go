package client

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"testing"
	"time"
)

func TestChannel_Use(t *testing.T) {
	t.Run("p.item != nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 642, item: &SendItem{}}
		assert(v.Use(&SendItem{}, 32)).IsFalse()
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 642}
		item := NewSendItem(0)
		assert(v.Use(item, 32)).IsTrue()
		assert(v.sequence).Equal(uint64(674))
		assert(v.item).Equal(item)
		assert(item.sendStream.GetCallbackID()).Equal(uint64(674))
		nowNS := base.TimeNow().UnixNano()
		assert(nowNS-v.item.sendTimeNS < int64(time.Second)).IsTrue()
		assert(nowNS-v.item.sendTimeNS > 0).IsTrue()
	})
}

func TestChannel_Free(t *testing.T) {
	t.Run("p.item == nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 642, item: nil}
		assert(v.Free(core.NewStream())).IsFalse()
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		item := NewSendItem(0)
		v := &Channel{sequence: 642, item: item}
		stream := core.NewStream()
		assert(len(item.returnCH)).Equal(0)
		assert(v.Free(stream)).IsTrue()
		assert(v.item).IsNil()
		assert(len(item.returnCH)).Equal(1)
		assert(<-item.returnCH).Equal(stream)
	})
}

func TestChannel_CheckTime(t *testing.T) {
	t.Run("p.item == nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 642, item: nil}
		assert(v.CheckTime(base.TimeNow().UnixNano())).IsFalse()
	})

	t.Run("CheckTime return false", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 642, item: NewSendItem(int64(time.Second))}
		assert(v.CheckTime(base.TimeNow().UnixNano())).IsFalse()
	})

	t.Run("CheckTime return true", func(t *testing.T) {
		assert := base.NewAssert(t)
		item := NewSendItem(int64(time.Millisecond))
		v := &Channel{sequence: 642, item: item}

		time.Sleep(10 * time.Millisecond)
		assert(len(item.returnCH)).Equal(0)
		assert(v.CheckTime(base.TimeNow().UnixNano())).IsTrue()
		assert(v.item).IsNil()
		assert(len(item.returnCH)).Equal(1)
	})
}
