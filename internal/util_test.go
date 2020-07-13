package internal

import (
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

func TestIsNil(t *testing.T) {
	assert := NewRPCAssert(t)
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
	assert := NewRPCAssert(t)

	assert(getFuncKind(nil)).Equals("", false)
	assert(getFuncKind(3)).Equals("", false)
	fn1 := func() {}
	assert(getFuncKind(fn1)).Equals("", false)
	fn2 := func(_ chan bool) {}
	assert(getFuncKind(fn2)).Equals("", false)
	fn3 := func(ctx *RPCContext, _ bool) *RPCReturn { return nilReturn }
	assert(getFuncKind(fn3)).Equals("B", true)
	fn4 := func(ctx *RPCContext, _ int64) *RPCReturn { return nilReturn }
	assert(getFuncKind(fn4)).Equals("I", true)
	fn5 := func(ctx *RPCContext, _ uint64) *RPCReturn { return nilReturn }
	assert(getFuncKind(fn5)).Equals("U", true)
	fn6 := func(ctx *RPCContext, _ float64) *RPCReturn { return nilReturn }
	assert(getFuncKind(fn6)).Equals("F", true)
	fn7 := func(ctx *RPCContext, _ string) *RPCReturn { return nilReturn }
	assert(getFuncKind(fn7)).Equals("S", true)
	fn8 := func(ctx *RPCContext, _ RPCBytes) *RPCReturn { return nilReturn }
	assert(getFuncKind(fn8)).Equals("X", true)
	fn9 := func(ctx *RPCContext, _ RPCArray) *RPCReturn { return nilReturn }
	assert(getFuncKind(fn9)).Equals("A", true)
	fn10 := func(ctx *RPCContext, _ RPCMap) *RPCReturn { return nilReturn }
	assert(getFuncKind(fn10)).Equals("M", true)

	fn11 := func(ctx *RPCContext) *RPCReturn { return nilReturn }
	assert(getFuncKind(fn11)).Equals("", true)

	// no return
	fn12 := func(ctx RPCContext, _ bool) {}
	assert(getFuncKind(fn12)).Equals("", false)

	// value type not supported
	fn13 := func(ctx *RPCContext, _ chan bool) *RPCReturn { return nilReturn }
	assert(getFuncKind(fn13)).Equals("", false)

	fn14 := func(
		ctx *RPCContext,
		_ bool, _ int64, _ uint64, _ float64, _ string,
		_ RPCBytes, _ RPCArray, _ RPCMap,
	) *RPCReturn {
		return nilReturn
	}
	assert(getFuncKind(fn14)).Equals("BIUFSXAM", true)
}

func TestConvertTypeToString(t *testing.T) {
	assert := NewRPCAssert(t)
	assert(convertTypeToString(nil)).Equals("<nil>")
	assert(convertTypeToString(bytesType)).Equals("rpc.Bytes")
	assert(convertTypeToString(arrayType)).Equals("rpc.Array")
	assert(convertTypeToString(mapType)).Equals("rpc.Map")
	assert(convertTypeToString(boolType)).Equals("rpc.Bool")
	assert(convertTypeToString(int64Type)).Equals("rpc.Int")
	assert(convertTypeToString(uint64Type)).Equals("rpc.Uint")
	assert(convertTypeToString(float64Type)).Equals("rpc.Float")
	assert(convertTypeToString(stringType)).Equals("rpc.String")
	assert(convertTypeToString(contextType)).Equals("rpc.Context")
	assert(convertTypeToString(returnType)).Equals("rpc.Return")
	assert(convertTypeToString(reflect.ValueOf(make(chan bool)).Type())).
		Equals("chan bool")
}

func TestGetArgumentsErrorPosition(t *testing.T) {
	assert := NewRPCAssert(t)

	fn1 := func() {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn1))).Equals(0)
	fn2 := func(_ chan bool) {}

	assert(getArgumentsErrorPosition(reflect.ValueOf(fn2))).Equals(0)
	fn3 := func(ctx RPCContext, _ bool, _ chan bool) {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn3))).Equals(2)
	fn4 := func(ctx RPCContext, _ int64, _ chan bool) {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn4))).Equals(2)
	fn5 := func(ctx RPCContext, _ uint64, _ chan bool) {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn5))).Equals(2)
	fn6 := func(ctx RPCContext, _ float64, _ chan bool) {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn6))).Equals(2)
	fn7 := func(ctx RPCContext, _ string, _ chan bool) {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn7))).Equals(2)
	fn8 := func(ctx RPCContext, _ RPCBytes, _ chan bool) {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn8))).Equals(2)
	fn9 := func(ctx RPCContext, _ RPCArray, _ chan bool) {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn9))).Equals(2)
	fn10 := func(ctx RPCContext, _ RPCMap, _ chan bool) {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn10))).Equals(2)

	fn11 := func(ctx RPCContext, _ bool) {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn11))).Equals(-1)
}

func TestConvertToIsoDateString(t *testing.T) {
	assert := NewRPCAssert(t)
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

func TestTimeNowNS(t *testing.T) {
	assert := NewRPCAssert(t)

	for i := 0; i < 500000; i++ {
		nowNS := TimeNowNS()
		assert(time.Now().UnixNano()-nowNS < int64(20*time.Millisecond)).IsTrue()
		assert(time.Now().UnixNano()-nowNS > int64(-20*time.Millisecond)).IsTrue()
	}

	for i := 0; i < 500; i++ {
		nowNS := TimeNowNS()
		assert(time.Now().UnixNano()-nowNS < int64(10*time.Millisecond)).IsTrue()
		assert(time.Now().UnixNano()-nowNS > int64(-10*time.Millisecond)).IsTrue()
		time.Sleep(time.Millisecond)
	}

	// hack timeNowPointer to nil
	atomic.StorePointer(&timeNowPointer, nil)
	for i := 0; i < 500; i++ {
		nowNS := TimeNowNS()
		assert(time.Now().UnixNano()-nowNS < int64(10*time.Millisecond)).IsTrue()
		assert(time.Now().UnixNano()-nowNS > int64(-10*time.Millisecond)).IsTrue()
		time.Sleep(time.Millisecond)
	}
}

func TestTimeNowMS(t *testing.T) {
	assert := NewRPCAssert(t)
	nowNS := TimeNowMS() * int64(time.Millisecond)
	assert(time.Now().UnixNano()-nowNS < int64(10*time.Millisecond)).IsTrue()
	assert(time.Now().UnixNano()-nowNS > int64(-10*time.Millisecond)).IsTrue()
}

func TestTimeNowISOString(t *testing.T) {
	assert := NewRPCAssert(t)

	for i := 0; i < 1000000; i++ {
		if nowNS, err := time.Parse(
			"2006-01-02T15:04:05.999Z07:00",
			TimeNowISOString(),
		); err == nil {
			assert(
				time.Now().UnixNano()-nowNS.UnixNano() < int64(30*time.Millisecond),
			).IsTrue()
			assert(
				time.Now().UnixNano()-nowNS.UnixNano() > int64(-30*time.Millisecond),
			).IsTrue()
		} else {
			assert().Fail()
		}
	}
}

func TestTimeSpanFrom(t *testing.T) {
	assert := NewRPCAssert(t)
	ns := TimeNowNS()
	time.Sleep(50 * time.Millisecond)
	dur := TimeSpanFrom(ns)
	assert(int64(dur) > int64(40*time.Millisecond)).IsTrue()
	assert(int64(dur) < int64(60*time.Millisecond)).IsTrue()
}

func TestTimeSpanBetween(t *testing.T) {
	assert := NewRPCAssert(t)
	start := TimeNowNS()
	time.Sleep(50 * time.Millisecond)
	dur := TimeSpanBetween(start, TimeNowNS())
	assert(int64(dur) > int64(40*time.Millisecond)).IsTrue()
	assert(int64(dur) < int64(60*time.Millisecond)).IsTrue()
}

func TestGetRandString(t *testing.T) {
	assert := NewRPCAssert(t)
	assert(GetRandString(-1)).Equals("")
	for i := 0; i < 100; i++ {
		assert(len(GetRandString(i))).Equals(i)
	}
}
func TestGetSeed(t *testing.T) {
	assert := NewRPCAssert(t)
	seed := GetSeed()
	assert(seed > 10000).IsTrue()

	for i := int64(0); i < 1000; i++ {
		assert(GetSeed()).Equals(seed + 1 + i)
	}
}

func TestAddPrefixPerLine(t *testing.T) {
	assert := NewRPCAssert(t)

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
	assert := NewRPCAssert(t)

	assert(ConcatString("", "")).Equals("")
	assert(ConcatString("a", "")).Equals("a")
	assert(ConcatString("", "b")).Equals("b")
	assert(ConcatString("a", "b")).Equals("ab")
	assert(ConcatString("a", "b", "")).Equals("ab")
	assert(ConcatString("a", "b", "c")).Equals("abc")
}

func TestGetStackString(t *testing.T) {
	assert := NewRPCAssert(t)
	assert(strings.Contains(FindLinesByPrefix(
		GetStackString(0),
		"-01",
	)[0], "TestGetStackString")).IsTrue()
	assert(strings.Contains(FindLinesByPrefix(
		GetStackString(0),
		"-01",
	)[0], "helper_test")).IsTrue()
}

func TestFindLinesByPrefix(t *testing.T) {
	assert := NewRPCAssert(t)

	ret := FindLinesByPrefix("", "")
	assert(len(ret)).Equals(1)
	assert(ret[0]).Equals("")

	ret = FindLinesByPrefix("", "hello")
	assert(len(ret)).Equals(0)

	ret = FindLinesByPrefix("hello", "dd")
	assert(len(ret)).Equals(0)

	ret = FindLinesByPrefix("  hello world", "hello")
	assert(len(ret)).Equals(1)
	assert(ret[0]).Equals("  hello world")

	ret = FindLinesByPrefix(" \t hello world", "hello")
	assert(len(ret)).Equals(1)
	assert(ret[0]).Equals(" \t hello world")

	ret = FindLinesByPrefix(" \t hello world\nhello\n", "hello")
	assert(len(ret)).Equals(2)
	assert(ret[0]).Equals(" \t hello world")
	assert(ret[1]).Equals("hello")
}

func TestConvertOrdinalToString(t *testing.T) {
	assert := NewRPCAssert(t)

	assert(ConvertOrdinalToString(0)).Equals("")
	assert(ConvertOrdinalToString(1)).Equals("1st")
	assert(ConvertOrdinalToString(2)).Equals("2nd")
	assert(ConvertOrdinalToString(3)).Equals("3rd")
	assert(ConvertOrdinalToString(4)).Equals("4th")
	assert(ConvertOrdinalToString(10)).Equals("10th")
	assert(ConvertOrdinalToString(100)).Equals("100th")
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

func BenchmarkGetStackString(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		GetStackString(0)
	}
}

func BenchmarkGetRandString(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		GetRandString(128)
	}
}

func BenchmarkTimeNowNS(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		TimeNowNS()
	}
}

func BenchmarkTimeNowISOString(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		TimeNowISOString()
	}
}
