package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"runtime"
	"testing"
)

func TestRTArray(t *testing.T) {
	t.Run("test thread safe", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(7)

		wait := make(chan bool)
		for i := 0; i < 1000; i++ {
			go func(idx int) {
				for j := 0; j < 1000; j++ {
					assert(v.Append(idx)).IsNil()
				}
				wait <- true
			}(i)

			runtime.GC()
		}

		for i := 0; i < 1000; i++ {
			<-wait
		}

		//  assert(v.Size()).Equal(1000000)
	})
}

func TestRTArrayNewRTArray(t *testing.T) {
	t.Run("thread is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(newRTArray(Runtime{}, 2)).Equal(RTArray{})
	})

	t.Run("no cache", func(t *testing.T) {
		assert := base.NewAssert(t)

		// consume cache
		testRuntime.thread.Reset()
		for i := 0; i < len(testRuntime.thread.cache); i++ {
			testRuntime.thread.malloc(1)
		}

		v := newRTArray(testRuntime, 2)
		assert(v.rt).Equal(testRuntime)
		assert(len(*v.items), cap(*v.items)).Equal(0, 2)
	})

	t.Run("size is less than zero", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		assert(newRTArray(testRuntime, -1)).Equal(RTArray{})
	})

	t.Run("size is zero", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := newRTArray(testRuntime, 0)
		assert(v.rt).Equal(testRuntime)
		assert(v.items != nil)
		assert(len(*v.items), cap(*v.items)).Equal(0, 0)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 0; i < 100; i++ {
			testRuntime.thread.Reset()
			v := newRTArray(testRuntime, i)
			assert(v.rt).Equal(testRuntime)
			assert(len(*v.items), cap(*v.items)).Equal(0, i)
		}
	})
}

func TestRTArray_Get(t *testing.T) {
	t.Run("invalid RTArray", func(t *testing.T) {
		assert := base.NewAssert(t)
		rtArray := RTArray{}
		assert(rtArray.Get(0).err).Equal(
			errors.ErrRuntimeIllegalInCurrentGoroutine,
		)
	})

	t.Run("index overflow 1", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(true)
		assert(v.Get(1).err).Equal(errors.ErrRTArrayIndexOverflow.
			AddDebug("RTArray index 1 out of range"))
	})

	t.Run("index overflow 2", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(true)
		assert(v.Get(-1).err).Equal(errors.ErrRTArrayIndexOverflow.
			AddDebug("RTArray index -1 out of range"))
	})

	t.Run("index ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(true)
		assert(v.Get(0).ToBool()).Equal(true, nil)
	})
}

func TestRTArray_Set(t *testing.T) {
	t.Run("invalid RTArray", func(t *testing.T) {
		assert := base.NewAssert(t)
		rtArray := RTArray{}
		assert(rtArray.Set(0, true)).Equal(
			errors.ErrRuntimeIllegalInCurrentGoroutine,
		)
	})

	t.Run("unsupported value", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(true)
		assert(v.Set(0, make(chan bool))).Equal(
			errors.ErrUnsupportedValue.AddDebug("value is not supported"),
		)
	})

	t.Run("index overflow 1", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(true)
		assert(v.Set(1, true)).Equal(errors.ErrRTArrayIndexOverflow.
			AddDebug("RTArray index 1 out of range"))
	})

	t.Run("index overflow 2", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(true)
		assert(v.Set(-1, true)).Equal(errors.ErrRTArrayIndexOverflow.
			AddDebug("RTArray index -1 out of range"))
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(true)
		_ = v.Append("kitty")
		assert(v.Get(0).ToBool()).Equal(true, nil)
		assert(v.Get(1).ToString()).Equal("kitty", nil)
		assert(v.Set(0, "doggy")).Equal(nil)
		assert(v.Set(1, false)).Equal(nil)
		assert(v.Get(0).ToString()).Equal("doggy", nil)
		assert(v.Get(1).ToBool()).Equal(false, nil)
	})
}

func TestRTArray_Append(t *testing.T) {
	t.Run("invalid RTArray", func(t *testing.T) {
		assert := base.NewAssert(t)
		rtArray := RTArray{}
		assert(rtArray.Append(true)).Equal(
			errors.ErrRuntimeIllegalInCurrentGoroutine,
		)
	})

	t.Run("unsupported value", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		assert(v.Append(make(chan bool))).Equal(
			errors.ErrUnsupportedValue.AddDebug("value is not supported"),
		)
		assert(v.Size()).Equal(0)
	})

	t.Run("test ok (string)", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		assert(v.Append("kitty")).Equal(nil)
		assert(v.Size()).Equal(1)
		assert(v.Get(0).ToString()).Equal("kitty", nil)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 0; i < 1000; i++ {
			testRuntime.thread.Reset()
			v := testRuntime.NewRTArray(0)
			for j := int64(0); j < 1000; j++ {
				assert(v.Append(j)).IsNil()
			}

			assert(v.Size()).Equal(1000)
			for j := 0; j < 1000; j++ {
				assert(v.Get(j).ToInt64()).Equal(int64(j), nil)
			}
		}
	})
}

func TestRTArray_Delete(t *testing.T) {
	t.Run("invalid RTArray", func(t *testing.T) {
		assert := base.NewAssert(t)
		rtArray := RTArray{}
		assert(rtArray.Delete(0)).Equal(
			errors.ErrRuntimeIllegalInCurrentGoroutine,
		)
	})

	t.Run("index overflow 1", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(true)
		assert(v.Delete(1)).Equal(errors.ErrRTArrayIndexOverflow.
			AddDebug("RTArray index 1 out of range"))
	})

	t.Run("index overflow 2", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(true)
		assert(v.Delete(-1)).Equal(errors.ErrRTArrayIndexOverflow.
			AddDebug("RTArray index -1 out of range"))
	})

	t.Run("delete first elem", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(0)
		_ = v.Append(1)
		_ = v.Append(2)
		assert(v.Delete(0)).Equal(nil)
		assert(v.Size()).Equal(2)
		assert(v.Get(0).ToInt64()).Equal(int64(1), nil)
		assert(v.Get(1).ToInt64()).Equal(int64(2), nil)
	})

	t.Run("delete middle elem", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(0)
		_ = v.Append(1)
		_ = v.Append(2)
		assert(v.Delete(1)).Equal(nil)
		assert(v.Size()).Equal(2)
		assert(v.Get(0).ToInt64()).Equal(int64(0), nil)
		assert(v.Get(1).ToInt64()).Equal(int64(2), nil)
	})

	t.Run("delete last elem", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(0)
		_ = v.Append(1)
		_ = v.Append(2)
		assert(v.Delete(2)).Equal(nil)
		assert(v.Size()).Equal(2)
		assert(v.Get(0).ToInt64()).Equal(int64(0), nil)
		assert(v.Get(1).ToInt64()).Equal(int64(1), nil)
	})
}

func TestRTArray_Size(t *testing.T) {
	t.Run("invalid RTArray", func(t *testing.T) {
		assert := base.NewAssert(t)
		rtArray := RTArray{}
		assert(rtArray.Size()).Equal(-1)
	})

	t.Run("valid RTArray 1", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		assert(v.Size()).Equal(0)
	})

	t.Run("valid RTArray 2", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(1)
		assert(v.Size()).Equal(1)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 500; i++ {
			testRuntime.thread.Reset()
			v := testRuntime.NewRTArray(0)
			for j := 0; j < i; j++ {
				assert(v.Append(true)).Equal(nil)
			}
			assert(v.Size()).Equal(i)
			for j := 0; j < i; j++ {
				assert(v.Delete(0)).Equal(nil)
			}
			assert(v.Size()).Equal(0)
		}
	})
}
