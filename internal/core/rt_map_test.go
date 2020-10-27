package core

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"math/rand"
	"sort"
	"testing"
	"unsafe"
)

func getTestMapItems(size int) []mapItem {
	ret := make([]mapItem, 0, size)
	mp := map[string]bool{}
	pos := 0
	for pos < size {
		str := base.GetRandString(rand.Int() % 6)
		if _, ok := mp[str]; !ok {
			mp[str] = true
			ret = append(ret, mapItem{str, getFastKey(str), posRecord(pos)})
			pos++
		}
	}
	return ret
}

func TestCompareMapItem(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		assert(compareMapItem(
			&mapItem{"b", getFastKey("b"), 0},
			"a", getFastKey("a")),
		).Equal(1)

		assert(compareMapItem(
			&mapItem{"a", getFastKey("a"), 0},
			"b", getFastKey("b")),
		).Equal(-1)

		assert(compareMapItem(
			&mapItem{"aaa", getFastKey("aaa"), 0},
			"a", getFastKey("a")),
		).Equal(1)

		assert(compareMapItem(
			&mapItem{"a", getFastKey("a"), 0},
			"aaa", getFastKey("aaa")),
		).Equal(-1)

		assert(compareMapItem(
			&mapItem{"", getFastKey(""), 0},
			"", getFastKey("")),
		).Equal(0)

		assert(compareMapItem(
			&mapItem{"a", getFastKey("a"), 0},
			"a", getFastKey("a")),
		).Equal(0)

		assert(compareMapItem(
			&mapItem{"hello", getFastKey("hello"), 0},
			"hello", getFastKey("hello")),
		).Equal(0)
	})
}

func TestIsMapItemLess(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		assert(isMapItemLess(
			&mapItem{"b", getFastKey("b"), 0},
			&mapItem{"a", getFastKey("a"), 0},
		)).IsFalse()

		assert(isMapItemLess(
			&mapItem{"a", getFastKey("a"), 0},
			&mapItem{"b", getFastKey("b"), 0},
		)).IsTrue()

		assert(isMapItemLess(
			&mapItem{"aaa", getFastKey("aaa"), 0},
			&mapItem{"a", getFastKey("a"), 0},
		)).IsFalse()

		assert(isMapItemLess(
			&mapItem{"a", getFastKey("a"), 0},
			&mapItem{"aaa", getFastKey("aaa"), 0},
		)).IsTrue()

		assert(isMapItemLess(
			&mapItem{"", getFastKey(""), 0},
			&mapItem{"", getFastKey(""), 0},
		)).IsFalse()

		assert(isMapItemLess(
			&mapItem{"a", getFastKey("a"), 0},
			&mapItem{"a", getFastKey("a"), 0},
		)).IsFalse()

		assert(isMapItemLess(
			&mapItem{"hello", getFastKey("hello"), 0},
			&mapItem{"hello", getFastKey("hello"), 0},
		)).IsFalse()
	})
}

func TestGetSort4(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 10000; i++ {
			items := getTestMapItems(4)
			v1 := getSort4(items, 0)
			sort.Slice(items, func(i, j int) bool {
				return isMapItemLess(&items[i], &items[j])
			})
			v2 := uint32(0)
			for i := len(items) - 1; i >= 0; i-- {
				v2 <<= 4
				v2 |= uint32(items[i].pos)
			}
			assert(v1).Equal(v2)
		}
	})
}

func TestGetSort8(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 10000; i++ {
			items := getTestMapItems(8)
			v1 := getSort8(items)
			sort.Slice(items, func(i, j int) bool {
				return isMapItemLess(&items[i], &items[j])
			})
			v2 := uint32(0)
			for i := len(items) - 1; i >= 0; i-- {
				v2 <<= 4
				v2 |= uint32(items[i].pos)
			}
			assert(v1).Equal(v2)
		}
	})
}

func TestRTMap(t *testing.T) {
	t.Run("check constant", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(sizeOfMapItem).Equal(int(unsafe.Sizeof(mapItem{})))
	})
}

func TestNewRTMap(t *testing.T) {
	t.Run("thread is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(newRTMap(Runtime{}, 2)).Equal(RTMap{})
	})

	t.Run("no cache", func(t *testing.T) {
		assert := base.NewAssert(t)

		// consume cache
		testRuntime.thread.rootFrame.Reset()
		for i := 0; i < len(testRuntime.thread.cache); i++ {
			testRuntime.thread.malloc(1)
		}

		v := newRTMap(testRuntime, 2)
		assert(v.rt).Equal(testRuntime)
		assert(len(*v.items), cap(*v.items)).Equal(0, 2)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 0; i < 100; i++ {
			testRuntime.thread.rootFrame.Reset()
			v := newRTMap(testRuntime, i)
			assert(v.rt).Equal(testRuntime)
			assert(len(*v.items), cap(*v.items)).Equal(0, i)
		}
	})
}

func TestRTMap_swapUint32(t *testing.T) {
	rtMap := testRuntime.NewRTMap()

	rtMap.appendValue("a2", 34234)
	rtMap.appendValue("v4", 34234)
	rtMap.appendValue("v6", 34234)
	rtMap.appendValue("a1", 34234)
	rtMap.appendValue("a4", 34234)
	rtMap.appendValue("a3", 34234)
	rtMap.appendValue("a5", 34234)
	rtMap.appendValue("a6", 34234)

	//rtMap.appendValue("v1", 34234)
	//rtMap.appendValue("v8", 34234)
	//rtMap.appendValue("v7", 34234)
	//rtMap.appendValue("a7", 34234)
	//rtMap.appendValue("a8", 34234)
	//rtMap.appendValue("v5", 34234)
	//rtMap.appendValue("v2", 34234)
	//rtMap.appendValue("v3", 34234)

	fmt.Println(rtMap.items)
}

func BenchmarkMakeRequestStream(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		testRuntime.thread.rtStream.Reset()
		testRuntime.thread.rootFrame.Reset()
		rtMap := testRuntime.NewRTMap()
		rtMap.appendValue("v1", 34234)
		rtMap.appendValue("v8", 34234)
		rtMap.appendValue("v7", 34234)
		rtMap.appendValue("a7", 34234)
		rtMap.appendValue("a8", 34234)
		rtMap.appendValue("v5", 34234)
		rtMap.appendValue("v2", 34234)
		rtMap.appendValue("v3", 34234)
		//rtMap.appendValue("a2", 34234)
		//rtMap.appendValue("v4", 34234)
		//rtMap.appendValue("v6", 34234)
		//rtMap.appendValue("a1", 34234)
		//rtMap.appendValue("a4", 34234)
		//rtMap.appendValue("a3", 34234)
		//rtMap.appendValue("a5", 34234)
		//rtMap.appendValue("a6", 34234)
	}
}
