package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"testing"
)

func testWithRTValue(fn func(v RTValue), v interface{}) {
	testRuntime.thread.Reset()
	rtArray := testRuntime.NewRTArray()
	_ = rtArray.Append(v)
	fn(rtArray.Get(0))
	testRuntime.thread.Reset()
}

func TestRTValue_ToBool(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: errors.ErrStream}.ToBool()).
			Equal(false, errors.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToBool()).
			Equal(false, errors.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToBool()).Equal(false, errors.ErrStream)
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
		assert(RTValue{err: errors.ErrStream}.ToInt64()).
			Equal(int64(0), errors.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToInt64()).
			Equal(int64(0), errors.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToInt64()).Equal(int64(0), errors.ErrStream)
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
		assert(RTValue{err: errors.ErrStream}.ToUint64()).
			Equal(uint64(0), errors.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToUint64()).
			Equal(uint64(0), errors.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToUint64()).Equal(uint64(0), errors.ErrStream)
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
		assert(RTValue{err: errors.ErrStream}.ToFloat64()).
			Equal(float64(0), errors.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToFloat64()).
			Equal(float64(0), errors.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToFloat64()).Equal(float64(0), errors.ErrStream)
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
		assert(RTValue{err: errors.ErrStream}.ToString()).
			Equal("", errors.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToString()).
			Equal("", errors.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToString()).Equal("", errors.ErrStream)
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
		assert(RTValue{err: errors.ErrStream}.ToBytes()).
			Equal([]byte{}, errors.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToBytes()).
			Equal([]byte{}, errors.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToBytes()).Equal([]byte{}, errors.ErrStream)
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
		assert(RTValue{err: errors.ErrStream}.ToArray()).
			Equal(Array{}, errors.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToArray()).
			Equal(Array{}, errors.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToArray()).Equal(Array{}, errors.ErrStream)
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
		assert(RTValue{err: errors.ErrStream}.ToRTArray()).
			Equal(RTArray{}, errors.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToRTArray()).
			Equal(RTArray{}, errors.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToRTArray()).Equal(RTArray{}, errors.ErrStream)
		}, true)
	})

	//t.Run("test ok", func(t *testing.T) {
	//  assert := base.NewAssert(t)
	//  testWithRTValue(func(v RTValue) {
	//    assert(v.ToArray()).Equal(Array{"1", true, int64(3)}, nil)
	//  }, Array{"1", true, int64(3)})
	//})
}
