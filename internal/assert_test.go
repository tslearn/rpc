package internal

import (
	"testing"
	"unsafe"
)

// test IsNil fail
func runWithFail(fn func(assert func(args ...interface{}) *RPCAssert)) bool {
	ch := make(chan bool, 1)
	fn(func(args ...interface{}) *RPCAssert {
		return &RPCAssert{
			t: nil,
			hookFail: func() {
				ch <- true
			},
			args: args,
		}
	})

	select {
	case <-ch:
		return true
	default:
		return false
	}
}

func TestNewAssert(t *testing.T) {
	assert := NewRPCAssert(t)

	assert(assert(3).args[0]).Equals(3)
	assert(assert(3).t).Equals(t)

	assert(assert(3, true, nil).args...).Equals(3, true, nil)
	assert(assert(3).t).Equals(t)
}

func TestAssert_Fail(t *testing.T) {
	assert := NewRPCAssert(t)
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert(nil).Fail()
	})).IsTrue()
}

func TestAssert_Equals(t *testing.T) {
	assert := NewRPCAssert(t)
	assert(nil).Equals(nil)
	assert(3).Equals(3)
	assert((interface{})(nil)).Equals(nil)
	assert((*RPCAssert)(nil)).Equals((*RPCAssert)(nil))
	assert(nil).Equals((interface{})(nil))
	assert((*RPCAssert)(nil)).Equals((*RPCAssert)(nil))
	assert([]int{1, 2, 3}).Equals([]int{1, 2, 3})
	assert(map[int]string{3: "OK", 4: "NO"}).
		Equals(map[int]string{4: "NO", 3: "OK"})

	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert(nil).Equals(0)
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert(3).Equals(uint(3))
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert((interface{})(nil)).Equals(0)
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert([]int{1, 2, 3}).Equals([]int64{1, 2, 3})
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert([]int{1, 2, 3}).Equals([]int32{1, 2, 3})
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert(map[int]string{3: "OK", 4: "NO"}).
			Equals(map[int64]string{4: "NO", 3: "OK"})
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert().Equals(3)
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert(3, 3).Equals(3)
	})).IsTrue()
}

func TestAssert_IsNil(t *testing.T) {
	assert := NewRPCAssert(t)

	assert(nil).IsNil()
	assert((*RPCAssert)(nil)).IsNil()
	assert((unsafe.Pointer)(nil)).IsNil()

	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert(uintptr(0)).IsNil()
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert(NewRPCAssert(t)).IsNil()
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert(32).IsNil()
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert(false).IsNil()
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert(0).IsNil()
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert().IsNil()
	})).IsTrue()
}

func TestAssert_IsNotNil(t *testing.T) {
	assert := NewRPCAssert(t)
	assert(t).IsNotNil()

	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert(nil).IsNotNil()
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert((*RPCAssert)(nil)).IsNotNil()
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert().IsNotNil()
	})).IsTrue()
}

func TestAssert_IsTrue(t *testing.T) {
	assert := NewRPCAssert(t)
	assert(true).IsTrue()

	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert((*RPCAssert)(nil)).IsTrue()
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert(32).IsTrue()
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert(false).IsTrue()
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert(0).IsTrue()
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert().IsTrue()
	})).IsTrue()
}

func TestAssert_IsFalse(t *testing.T) {
	assert := NewRPCAssert(t)
	assert(false).IsFalse()

	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert(32).IsFalse()
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert(true).IsFalse()
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert(0).IsFalse()
	})).IsTrue()
	assert(runWithFail(func(assert func(args ...interface{}) *RPCAssert) {
		assert().IsFalse()
	})).IsTrue()
}

type testAssertFail struct {
	ch chan bool
}

func (p *testAssertFail) Fail() {
	p.ch <- true
}

func Test_reportFail(t *testing.T) {
	assert := NewRPCAssert(t)

	ch := make(chan bool, 1)
	target := NewRPCAssert(t)(3)
	target.t = &testAssertFail{ch: ch}
	target.Equals(4)

	assert(<-target.t.(*testAssertFail).ch).IsTrue()
}
