package base

import (
	"bytes"
	"os"
	"testing"
)

func TestSyncPoolDebug_Put(t *testing.T) {
	t.Run("put unmanaged value", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := &SyncPoolDebug{
			New: func() interface{} {
				ret := make([]byte, 512)
				return &ret
			},
		}
		assert(RunWithCatchPanic(func() {
			ret := make([]byte, 512)
			v1.Put(&ret)
		})).Equal("SyncPoolDebug Put check failed")
	})

	t.Run("put managed value", func(t *testing.T) {
		assert := NewAssert(t)

		v1 := &SyncPoolDebug{
			New: func() interface{} {
				ret := make([]byte, 512)
				return &ret
			},
		}
		assert(RunWithCatchPanic(func() {
			var buf bytes.Buffer
			SetLogWriter(&buf)
			defer SetLogWriter(os.Stdout)
			v1.Put(v1.Get())
		})).IsNil()
	})
}

func TestSyncPoolDebug_Get(t *testing.T) {
	t.Run("get value twice", func(t *testing.T) {
		assert := NewAssert(t)
		var buf bytes.Buffer
		assert(RunWithCatchPanic(func() {
			SetLogWriter(&buf)
			defer SetLogWriter(os.Stdout)
			v1 := &SyncPoolDebug{
				New: func() interface{} {
					return &buf
				},
			}
			v1.Get()
			v1.Get()
		})).Equal("SyncPoolDebug Get check failed")
		assert(buf.String()).Equal(
			"Warn: SyncPool is in debug mode, which may slow down the program",
		)
	})

	t.Run("get value ok", func(t *testing.T) {
		assert := NewAssert(t)
		var buf bytes.Buffer
		SetLogWriter(&buf)
		defer SetLogWriter(os.Stdout)

		v1 := &SyncPoolDebug{
			New: func() interface{} {
				ret := make([]byte, 512)
				return &ret
			},
		}
		assert(v1.Get()).IsNotNil()
		assert(buf.String()).Equal(
			"Warn: SyncPool is in debug mode, which may slow down the program",
		)
	})
}
