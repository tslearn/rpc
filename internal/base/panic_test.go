package base

import "testing"

func TestReportPanic(t *testing.T) {
	//assert := base.NewAssert(t)
	//
	//// Test(1)
	//assert(testRunWithSubscribePanic(
	//	func() {
	//		ReportPanic(NewReplyPanic("reply panic error"))
	//	},
	//)).Equal(NewReplyPanic("reply panic error"))
}

func TestSubscribePanic(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := SubscribePanic(nil)
	assert(o1).IsNil()

	// Test(2)
	o2 := SubscribePanic(func(v interface{}) {})
	defer o2.Close()
	assert(o2).IsNotNil()
	assert(o2.id > 0).IsTrue()
	assert(o2.onPanic).IsNotNil()
	assert(len(gPanicSubscriptions)).Equal(1)
	assert(gPanicSubscriptions[0]).Equal(o2)
}

func TestRpcPanicSubscription_Close(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := (*PanicSubscription)(nil)
	assert(o1.Close()).IsFalse()

	// Test(2)
	o2 := SubscribePanic(func(v interface{}) {})
	assert(o2.Close()).IsTrue()

	// Test(3)
	o3 := SubscribePanic(func(v interface{}) {})
	o3.Close()
	assert(o3.Close()).IsFalse()
}
