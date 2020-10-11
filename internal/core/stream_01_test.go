package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"testing"
)

func TestStream_ReadRTArray(t *testing.T) {
	t.Run("runtime is not available", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		type R = Runtime
		s := ""
		f := base.GetFileLine
		assert(stream.ReadRTArray((func() R { s = f(0); return R{} })())).
			Equal(RTArray{}, errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(s))
	})

}

func TestStream_ReadRTMap(t *testing.T) {
	t.Run("runtime is not available", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		type R = Runtime
		s := ""
		f := base.GetFileLine
		assert(stream.ReadRTMap((func() R { s = f(0); return R{} })())).
			Equal(RTMap{}, errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(s))
	})
}

func TestStream_ReadRTValue(t *testing.T) {
	t.Run("runtime is not available", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		type R = Runtime
		s := ""
		f := base.GetFileLine
		assert(stream.ReadRTValue((func() R { s = f(0); return R{} })())).Equal(
			RTValue{err: errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(s)},
		)
	})
}
