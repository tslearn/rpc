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
		assert(o(true)).Equal(&rpcAssert{t: nil, args: []interface{}{true}})
	})

	t.Run("NewAssert args is nil", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewAssert(t)
		assert(o()).Equal(&rpcAssert{t: t, args: nil})
	})

	t.Run("NewAssert ok", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewAssert(t)
		assert(o(true, 1)).Equal(&rpcAssert{t: t, args: []interface{}{true, 1}})
	})
}

func TestRpcAssert_Fail(t *testing.T) {
	t.Run("Report reason and source", func(t *testing.T) {
		assert := NewAssert(t)
		source := ""
		assert(testFailHelper(func(o func(_ ...interface{}) Assert) {
			func() { o().Fail("error"); source = GetFileLine(0) }()
		})).Equal(true, "\terror\n\t"+source+"\n")
	})
}

//func TestAssert_Equals(t *testing.T) {
//	assert := NewAssert(t)
//	assert(nil).Equal(nil)
//	assert(3).Equal(3)
//	assert((interface{})(nil)).Equal(nil)
//	assert((Assert)(nil)).Equal((Assert)(nil))
//	assert(nil).Equal((interface{})(nil))
//	assert((Assert)(nil)).Equal((Assert)(nil))
//	assert([]int{1, 2, 3}).Equal([]int{1, 2, 3})
//	assert(map[int]string{3: "OK", 4: "NO"}).
//		Equal(map[int]string{4: "NO", 3: "OK"})
//
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(nil).Equal(0)
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(3).Equal(uint(3))
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert((interface{})(nil)).Equal(0)
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert([]int{1, 2, 3}).Equal([]int64{1, 2, 3})
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert([]int{1, 2, 3}).Equal([]int32{1, 2, 3})
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(map[int]string{3: "OK", 4: "NO"}).
//			Equal(map[int64]string{4: "NO", 3: "OK"})
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert().Equal(3)
//	})).IsTrue()
//	assert(runWithFail(func(assert func(args ...interface{}) Assert) {
//		assert(3, 3).Equal(3)
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
