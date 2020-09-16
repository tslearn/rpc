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

//
//func TestReportPanic(t *testing.T) {
//	//assert := base.NewAssert(t)
//	//
//	//// Test(1)
//	//assert(testRunWithSubscribePanic(
//	//	func() {
//	//		PublishPanic(NewReplyPanic("reply panic error"))
//	//	},
//	//)).Equal(NewReplyPanic("reply panic error"))
//}
//
//func TestSubscribePanic(t *testing.T) {
//	assert := NewAssert(t)
//
//	// Test(1)
//	o1 := SubscribePanic(nil)
//	assert(o1).IsNil()
//
//	// Test(2)
//	o2 := SubscribePanic(func(v Error) {})
//	defer o2.Close()
//	assert(o2).IsNotNil()
//	assert(o2.id > 0).IsTrue()
//	assert(o2.onPanic).IsNotNil()
//	assert(len(gPanicSubscriptions)).Equal(1)
//	assert(gPanicSubscriptions[0]).Equal(o2)
//}
//
//func TestRpcPanicSubscription_Close(t *testing.T) {
//	assert := NewAssert(t)
//
//	// Test(1)
//	o1 := (*PanicSubscription)(nil)
//	assert(o1.Close()).IsFalse()
//
//	// Test(2)
//	o2 := SubscribePanic(func(e Error) {})
//	assert(o2.Close()).IsTrue()
//
//	// Test(3)
//	o3 := SubscribePanic(func(e Error) {})
//	o3.Close()
//	assert(o3.Close()).IsFalse()
//}
