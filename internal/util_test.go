package internal

import (
	"errors"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
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

func TestGetFuncKind(t *testing.T) {
	assert := NewAssert(t)

	fn1 := 3
	assert(getFuncKind(reflect.ValueOf(fn1))).
		Equals("", errors.New("handler must be a function"))

	fn2 := func() {}
	assert(getFuncKind(reflect.ValueOf(fn2))).
		Equals("", errors.New("handler 1st argument type must be rpc.Context"))

	fn3 := func(_ chan bool) {}
	assert(getFuncKind(reflect.ValueOf(fn3))).
		Equals("", errors.New("handler 1st argument type must be rpc.Context"))

	fn4 := func(ctx Context, _ bool) {}
	assert(getFuncKind(reflect.ValueOf(fn4))).
		Equals("", errors.New("handler return type must be rpc.Return"))

	fn5 := func(ctx Context, _ bool) (Return, bool) { return nilReturn, true }
	assert(getFuncKind(reflect.ValueOf(fn5))).
		Equals("", errors.New("handler return type must be rpc.Return"))

	fn6 := func(ctx Context, _ bool) bool { return true }
	assert(getFuncKind(reflect.ValueOf(fn6))).
		Equals("", errors.New("handler return type must be rpc.Return"))

	fn7 := func(ctx Context,
		_ bool, _ int64, _ uint64, _ float64,
		_ string, _ Bytes, _ Array, _ Map,
	) Return {
		return ctx.OK(true)
	}
	assert(getFuncKind(reflect.ValueOf(fn7))).Equals("BIUFSXAM", nil)

	fn8 := func(ctx Context,
		_ int32, _ int64, _ uint64, _ float64,
		_ string, _ Bytes, _ Array, _ Map,
	) Return {
		return ctx.OK(true)
	}
	assert(getFuncKind(reflect.ValueOf(fn8))).
		Equals("", errors.New("handler 2nd argument type int32 is not supported"))

	fn9 := func(ctx Context,
		_ bool, _ int32, _ uint64, _ float64,
		_ string, _ Bytes, _ Array, _ Map,
	) Return {
		return ctx.OK(true)
	}
	assert(getFuncKind(reflect.ValueOf(fn9))).
		Equals("", errors.New("handler 3rd argument type int32 is not supported"))

	fn10 := func(ctx Context,
		_ bool, _ int64, _ int32, _ float64,
		_ string, _ Bytes, _ Array, _ Map,
	) Return {
		return ctx.OK(true)
	}
	assert(getFuncKind(reflect.ValueOf(fn10))).
		Equals("", errors.New("handler 4th argument type int32 is not supported"))

	fn11 := func(ctx Context,
		_ bool, _ int64, _ uint64, _ int32,
		_ string, _ Bytes, _ Array, _ Map,
	) Return {
		return ctx.OK(true)
	}
	assert(getFuncKind(reflect.ValueOf(fn11))).
		Equals("", errors.New("handler 5th argument type int32 is not supported"))

	fn12 := func(ctx Context,
		_ bool, _ int64, _ uint64, _ float64,
		_ int32, _ Bytes, _ Array, _ Map,
	) Return {
		return ctx.OK(true)
	}
	assert(getFuncKind(reflect.ValueOf(fn12))).
		Equals("", errors.New("handler 6th argument type int32 is not supported"))

	fn13 := func(ctx Context,
		_ bool, _ int64, _ uint64, _ float64,
		_ string, _ int32, _ Array, _ Map,
	) Return {
		return ctx.OK(true)
	}
	assert(getFuncKind(reflect.ValueOf(fn13))).
		Equals("", errors.New("handler 7th argument type int32 is not supported"))

	fn14 := func(ctx Context,
		_ bool, _ int64, _ uint64, _ float64,
		_ string, _ Bytes, _ int32, _ Map,
	) Return {
		return ctx.OK(true)
	}
	assert(getFuncKind(reflect.ValueOf(fn14))).
		Equals("", errors.New("handler 8th argument type int32 is not supported"))

	fn15 := func(ctx Context,
		_ bool, _ int64, _ uint64, _ float64,
		_ string, _ Bytes, _ Array, _ int32,
	) Return {
		return ctx.OK(true)
	}
	assert(getFuncKind(reflect.ValueOf(fn15))).
		Equals("", errors.New("handler 9th argument type int32 is not supported"))
}

func TestConvertTypeToString(t *testing.T) {
	assert := NewAssert(t)
	assert(convertTypeToString(nil)).Equals("<nil>")
	assert(convertTypeToString(bytesType)).Equals("rpc.Bytes")
	assert(convertTypeToString(arrayType)).Equals("rpc.Array")
	assert(convertTypeToString(mapType)).Equals("rpc.Map")
	assert(convertTypeToString(boolType)).Equals("rpc.Bool")
	assert(convertTypeToString(int64Type)).Equals("rpc.Int64")
	assert(convertTypeToString(uint64Type)).Equals("rpc.Uint64")
	assert(convertTypeToString(float64Type)).Equals("rpc.Float64")
	assert(convertTypeToString(stringType)).Equals("rpc.String")
	assert(convertTypeToString(contextType)).Equals("rpc.Context")
	assert(convertTypeToString(returnType)).Equals("rpc.Return")
	assert(convertTypeToString(reflect.ValueOf(make(chan bool)).Type())).
		Equals("chan bool")
}

func TestConvertToIsoDateString(t *testing.T) {
	assert := NewAssert(t)
	start, _ := time.Parse(
		"2006-01-02T15:04:05.999Z07:00",
		"0001-01-01T00:00:00+00:00",
	)

	for i := 0; i < 1000000; i++ {
		parseTime, err := time.Parse(
			"2006-01-02T15:04:05.999Z07:00",
			ConvertToIsoDateString(start),
		)
		assert(err).IsNil()
		assert(parseTime.UnixNano()).Equals(start.UnixNano())
		start = start.Add(271099197000000)
	}

	smallTime, _ := time.Parse(
		"2006-01-02T15:04:05.999Z07:00",
		"0000-01-01T00:00:00+00:00",
	)
	assert(ConvertToIsoDateString(smallTime)).
		Equals("0000-01-01T00:00:00.000+00:00")

	largeTime, _ := time.Parse(
		"2006-01-02T15:04:05.999Z07:00",
		"9998-01-01T00:00:00+00:00",
	)
	largeTime = largeTime.Add(1000000 * time.Hour)
	assert(ConvertToIsoDateString(largeTime)).
		Equals("9999-01-30T16:00:00.000+00:00")

	time1, _ := time.Parse(
		"2006-01-02T15:04:05.999Z07:00",
		"2222-12-22T11:11:11.333-11:59",
	)
	assert(ConvertToIsoDateString(time1)).
		Equals("2222-12-22T11:11:11.333-11:59")

	time2, _ := time.Parse(
		"2006-01-02T15:04:05.999Z07:00",
		"2222-12-22T11:11:11.333+11:59",
	)
	assert(ConvertToIsoDateString(time2)).
		Equals("2222-12-22T11:11:11.333+11:59")

	time3, _ := time.Parse(
		"2006-01-02T15:04:05.999Z07:00",
		"2222-12-22T11:11:11.333+00:00",
	)
	assert(ConvertToIsoDateString(time3)).
		Equals("2222-12-22T11:11:11.333+00:00")

	time4, _ := time.Parse(
		"2006-01-02T15:04:05.999Z07:00",
		"2222-12-22T11:11:11.333-00:00",
	)
	assert(ConvertToIsoDateString(time4)).
		Equals("2222-12-22T11:11:11.333+00:00")
}

func TestTimeNow(t *testing.T) {
	assert := NewAssert(t)

	for i := 0; i < 10000000; i++ {
		now := TimeNow()
		assert(TimeNow().Sub(now) < 30*time.Millisecond).IsTrue()
		assert(TimeNow().Sub(now) > -30*time.Millisecond).IsTrue()
	}

	for i := 0; i < 10; i++ {
		now := TimeNow()
		time.Sleep(50 * time.Millisecond)
		assert(TimeNow().Sub(now) < 70*time.Millisecond).IsTrue()
		assert(TimeNow().Sub(now) > 30*time.Millisecond).IsTrue()
	}
}

func TestTimeNowISOString(t *testing.T) {
	assert := NewAssert(t)

	for i := 0; i < 1000000; i++ {
		if now, err := time.Parse(
			"2006-01-02T15:04:05.999Z07:00",
			TimeNowISOString(),
		); err == nil {
			assert(time.Now().Sub(now) < 30*time.Millisecond).IsTrue()
			assert(time.Now().Sub(now) > -20*time.Millisecond).IsTrue()
		} else {
			assert().Fail("time parse error")
		}
	}

	for i := 0; i < 1000000; i++ {
		atomic.StorePointer(&timeNowPointer, nil)
		if now, err := time.Parse(
			"2006-01-02T15:04:05.999Z07:00",
			TimeNowISOString(),
		); err == nil {
			assert(time.Now().Sub(now) < 30*time.Millisecond).IsTrue()
			assert(time.Now().Sub(now) > -20*time.Millisecond).IsTrue()
		} else {
			assert().Fail("time parse error")
		}
	}
}

func TestGetRandString(t *testing.T) {
	assert := NewAssert(t)
	assert(GetRandString(-1)).Equals("")
	for i := 0; i < 100; i++ {
		assert(len(GetRandString(i))).Equals(i)
	}
}
func TestGetSeed(t *testing.T) {
	assert := NewAssert(t)
	seed := GetSeed()
	assert(seed > 10000).IsTrue()

	for i := int64(0); i < 1000; i++ {
		assert(GetSeed()).Equals(seed + 1 + i)
	}
}

func TestAddPrefixPerLine(t *testing.T) {
	assert := NewAssert(t)

	assert(AddPrefixPerLine("", "")).Equals("")
	assert(AddPrefixPerLine("a", "")).Equals("a")
	assert(AddPrefixPerLine("\n", "")).Equals("\n")
	assert(AddPrefixPerLine("a\n", "")).Equals("a\n")
	assert(AddPrefixPerLine("a\nb", "")).Equals("a\nb")
	assert(AddPrefixPerLine("", "-")).Equals("-")
	assert(AddPrefixPerLine("a", "-")).Equals("-a")
	assert(AddPrefixPerLine("\n", "-")).Equals("-\n")
	assert(AddPrefixPerLine("a\n", "-")).Equals("-a\n")
	assert(AddPrefixPerLine("a\nb", "-")).Equals("-a\n-b")
}

func TestConcatString(t *testing.T) {
	assert := NewAssert(t)

	assert(ConcatString("", "")).Equals("")
	assert(ConcatString("a", "")).Equals("a")
	assert(ConcatString("", "b")).Equals("b")
	assert(ConcatString("a", "b")).Equals("ab")
	assert(ConcatString("a", "b", "")).Equals("ab")
	assert(ConcatString("a", "b", "c")).Equals("abc")
}

func TestConvertOrdinalToString(t *testing.T) {
	assert := NewAssert(t)

	assert(ConvertOrdinalToString(0)).Equals("")
	assert(ConvertOrdinalToString(1)).Equals("1st")
	assert(ConvertOrdinalToString(2)).Equals("2nd")
	assert(ConvertOrdinalToString(3)).Equals("3rd")
	assert(ConvertOrdinalToString(4)).Equals("4th")
	assert(ConvertOrdinalToString(10)).Equals("10th")
	assert(ConvertOrdinalToString(100)).Equals("100th")
}

func TestAddFileLine(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	fileLine1 := AddFileLine("header", 0)
	assert(strings.HasPrefix(fileLine1, "header ")).IsTrue()
	assert(strings.Contains(fileLine1, "util_test.go")).IsTrue()

	// Test(2)
	fileLine2 := AddFileLine("", 0)
	assert(strings.HasPrefix(fileLine2, " ")).IsFalse()
	assert(strings.Contains(fileLine2, "util_test.go")).IsTrue()

	// Test(3)
	assert(AddFileLine("header", 1000)).Equals("header")
}

func TestGetFileLine(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	fileLine1 := GetFileLine(0)
	assert(strings.Contains(fileLine1, "util_test.go")).IsTrue()
}

func TestCurrentGoroutineID(t *testing.T) {
	assert := NewAssert(t)
	idMap := make(map[int64]bool)
	lock := NewLock()
	waitCH := make(chan bool)
	testCount := 100000

	for i := 0; i < testCount; i++ {
		go func() {
			id := CurrentGoroutineID()
			assert(id > 0).IsTrue()

			lock.DoWithLock(func() {
				idMap[id] = true
				waitCH <- true
			})
		}()
	}

	for i := 0; i < testCount; i++ {
		<-waitCH
	}
	assert(len(idMap)).Equals(testCount)

	// make fake error
	temp := goroutinePrefix
	goroutinePrefix = "fake "
	assert(CurrentGoroutineID()).Equals(int64(0))
	goroutinePrefix = temp
}

func BenchmarkAddPrefixPerLine(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		AddPrefixPerLine("a\nb\nc", "test")
	}
}

func BenchmarkConcatString(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		ConcatString("a", "b")
	}
}

func BenchmarkGetCodePosition(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			AddFileLine("test", 0)
		}
	})
}

func BenchmarkGetRandString(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		GetRandString(128)
	}
}

func BenchmarkTimeNow(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		TimeNow()
	}
}

func BenchmarkTimeNowISOString(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		TimeNowISOString()
	}
}

func BenchmarkCurGoroutineID(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			CurrentGoroutineID()
		}
	})
}

func BenchmarkRunWithPanicCatch(b *testing.B) {
	a := uint64(0)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			testRunWithCatchPanic(func() {
				a = a + 1
			})
		}
	})
}
