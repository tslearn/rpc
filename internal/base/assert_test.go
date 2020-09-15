package base

import (
	"bytes"
	"os"
	"testing"
)

type fakeTesting struct {
	onFail func()
}

func (p *fakeTesting) Fail() {
	if p.onFail != nil {
		p.onFail()
	}
}

func testFailHelper(fn func(_ func(_ ...interface{}) Assert)) (bool, string) {
	var buf bytes.Buffer
	retCH := make(chan bool, 1)

	logWriter = &buf
	defer func() {
		logWriter = os.Stdout
	}()

	fn(func(args ...interface{}) Assert {
		return &rpcAssert{
			t: &fakeTesting{
				onFail: func() {
					retCH <- true
				},
			},
			args: args,
		}
	})

	select {
	case <-retCH:
		return true, buf.String()
	default:
		return false, buf.String()
	}
}

func TestNewAssert(t *testing.T) {
	t.Run("NewAssert t is nil", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewAssert(nil)
		assert(o(true)).Equals(&rpcAssert{t: nil, args: []interface{}{true}})
	})

	t.Run("NewAssert args is nil", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewAssert(t)
		assert(o()).Equals(&rpcAssert{t: t, args: nil})
	})

	t.Run("NewAssert ok", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewAssert(t)
		assert(o(true, 1)).Equals(&rpcAssert{t: t, args: []interface{}{true, 1}})
	})
}

func TestRpcAssert_Fail(t *testing.T) {
	t.Run("Report reason and source", func(t *testing.T) {
		assert := NewAssert(t)
		source := ""

		assert(testFailHelper(func(o func(_ ...interface{}) Assert) {
			func() { o().Fail("error"); source = GetFileLine(0) }()
		})).Equals(true, "\terror\n\t"+source+"")
	})
}

//
//func TestAssert_Fail(t *testing.T) {
//  assert := NewAssert(t)
//  assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//    assert(nil).Fail("test")
//  })).IsTrue()
//}

//func TestAssert_Equals(t *testing.T) {
//	assert := NewAssert(t)
//	assert(nil).Equals(nil)
//	assert(3).Equals(3)
//	assert((interface{})(nil)).Equals(nil)
//	assert((Assert)(nil)).Equals((Assert)(nil))
//	assert(nil).Equals((interface{})(nil))
//	assert((Assert)(nil)).Equals((Assert)(nil))
//	assert([]int{1, 2, 3}).Equals([]int{1, 2, 3})
//	assert(map[int]string{3: "OK", 4: "NO"}).
//		Equals(map[int]string{4: "NO", 3: "OK"})
//
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(nil).Equals(0)
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(3).Equals(uint(3))
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert((interface{})(nil)).Equals(0)
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert([]int{1, 2, 3}).Equals([]int64{1, 2, 3})
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert([]int{1, 2, 3}).Equals([]int32{1, 2, 3})
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(map[int]string{3: "OK", 4: "NO"}).
//			Equals(map[int64]string{4: "NO", 3: "OK"})
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert().Equals(3)
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(3, 3).Equals(3)
//	})).IsTrue()
//}
//
//func TestAssert_IsNil(t *testing.T) {
//	assert := NewAssert(t)
//
//	assert(nil).IsNil()
//	assert((Assert)(nil)).IsNil()
//	assert((unsafe.Pointer)(nil)).IsNil()
//
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(uintptr(0)).IsNil()
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(NewAssert(t)).IsNil()
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(32).IsNil()
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(false).IsNil()
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(0).IsNil()
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert().IsNil()
//	})).IsTrue()
//}
//
//func TestAssert_IsNotNil(t *testing.T) {
//	assert := NewAssert(t)
//	assert(t).IsNotNil()
//
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(nil).IsNotNil()
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert((Assert)(nil)).IsNotNil()
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert().IsNotNil()
//	})).IsTrue()
//}
//
//func TestAssert_IsTrue(t *testing.T) {
//	assert := NewAssert(t)
//	assert(true).IsTrue()
//
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert((Assert)(nil)).IsTrue()
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(32).IsTrue()
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(false).IsTrue()
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(0).IsTrue()
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert().IsTrue()
//	})).IsTrue()
//}
//
//func TestAssert_IsFalse(t *testing.T) {
//	assert := NewAssert(t)
//	assert(false).IsFalse()
//
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(32).IsFalse()
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(true).IsFalse()
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(0).IsFalse()
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert().IsFalse()
//	})).IsTrue()
//}
