package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"testing"
)

func testWithRTValue(fn func(v RTValue), v interface{}) {
	testRuntime.thread.Reset()
	rtArray := testRuntime.NewRTArray(1)
	_ = rtArray.Append(v)
	fn(rtArray.Get(0))
	testRuntime.thread.Reset()
}

func TestRTValue_ToBool(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToBool()).
			Equal(false, base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToBool()).
			Equal(false, base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToBool()).Equal(false, base.ErrStream)
		}, "kitty")
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToBool()).Equal(true, nil)
		}, true)
	})
}

func TestRTValue_ToInt64(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToInt64()).
			Equal(int64(0), base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToInt64()).
			Equal(int64(0), base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToInt64()).Equal(int64(0), base.ErrStream)
		}, "kitty")
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToInt64()).Equal(int64(12), nil)
		}, 12)
	})
}

func TestRTValue_ToUint64(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToUint64()).
			Equal(uint64(0), base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToUint64()).
			Equal(uint64(0), base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToUint64()).Equal(uint64(0), base.ErrStream)
		}, "kitty")
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToUint64()).Equal(uint64(12), nil)
		}, uint64(12))
	})
}

func TestRTValue_ToFloat64(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToFloat64()).
			Equal(float64(0), base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToFloat64()).
			Equal(float64(0), base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToFloat64()).Equal(float64(0), base.ErrStream)
		}, "kitty")
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToFloat64()).Equal(float64(12), nil)
		}, float64(12))
	})
}

func TestRTValue_ToString(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToString()).
			Equal("", base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToString()).
			Equal("", base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToString()).Equal("", base.ErrStream)
		}, true)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToString()).Equal("kitty", nil)
		}, "kitty")
	})
}

func TestRTValue_ToBytes(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToBytes()).
			Equal([]byte{}, base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToBytes()).
			Equal([]byte{}, base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToBytes()).Equal([]byte{}, base.ErrStream)
		}, true)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToBytes()).Equal([]byte{1, 2, 3, 4, 5}, nil)
		}, []byte{1, 2, 3, 4, 5})
	})
}

func TestRTValue_ToArray(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToArray()).
			Equal(Array{}, base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToArray()).
			Equal(Array{}, base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToArray()).Equal(Array{}, base.ErrStream)
		}, true)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToArray()).Equal(Array{"1", true, int64(3)}, nil)
		}, Array{"1", true, int64(3)})
	})
}

func TestRTValue_ToRTArray(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToRTArray()).
			Equal(RTArray{}, base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToRTArray()).
			Equal(RTArray{}, base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToRTArray()).Equal(RTArray{}, base.ErrStream)
		}, true)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			rtArray, err := v.ToRTArray()
			assert(err).IsNil()
			assert(rtArray.Size()).Equal(3)
			assert(rtArray.Get(0).ToString()).Equal("1", nil)
			assert(rtArray.Get(1).ToBool()).Equal(true, nil)
			assert(rtArray.Get(2).ToInt64()).Equal(int64(3), nil)
		}, Array{"1", true, int64(3)})
	})
}

func TestRTValue_ToAMap(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToMap()).
			Equal(Map{}, base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToMap()).
			Equal(Map{}, base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToMap()).Equal(Map{}, base.ErrStream)
		}, true)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToMap()).Equal(Map{"name": "kitty", "age": int64(18)}, nil)
		}, Map{"name": "kitty", "age": int64(18)})
	})
}

func TestRTValue_ToRTMap(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToRTMap()).
			Equal(RTMap{}, base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToRTMap()).
			Equal(RTMap{}, base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToRTMap()).Equal(RTMap{}, base.ErrStream)
		}, true)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			rtMap, err := v.ToRTMap()
			assert(err).IsNil()
			assert(rtMap.Size()).Equal(2)
			assert(rtMap.Get("name").ToString()).Equal("kitty", nil)
			assert(rtMap.Get("age").ToInt64()).Equal(int64(18), nil)
		}, Map{"name": "kitty", "age": int64(18)})
	})
}
