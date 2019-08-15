package core

import (
	"testing"
	"unsafe"
)

// test IsNil fail
func assertFailedFn(fn func()) {
	ch := make(chan bool, 1)
	originReportFail := reportFail
	reportFail = func(p *Assert) {
		ch <- true
	}
	fn()
	reportFail = originReportFail
	<-ch
}

func TestNewAssert(t *testing.T) {
	assert := NewAssert(t)

	assert(assert(3).args[0]).Equals(3)
	assert(assert(3).t).Equals(t)
}

func TestAssert_Equals(t *testing.T) {
	assert := NewAssert(t)
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
	assert := NewAssert(t)
	ctx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}
	assert("hello").Contains("")
	assert("hello").Contains("o")

	assert(toRPCArray([]interface{}{1, 2}, ctx)).Contains(int64(2))

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
		assert(toRPCArray([]interface{}{1, 3}, ctx)).Contains(nil)
	})
	assertFailedFn(func() {
		assert().Contains(2)
	})
}

func TestAssert_IsNil(t *testing.T) {
	assert := NewAssert(t)

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
	assert := NewAssert(t)
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
	assert := NewAssert(t)
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
	assert := NewAssert(t)
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
	assert := NewAssert(t)

	ch := make(chan bool, 1)
	target := NewAssert(t)(3)
	target.t = &testAssertFail{ch: ch}
	target.Equals(4)

	assert(<-target.t.(*testAssertFail).ch).IsTrue()
}

func Test_isNil(t *testing.T) {
	assert := NewAssert(t)

	assert(assertIsNil(nil)).IsTrue()
	assert(assertIsNil((*rpcStream)(nil))).IsTrue()
	assert(assertIsNil((*rpcArray)(nil))).IsTrue()
	assert(assertIsNil((*rpcMap)(nil))).IsTrue()

	assert(assertIsNil(nilRPCArray)).IsFalse()
	assert(assertIsNil(nilRPCMap)).IsFalse()

	unsafeNil := unsafe.Pointer(nil)
	uintptrNil := uintptr(0)

	assert(assertIsNil(unsafeNil)).IsTrue()
	assert(assertIsNil(uintptrNil)).IsTrue()
}

func Test_equals(t *testing.T) {
	assert := NewAssert(t)
	ctx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}

	loggerPtr := NewLogger()
	testCollection := [][3]interface{}{
		{true, true, true},
		{false, false, true},
		{false, true, false},
		{false, 0, false},
		{true, 1, false},
		{true, nil, false},
		{0, 0, true},
		{3, 4, false},
		{3, int(3), true},
		{3, int32(3), false},
		{3, nil, false},
		{3.14, 3.14, true},
		{3.14, 3.15, false},
		{3.14, float32(3.14), false},
		{3.14, float64(3.14), true},
		{3.14, nil, false},
		{"", "", true},
		{"abc", "abc", true},
		{"abc", "ab", false},
		{"", nil, false},
		{"", 6, false},

		{[]byte{}, []byte{}, true},
		{[]byte{12}, []byte{12}, true},
		{[]byte{12, 13}, []byte{12, 13}, true},
		{[]byte{12, 13}, 12, false},
		{[]byte{13, 12}, []byte{12, 13}, false},
		{[]byte{12}, []byte{12, 13}, false},
		{[]byte{12, 13}, []byte{12}, false},
		{[]byte{13, 12}, nil, false},
		{[]byte{}, nil, false},

		{nilRPCMap, nilRPCMap, true},
		{toRPCMap(map[string]interface{}{}, ctx), toRPCMap(map[string]interface{}{}, ctx), true},
		{
			toRPCMap(map[string]interface{}{"test": 9007199254740991}, ctx),
			toRPCMap(map[string]interface{}{"test": 9007199254740991}, ctx),
			true,
		},
		{
			toRPCMap(map[string]interface{}{"test": 9007199254740991}, ctx),
			toRPCArray([]interface{}{9007199254740991}, ctx),
			false,
		},
		{
			toRPCMap(map[string]interface{}{"test": 9007199254740991}, ctx),
			toRPCMap(map[string]interface{}{"test": 9007199254740991, "3": 9007199254740991}, ctx),
			false,
		},
		{
			toRPCMap(map[string]interface{}{"test": 9007199254740991, "3": 9007199254740991}, ctx),
			toRPCMap(map[string]interface{}{"test": 9007199254740991}, ctx),
			false,
		},
		{
			toRPCMap(map[string]interface{}{"test": 9007199254740991}, ctx),
			toRPCMap(map[string]interface{}{"test": 9007199254740990}, ctx),
			false,
		},
		{
			toRPCMap(map[string]interface{}{"test": 9007199254740991}, ctx),
			toRPCMap(nil, ctx),
			false,
		},
		{toRPCMap(map[string]interface{}{}, ctx), nil, false},
		{nilRPCArray, nilRPCArray, true},
		{toRPCArray([]interface{}{}, ctx), toRPCArray([]interface{}{}, ctx), true},
		{toRPCArray([]interface{}{1}, ctx), toRPCArray([]interface{}{1}, ctx), true},
		{toRPCArray([]interface{}{1, 2}, ctx), toRPCArray([]interface{}{1, 2}, ctx), true},
		{toRPCArray([]interface{}{1, 2}, ctx), 3, false},
		{toRPCArray([]interface{}{1, 2}, ctx), toRPCArray([]interface{}{1}, ctx), false},
		{toRPCArray([]interface{}{1}, ctx), toRPCArray([]interface{}{1, 2}, ctx), false},
		{toRPCArray([]interface{}{1, 2}, ctx), toRPCArray([]interface{}{2, 1}, ctx), false},
		{toRPCArray([]interface{}{1, 2}, ctx), toRPCArray(nil, ctx), false},
		{toRPCArray([]interface{}{}, ctx), toRPCArray(nil, ctx), false},

		{nil, nil, true},
		{nil, (*Logger)(nil), true},
		{(*Logger)(nil), nil, true},
		{nil, []interface{}(nil), true},
		{nil, map[string]interface{}(nil), true},
		{nil, []byte(nil), true},
		{nil, []byte{}, false},
		{rpcArray{}, nil, false},
		{rpcMap{}, nil, false},
		{[]byte{}, nil, false},

		{NewRPCErrorWithDebug("m1", "d1"), NewRPCErrorWithDebug("m1", "d1"), true},
		{NewRPCErrorWithDebug("", "d1"), NewRPCErrorWithDebug("m1", "d1"), false},
		{NewRPCErrorWithDebug("m1", ""), NewRPCErrorWithDebug("m1", "d1"), false},
		{NewRPCErrorWithDebug("m1", ""), nil, false},
		{NewRPCErrorWithDebug("m1", ""), 3, false},

		{loggerPtr, loggerPtr, true},
		{NewLogger(), NewLogger(), false},

		{nilRPCArray, nilRPCArray, true},
		{nilRPCMap, nilRPCMap, true},
	}

	for _, item := range testCollection {
		assert(assertEquals(item[0], item[1]) == item[2]).IsTrue()
	}
}

func Test_equals_exceptions(t *testing.T) {
	assert := NewAssert(t)

	ctx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}

	rightArray := newRPCArray(ctx)
	errorArray := newRPCArray(ctx)
	rightArray.Append(true)
	errorArray.Append(true)
	(*errorArray.ctx.getCacheStream().frames[0])[1] = 13
	assert(assertEquals(rightArray, errorArray)).IsFalse()
	assert(assertEquals(errorArray, rightArray)).IsFalse()

	ctx.inner.stream = NewRPCStream()
	rightMap := newRPCMap(ctx)
	errorMap := newRPCMap(ctx)
	rightMap.Set("0", true)
	errorMap.Set("0", true)
	(*errorMap.ctx.getCacheStream().frames[0])[1] = 13
	assert(assertEquals(rightMap, errorMap)).IsFalse()
	assert(assertEquals(errorMap, rightMap)).IsFalse()
}

func Test_contains(t *testing.T) {
	assert := NewAssert(t)
	ctx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}

	testCollection := [][3]interface{}{
		{"hello world", "world", true},
		{"hello world", "you", false},
		{"hello world", 3, false},
		{"hello world", nil, false},
		{toRPCArray([]interface{}{1, 2, int64(3)}, ctx), int64(3), true},
		{toRPCArray([]interface{}{1, 2, int64(3)}, ctx), int(3), false},
		{toRPCArray([]interface{}{1, 2, 3}, ctx), 0, false},
		{toRPCArray([]interface{}{1, 2, 3}, ctx), nil, false},
		{toRPCArray([]interface{}{1, 2, 3}, ctx), true, false},
		{toRPCMap(map[string]interface{}{"1": 1, "2": 2}, ctx), "1", false},
		{toRPCMap(map[string]interface{}{"1": 1, "2": 2}, ctx), "3", false},
		{toRPCMap(map[string]interface{}{"1": 1, "2": 2}, ctx), true, false},
		{toRPCMap(map[string]interface{}{"1": 1, "2": 2}, ctx), nil, false},
		{[]byte{}, []byte{}, true},
		{[]byte{1, 2, 3, 4}, []byte{}, true},
		{[]byte{1, 2, 3, 4}, []byte{2, 3}, true},
		{[]byte{1, 2}, []byte{1, 2}, true},
		{[]byte{1, 2}, []byte{1, 2, 3}, false},
		{[]byte{1, 2, 3, 4}, []byte{2, 4}, false},
		{[]byte{1, 2}, 1, false},
		{[]byte{1, 2}, true, false},
		{[]byte{1, 2}, nil, false},

		{nil, "3", false},
		{nil, nil, false},
		{true, 3, false},
		{float64(0), float64(0), false},
	}

	for _, v := range testCollection {
		assert(contains(v[0], v[1])).Equals(v[2])
	}
}

func Test_contains_exceptions(t *testing.T) {
	assert := NewAssert(t)
	ctx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}

	errorArray := newRPCArray(ctx)
	errorArray.Append(true)
	(*errorArray.ctx.getCacheStream().frames[0])[1] = 13

	assert(contains(errorArray, true)).Equals(false)
}
