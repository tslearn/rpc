package common

import (
	"strconv"
	"testing"
)

func Test_RPCString_OK(t *testing.T) {
	assert := NewAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}
	invalidCtx := &rpcContext{
		inner: nil,
	}

	assert(RPCString{
		ctx:    validCtx,
		status: rpcStatusAllocated,
		bytes:  ([]byte)("hello"),
	}.OK()).IsTrue()
	assert(RPCString{
		ctx:    invalidCtx,
		status: rpcStatusAllocated,
		bytes:  ([]byte)("hello"),
	}.OK()).IsFalse()
}

func Test_RPCBytes_OK(t *testing.T) {
	assert := NewAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}
	invalidCtx := &rpcContext{
		inner: nil,
	}

	assert(RPCBytes{ctx: validCtx, status: rpcStatusAllocated, bytes: []byte{1, 2}}.OK()).IsTrue()
	assert(RPCBytes{ctx: invalidCtx, status: rpcStatusAllocated, bytes: []byte{1, 2}}.OK()).IsFalse()
}

func Test_RPCArray_newRPCArray(t *testing.T) {
	assert := NewAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}
	invalidCtx := &rpcContext{
		inner: nil,
	}
	assert(newRPCArray(nil)).Equals(newRPCArray(validCtx))
	assert(newRPCArray(invalidCtx)).Equals(nilRPCArray)
}

func Test_RPCArray_Release(t *testing.T) {
	assert := NewAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}

	array0 := newRPCArray(validCtx)
	array0.release()
	assert(array0.in).IsNil()

	array1 := newRPCArray(nil)
	for i := 0; i < 100; i++ {
		array1.Append(20)
	}
	array1.release()
	assert(array1.in).IsNil()
}

func Test_RPCArray_getStream(t *testing.T) {
	assert := NewAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}
	invalidCtx := &rpcContext{
		inner: nil,
	}
	validArray := newRPCArray(validCtx)
	invalidArray := newRPCArray(invalidCtx)
	assert(validArray.ctx.getCacheStream()).IsNotNil()
	assert(invalidArray.ctx.getCacheStream()).IsNil()
	assert(invalidArray.Size()).Equals(0)
}

func Test_RPCArray_Get(t *testing.T) {
	assert := NewAssert(t)
	testArray := make([]interface{}, 16, 16)

	testArray[0] = nil
	testArray[1] = false
	testArray[2] = float64(3.14)
	testArray[3] = int64(30000)
	testArray[4] = uint64(30000)
	testArray[5] = ""
	testArray[6] = "hello"
	testArray[7] = []byte{}
	testArray[8] = []byte{0x53}
	testArray[9] = newRPCArray(nil)
	testArray10 := newRPCArray(nil)
	testArray10.Append("world")
	testArray[10] = testArray10
	testArray[11] = newRPCMap(nil)
	testSmallArray12 := newRPCMap(nil)
	testSmallArray12.Set("hello", "world")
	testArray[12] = testSmallArray12

	fnTestArray := func(array []interface{}, index int, tp string) {
		array[len(array)-1] = array[index]

		inner := &rpcInnerContext{
			stream: NewRPCStream(),
		}
		ctx := &rpcContext{
			inner: inner,
		}
		rpcArray := newRPCArray(nil)
		for _, v := range array {
			rpcArray.Append(v)
		}

		stream := NewRPCStream()
		stream.Write(rpcArray)
		arr, _ := stream.ReadRPCArray(ctx)

		switch tp {
		case "nil":
			assert(arr.GetNil(arr.Size() - 1)).Equals(true)
			assert(arr.ctx.getCacheStream().GetWritePos()).Equals(arr.ctx.getCacheStream().GetReadPos())
			assert(arr.GetNil(index)).Equals(true)
			assert(arr.GetNil(arr.Size())).Equals(false)
			ctx.close()
			assert(arr.GetNil(arr.Size() - 1)).Equals(false)
			assert(arr.GetNil(index)).Equals(false)
		case "bool":
			assert(arr.GetBool(arr.Size()-1)).Equals(array[index], true)
			assert(arr.ctx.getCacheStream().GetWritePos()).Equals(arr.ctx.getCacheStream().GetReadPos())
			assert(arr.GetBool(index)).Equals(array[index], true)
			assert(arr.GetBool(arr.Size())).Equals(false, false)
			ctx.close()
			assert(arr.GetBool(arr.Size()-1)).Equals(false, false)
			assert(arr.GetBool(index)).Equals(false, false)
		case "float64":
			assert(arr.GetFloat64(arr.Size()-1)).Equals(array[index], true)
			assert(arr.ctx.getCacheStream().GetWritePos()).Equals(arr.ctx.getCacheStream().GetReadPos())
			assert(arr.GetFloat64(index)).Equals(array[index], true)
			assert(arr.GetFloat64(arr.Size())).Equals(float64(0), false)
			ctx.close()
			assert(arr.GetFloat64(arr.Size()-1)).Equals(float64(0), false)
			assert(arr.GetFloat64(index)).Equals(float64(0), false)
		case "int64":
			assert(arr.GetInt64(arr.Size()-1)).Equals(array[index], true)
			assert(arr.ctx.getCacheStream().GetWritePos()).Equals(arr.ctx.getCacheStream().GetReadPos())
			assert(arr.GetInt64(index)).Equals(array[index], true)
			assert(arr.GetInt64(arr.Size())).Equals(int64(0), false)
			ctx.close()
			assert(arr.GetInt64(arr.Size()-1)).Equals(int64(0), false)
			assert(arr.GetInt64(index)).Equals(int64(0), false)
		case "uint64":
			assert(arr.GetUint64(arr.Size()-1)).Equals(array[index], true)
			assert(arr.ctx.getCacheStream().GetWritePos()).Equals(arr.ctx.getCacheStream().GetReadPos())
			assert(arr.GetUint64(index)).Equals(array[index], true)
			assert(arr.GetUint64(arr.Size())).Equals(uint64(0), false)
			ctx.close()
			assert(arr.GetUint64(arr.Size()-1)).Equals(uint64(0), false)
			assert(arr.GetUint64(index)).Equals(uint64(0), false)
		case "rpcString":
			assert(arr.GetRPCString(arr.Size()-1).ToString()).Equals(array[index], true)
			assert(arr.ctx.getCacheStream().GetWritePos()).Equals(arr.ctx.getCacheStream().GetReadPos())
			assert(arr.GetRPCString(index).ToString()).Equals(array[index], true)
			assert(arr.GetRPCString(arr.Size()).ToString()).Equals("", false)
			ctx.close()
			assert(arr.GetRPCString(arr.Size()-1).ToString()).Equals("", false)
			assert(arr.GetRPCString(index).ToString()).Equals("", false)
		case "rpcBytes":
			assert(arr.GetRPCBytes(arr.Size()-1).ToBytes()).Equals(array[index], true)
			assert(arr.ctx.getCacheStream().GetWritePos()).Equals(arr.ctx.getCacheStream().GetReadPos())
			assert(arr.GetRPCBytes(index).ToBytes()).Equals(array[index], true)
			assert(arr.GetRPCBytes(arr.Size()).ToBytes()).Equals(nil, false)
			ctx.close()
			assert(arr.GetRPCBytes(arr.Size()-1).ToBytes()).Equals(nil, false)
			assert(arr.GetRPCBytes(index).ToBytes()).Equals(nil, false)
		case "rpcArray":
			target1, ok := arr.GetRPCArray(arr.Size() - 1)
			assert(ok, arr.ctx.getCacheStream().GetWritePos()).Equals(true, arr.ctx.getCacheStream().GetReadPos())
			target2, ok := arr.GetRPCArray(index)
			assert(arr.GetRPCArray(arr.Size())).Equals(nilRPCArray, false)
			ctx.close()
			assert(arr.GetRPCArray(arr.Size()-1)).Equals(nilRPCArray, false)
			assert(arr.GetRPCArray(index)).Equals(nilRPCArray, false)
			assert(target1.ctx).Equals(ctx)
			assert(target2.ctx).Equals(ctx)
		case "rpcMap":
			target1, ok := arr.GetRPCMap(arr.Size() - 1)
			assert(ok, arr.ctx.getCacheStream().GetWritePos()).Equals(true, arr.ctx.getCacheStream().GetReadPos())
			target2, ok := arr.GetRPCMap(index)
			assert(arr.GetRPCMap(arr.Size())).Equals(nilRPCMap, false)
			ctx.close()
			assert(arr.GetRPCMap(arr.Size()-1)).Equals(nilRPCMap, false)
			assert(arr.GetRPCMap(index)).Equals(nilRPCMap, false)
			assert(target1.ctx).Equals(ctx)
			assert(target2.ctx).Equals(ctx)
		}

		ctx.inner = inner
		assert(arr.Get(arr.Size()-1)).Equals(array[index], true)
		assert(arr.ctx.getCacheStream().GetWritePos()).Equals(arr.ctx.getCacheStream().GetReadPos())
		assert(arr.Get(index)).Equals(array[index], true)
		assert(arr.Get(arr.Size())).Equals(nil, false)
		ctx.close()
		assert(arr.Get(arr.Size()-1)).Equals(nil, false)
		assert(arr.Get(index)).Equals(nil, false)
		assert(ctx.close()).IsFalse()
	}

	fnTestArray(testArray, 0, "nil")
	fnTestArray(testArray, 1, "bool")
	fnTestArray(testArray, 2, "float64")
	fnTestArray(testArray, 3, "int64")
	fnTestArray(testArray, 4, "uint64")
	fnTestArray(testArray, 5, "rpcString")
	fnTestArray(testArray, 6, "rpcString")
	fnTestArray(testArray, 7, "rpcBytes")
	fnTestArray(testArray, 8, "rpcBytes")
	fnTestArray(testArray, 9, "rpcArray")
	fnTestArray(testArray, 10, "rpcArray")
	fnTestArray(testArray, 11, "rpcMap")
	fnTestArray(testArray, 12, "rpcMap")
}

func Test_RPCArray_Set(t *testing.T) {
	assert := NewAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}
	invalidCtx := &rpcContext{
		inner: nil,
	}

	array1 := newRPCArray(nil)
	array2 := newRPCArray(validCtx)
	array3 := newRPCArray(invalidCtx)

	assert(array1.AppendNil()).IsTrue()
	assert(array1.SetNil(0)).IsTrue()
	assert(array2.AppendNil()).IsTrue()
	assert(array2.SetNil(0)).IsTrue()
	assert(array3.AppendNil()).IsFalse()
	assert(array3.SetNil(0)).IsFalse()

	assert(array1.AppendBool(false)).IsTrue()
	assert(array1.SetBool(0, false)).IsTrue()
	assert(array2.AppendBool(false)).IsTrue()
	assert(array2.SetBool(0, false)).IsTrue()
	assert(array3.AppendBool(false)).IsFalse()
	assert(array3.SetBool(0, false)).IsFalse()

	assert(array1.AppendFloat64(3.14)).IsTrue()
	assert(array1.SetFloat64(0, 3.14)).IsTrue()
	assert(array2.AppendFloat64(3.14)).IsTrue()
	assert(array2.SetFloat64(0, 3.14)).IsTrue()
	assert(array3.AppendFloat64(3.14)).IsFalse()
	assert(array3.SetFloat64(0, 3.14)).IsFalse()

	assert(array1.AppendInt64(100)).IsTrue()
	assert(array1.SetInt64(0, 100)).IsTrue()
	assert(array2.AppendInt64(100)).IsTrue()
	assert(array2.SetInt64(0, 100)).IsTrue()
	assert(array3.AppendInt64(100)).IsFalse()
	assert(array3.SetInt64(0, 100)).IsFalse()

	assert(array1.AppendUint64(100)).IsTrue()
	assert(array1.SetUint64(0, 100)).IsTrue()
	assert(array2.AppendUint64(100)).IsTrue()
	assert(array2.SetUint64(0, 100)).IsTrue()
	assert(array3.AppendUint64(100)).IsFalse()
	assert(array3.SetUint64(0, 100)).IsFalse()

	rpcString := RPCString{status: rpcStatusAllocated, bytes: ([]byte)("hello")}
	assert(array1.AppendRPCString(rpcString)).IsTrue()
	assert(array1.SetRPCString(0, rpcString)).IsTrue()
	assert(array2.AppendRPCString(rpcString)).IsTrue()
	assert(array2.SetRPCString(0, rpcString)).IsTrue()
	assert(array3.AppendRPCString(rpcString)).IsFalse()
	assert(array3.SetRPCString(0, rpcString)).IsFalse()

	rpcBytes := RPCBytes{status: rpcStatusAllocated, bytes: []byte{1, 2, 3}}
	assert(array1.AppendRPCBytes(rpcBytes)).IsTrue()
	assert(array1.SetRPCBytes(0, rpcBytes)).IsTrue()
	assert(array2.AppendRPCBytes(rpcBytes)).IsTrue()
	assert(array2.SetRPCBytes(0, rpcBytes)).IsTrue()
	assert(array3.AppendRPCBytes(rpcBytes)).IsFalse()
	assert(array3.SetRPCBytes(0, rpcBytes)).IsFalse()

	rpcArray := newRPCArray(nil)
	assert(array1.AppendRPCArray(rpcArray)).IsTrue()
	assert(array1.SetRPCArray(0, rpcArray)).IsTrue()
	assert(array2.AppendRPCArray(rpcArray)).IsTrue()
	assert(array2.SetRPCArray(0, rpcArray)).IsTrue()
	assert(array3.AppendRPCArray(rpcArray)).IsFalse()
	assert(array3.SetRPCArray(0, rpcArray)).IsFalse()

	rpcMap := newRPCMap(nil)
	assert(array1.AppendRPCMap(rpcMap)).IsTrue()
	assert(array1.SetRPCMap(0, rpcMap)).IsTrue()
	assert(array2.AppendRPCMap(rpcMap)).IsTrue()
	assert(array2.SetRPCMap(0, rpcMap)).IsTrue()
	assert(array3.AppendRPCMap(rpcMap)).IsFalse()
	assert(array3.SetRPCMap(0, rpcMap)).IsFalse()

	assert(array1.Append("hello")).IsTrue()
	assert(array1.Set(0, "hello")).IsTrue()
	assert(array2.Append("hello")).IsTrue()
	assert(array2.Set(0, "hello")).IsTrue()
	assert(array3.Append("hello")).IsFalse()
	assert(array3.Set(0, "hello")).IsFalse()
}

func Test_RPCMap_newRPCArray(t *testing.T) {
	assert := NewAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}
	invalidCtx := &rpcContext{
		inner: nil,
	}
	assert(newRPCMap(nil)).Equals(newRPCMap(validCtx))
	assert(newRPCMap(invalidCtx)).Equals(nilRPCMap)
}

func Test_RPCMap_Release(t *testing.T) {
	assert := NewAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}

	mp0 := newRPCMap(validCtx)
	mp0.release()
	assert(mp0.in).IsNil()

	mp1 := newRPCMap(nil)
	for i := 0; i < 100; i++ {
		mp1.Set(strconv.Itoa(i), i)
	}

	for i := 0; i < 90; i++ {
		mp1.Delete(strconv.Itoa(i))
	}
	mp1.release()
	assert(mp1.in).IsNil()
}

func Test_RPCMap_getStream(t *testing.T) {
	assert := NewAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}
	invalidCtx := &rpcContext{
		inner: nil,
	}
	validMap := newRPCMap(validCtx)
	invalidMap := newRPCMap(invalidCtx)
	assert(validMap.ctx.getCacheStream()).IsNotNil()
	assert(invalidMap.ctx.getCacheStream()).IsNil()
	assert(invalidMap.Size()).Equals(0)
	assert(len(invalidMap.Keys())).Equals(0)
}

func Test_RPCMap_Get(t *testing.T) {
	assert := NewAssert(t)
	testSmallMap := make(map[string]interface{})
	testLargeMap := make(map[string]interface{})

	testSmallMap["0"] = nil
	testSmallMap["1"] = false
	testSmallMap["2"] = float64(3.14)
	testSmallMap["3"] = int64(30000)
	testSmallMap["4"] = uint64(30000)
	testSmallMap["5"] = ""
	testSmallMap["6"] = "hello"
	testSmallMap["7"] = []byte{}
	testSmallMap["8"] = []byte{0x53}
	testSmallMap["9"] = newRPCArray(nil)
	testSmallMap10 := newRPCArray(nil)
	testSmallMap10.Append("world")
	testSmallMap["10"] = testSmallMap10
	testSmallMap["11"] = newRPCMap(nil)
	testSmallMap12 := newRPCMap(nil)
	testSmallMap12.Set("hello", "world")
	testSmallMap["12"] = testSmallMap12
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
	testLargeMap["9"] = newRPCArray(nil)
	testLargeMap10 := newRPCArray(nil)
	testLargeMap10.Append("world")
	testLargeMap["10"] = testLargeMap10
	testLargeMap["11"] = newRPCMap(nil)
	testLargeMap12 := newRPCMap(nil)
	testLargeMap12.Set("hello", "world")
	testLargeMap["12"] = testLargeMap12
	testLargeMap["13"] = nil
	testLargeMap["14"] = nil
	testLargeMap["15"] = nil
	testLargeMap["16"] = nil

	fnTestMap := func(mp map[string]interface{}, name string, tp string) {
		inner := &rpcInnerContext{
			stream: NewRPCStream(),
		}
		ctx := &rpcContext{
			inner: inner,
		}
		rpcMap := newRPCMap(nil)
		for k, v := range mp {
			rpcMap.Set(k, v)
		}

		stream := NewRPCStream()
		stream.Write(rpcMap)
		sm, _ := stream.ReadRPCMap(ctx)

		switch tp {
		case "nil":
			assert(sm.GetNil(name)).Equals(true)
			assert(sm.GetNil("no")).Equals(false)
			ctx.close()
			assert(sm.GetNil(name)).Equals(false)
		case "bool":
			assert(sm.GetBool(name)).Equals(mp[name], true)
			assert(sm.GetBool("no")).Equals(false, false)
			ctx.close()
			assert(sm.GetBool(name)).Equals(false, false)
		case "float64":
			assert(sm.GetFloat64(name)).Equals(mp[name], true)
			assert(sm.GetFloat64("no")).Equals(float64(0), false)
			ctx.close()
			assert(sm.GetFloat64(name)).Equals(float64(0), false)
		case "int64":
			assert(sm.GetInt64(name)).Equals(mp[name], true)
			assert(sm.GetInt64("no")).Equals(int64(0), false)
			ctx.close()
			assert(sm.GetInt64(name)).Equals(int64(0), false)
		case "uint64":
			assert(sm.GetUint64(name)).Equals(mp[name], true)
			assert(sm.GetUint64("no")).Equals(uint64(0), false)
			ctx.close()
			assert(sm.GetUint64(name)).Equals(uint64(0), false)
		case "rpcString":
			sc := sm.GetRPCString(name)
			assert(sc.ToString()).Equals(mp[name], true)
			assert(sm.GetRPCString("no").ToString()).Equals("", false)
			ctx.close()
			assert(sc.ToString()).Equals("", false)
			assert(sm.GetRPCString(name).ToString()).Equals("", false)
			assert(sc.ctx).Equals(ctx)
		case "rpcBytes":
			sc := sm.GetRPCBytes(name)
			assert(sc.ToBytes()).Equals(mp[name], true)
			assert(sm.GetRPCBytes("no").ToBytes()).Equals(nil, false)
			ctx.close()
			assert(sc.ToBytes()).Equals(nil, false)
			assert(sm.GetRPCBytes(name).ToBytes()).Equals(nil, false)
			assert(sc.ctx).Equals(ctx)
		case "rpcArray":
			target1, ok := sm.GetRPCArray(name)
			assert(ok).Equals(true)
			_, ok = sm.GetRPCArray("no")
			assert(ok).Equals(false)
			ctx.close()
			assert(sm.GetRPCArray(name)).Equals(nilRPCArray, false)
			assert(target1.ctx).Equals(ctx)
		case "rpcMap":
			target1, ok := sm.GetRPCMap(name)
			assert(ok).Equals(true)
			_, ok = sm.GetRPCMap("no")
			assert(ok).Equals(false)
			ctx.close()
			assert(sm.GetRPCMap(name)).Equals(nilRPCMap, false)
			assert(target1.ctx).Equals(ctx)
		}

		ctx.inner = inner
		assert(sm.Get(name)).Equals(mp[name], true)
		assert(sm.Get("no")).Equals(nil, false)
		ctx.close()
		assert(sm.Get(name)).Equals(nil, false)
		assert(ctx.close()).IsFalse()
	}

	fnTestMap(testSmallMap, "0", "nil")
	fnTestMap(testLargeMap, "0", "nil")
	fnTestMap(testSmallMap, "1", "bool")
	fnTestMap(testLargeMap, "1", "bool")
	fnTestMap(testSmallMap, "2", "float64")
	fnTestMap(testLargeMap, "2", "float64")
	fnTestMap(testSmallMap, "3", "int64")
	fnTestMap(testLargeMap, "3", "int64")
	fnTestMap(testSmallMap, "4", "uint64")
	fnTestMap(testLargeMap, "4", "uint64")
	fnTestMap(testSmallMap, "5", "rpcString")
	fnTestMap(testLargeMap, "5", "rpcString")
	fnTestMap(testSmallMap, "6", "rpcString")
	fnTestMap(testLargeMap, "6", "rpcString")
	fnTestMap(testSmallMap, "7", "rpcBytes")
	fnTestMap(testLargeMap, "7", "rpcBytes")
	fnTestMap(testSmallMap, "8", "rpcBytes")
	fnTestMap(testLargeMap, "8", "rpcBytes")
	fnTestMap(testSmallMap, "9", "rpcArray")
	fnTestMap(testLargeMap, "9", "rpcArray")
	fnTestMap(testSmallMap, "10", "rpcArray")
	fnTestMap(testLargeMap, "10", "rpcArray")
	fnTestMap(testSmallMap, "11", "rpcMap")
	fnTestMap(testLargeMap, "11", "rpcMap")
	fnTestMap(testSmallMap, "12", "rpcMap")
	fnTestMap(testLargeMap, "12", "rpcMap")
}

func Test_RPCMap_Set(t *testing.T) {
	assert := NewAssert(t)

	fnTest := func(tp string, name string, value interface{}) {
		invalidCtx := &rpcContext{
			inner: nil,
		}

		map0 := newRPCMap(nil)
		map16 := newRPCMap(nil)
		map100 := newRPCMap(nil)

		for i := 0; i < 16; i++ {
			map16.Set(strconv.Itoa(i), i)
		}
		for i := 0; i < 100; i++ {
			map100.Set(strconv.Itoa(i), i)
		}
		invalidMap := newRPCMap(invalidCtx)

		switch tp {
		case "nil":
			assert(map0.SetNil(name)).IsTrue()
			assert(map16.SetNil(name)).IsTrue()
			assert(map100.SetNil(name)).IsTrue()
			assert(invalidMap.SetNil(name)).IsFalse()
		case "bool":
			assert(map0.SetBool(name, value.(bool))).IsTrue()
			assert(map16.SetBool(name, value.(bool))).IsTrue()
			assert(map100.SetBool(name, value.(bool))).IsTrue()
			assert(invalidMap.SetBool(name, value.(bool))).IsFalse()
		case "int64":
			assert(map0.SetInt64(name, value.(int64))).IsTrue()
			assert(map16.SetInt64(name, value.(int64))).IsTrue()
			assert(map100.SetInt64(name, value.(int64))).IsTrue()
			assert(invalidMap.SetInt64(name, value.(int64))).IsFalse()
		case "uint64":
			assert(map0.SetUint64(name, value.(uint64))).IsTrue()
			assert(map16.SetUint64(name, value.(uint64))).IsTrue()
			assert(map100.SetUint64(name, value.(uint64))).IsTrue()
			assert(invalidMap.SetUint64(name, value.(uint64))).IsFalse()
		case "float64":
			assert(map0.SetFloat64(name, value.(float64))).IsTrue()
			assert(map16.SetFloat64(name, value.(float64))).IsTrue()
			assert(map100.SetFloat64(name, value.(float64))).IsTrue()
			assert(invalidMap.SetFloat64(name, value.(float64))).IsFalse()
		case "rpcString":
			assert(map0.SetRPCString(name, value.(RPCString))).IsTrue()
			assert(map16.SetRPCString(name, value.(RPCString))).IsTrue()
			assert(map100.SetRPCString(name, value.(RPCString))).IsTrue()
			assert(invalidMap.SetRPCString(name, value.(RPCString))).IsFalse()
		case "rpcBytes":
			assert(map0.SetRPCBytes(name, value.(RPCBytes))).IsTrue()
			assert(map16.SetRPCBytes(name, value.(RPCBytes))).IsTrue()
			assert(map100.SetRPCBytes(name, value.(RPCBytes))).IsTrue()
			assert(invalidMap.SetRPCBytes(name, value.(RPCBytes))).IsFalse()
		case "rpcArray":
			assert(map0.SetRPCArray(name, value.(RPCArray))).IsTrue()
			assert(map16.SetRPCArray(name, value.(RPCArray))).IsTrue()
			assert(map100.SetRPCArray(name, value.(RPCArray))).IsTrue()
			assert(invalidMap.SetRPCArray(name, value.(RPCArray))).IsFalse()
		case "rpcMap":
			assert(map0.SetRPCMap(name, value.(RPCMap))).IsTrue()
			assert(map16.SetRPCMap(name, value.(RPCMap))).IsTrue()
			assert(map100.SetRPCMap(name, value.(RPCMap))).IsTrue()
			assert(invalidMap.SetRPCMap(name, value.(RPCMap))).IsFalse()
		}

		assert(map0.Set(name, value)).IsTrue()
		assert(map16.Set(name, value)).IsTrue()
		assert(map100.Set(name, value)).IsTrue()
		assert(invalidMap.Set(name, value)).IsFalse()
	}

	fnTest("nil", "1", nil)
	fnTest("bool", "2", false)
	fnTest("float64", "3", float64(3.14))
	fnTest("int64", "4", int64(23))
	fnTest("uint64", "5", uint64(324))
	fnTest("rpcString", "6", RPCString{status: rpcStatusAllocated, bytes: ([]byte)("hello")})
	fnTest("rpcBytes", "7", RPCBytes{status: rpcStatusAllocated, bytes: []byte{123, 1}})
	fnTest("rpcArray", "8", newRPCArray(nil))
	fnTest("rpcMap", "9", newRPCMap(nil))

	fnTest("nil", "t1", nil)
	fnTest("bool", "t2", false)
	fnTest("float64", "t3", float64(3.14))
	fnTest("int64", "t4", int64(23))
	fnTest("uint64", "t5", uint64(324))
	fnTest("rpcString", "t6", RPCString{status: rpcStatusAllocated, bytes: ([]byte)("hello")})
	fnTest("rpcBytes", "t7", RPCBytes{status: rpcStatusAllocated, bytes: []byte{123, 1}})
	fnTest("rpcArray", "t8", newRPCArray(nil))
	fnTest("rpcMap", "t9", newRPCMap(nil))

	mp := newRPCMap(nil)
	assert(mp.SetRPCString("error", errorRPCString)).IsFalse()
	assert(mp.SetRPCBytes("error", errorRPCBytes)).IsFalse()
	invalidCtx := &rpcContext{
		inner: nil,
	}
	assert(mp.SetRPCArray("error", newRPCArray(invalidCtx))).IsFalse()
	assert(mp.SetRPCMap("error", newRPCMap(invalidCtx))).IsFalse()
	assert(mp.Set("error", make(chan bool))).IsFalse()
}

func Test_RPCMap_Delete(t *testing.T) {
	assert := NewAssert(t)

	rpcMap := newRPCMap(nil)

	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}
	invalidCtx := &rpcContext{
		inner: nil,
	}
	validMap := newRPCMap(validCtx)
	invalidMap := newRPCMap(invalidCtx)

	assert(rpcMap.Delete("")).Equals(false)
	assert(validMap.Delete("")).Equals(false)
	assert(invalidMap.Delete("")).Equals(false)
	assert(nilRPCMap.Delete("")).Equals(false)

	assert(rpcMap.Delete("hi")).Equals(false)
	assert(validMap.Delete("hi")).Equals(false)
	assert(invalidMap.Delete("hi")).Equals(false)
	assert(nilRPCMap.Delete("hi")).Equals(false)
}
