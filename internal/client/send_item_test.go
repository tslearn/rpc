package client

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"testing"
	"time"
)

func TestSendItem_NewSendItem(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSendItem(5432)
		assert(v.isRunning).IsTrue()
		assert(base.TimeNow().UnixNano()-v.startTimeNS < int64(time.Second)).
			IsTrue()
		assert(base.TimeNow().UnixNano()-v.startTimeNS >= 0).IsTrue()
		assert(v.sendTimeNS).Equal(int64(0))
		assert(v.timeoutNS).Equal(int64(5432))
		assert(len(v.returnCH)).Equal(0)
		assert(cap(v.returnCH)).Equal(1)
		assert(v.sendStream).IsNotNil()
		assert(v.next).IsNil()
	})
}

func TestSendItem_Return(t *testing.T) {
	t.Run("stream is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSendItem(0)
		assert(v.Return(nil)).IsFalse()
		assert(len(v.returnCH)).Equal(0)
	})

	t.Run("item is not running", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSendItem(0)
		v.isRunning = false
		assert(v.Return(core.NewStream())).IsFalse()
		assert(len(v.returnCH)).Equal(0)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSendItem(0)
		stream := core.NewStream()
		assert(v.Return(stream)).IsTrue()
		assert(len(v.returnCH)).Equal(1)
		assert(<-v.returnCH).Equal(stream)
	})
}

func TestSendItem_CheckTime(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSendItem(int64(time.Millisecond))
		v.sendStream.SetCallbackID(15)
		time.Sleep(100 * time.Millisecond)
		assert(v.CheckTime(base.TimeNow().UnixNano())).IsTrue()
		assert(v.isRunning).IsFalse()
		assert(len(v.returnCH)).Equal(1)
		stream := <-v.returnCH
		assert(stream.GetCallbackID()).Equal(uint64(15))
		assert(stream.ReadUint64()).
			Equal(errors.ErrClientTimeout.GetCode(), nil)
		assert(stream.ReadString()).
			Equal(errors.ErrClientTimeout.GetMessage(), nil)
		assert(stream.IsReadFinish()).IsTrue()
	})

	t.Run("it is not timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSendItem(int64(time.Second))
		v.sendStream.SetCallbackID(15)
		time.Sleep(10 * time.Millisecond)
		assert(v.CheckTime(base.TimeNow().UnixNano())).IsFalse()
		assert(v.isRunning).IsTrue()
		assert(len(v.returnCH)).Equal(0)
	})

	t.Run("it is not running", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSendItem(int64(time.Millisecond))
		v.sendStream.SetCallbackID(15)
		time.Sleep(10 * time.Millisecond)
		v.isRunning = false
		assert(v.CheckTime(base.TimeNow().UnixNano())).IsFalse()
		assert(v.isRunning).IsFalse()
		assert(len(v.returnCH)).Equal(0)
	})
}

func TestSendItem_Release(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSendItem(0)
		for i := 0; i < 10000; i++ {
			v.sendStream.PutBytes([]byte{1})
		}

		v.Release()
		assert(v.sendStream.GetWritePos()).Equal(core.StreamHeadSize)
	})

	t.Run("test put back", func(t *testing.T) {
		assert := base.NewAssert(t)
		mp := map[string]bool{}
		for i := 0; i < 1000; i++ {
			v := NewSendItem(0)
			mp[fmt.Sprintf("%p", v)] = true
			v.Release()
		}
		assert(len(mp) < 1000).IsTrue()
	})
}
