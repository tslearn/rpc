package core

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"math/rand"
	"runtime"
	"sort"
	"strconv"
	"testing"
	"unsafe"
)

func getTestMapItems(size int, sequencePos bool) []mapItem {
	ret := make([]mapItem, 0, size)
	mp := map[string]bool{}
	pos := 0
	for pos < size {
		str := base.GetRandString(rand.Int() % 6)
		if _, ok := mp[str]; !ok {
			mp[str] = true
			if sequencePos {
				ret = append(ret, mapItem{str, getFastKey(str), posRecord(pos)})
			} else {
				ret = append(ret, mapItem{str, getFastKey(str), posRecord(1)})
			}

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
			items := getTestMapItems(4, true)
			v1 := getSort4(items, 0)
			sort.Slice(items, func(i, j int) bool {
				return isMapItemLess(&items[i], &items[j])
			})
			v2 := uint64(0)
			for i := len(items) - 1; i >= 0; i-- {
				v2 <<= 4
				v2 |= uint64(0xFFFFFFFFFFFF0000 | items[i].pos)
			}
			assert(v1).Equal(v2)
		}
	})
}

func TestGetSort8(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 10000; i++ {
			items := getTestMapItems(8, true)
			v1 := getSort8(items, 0)
			sort.Slice(items, func(i, j int) bool {
				return isMapItemLess(&items[i], &items[j])
			})
			v2 := uint64(0)
			for i := len(items) - 1; i >= 0; i-- {
				v2 <<= 4
				v2 |= uint64(items[i].pos)
			}
			assert(v1).Equal(0xFFFFFFFF00000000 | v2)
		}
	})
}

func TestGetSort16(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 10000; i++ {
			items := getTestMapItems(16, true)
			v1 := getSort16(items)

			sort.Slice(items, func(i, j int) bool {
				return isMapItemLess(&items[i], &items[j])
			})
			v2 := uint64(0)
			for i := len(items) - 1; i >= 0; i-- {
				v2 <<= 4
				v2 |= uint64(items[i].pos)
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

	t.Run("test thread safe", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTMap(7)
		wait := make(chan bool)
		for i := 0; i < 20; i++ {
			go func(idx int) {
				for j := 0; j < 100; j++ {
					assert(v.Set(fmt.Sprintf("%d-%d", idx, j), idx)).IsNil()
				}
				wait <- true
			}(i)
			runtime.GC()
		}
		for i := 0; i < 20; i++ {
			<-wait
		}
		assert(v.Size()).Equal(2000)
		sum := int64(0)
		for i := 0; i < 20; i++ {
			for j := 0; j < 100; j++ {
				v, _ := v.Get(fmt.Sprintf("%d-%d", i, j)).ToInt64()
				sum += v
			}
		}
		assert(sum).Equal(int64(19000))
	})
}

func TestRTMap_Get(t *testing.T) {
	t.Run("invalid RTMap", func(t *testing.T) {
		assert := base.NewAssert(t)
		rtMap := RTMap{}
		assert(rtMap.Get("name").err).Equal(
			errors.ErrRuntimeIllegalInCurrentGoroutine,
		)
	})

	t.Run("key exists", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTMap(0)
		_ = v.Set("name", "kitty")
		_ = v.Set("age", uint64(18))
		assert(v.Get("name").ToString()).Equal("kitty", nil)
		assert(v.Get("age").ToUint64()).Equal(uint64(18), nil)
		assert(v.Get("noKey").ToString()).Equal(
			"",
			errors.ErrRTMapNameNotFound.AddDebug("RTMap key noKey does not exist"))
	})

	t.Run("key does not exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTMap(0)
		assert(v.Get("name").err).Equal(
			errors.ErrRTMapNameNotFound.AddDebug("RTMap key name does not exist"),
		)
		assert(v.Get("age").err).Equal(
			errors.ErrRTMapNameNotFound.AddDebug("RTMap key age does not exist"),
		)
	})
}

func TestRTMap_Set(t *testing.T) {
	t.Run("invalid RTMap", func(t *testing.T) {
		assert := base.NewAssert(t)
		rtMap := RTMap{}
		assert(rtMap.Set("name", "kitty")).Equal(
			errors.ErrRuntimeIllegalInCurrentGoroutine,
		)
	})

	t.Run("unsupported value", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTMap(0)
		assert(v.Set("name", make(chan bool))).Equal(
			errors.ErrUnsupportedValue.AddDebug(
				"value type(chan bool) is not supported",
			),
		)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTMap(0)
		_ = v.Set("name", "kitty")
		assert(v.Get("name").ToString()).Equal("kitty", nil)
		assert(v.Get("name").cacheBytes).Equal([]byte("kitty"))
		_ = v.Set("name", "doggy")
		assert(v.Get("name").ToString()).Equal("doggy", nil)
		assert(v.Get("name").cacheBytes).Equal([]byte("doggy"))

		_ = v.Set("age", 3)
		assert(v.Get("age").ToInt64()).Equal(int64(3), nil)
		assert(v.Get("age").cacheSafe).Equal(true)
		assert(v.Get("age").cacheBytes).Equal([]byte(nil))
		assert(v.Get("age").cacheError).Equal(errors.ErrStream)
		_ = v.Set("age", 6)
		assert(v.Get("age").ToInt64()).Equal(int64(6), nil)
		assert(v.Get("age").cacheSafe).Equal(true)
		assert(v.Get("age").cacheBytes).Equal([]byte(nil))
		assert(v.Get("age").cacheError).Equal(errors.ErrStream)
	})
}

func TestRTMap_Delete(t *testing.T) {
	t.Run("invalid RTMap", func(t *testing.T) {
		assert := base.NewAssert(t)
		rtMap := RTMap{}
		assert(rtMap.Delete("name")).Equal(
			errors.ErrRuntimeIllegalInCurrentGoroutine,
		)
	})

	t.Run("name does not exist 1", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTMap(0)
		assert(v.Delete("name")).Equal(
			errors.ErrRTMapNameNotFound.AddDebug("RTMap key name does not exist"),
		)
	})

	t.Run("name does not exist 2", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTMap(0)
		_ = v.Set("name", "kitty")
		_ = v.Delete("name")
		assert(v.Delete("name")).Equal(
			errors.ErrRTMapNameNotFound.AddDebug("RTMap key name does not exist"),
		)
	})

	t.Run("name exists", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTMap(0)
		_ = v.Set("name", "kitty")
		assert(v.Delete("name")).Equal(nil)
	})
}

func TestRTMap_DeleteAll(t *testing.T) {
	t.Run("invalid RTMap", func(t *testing.T) {
		assert := base.NewAssert(t)
		rtMap := RTMap{}
		assert(rtMap.DeleteAll()).Equal(
			errors.ErrRuntimeIllegalInCurrentGoroutine,
		)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		rtMap := testRuntime.NewRTMap(0)
		for i := 0; i < 100; i++ {
			for j := 0; j < 100; j++ {
				_ = rtMap.Set(strconv.Itoa(j), j)
			}
			assert(rtMap.Size()).Equal(100)
			preCap := cap(*rtMap.items)
			assert(rtMap.DeleteAll()).Equal(nil)
			assert(rtMap.Size()).Equal(0)
			assert(len(*rtMap.items), cap(*rtMap.items)).Equal(0, preCap)
		}
	})
}

func TestRTMap_Size(t *testing.T) {
	t.Run("invalid RTMap", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := RTMap{}
		assert(v.Size()).Equal(-1)
	})

	t.Run("valid RTMap 1", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTMap(0)
		assert(v.Size()).Equal(0)
	})

	t.Run("valid RTMap 2", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTMap(0)
		_ = v.Set("name", "kitty")
		assert(v.Size()).Equal(1)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 100; i++ {
			testRuntime.thread.Reset()
			v := testRuntime.NewRTMap(0)
			items := getTestMapItems(i, false)
			for _, it := range items {
				assert(v.Set(it.key, true)).Equal(nil)
			}
			assert(v.Size()).Equal(i)
			for _, it := range items {
				assert(v.Delete(it.key)).Equal(nil)
			}
			assert(v.Size()).Equal(0)
		}
	})
}

func TestRTMap_getPosRecord(t *testing.T) {
	t.Run("key exists", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 1; i < 600; i += 13 {
			testRuntime.thread.Reset()
			v := testRuntime.NewRTMap(0)

			items := getTestMapItems(i, false)

			for _, it := range items {
				_ = v.Set(it.key, true)
			}

			for _, it := range items {
				idx, pos := v.getPosRecord(it.key, getFastKey(it.key))
				assert(pos > 0).IsTrue()
				assert((*v.items)[idx].key).Equal(it.key)
			}
		}
	})

	t.Run("key does not exist", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 1; i < 600; i += 13 {
			testRuntime.thread.Reset()
			v := testRuntime.NewRTMap(0)
			items := getTestMapItems(i, false)
			for _, it := range items {
				_ = v.Set(it.key, true)
			}

			for j := 0; j < 600; j++ {
				key := base.GetRandString(6 + rand.Int()%6)
				assert(v.getPosRecord(key, getFastKey(key))).Equal(-1, posRecord(0))
			}
		}
	})
}

func TestRTMap_appendValue(t *testing.T) {
	t.Run("key does not exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTMap(0)
		v.appendValue("name", 1)
		assert(len(*v.items)).Equal(1)
		assert((*v.items)[0].pos).Equal(posRecord(1))
	})

	t.Run("key exists", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTMap(0)
		v.appendValue("name", 1)
		v.appendValue("name", 2)
		assert(len(*v.items)).Equal(1)
		assert((*v.items)[0].pos).Equal(posRecord(2))
	})

	t.Run("sort ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 50; i++ {
			testRuntime.thread.Reset()
			v := testRuntime.NewRTMap(0)
			items := getTestMapItems(i*16, false)
			for _, it := range items {
				v.appendValue(it.key, it.pos)
			}
			sort.Slice(items, func(i, j int) bool {
				return isMapItemLess(&items[i], &items[j])
			})
			assert(*v.items).Equal(items)
		}
	})
}

func TestRTMap_sort(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 50; i++ {
			testRuntime.thread.Reset()
			v := testRuntime.NewRTMap(0)
			items := getTestMapItems(i*16, false)
			for _, it := range items {
				v.appendValue(it.key, it.pos)
			}
			sort.Slice(items, func(i, j int) bool {
				return isMapItemLess(&items[i], &items[j])
			})
			assert(*v.items).Equal(items)
		}
	})
}
