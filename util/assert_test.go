package util

import (
	"testing"
	"unsafe"
)

func TestIsNil(t *testing.T) {
	assert := NewAssert(t)
	assert(isNil(nil)).IsTrue()
	assert(isNil(t)).IsFalse()
	assert(isNil(3)).IsFalse()
	assert(isNil(0)).IsFalse()
	assert(isNil(uintptr(0))).IsFalse()
	assert(isNil(uintptr(1))).IsFalse()
	assert(isNil(unsafe.Pointer(nil))).IsTrue()
	assert(isNil(unsafe.Pointer(t))).IsFalse()
}

// test IsNil fail
func runReportAssertFail(fn func(assert func(args ...interface{}) *rpcAssert)) {
	ch := make(chan bool, 1)
	fn(func(args ...interface{}) *rpcAssert {
		return &rpcAssert{
			t: nil,
			hookFN: func() {
				ch <- true
			},
			args: args,
		}
	})
	<-ch
}

func TestNewAssert(t *testing.T) {
	assert := NewAssert(t)

	assert(assert(3).args[0]).Equals(3)
	assert(assert(3).t).Equals(t)

	assert(assert(3, true, nil).args...).Equals(3, true, nil)
	assert(assert(3).t).Equals(t)
}

func TestAssert_Fail(t *testing.T) {
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert(nil).Fail()
	})
}

func TestAssert_Equals(t *testing.T) {
	assert := NewAssert(t)
	assert(nil).Equals(nil)
	assert(3).Equals(3)
	assert((interface{})(nil)).Equals(nil)
	assert((*rpcAssert)(nil)).Equals((*rpcAssert)(nil))
	assert(nil).Equals((interface{})(nil))
	assert((*rpcAssert)(nil)).Equals((*rpcAssert)(nil))
	assert([]int{1, 2, 3}).Equals([]int{1, 2, 3})
	assert(map[int]string{3: "OK", 4: "NO"}).
		Equals(map[int]string{4: "NO", 3: "OK"})

	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert(nil).Equals(0)
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert(3).Equals(uint(3))
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert((interface{})(nil)).Equals(0)
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert([]int{1, 2, 3}).Equals([]int64{1, 2, 3})
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert([]int{1, 2, 3}).Equals([]int32{1, 2, 3})
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert(map[int]string{3: "OK", 4: "NO"}).
			Equals(map[int64]string{4: "NO", 3: "OK"})
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert().Equals(3)
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert(3, 3).Equals(3)
	})
}

func TestAssert_IsNil(t *testing.T) {
	assert := NewAssert(t)

	assert(nil).IsNil()
	assert((*rpcAssert)(nil)).IsNil()
	assert((unsafe.Pointer)(nil)).IsNil()

	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert(uintptr(0)).IsNil()
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert(NewAssert(t)).IsNil()
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert(32).IsNil()
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert(false).IsNil()
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert(0).IsNil()
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert().IsNil()
	})
}

func TestAssert_IsNotNil(t *testing.T) {
	assert := NewAssert(t)
	assert(t).IsNotNil()

	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert(nil).IsNotNil()
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert((*rpcAssert)(nil)).IsNotNil()
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert().IsNotNil()
	})
}

func TestAssert_IsTrue(t *testing.T) {
	assert := NewAssert(t)
	assert(true).IsTrue()

	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert((*rpcAssert)(nil)).IsTrue()
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert(32).IsTrue()
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert(false).IsTrue()
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert(0).IsTrue()
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert().IsTrue()
	})
}

func TestAssert_IsFalse(t *testing.T) {
	assert := NewAssert(t)
	assert(false).IsFalse()

	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert(32).IsFalse()
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert(true).IsFalse()
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
		assert(0).IsFalse()
	})
	runReportAssertFail(func(assert func(args ...interface{}) *rpcAssert) {
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
	assert := NewAssert(t)

	ch := make(chan bool, 1)
	target := NewAssert(t)(3)
	target.t = &testAssertFail{ch: ch}
	target.Equals(4)

	assert(<-target.t.(*testAssertFail).ch).IsTrue()
}
