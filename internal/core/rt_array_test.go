package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"testing"
)

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
		v := testRuntime.NewRTArray()
		_ = v.Append(true)
		assert(v.Get(1).err).Equal(errors.ErrRTArrayIndexOverflow.
			AddDebug("RTArray index 1 out of range"))
	})

	t.Run("index overflow 2", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray()
		_ = v.Append(true)
		assert(v.Get(-1).err).Equal(errors.ErrRTArrayIndexOverflow.
			AddDebug("RTArray index -1 out of range"))
	})

	t.Run("index ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray()
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
		v := testRuntime.NewRTArray()
		_ = v.Append(true)
		assert(v.Set(0, make(chan bool))).Equal(
			errors.ErrUnsupportedValue.AddDebug("value is not supported"),
		)
	})

	t.Run("index overflow 1", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray()
		_ = v.Append(true)
		assert(v.Set(1, true)).Equal(errors.ErrRTArrayIndexOverflow.
			AddDebug("RTArray index 1 out of range"))
	})

	t.Run("index overflow 2", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray()
		_ = v.Append(true)
		assert(v.Set(-1, true)).Equal(errors.ErrRTArrayIndexOverflow.
			AddDebug("RTArray index -1 out of range"))
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray()
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
		assert(rtArray.Set(0, true)).Equal(
			errors.ErrRuntimeIllegalInCurrentGoroutine,
		)
	})
}

func TestRTArray_Delete(t *testing.T) {
	t.Run("invalid RTArray", func(t *testing.T) {
		assert := base.NewAssert(t)
		rtArray := RTArray{}
		assert(rtArray.Set(0, true)).Equal(
			errors.ErrRuntimeIllegalInCurrentGoroutine,
		)
	})
}

func TestRTArray_Size(t *testing.T) {
	t.Run("invalid RTArray", func(t *testing.T) {
		assert := base.NewAssert(t)
		rtArray := RTArray{}
		assert(rtArray.Set(0, true)).Equal(
			errors.ErrRuntimeIllegalInCurrentGoroutine,
		)
	})
}
