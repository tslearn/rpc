package internal

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"runtime"
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

	assert(getFuncKind(nil)).Equals("", false)
	assert(getFuncKind(3)).Equals("", false)
	fn1 := func() {}
	assert(getFuncKind(fn1)).Equals("", false)
	fn2 := func(_ chan bool) {}
	assert(getFuncKind(fn2)).Equals("", false)
	fn3 := func(ctx Context, _ bool) Return { return nilReturn }
	assert(getFuncKind(fn3)).Equals("B", true)
	fn4 := func(ctx Context, _ int64) Return { return nilReturn }
	assert(getFuncKind(fn4)).Equals("I", true)
	fn5 := func(ctx Context, _ uint64) Return { return nilReturn }
	assert(getFuncKind(fn5)).Equals("U", true)
	fn6 := func(ctx Context, _ float64) Return { return nilReturn }
	assert(getFuncKind(fn6)).Equals("F", true)
	fn7 := func(ctx Context, _ string) Return { return nilReturn }
	assert(getFuncKind(fn7)).Equals("S", true)
	fn8 := func(ctx Context, _ Bytes) Return { return nilReturn }
	assert(getFuncKind(fn8)).Equals("X", true)
	fn9 := func(ctx Context, _ Array) Return { return nilReturn }
	assert(getFuncKind(fn9)).Equals("A", true)
	fn10 := func(ctx Context, _ Map) Return { return nilReturn }
	assert(getFuncKind(fn10)).Equals("M", true)

	fn11 := func(ctx Context) Return { return nilReturn }
	assert(getFuncKind(fn11)).Equals("", true)

	// no return
	fn12 := func(ctx Context, _ bool) {}
	assert(getFuncKind(fn12)).Equals("", false)

	// value type not supported
	fn13 := func(ctx Context, _ chan bool) Return { return nilReturn }
	assert(getFuncKind(fn13)).Equals("", false)

	fn14 := func(
		ctx Context,
		_ bool, _ int64, _ uint64, _ float64, _ string,
		_ Bytes, _ Array, _ Map,
	) Return {
		return nilReturn
	}
	assert(getFuncKind(fn14)).Equals("BIUFSXAM", true)
}

func TestBuildFuncCache(t *testing.T) {
	assert := NewAssert(t)
	_, file, _, _ := runtime.Caller(0)
	fnReadFromFile := func(filePath string) (string, error) {
		ret, err := ioutil.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		return string(ret), nil
	}
	defer func() {
		_ = os.RemoveAll(path.Join(path.Dir(file), "_tmp_"))
	}()
	// Test(1)
	tmpFile1 := path.Join(path.Dir(file), "_tmp_/test-cache-01.go")
	snapshotFile1 := path.Join(path.Dir(file), "snapshot/test-cache-01.snapshot")
	assert(buildFuncCache("pkgName", tmpFile1, []string{})).IsNil()
	assert(fnReadFromFile(tmpFile1)).Equals(fnReadFromFile(snapshotFile1))

	// Test(2)
	tmpFile2 := path.Join(path.Dir(file), "_tmp_/test-cache-02.go")
	snapshotFile2 := path.Join(path.Dir(file), "snapshot/test-cache-02.snapshot")
	assert(buildFuncCache("pkgName", tmpFile2, []string{
		"BMUF", "UUIB", "MSXA", "FFFFF",
		"",
		"B", "I", "U", "F", "S", "X", "A", "M",
		"BU", "FI", "AU", "FM", "SX", "BX", "MA", "MI",
		"BUF", "ABM", "UFS", "XAA", "MMS", "MMM", "AAA", "MFF",
		"BIUFSXAM", "AAAAAAAA", "MAXSFUIB",
	})).IsNil()
	assert(fnReadFromFile(tmpFile2)).Equals(fnReadFromFile(snapshotFile2))

	// Test(3)
	tmpFile3 := path.Join(path.Dir(file), "_tmp_/test-cache-03.go")
	assert(buildFuncCache("pkgName", tmpFile3, []string{"A", "A"})).
		Equals(errors.New("duplicate kind A"))

	// Test(4)
	tmpFile4 := path.Join(path.Dir(file), "_tmp_/test-cache-04.go")
	assert(buildFuncCache("pkgName", tmpFile4, []string{"Y", "A"})).
		Equals(errors.New("error kind Y"))

	// Test(5)
	tmpFile5 := path.Join(path.Dir(file), "/util_test.go/test-cache-05.go")
	assert(strings.Contains(
		buildFuncCache("pkgName", tmpFile5, []string{"A"}).Error(),
		"util_test.go: not a directory",
	)).IsTrue()

	// Test(6)
	assert(strings.Contains(
		buildFuncCache("pkgName", "/", []string{"A"}).Error(),
		"is a directory",
	)).IsTrue()
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

func TestGetArgumentsErrorPosition(t *testing.T) {
	assert := NewAssert(t)

	fn1 := func() {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn1))).Equals(0)
	fn2 := func(_ chan bool) {}

	assert(getArgumentsErrorPosition(reflect.ValueOf(fn2))).Equals(0)
	fn3 := func(ctx Context, _ bool, _ chan bool) {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn3))).Equals(2)
	fn4 := func(ctx Context, _ int64, _ chan bool) {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn4))).Equals(2)
	fn5 := func(ctx Context, _ uint64, _ chan bool) {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn5))).Equals(2)
	fn6 := func(ctx Context, _ float64, _ chan bool) {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn6))).Equals(2)
	fn7 := func(ctx Context, _ string, _ chan bool) {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn7))).Equals(2)
	fn8 := func(ctx Context, _ Bytes, _ chan bool) {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn8))).Equals(2)
	fn9 := func(ctx Context, _ Array, _ chan bool) {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn9))).Equals(2)
	fn10 := func(ctx Context, _ Map, _ chan bool) {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn10))).Equals(2)

	fn11 := func(ctx Context, _ bool) {}
	assert(getArgumentsErrorPosition(reflect.ValueOf(fn11))).Equals(-1)
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
		assert(time.Now().Sub(now) < 20*time.Millisecond).IsTrue()
		assert(time.Now().Sub(now) > -20*time.Millisecond).IsTrue()
	}

	for i := 0; i < 10; i++ {
		now := TimeNow()
		time.Sleep(50 * time.Millisecond)
		assert(time.Now().Sub(now) < 70*time.Millisecond).IsTrue()
		assert(time.Now().Sub(now) > 30*time.Millisecond).IsTrue()
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

func TestCheckValue(t *testing.T) {
	assert := NewAssert(t)
	assert(CheckValue(nil, 1)).Equals("")
	assert(CheckValue(Array(nil), 1)).Equals("")
	assert(CheckValue(Map(nil), 1)).Equals("")
	assert(CheckValue(true, 1)).Equals("")
	assert(CheckValue(1, 1)).Equals("")
	assert(CheckValue(int8(1), 1)).Equals("")
	assert(CheckValue(int16(1), 1)).Equals("")
	assert(CheckValue(int32(1), 1)).Equals("")
	assert(CheckValue(int64(1), 1)).Equals("")
	assert(CheckValue(uint(1), 1)).Equals("")
	assert(CheckValue(uint8(1), 1)).Equals("")
	assert(CheckValue(uint16(1), 1)).Equals("")
	assert(CheckValue(uint32(1), 1)).Equals("")
	assert(CheckValue(uint64(1), 1)).Equals("")
	assert(CheckValue(float32(1), 1)).Equals("")
	assert(CheckValue(float64(1), 1)).Equals("")
	assert(CheckValue("", 1)).Equals("")
	assert(CheckValue(Bytes{}, 1)).Equals("")
	assert(CheckValue(Array{}, 1)).Equals("")
	assert(CheckValue(Map{}, 1)).Equals("")

	assert(CheckValue(nil, 0)).
		Equals("value is too complicated")
	assert(CheckValue(make(chan bool), 1)).
		Equals("value type (chan bool) is not supported")

	assert(CheckValue(Array{true}, 1)).
		Equals("value[0] is too complicated")
	assert(CheckValue(Array{true}, 2)).Equals("")
	assert(CheckValue(Map{"key": "value"}, 1)).
		Equals("value[\"key\"] is too complicated")
	assert(CheckValue(Map{"key": "value"}, 2)).Equals("")
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
			testRunWithPanicCatch(func() {
				a = a + 1
			})
		}
	})
}
