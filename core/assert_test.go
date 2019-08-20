package core

import (
	"testing"
)

// test IsNil fail
func assertFailedFn(fn func()) {
	ch := make(chan bool, 1)
	originReportFail := reportFail
	reportFail = func(p *rpcAssert) {
		ch <- true
	}
	fn()
	reportFail = originReportFail
	<-ch
}

func TestNewAssert(t *testing.T) {
	assert := newAssert(t)

	assert(assert(3).args[0]).Equals(3)
	assert(assert(3).t).Equals(t)
}

func TestAssert_Equals(t *testing.T) {
	assert := newAssert(t)
	assert(3).Equals(3)
	assert(nil).Equals(nil)
	assert((*rpcError)(nil)).Equals(nil)

	assertFailedFn(func() {
		assert(3).Equals(4)
	})
	assertFailedFn(func() {
		assert(3).Equals(true)
	})
	assertFailedFn(func() {
		assert(3).Equals(nil)
	})
	assertFailedFn(func() {
		assert(nil).Equals(3)
	})
	assertFailedFn(func() {
		assert().Equals(3)
	})
	assertFailedFn(func() {
		assert(3, 3).Equals(3)
	})
}

func TestAssert_Contains(t *testing.T) {
	assert := newAssert(t)
	ctx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}
	assert("hello").Contains("")
	assert("hello").Contains("o")

	assert(newRPCArrayByArray(ctx, []interface{}{1, 2})).Contains(int64(2))

	assertFailedFn(func() {
		assert(3).Contains(3)
	})
	assertFailedFn(func() {
		assert(3).Contains(4)
	})
	assertFailedFn(func() {
		assert(nil).Contains(3)
	})
	assertFailedFn(func() {
		assert(nil).Contains(nil)
	})
	assertFailedFn(func() {
		assert(newRPCArrayByArray(ctx, []interface{}{1, 3})).Contains(nil)
	})
	assertFailedFn(func() {
		assert().Contains(2)
	})
}

func TestAssert_IsNil(t *testing.T) {
	assert := newAssert(t)

	assert(nil).IsNil()
	assert((*rpcError)(nil)).IsNil()

	assertFailedFn(func() {
		assert(NewRPCError("")).IsNil()
	})
	assertFailedFn(func() {
		assert(32).IsNil()
	})
	assertFailedFn(func() {
		assert(false).IsNil()
	})
	assertFailedFn(func() {
		assert(0).IsNil()
	})
	assertFailedFn(func() {
		assert(nilRPCArray).IsNil()
	})
	assertFailedFn(func() {
		assert(nilRPCMap).IsNil()
	})
	assertFailedFn(func() {
		assert().IsNil()
	})
}

func TestAssert_IsNotNil(t *testing.T) {
	assert := newAssert(t)
	assert(t).IsNotNil()

	assertFailedFn(func() {
		assert(nil).IsNotNil()
	})
	assertFailedFn(func() {
		assert((*rpcError)(nil)).IsNotNil()
	})
	assertFailedFn(func() {
		assert().IsNotNil()
	})
}

func TestAssert_IsTrue(t *testing.T) {
	assert := newAssert(t)
	assert(true).IsTrue()

	assertFailedFn(func() {
		assert((*rpcError)(nil)).IsTrue()
	})
	assertFailedFn(func() {
		assert(32).IsTrue()
	})
	assertFailedFn(func() {
		assert(false).IsTrue()
	})
	assertFailedFn(func() {
		assert(0).IsTrue()
	})
	assertFailedFn(func() {
		assert().IsTrue()
	})
}

func TestAssert_IsFalse(t *testing.T) {
	assert := newAssert(t)
	assert(false).IsFalse()

	assertFailedFn(func() {
		assert(32).IsFalse()
	})
	assertFailedFn(func() {
		assert(true).IsFalse()
	})
	assertFailedFn(func() {
		assert(0).IsFalse()
	})
	assertFailedFn(func() {
		assert().IsFalse()
	})
}

type testAssertFail struct {
	ch chan bool
}

func (p *testAssertFail) Fail() {
	p.ch <- true
}

func Test_reportFail(t *testing.T) {
	assert := newAssert(t)

	ch := make(chan bool, 1)
	target := newAssert(t)(3)
	target.t = &testAssertFail{ch: ch}
	target.Equals(4)

	assert(<-target.t.(*testAssertFail).ch).IsTrue()
}
