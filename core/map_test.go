package core

import (
	"fmt"
	"sort"
	"strconv"
	"testing"
)

func TestRpcMapInner_hasKey(t *testing.T) {
	assert := newAssert(t)
	ctx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}

	sMap := newRPCMapByMap(ctx, Map{"0": 0, "1": 1, "2": 2})
	assert(sMap.in.hasKey("0")).IsTrue()
	assert(sMap.in.hasKey("2")).IsTrue()
	assert(sMap.in.hasKey("no")).IsFalse()

	bMap := newRPCMapByMap(ctx, Map{
		"1": true, "2": true, "3": true, "4": true,
		"5": true, "6": true, "7": true, "8": true,
		"9": true, "a": true, "b": true, "c": true,
		"d": true, "e": true, "f": true, "g": true,
		"h": true, "i": true, "j": true, "k": true,
		"l": true, "m": true, "n": true, "o": true,
		"p": true, "q": true, "r": true, "s": true,
		"t": true, "u": true, "v": true, "w": true,
	})

	assert(bMap.in.hasKey("1")).IsTrue()
	assert(bMap.in.hasKey("2")).IsTrue()
	assert(bMap.in.hasKey("v")).IsTrue()
	assert(bMap.in.hasKey("no")).IsFalse()
}

func TestRpcMapInner_getItemPos_setItemPos_deleteItem(t *testing.T) {
	assert := newAssert(t)

	fnTest := func(size int) {
		mapInner := rpcMapInnerCache.Get().(*rpcMapInner)
		assert(mapInner).IsNotNil()

		// set value and get
		for i := 0; i < size; i++ {
			assert(mapInner.setItemPos(strconv.Itoa(i), i)).IsTrue()
		}
		for i := 0; i < size; i++ {
			assert(mapInner.getItemPos(strconv.Itoa(i))).Equals(i)
		}

		// get unset value
		for i := size; i < 2*size; i++ {
			assert(mapInner.getItemPos(strconv.Itoa(i))).Equals(-1)
		}

		// Reset new value
		for i := 0; i < size; i++ {
			assert(mapInner.setItemPos(strconv.Itoa(i), 2*i)).IsTrue()
		}
		for i := 0; i < size; i++ {
			assert(mapInner.getItemPos(strconv.Itoa(i))).Equals(2 * i)
		}

		// delete
		for i := 0; i < size; i++ {
			assert(mapInner.deleteItem(strconv.Itoa(size + i))).IsFalse()
			assert(mapInner.deleteItem(strconv.Itoa(i))).IsTrue()
			assert(mapInner.deleteItem(strconv.Itoa(i))).IsFalse()
		}
	}

	fnTest(0)
	fnTest(1)
	fnTest(8)
	fnTest(15)
	fnTest(16)
	fnTest(17)
	fnTest(47)
	fnTest(99)
	fnTest(1000)
}

func TestRpcMapInner_free(t *testing.T) {
	assert := newAssert(t)

	fnTest := func(size int) {
		mapInner := rpcMapInnerCache.Get().(*rpcMapInner)
		assert(mapInner).IsNotNil()
		assert(len(mapInner.smallMap)).Equals(0)
		assert(cap(mapInner.smallMap)).Equals(16)
		assert(mapInner.largeMap).IsNil()

		// set value and get
		for i := 0; i < size; i++ {
			assert(mapInner.setItemPos(strconv.Itoa(i), i)).IsTrue()
		}

		mapInner.free()
		assert(len(mapInner.smallMap)).Equals(0)
		assert(cap(mapInner.smallMap)).Equals(16)
		assert(mapInner.largeMap).IsNil()
	}

	for i := 0; i < 200; i++ {
		fnTest(i)
	}
}

func TestRpcMap_newRPCMap(t *testing.T) {
	assert := newAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}
	invalidCtx := &rpcContext{}

	assert(newRPCMap(nil).ctx).IsNil()
	assert(newRPCMap(nil).in).IsNil()
	assert(newRPCMap(nil).ok()).IsFalse()

	assert(newRPCMap(invalidCtx).ctx).IsNil()
	assert(newRPCMap(invalidCtx).in).IsNil()
	assert(newRPCMap(invalidCtx).ok()).IsFalse()

	assert(newRPCMap(validCtx).ctx).IsNotNil()
	assert(newRPCMap(validCtx).in).IsNotNil()
	assert(newRPCMap(validCtx).ok()).IsTrue()

	assert(newRPCMapByMap(validCtx, nil).ctx).IsNil()
	assert(newRPCMapByMap(validCtx, nil).in).IsNil()
	assert(newRPCMapByMap(validCtx, nil).ok()).IsFalse()
	assert(newRPCMapByMap(validCtx, nil).Size()).Equals(-1)

	assert(newRPCMapByMap(validCtx, Map{}).ctx).IsNotNil()
	assert(newRPCMapByMap(validCtx, Map{}).in).IsNotNil()
	assert(newRPCMapByMap(validCtx, Map{}).ok()).IsTrue()
	assert(newRPCMapByMap(validCtx, Map{}).Size()).Equals(0)

	assert(newRPCMapByMap(validCtx, Map{"0": true}).ctx).IsNotNil()
	assert(newRPCMapByMap(validCtx, Map{"0": true}).in).IsNotNil()
	assert(newRPCMapByMap(validCtx, Map{"0": true}).ok()).IsTrue()
	assert(newRPCMapByMap(validCtx, Map{"0": true}).Size()).Equals(1)

	assert(newRPCMapByMap(validCtx, Map{"0": nilReturn}).ctx).IsNil()
	assert(newRPCMapByMap(validCtx, Map{"0": nilReturn}).in).IsNil()
	assert(newRPCMapByMap(validCtx, Map{"0": nilReturn}).ok()).IsFalse()
	assert(newRPCMapByMap(validCtx, Map{"0": nilReturn}).Size()).Equals(-1)
}

func TestRpcMap_ok(t *testing.T) {
	assert := newAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}
	invalidCtx := &rpcContext{}

	assert(newRPCMap(validCtx).ok()).IsTrue()
	assert(newRPCMap(nil).ok()).IsFalse()
	assert(newRPCMap(invalidCtx).ok()).IsFalse()
}

func TestRpcMap_release(t *testing.T) {
	assert := newAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}

	nilRPCMap := rpcMap{}
	assert(nilRPCMap.Size()).Equals(-1)
	nilRPCMap.release()
	assert(nilRPCMap.ctx).IsNil()
	assert(nilRPCMap.in).IsNil()

	emptyRPCMap := newRPCMap(validCtx)
	assert(emptyRPCMap.Size()).Equals(0)
	emptyRPCMap.release()
	assert(emptyRPCMap.ctx).IsNil()
	assert(emptyRPCMap.in).IsNil()

	bugRPCMap1 := rpcMap{
		ctx: nil,
		in:  rpcMapInnerCache.Get().(*rpcMapInner),
	}
	assert(bugRPCMap1.Size()).Equals(-1)
	bugRPCMap1.release()
	assert(bugRPCMap1.ctx).IsNil()
	assert(bugRPCMap1.in).IsNil()
}

func TestRpcMap_getIS(t *testing.T) {
	assert := newAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}

	nilRPCMap := rpcMap{}
	assert(nilRPCMap.getIS()).IsNil()

	emptyRPCMap := newRPCMap(validCtx)
	assert(emptyRPCMap.getIS()).IsNotNil()

	bugRPCMap1 := rpcMap{
		ctx: nil,
		in:  rpcMapInnerCache.Get().(*rpcMapInner),
	}
	assert(bugRPCMap1.getIS()).Equals(bugRPCMap1.in, nil)

	bugRPCMap2 := rpcMap{
		ctx: validCtx,
		in:  nil,
	}
	assert(bugRPCMap2.getIS()).Equals(nil, validCtx.inner.stream)
}

func TestRpcMap_equals(t *testing.T) {
	assert := newAssert(t)
	ctx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}

	mp0 := newRPCMap(ctx)
	mp1 := newRPCMapByMap(ctx, Map{"0": 0})
	mp2 := newRPCMapByMap(ctx, Map{"0": 0, "1": 1})
	mp3 := newRPCMapByMap(ctx, Map{"1": 1, "0": 0})
	mp4 := newRPCMapByMap(ctx, Map{"1": 1, "0": 2})

	bugMap := newRPCMapByMap(ctx, Map{"1": 1, "0": 0})
	ctx.getCacheStream().setWritePosUnsafe(bugMap.in.smallMap[0].pos)
	ctx.getCacheStream().putBytes([]byte{11})

	assert(nilRPCMap.equals(rpcMap{})).IsTrue()
	assert(nilRPCMap.equals(mp0)).IsFalse()
	assert(mp0.equals(nilRPCMap)).IsFalse()
	assert(mp0.equals(mp1)).IsFalse()
	assert(mp1.equals(mp2)).IsFalse()
	assert(mp2.equals(mp3)).IsTrue()
	assert(mp3.equals(mp4)).IsFalse()

	assert(mp2.equals(bugMap)).IsFalse()
	assert(bugMap.equals(mp2)).IsFalse()
}

func TestRpcMap_contains(t *testing.T) {
	assert := newAssert(t)
	ctx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}
	mp0 := newRPCMap(ctx)
	mp1 := newRPCMapByMap(ctx, Map{"0": 0, "1": 1})
	bugMap := newRPCMapByMap(ctx, Map{"1": 1, "0": 0})
	ctx.getCacheStream().setWritePosUnsafe(bugMap.in.smallMap[0].pos)
	ctx.getCacheStream().putBytes([]byte{11})

	assert(nilRPCMap.contains(int64(0))).IsFalse()
	assert(mp0.contains(int64(0))).IsFalse()
	assert(mp1.contains(int64(0))).IsTrue()
	assert(bugMap.contains(int64(0))).IsFalse()
}

func TestRpcMap_Size(t *testing.T) {
	assert := newAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}
	invalidCtx := &rpcContext{
		inner: nil,
	}
	validRPCMap := newRPCMap(validCtx)
	invalidRPCMap1 := newRPCMap(invalidCtx)
	invalidRPCMap2 := rpcMap{
		ctx: nil,
		in:  rpcMapInnerCache.Get().(*rpcMapInner),
	}

	for i := 1; i < 522; i++ {
		assert(validRPCMap.Set(strconv.Itoa(i), i)).IsTrue()
		assert(validRPCMap.Size()).Equals(i)

		assert(invalidRPCMap1.Set(strconv.Itoa(i), i)).IsFalse()
		assert(invalidRPCMap1.Size()).Equals(-1)

		assert(invalidRPCMap2.Set(strconv.Itoa(i), i)).IsFalse()
		assert(invalidRPCMap2.Size()).Equals(-1)
	}
}

func TestRpcMap_Keys(t *testing.T) {
	assert := newAssert(t)

	fnTest := func(size int) {
		mp := newRPCMap(&rpcContext{
			inner: &rpcInnerContext{
				stream: newRPCStream(),
			},
		})

		keys := make([]string, size, size)
		for i := 0; i < size; i++ {
			key := strconv.Itoa(i)
			mp.Set(key, i)
			keys[i] = key
		}
		retKeys := mp.Keys()
		sort.Strings(retKeys)
		sort.Strings(keys)

		assert(fmt.Sprint(keys) == fmt.Sprint(retKeys)).IsTrue()
	}

	for i := 0; i < 252; i++ {
		fnTest(i)
	}

	bugMap0 := rpcMap{}
	bugMap1 := rpcMap{
		ctx: &rpcContext{
			inner: &rpcInnerContext{
				stream: newRPCStream(),
			},
		},
		in: nil,
	}
	bugMap2 := rpcMap{
		ctx: nil,
		in:  rpcMapInnerCache.Get().(*rpcMapInner),
	}
	assert(len(bugMap0.Keys())).Equals(0)
	assert(len(bugMap1.Keys())).Equals(0)
	assert(len(bugMap2.Keys())).Equals(0)
}

func TestRpcMap_Get(t *testing.T) {
	assert := newAssert(t)
	testSmallMap := make(Map)
	testLargeMap := make(Map)
	testCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}

	testSmallMap["0"] = nil
	testSmallMap["1"] = false
	testSmallMap["2"] = float64(3.14)
	testSmallMap["3"] = int64(30000)
	testSmallMap["4"] = uint64(30000)
	testSmallMap["5"] = ""
	testSmallMap["6"] = "hello"
	testSmallMap["7"] = []byte{}
	testSmallMap["8"] = []byte{0x53}
	testSmallMap["9"] = newRPCArray(testCtx)
	testSmallMap["10"] = newRPCArrayByArray(testCtx, Array{"world"})
	testSmallMap["11"] = newRPCMap(testCtx)
	testSmallMap["12"] = newRPCMapByMap(testCtx, Map{"hello": "world"})
	testSmallMap["13"] = nil
	testSmallMap["14"] = nil
	testSmallMap["15"] = nil

	testLargeMap["0"] = nil
	testLargeMap["1"] = false
	testLargeMap["2"] = float64(3.14)
	testLargeMap["3"] = int64(30000)
	testLargeMap["4"] = uint64(30000)
	testLargeMap["5"] = ""
	testLargeMap["6"] = "hello"
	testLargeMap["7"] = []byte{}
	testLargeMap["8"] = []byte{0x53}
	testLargeMap["9"] = newRPCArray(testCtx)
	testLargeMap["10"] = newRPCArrayByArray(testCtx, Array{"world"})
	testLargeMap["11"] = newRPCMap(testCtx)
	testLargeMap["12"] = newRPCMapByMap(testCtx, Map{"hello": "world"})
	testLargeMap["13"] = nil
	testLargeMap["14"] = nil
	testLargeMap["15"] = nil
	testLargeMap["16"] = nil

	fnTestMap := func(mp map[string]interface{}, name string, tp string) {
		ctx := &rpcContext{
			inner: &rpcInnerContext{
				stream: newRPCStream(),
			},
		}

		sm0 := newRPCMapByMap(ctx, mp)
		sm1 := rpcMap{
			ctx: ctx,
			in:  nil,
		}
		sm2 := rpcMap{
			ctx: nil,
			in:  rpcMapInnerCache.Get().(*rpcMapInner),
		}
		sm3 := rpcMap{
			ctx: nil,
			in:  nil,
		}

		assert(sm0.Get(name)).Equals(mp[name], true)
		assert(sm0.Get("")).Equals(nil, false)
		assert(sm0.Get("no")).Equals(nil, false)
		assert(sm1.Get(name)).Equals(nil, false)
		assert(sm2.Get(name)).Equals(nil, false)
		assert(sm3.Get(name)).Equals(nil, false)
		switch tp {
		case "nil":
			assert(sm0.GetNil(name)).Equals(true)
			assert(sm0.GetNil("")).Equals(false)
			assert(sm0.GetNil("no")).Equals(false)
			assert(sm1.GetNil(name)).Equals(false)
			assert(sm2.GetNil(name)).Equals(false)
			assert(sm3.GetNil(name)).Equals(false)
			ctx.close()
			assert(sm0.GetNil(name)).Equals(false)
		case "bool":
			assert(sm0.GetBool(name)).Equals(mp[name], true)
			assert(sm0.GetBool("")).Equals(false, false)
			assert(sm0.GetBool("no")).Equals(false, false)
			assert(sm1.GetBool(name)).Equals(false, false)
			assert(sm2.GetBool(name)).Equals(false, false)
			assert(sm3.GetBool(name)).Equals(false, false)
			ctx.close()
			assert(sm0.GetBool(name)).Equals(false, false)
		case "float64":
			assert(sm0.GetFloat64(name)).Equals(mp[name], true)
			assert(sm0.GetFloat64("")).Equals(float64(0), false)
			assert(sm0.GetFloat64("no")).Equals(float64(0), false)
			assert(sm1.GetFloat64(name)).Equals(float64(0), false)
			assert(sm2.GetFloat64(name)).Equals(float64(0), false)
			assert(sm3.GetBool(name)).Equals(false, false)
			ctx.close()
			assert(sm0.GetFloat64(name)).Equals(float64(0), false)
		case "int64":
			assert(sm0.GetInt64(name)).Equals(mp[name], true)
			assert(sm0.GetInt64("")).Equals(int64(0), false)
			assert(sm0.GetInt64("no")).Equals(int64(0), false)
			assert(sm1.GetInt64(name)).Equals(int64(0), false)
			assert(sm2.GetInt64(name)).Equals(int64(0), false)
			assert(sm3.GetInt64(name)).Equals(int64(0), false)
			ctx.close()
			assert(sm0.GetInt64(name)).Equals(int64(0), false)
		case "uint64":
			assert(sm0.GetUint64(name)).Equals(mp[name], true)
			assert(sm0.GetUint64("")).Equals(uint64(0), false)
			assert(sm0.GetUint64("no")).Equals(uint64(0), false)
			assert(sm1.GetUint64(name)).Equals(uint64(0), false)
			assert(sm2.GetUint64(name)).Equals(uint64(0), false)
			assert(sm3.GetUint64(name)).Equals(uint64(0), false)
			ctx.close()
			assert(sm0.GetUint64(name)).Equals(uint64(0), false)
		case "string":
			assert(sm0.GetString(name)).Equals(mp[name], true)
			assert(sm0.GetString("")).Equals("", false)
			assert(sm0.GetString("no")).Equals("", false)
			assert(sm1.GetString(name)).Equals("", false)
			assert(sm2.GetString(name)).Equals("", false)
			assert(sm3.GetString(name)).Equals("", false)
			ctx.close()
			assert(sm0.GetString(name)).Equals("", false)
		case "bytes":
			assert(sm0.GetBytes(name)).Equals(mp[name], true)
			assert(sm0.GetBytes("")).Equals(emptyBytes, false)
			assert(sm0.GetBytes("no")).Equals(emptyBytes, false)
			assert(sm1.GetBytes(name)).Equals(emptyBytes, false)
			assert(sm2.GetBytes(name)).Equals(emptyBytes, false)
			assert(sm3.GetBytes(name)).Equals(emptyBytes, false)
			ctx.close()
			assert(sm0.GetBytes(name)).Equals(emptyBytes, false)
		case "rpcArray":
			target1, ok := sm0.GetRPCArray(name)
			assert(ok).IsTrue()
			assert(sm0.GetRPCArray("")).Equals(nilRPCArray, false)
			assert(sm0.GetRPCArray("no")).Equals(nilRPCArray, false)
			assert(sm1.GetRPCArray(name)).Equals(nilRPCArray, false)
			assert(sm2.GetRPCArray(name)).Equals(nilRPCArray, false)
			assert(sm3.GetRPCArray(name)).Equals(nilRPCArray, false)
			ctx.close()
			assert(sm0.GetRPCArray(name)).Equals(nilRPCArray, false)
			assert(target1.ctx).Equals(ctx)
		case "rpcMap":
			target1, ok := sm0.GetRPCMap(name)
			assert(ok).Equals(true)
			assert(sm0.GetRPCMap("")).Equals(nilRPCMap, false)
			assert(sm0.GetRPCMap("no")).Equals(nilRPCMap, false)
			assert(sm1.GetRPCMap(name)).Equals(nilRPCMap, false)
			assert(sm2.GetRPCMap(name)).Equals(nilRPCMap, false)
			assert(sm3.GetRPCMap(name)).Equals(nilRPCMap, false)
			ctx.close()
			assert(sm0.GetRPCMap(name)).Equals(nilRPCMap, false)
			assert(target1.ctx).Equals(ctx)
		}

		assert(sm0.Get(name)).Equals(nil, false)
		assert(ctx.close()).IsFalse()
	}

	fnTestMap(testSmallMap, "0", "nil")
	fnTestMap(testSmallMap, "1", "bool")
	fnTestMap(testSmallMap, "2", "float64")
	fnTestMap(testSmallMap, "3", "int64")
	fnTestMap(testSmallMap, "4", "uint64")
	fnTestMap(testSmallMap, "5", "string")
	fnTestMap(testSmallMap, "6", "string")
	fnTestMap(testSmallMap, "7", "bytes")
	fnTestMap(testSmallMap, "8", "bytes")
	fnTestMap(testSmallMap, "9", "rpcArray")
	fnTestMap(testSmallMap, "10", "rpcArray")
	fnTestMap(testSmallMap, "11", "rpcMap")
	fnTestMap(testSmallMap, "12", "rpcMap")

	fnTestMap(testLargeMap, "0", "nil")
	fnTestMap(testLargeMap, "1", "bool")
	fnTestMap(testLargeMap, "2", "float64")
	fnTestMap(testLargeMap, "3", "int64")
	fnTestMap(testLargeMap, "4", "uint64")
	fnTestMap(testLargeMap, "5", "string")
	fnTestMap(testLargeMap, "6", "string")
	fnTestMap(testLargeMap, "7", "bytes")
	fnTestMap(testLargeMap, "8", "bytes")
	fnTestMap(testLargeMap, "9", "rpcArray")
	fnTestMap(testLargeMap, "10", "rpcArray")
	fnTestMap(testLargeMap, "11", "rpcMap")
	fnTestMap(testLargeMap, "12", "rpcMap")
}

func Test_RPCMap_Set(t *testing.T) {
	assert := newAssert(t)
	ctx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}

	fnTest := func(tp string, name string, value interface{}) {
		validCtx := &rpcContext{
			inner: &rpcInnerContext{
				stream: newRPCStream(),
			},
		}
		invalidCtx := &rpcContext{
			inner: nil,
		}

		map0 := newRPCMap(validCtx)
		map16 := newRPCMap(validCtx)
		map100 := newRPCMap(validCtx)

		for i := 0; i < 16; i++ {
			map16.Set(strconv.Itoa(i), i)
		}
		for i := 0; i < 100; i++ {
			map100.Set(strconv.Itoa(i), i)
		}
		invalidMap0 := newRPCMap(invalidCtx)
		invalidMap1 := rpcMap{
			ctx: validCtx,
			in:  nil,
		}
		invalidMap2 := rpcMap{
			ctx: nil,
			in:  rpcMapInnerCache.Get().(*rpcMapInner),
		}

		switch tp {
		case "nil":
			assert(map0.SetNil(name)).IsTrue()
			assert(map16.SetNil(name)).IsTrue()
			assert(map100.SetNil(name)).IsTrue()
			assert(invalidMap0.SetNil(name)).IsFalse()
			assert(invalidMap1.SetNil(name)).IsFalse()
			assert(invalidMap2.SetNil(name)).IsFalse()
		case "bool":
			assert(map0.SetBool(name, value.(bool))).IsTrue()
			assert(map16.SetBool(name, value.(bool))).IsTrue()
			assert(map100.SetBool(name, value.(bool))).IsTrue()
			assert(invalidMap0.SetBool(name, value.(bool))).IsFalse()
			assert(invalidMap1.SetBool(name, value.(bool))).IsFalse()
			assert(invalidMap2.SetBool(name, value.(bool))).IsFalse()
		case "int64":
			assert(map0.SetInt64(name, value.(int64))).IsTrue()
			assert(map16.SetInt64(name, value.(int64))).IsTrue()
			assert(map100.SetInt64(name, value.(int64))).IsTrue()
			assert(invalidMap0.SetInt64(name, value.(int64))).IsFalse()
			assert(invalidMap1.SetInt64(name, value.(int64))).IsFalse()
			assert(invalidMap2.SetInt64(name, value.(int64))).IsFalse()
		case "uint64":
			assert(map0.SetUint64(name, value.(uint64))).IsTrue()
			assert(map16.SetUint64(name, value.(uint64))).IsTrue()
			assert(map100.SetUint64(name, value.(uint64))).IsTrue()
			assert(invalidMap0.SetUint64(name, value.(uint64))).IsFalse()
			assert(invalidMap1.SetUint64(name, value.(uint64))).IsFalse()
			assert(invalidMap2.SetUint64(name, value.(uint64))).IsFalse()
		case "float64":
			assert(map0.SetFloat64(name, value.(float64))).IsTrue()
			assert(map16.SetFloat64(name, value.(float64))).IsTrue()
			assert(map100.SetFloat64(name, value.(float64))).IsTrue()
			assert(invalidMap0.SetFloat64(name, value.(float64))).IsFalse()
			assert(invalidMap1.SetFloat64(name, value.(float64))).IsFalse()
			assert(invalidMap2.SetFloat64(name, value.(float64))).IsFalse()
		case "string":
			assert(map0.SetString(name, value.(string))).IsTrue()
			assert(map16.SetString(name, value.(string))).IsTrue()
			assert(map100.SetString(name, value.(string))).IsTrue()
			assert(invalidMap0.SetString(name, value.(string))).IsFalse()
			assert(invalidMap1.SetString(name, value.(string))).IsFalse()
			assert(invalidMap2.SetString(name, value.(string))).IsFalse()
		case "bytes":
			assert(map0.SetBytes(name, value.([]byte))).IsTrue()
			assert(map16.SetBytes(name, value.([]byte))).IsTrue()
			assert(map100.SetBytes(name, value.([]byte))).IsTrue()
			assert(invalidMap0.SetBytes(name, value.([]byte))).IsFalse()
			assert(invalidMap1.SetBytes(name, value.([]byte))).IsFalse()
			assert(invalidMap2.SetBytes(name, value.([]byte))).IsFalse()
		case "rpcArray":
			assert(map0.SetRPCArray(name, value.(rpcArray))).IsTrue()
			assert(map16.SetRPCArray(name, value.(rpcArray))).IsTrue()
			assert(map100.SetRPCArray(name, value.(rpcArray))).IsTrue()
			assert(map0.SetRPCArray(name, rpcArray{})).IsFalse()
			assert(map16.SetRPCArray(name, rpcArray{})).IsFalse()
			assert(map100.SetRPCArray(name, rpcArray{})).IsFalse()
			assert(invalidMap0.SetRPCArray(name, value.(rpcArray))).IsFalse()
			assert(invalidMap1.SetRPCArray(name, value.(rpcArray))).IsFalse()
			assert(invalidMap2.SetRPCArray(name, value.(rpcArray))).IsFalse()
		case "rpcMap":
			assert(map0.SetRPCMap(name, value.(rpcMap))).IsTrue()
			assert(map16.SetRPCMap(name, value.(rpcMap))).IsTrue()
			assert(map100.SetRPCMap(name, value.(rpcMap))).IsTrue()
			assert(map0.SetRPCMap(name, rpcMap{})).IsFalse()
			assert(map16.SetRPCMap(name, rpcMap{})).IsFalse()
			assert(map100.SetRPCMap(name, rpcMap{})).IsFalse()
			assert(invalidMap0.SetRPCMap(name, value.(rpcMap))).IsFalse()
			assert(invalidMap1.SetRPCMap(name, value.(rpcMap))).IsFalse()
			assert(invalidMap2.SetRPCMap(name, value.(rpcMap))).IsFalse()
		}

		assert(map0.Set(name, value)).IsTrue()
		assert(map16.Set(name, value)).IsTrue()
		assert(map100.Set(name, value)).IsTrue()
		assert(invalidMap0.Set(name, value)).IsFalse()
		assert(invalidMap1.Set(name, value)).IsFalse()
		assert(invalidMap2.Set(name, value)).IsFalse()
	}

	fnTest("nil", "1", nil)
	fnTest("bool", "2", false)
	fnTest("float64", "3", float64(3.14))
	fnTest("int64", "4", int64(23))
	fnTest("uint64", "5", uint64(324))
	fnTest("string", "6", "hello")
	fnTest("bytes", "7", []byte{123, 1})
	fnTest("rpcArray", "8", newRPCArray(ctx))
	fnTest("rpcMap", "9", newRPCMap(ctx))

	fnTest("nil", "t1", nil)
	fnTest("bool", "t2", false)
	fnTest("float64", "t3", float64(3.14))
	fnTest("int64", "t4", int64(23))
	fnTest("uint64", "t5", uint64(324))
	fnTest("string", "t6", "hello")
	fnTest("bytes", "t7", []byte{123, 1})
	fnTest("rpcArray", "t8", newRPCArray(ctx))
	fnTest("rpcMap", "t9", newRPCMap(ctx))
}

func Test_RPCMap_Delete(t *testing.T) {
	assert := newAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}

	map0 := newRPCMap(validCtx)
	map16 := newRPCMap(validCtx)
	map17 := newRPCMap(validCtx)
	map100 := newRPCMap(validCtx)
	bugMap0 := rpcMap{}
	bugMap1 := rpcMap{
		ctx: validCtx,
		in:  nil,
	}
	bugMap2 := rpcMap{
		ctx: nil,
		in:  rpcMapInnerCache.Get().(*rpcMapInner),
	}

	for i := 1; i <= 16; i++ {
		map16.Set(strconv.Itoa(i), i)
	}
	for i := 1; i <= 17; i++ {
		map17.Set(strconv.Itoa(i), i)
	}
	for i := 1; i <= 100; i++ {
		map100.Set(strconv.Itoa(i), i)
	}

	assert(map0.Delete("")).IsFalse()
	assert(map16.Delete("")).IsFalse()
	assert(map17.Delete("")).IsFalse()
	assert(map100.Delete("")).IsFalse()
	assert(bugMap0.Delete("")).IsFalse()
	assert(bugMap1.Delete("")).IsFalse()
	assert(bugMap2.Delete("")).IsFalse()

	assert(map0.Delete("no")).IsFalse()
	assert(map16.Delete("no")).IsFalse()
	assert(map17.Delete("no")).IsFalse()
	assert(map100.Delete("no")).IsFalse()
	assert(bugMap0.Delete("no")).IsFalse()
	assert(bugMap1.Delete("no")).IsFalse()
	assert(bugMap2.Delete("no")).IsFalse()

	assert(map0.Delete("1")).IsFalse()
	assert(map16.Delete("1")).IsTrue()
	assert(map17.Delete("1")).IsTrue()
	assert(map100.Delete("1")).IsTrue()

	assert(map16.Delete("16")).IsTrue()
	assert(map17.Delete("16")).IsTrue()
	assert(map100.Delete("16")).IsTrue()

	assert(map17.Delete("17")).IsTrue()
	assert(map100.Delete("17")).IsTrue()

}
