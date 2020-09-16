package base

import (
	"testing"
)

func TestPanicSubscription_Close(t *testing.T) {
	t.Run("object is nil", func(t *testing.T) {
		assert := NewAssert(t)
		assert((*PanicSubscription)(nil).Close()).IsFalse()
	})

	t.Run("id not found", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := SubscribePanic(func(e Error) {})
		defer v1.Close()
		v2 := SubscribePanic(func(e Error) {})
		defer v2.Close()
		assert((&PanicSubscription{id: 8273}).Close()).IsFalse()
	})

	t.Run("ok", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := SubscribePanic(func(e Error) {})
		assert(v1.Close()).IsTrue()
	})
}

func TestSubscribePanic(t *testing.T) {
	t.Run("onPanic is nil", func(t *testing.T) {
		assert := NewAssert(t)
		assert(SubscribePanic(nil)).IsNil()
	})

	t.Run("ok", func(t *testing.T) {
		assert := NewAssert(t)
		onPanic := func(e Error) {}
		v1 := SubscribePanic(onPanic)
		defer v1.Close()
		assert(v1).IsNotNil()
		assert(v1.id > 0).IsTrue()
		assert(v1.onPanic).IsNotNil()
	})
}

func TestPublishPanic(t *testing.T) {
	t.Run("onPanic goes panic", func(t *testing.T) {
		assert := NewAssert(t)
		retCH := make(chan Error, 1)
		err := NewError(ErrorKindKernelPanic, "message", "debug")
		v1 := SubscribePanic(func(e Error) {
			retCH <- e
			panic("error")
		})
		defer v1.Close()
		PublishPanic(err)
		assert(<-retCH).Equal(err)
	})

	t.Run("ok", func(t *testing.T) {
		assert := NewAssert(t)
		retCH := make(chan Error, 1)
		err := NewError(ErrorKindKernelPanic, "message", "debug")
		v1 := SubscribePanic(func(e Error) {
			retCH <- e
		})
		defer v1.Close()
		PublishPanic(err)
		assert(<-retCH).Equal(err)
	})
}

func TestRunWithCatchPanic(t *testing.T) {
	t.Run("func with panic", func(t *testing.T) {
		assert := NewAssert(t)
		assert(RunWithCatchPanic(func() {
			panic("error")
		})).Equal("error")
	})

	t.Run("func without panic", func(t *testing.T) {
		assert := NewAssert(t)
		assert(RunWithCatchPanic(func() {})).IsNil()
	})
}

func TestRunWithSubscribePanic(t *testing.T) {
	t.Run("func with PublishPanic", func(t *testing.T) {
		assert := NewAssert(t)
		err := NewError(ErrorKindKernelPanic, "message", "debug")
		assert(RunWithSubscribePanic(func() {
			PublishPanic(err)
		})).Equal(err)
	})

	t.Run("func without PublishPanic", func(t *testing.T) {
		assert := NewAssert(t)
		assert(RunWithSubscribePanic(func() {})).IsNil()
	})
}