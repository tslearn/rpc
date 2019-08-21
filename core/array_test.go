package core

import (
	"testing"
)

func TestRpcArrayInner_free(t *testing.T) {
	assert := newAssert(t)
	for i := 0; i < 522; i++ {
		arrayInner := rpcArrayInnerCache.New().(*rpcArrayInner)
		assert(arrayInner).IsNotNil()
		assert(len(arrayInner.items), cap(arrayInner.items)).Equals(0, 32)
		for n := 0; n < i; n++ {
			arrayInner.items = append(arrayInner.items, n)
		}
		assert(len(arrayInner.items)).Equals(i)
		assert(len(arrayInner.items), cap(arrayInner.items) >= i).Equals(i, true)
		arrayInner.free()
		assert(len(arrayInner.items), cap(arrayInner.items)).Equals(0, 32)
	}
}

func TestRpcArray_newRPCArray(t *testing.T) {
	assert := newAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}
	invalidCtx := &rpcContext{
		inner: nil,
	}
	assert(newRPCArray(validCtx).ctx).Equals(validCtx)
	assert(newRPCArray(validCtx).in).IsNotNil()
	assert(newRPCArray(nil).ctx).IsNil()
	assert(newRPCArray(nil).in).IsNil()
	assert(newRPCArray(invalidCtx).ctx).IsNil()
	assert(newRPCArray(invalidCtx).in).IsNil()
	assert(nilRPCArray.ctx).IsNil()
	assert(nilRPCArray.in).IsNil()

	assert(newRPCArrayByArray(validCtx, nil).ctx).IsNil()
	assert(newRPCArrayByArray(validCtx, nil).in).IsNil()

	assert(newRPCArrayByArray(validCtx, []interface{}{nilContext}).ctx).IsNil()
	assert(newRPCArrayByArray(validCtx, []interface{}{nilContext}).in).IsNil()

	assert(newRPCArrayByArray(validCtx, []interface{}{}).ctx).Equals(validCtx)
	assert(newRPCArrayByArray(validCtx, []interface{}{}).ok()).IsTrue()
	assert(newRPCArrayByArray(validCtx, []interface{}{}).Size()).Equals(0)

	assert(newRPCArrayByArray(validCtx, []interface{}{1}).ctx).Equals(validCtx)
	assert(newRPCArrayByArray(validCtx, []interface{}{1}).ok()).IsTrue()
	assert(newRPCArrayByArray(validCtx, []interface{}{1}).Size()).Equals(1)
}

func TestRpcArray_ok(t *testing.T) {
	assert := newAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}

	assert(nilRPCArray.ok()).IsFalse()
	assert(newRPCArray(validCtx).ok()).IsTrue()
	assert(rpcArray{ctx: validCtx}.ok()).IsFalse()
}

func TestRpcArray_release(t *testing.T) {
	assert := newAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}

	nilRPCArray := rpcArray{}
	assert(nilRPCArray.Size()).Equals(-1)
	nilRPCArray.release()
	assert(nilRPCArray.ctx).IsNil()
	assert(nilRPCArray.in).IsNil()

	emptyRPCArray := newRPCArray(validCtx)
	assert(emptyRPCArray.Size()).Equals(0)
	emptyRPCArray.release()
	assert(emptyRPCArray.ctx).IsNil()
	assert(emptyRPCArray.in).IsNil()

	bugRPCArray1 := rpcArray{
		ctx: nil,
		in:  rpcArrayInnerCache.Get().(*rpcArrayInner),
	}
	assert(bugRPCArray1.Size()).Equals(-1)
	bugRPCArray1.release()
	assert(bugRPCArray1.ctx).IsNil()
	assert(bugRPCArray1.in).IsNil()
}

func TestRpcArray_getIS(t *testing.T) {
	assert := newAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}

	nilRPCArray := rpcArray{}
	assert(nilRPCArray.getIS()).IsNil()

	emptyRPCArray := newRPCArray(validCtx)
	assert(emptyRPCArray.getIS()).IsNotNil()

	bugRPCArray1 := rpcArray{
		ctx: nil,
		in:  rpcArrayInnerCache.Get().(*rpcArrayInner),
	}
	assert(bugRPCArray1.getIS()).Equals(bugRPCArray1.in, nil)
	bugRPCArray2 := rpcArray{
		ctx: validCtx,
		in:  nil,
	}
	assert(bugRPCArray2.getIS()).Equals(nil, validCtx.inner.stream)
}

func TestRpcArray_equals(t *testing.T) {
	assert := newAssert(t)
	ctx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}
	nilArray := rpcArray{}
	emptyArray := newRPCArray(ctx)
	array1 := newRPCArrayByArray(ctx, Array{0})
	array2 := newRPCArrayByArray(ctx, Array{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	array3 := newRPCArrayByArray(ctx, Array{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	array4 := newRPCArrayByArray(ctx, Array{9, 8, 7, 6, 5, 4, 3, 2, 1, 0})

	bugArray := newRPCArrayByArray(ctx, Array{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	ctx.getCacheStream().setWritePosUnsafe(bugArray.in.items[0])
	ctx.getCacheStream().PutBytes([]byte{11})

	assert(rpcArray{}.equals(rpcArray{})).IsTrue()
	assert(emptyArray.equals(nilArray)).IsFalse()
	assert(nilArray.equals(emptyArray)).IsFalse()
	assert(array1.equals(array2)).IsFalse()
	assert(array2.equals(array3)).IsTrue()
	assert(array2.equals(array4)).IsFalse()
	assert(array2.equals(bugArray)).IsFalse()
	assert(bugArray.equals(array2)).IsFalse()
}

func TestRpcArray_contains(t *testing.T) {
	assert := newAssert(t)
	ctx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}

	nilArray := rpcArray{}
	emptyArray := newRPCArray(ctx)
	array1 := newRPCArrayByArray(ctx, Array{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	bugArray := newRPCArrayByArray(ctx, Array{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	ctx.getCacheStream().setWritePosUnsafe(bugArray.in.items[0])
	ctx.getCacheStream().PutBytes([]byte{11})

	assert(nilArray.contains(int64(0))).IsFalse()
	assert(emptyArray.contains(int64(0))).IsFalse()
	assert(array1.contains(int64(0))).IsTrue()
	assert(bugArray.contains(int64(0))).IsFalse()
}

func TestRpcArray_Size(t *testing.T) {
	assert := newAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}
	invalidCtx := &rpcContext{
		inner: nil,
	}
	validRPCArray := newRPCArray(validCtx)
	invalidRPCArray := newRPCArray(invalidCtx)

	for i := 1; i < 522; i++ {
		assert(validRPCArray.AppendBool(true)).IsTrue()
		assert(invalidRPCArray.AppendBool(true)).IsFalse()
		assert(validRPCArray.Size()).Equals(i)
		assert(invalidRPCArray.Size()).Equals(-1)
	}
}

func TestRpcArray_Get(t *testing.T) {
	assert := newAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}

	testArray := make([]interface{}, 14, 14)
	testArray[0] = nil
	testArray[1] = false
	testArray[2] = float64(3.14)
	testArray[3] = int64(30000)
	testArray[4] = uint64(30000)
	testArray[5] = ""
	testArray[6] = "hello"
	testArray[7] = []byte{}
	testArray[8] = []byte{0x53}
	testArray[9] = newRPCArray(validCtx)
	testArray[10] = newRPCArrayByArray(validCtx, Array{"world"})
	testArray[11] = newRPCMap(validCtx)
	testArray[12] = newRPCMapByMap(validCtx, Map{"hello": "world"})

	fnTestArray := func(array Array, index int, tp string) {
		// set the last one to the current index
		array[len(array)-1] = array[index]

		ctx := &rpcContext{
			inner: &rpcInnerContext{
				stream: newRPCStream(),
			},
		}

		arr := newRPCArrayByArray(ctx, array)
		bugArray0 := rpcArray{}
		bugArray1 := rpcArray{
			ctx: ctx,
			in:  nil,
		}
		bugArray2 := rpcArray{
			ctx: nil,
			in:  rpcArrayInnerCache.Get().(*rpcArrayInner),
		}

		// ctx is valid
		assert(arr.Get(arr.Size()-1)).Equals(array[index], true)
		assert(arr.ctx.getCacheStream().GetWritePos()).
			Equals(arr.ctx.getCacheStream().GetReadPos())
		assert(arr.Get(index)).Equals(array[index], true)
		assert(arr.Get(arr.Size())).Equals(nil, false)
		assert(bugArray0.Get(index)).Equals(nil, false)
		assert(bugArray1.Get(index)).Equals(nil, false)
		assert(bugArray2.Get(index)).Equals(nil, false)

		switch tp {
		case "nil":
			assert(arr.GetNil(arr.Size() - 1)).Equals(true)
			assert(arr.ctx.getCacheStream().GetWritePos()).
				Equals(arr.ctx.getCacheStream().GetReadPos())
			assert(arr.GetNil(index)).Equals(true)
			assert(arr.GetNil(arr.Size())).Equals(false)
			assert(bugArray0.GetNil(index)).Equals(false)
			assert(bugArray1.GetNil(index)).Equals(false)
			assert(bugArray2.GetNil(index)).Equals(false)
			ctx.close()
			assert(arr.GetNil(arr.Size() - 1)).Equals(false)
			assert(arr.GetNil(index)).Equals(false)
		case "bool":
			assert(arr.GetBool(arr.Size()-1)).Equals(array[index], true)
			assert(arr.ctx.getCacheStream().GetWritePos()).
				Equals(arr.ctx.getCacheStream().GetReadPos())
			assert(arr.GetBool(index)).Equals(array[index], true)
			assert(arr.GetBool(arr.Size())).Equals(false, false)
			assert(bugArray0.GetBool(index)).Equals(false, false)
			assert(bugArray1.GetBool(index)).Equals(false, false)
			assert(bugArray2.GetBool(index)).Equals(false, false)
			ctx.close()
			assert(arr.GetBool(arr.Size()-1)).Equals(false, false)
			assert(arr.GetBool(index)).Equals(false, false)
		case "float64":
			assert(arr.GetFloat64(arr.Size()-1)).Equals(array[index], true)
			assert(arr.ctx.getCacheStream().GetWritePos()).
				Equals(arr.ctx.getCacheStream().GetReadPos())
			assert(arr.GetFloat64(index)).Equals(array[index], true)
			assert(arr.GetFloat64(arr.Size())).Equals(float64(0), false)
			assert(bugArray0.GetFloat64(index)).Equals(float64(0), false)
			assert(bugArray1.GetFloat64(index)).Equals(float64(0), false)
			assert(bugArray2.GetFloat64(index)).Equals(float64(0), false)
			ctx.close()
			assert(arr.GetFloat64(arr.Size()-1)).Equals(float64(0), false)
			assert(arr.GetFloat64(index)).Equals(float64(0), false)
		case "int64":
			assert(arr.GetInt64(arr.Size()-1)).Equals(array[index], true)
			assert(arr.ctx.getCacheStream().GetWritePos()).
				Equals(arr.ctx.getCacheStream().GetReadPos())
			assert(arr.GetInt64(index)).Equals(array[index], true)
			assert(arr.GetInt64(arr.Size())).Equals(int64(0), false)
			assert(bugArray0.GetInt64(index)).Equals(int64(0), false)
			assert(bugArray1.GetInt64(index)).Equals(int64(0), false)
			assert(bugArray2.GetInt64(index)).Equals(int64(0), false)
			ctx.close()
			assert(arr.GetInt64(arr.Size()-1)).Equals(int64(0), false)
			assert(arr.GetInt64(index)).Equals(int64(0), false)
		case "uint64":
			assert(arr.GetUint64(arr.Size()-1)).Equals(array[index], true)
			assert(arr.ctx.getCacheStream().GetWritePos()).
				Equals(arr.ctx.getCacheStream().GetReadPos())
			assert(arr.GetUint64(index)).Equals(array[index], true)
			assert(arr.GetUint64(arr.Size())).Equals(uint64(0), false)
			assert(bugArray0.GetUint64(index)).Equals(uint64(0), false)
			assert(bugArray1.GetUint64(index)).Equals(uint64(0), false)
			assert(bugArray2.GetUint64(index)).Equals(uint64(0), false)
			ctx.close()
			assert(arr.GetUint64(arr.Size()-1)).Equals(uint64(0), false)
			assert(arr.GetUint64(index)).Equals(uint64(0), false)
		case "string":
			assert(arr.GetString(arr.Size()-1)).Equals(array[index], true)
			assert(arr.ctx.getCacheStream().GetWritePos()).
				Equals(arr.ctx.getCacheStream().GetReadPos())
			assert(arr.GetString(index)).Equals(array[index], true)
			assert(arr.GetString(arr.Size())).Equals("", false)
			assert(bugArray0.GetString(index)).Equals("", false)
			assert(bugArray1.GetString(index)).Equals("", false)
			assert(bugArray2.GetString(index)).Equals("", false)
			ctx.close()
			assert(arr.GetString(arr.Size()-1)).Equals("", false)
			assert(arr.GetString(index)).Equals("", false)
		case "bytes":
			assert(arr.GetBytes(arr.Size()-1)).Equals(array[index], true)
			assert(arr.ctx.getCacheStream().GetWritePos()).
				Equals(arr.ctx.getCacheStream().GetReadPos())
			assert(arr.GetBytes(index)).Equals(array[index], true)
			assert(arr.GetBytes(arr.Size())).Equals(emptyBytes, false)
			assert(bugArray0.GetBytes(index)).Equals(emptyBytes, false)
			assert(bugArray1.GetBytes(index)).Equals(emptyBytes, false)
			assert(bugArray2.GetBytes(index)).Equals(emptyBytes, false)
			ctx.close()
			assert(arr.GetBytes(arr.Size()-1)).Equals(emptyBytes, false)
			assert(arr.GetBytes(index)).Equals(emptyBytes, false)
		case "rpcArray":
			target1, ok := arr.GetRPCArray(arr.Size() - 1)
			assert(ok, arr.ctx.getCacheStream().GetWritePos()).
				Equals(true, arr.ctx.getCacheStream().GetReadPos())
			target2, ok := arr.GetRPCArray(index)
			assert(arr.GetRPCArray(arr.Size())).Equals(nilRPCArray, false)
			assert(bugArray0.GetRPCArray(index)).Equals(nilRPCArray, false)
			assert(bugArray1.GetRPCArray(index)).Equals(nilRPCArray, false)
			assert(bugArray2.GetRPCArray(index)).Equals(nilRPCArray, false)
			ctx.close()
			assert(arr.GetRPCArray(arr.Size()-1)).Equals(nilRPCArray, false)
			assert(arr.GetRPCArray(index)).Equals(nilRPCArray, false)
			assert(target1.ctx).Equals(ctx)
			assert(target2.ctx).Equals(ctx)
		case "rpcMap":
			target1, ok := arr.GetRPCMap(arr.Size() - 1)
			assert(ok, arr.ctx.getCacheStream().GetWritePos()).
				Equals(true, arr.ctx.getCacheStream().GetReadPos())
			target2, ok := arr.GetRPCMap(index)
			assert(arr.GetRPCMap(arr.Size())).Equals(nilRPCMap, false)
			assert(bugArray0.GetRPCMap(index)).Equals(nilRPCMap, false)
			assert(bugArray1.GetRPCMap(index)).Equals(nilRPCMap, false)
			assert(bugArray2.GetRPCMap(index)).Equals(nilRPCMap, false)
			ctx.close()
			assert(arr.GetRPCMap(arr.Size()-1)).Equals(nilRPCMap, false)
			assert(arr.GetRPCMap(index)).Equals(nilRPCMap, false)
			assert(target1.ctx).Equals(ctx)
			assert(target2.ctx).Equals(ctx)
		default:
			panic("unknown token")
		}

		// ctx is closed
		assert(arr.Get(arr.Size()-1)).Equals(nil, false)
		assert(arr.Get(index)).Equals(nil, false)
	}

	fnTestArray(testArray, 0, "nil")
	fnTestArray(testArray, 1, "bool")
	fnTestArray(testArray, 2, "float64")
	fnTestArray(testArray, 3, "int64")
	fnTestArray(testArray, 4, "uint64")
	fnTestArray(testArray, 5, "string")
	fnTestArray(testArray, 6, "string")
	fnTestArray(testArray, 7, "bytes")
	fnTestArray(testArray, 8, "bytes")
	fnTestArray(testArray, 9, "rpcArray")
	fnTestArray(testArray, 10, "rpcArray")
	fnTestArray(testArray, 11, "rpcMap")
	fnTestArray(testArray, 12, "rpcMap")
}

func TestRpcArray_Set_Append(t *testing.T) {
	assert := newAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: newRPCStream(),
		},
	}

	array1 := newRPCArray(validCtx)
	bugArray0 := rpcArray{}
	bugArray1 := rpcArray{
		ctx: validCtx,
		in:  nil,
	}
	bugArray2 := rpcArray{
		ctx: nil,
		in:  rpcArrayInnerCache.Get().(*rpcArrayInner),
	}

	fnTest := func(array rpcArray, canSet bool, canAppend bool) {
		assert(array.AppendNil()).Equals(canAppend)
		assert(array.SetNil(0)).Equals(canSet)
		assert(array.SetNil(-1)).IsFalse()
		assert(array.SetNil(1000)).IsFalse()

		assert(array.AppendBool(false)).Equals(canAppend)
		assert(array.SetBool(0, false)).Equals(canSet)
		assert(array.SetBool(-1, false)).IsFalse()
		assert(array.SetBool(1000, false)).IsFalse()

		assert(array.AppendFloat64(3.14)).Equals(canAppend)
		assert(array.SetFloat64(0, 3.14)).Equals(canSet)
		assert(array.SetFloat64(-1, 3.14)).IsFalse()
		assert(array.SetFloat64(1000, 3.14)).IsFalse()

		assert(array.AppendInt64(100)).Equals(canAppend)
		assert(array.SetInt64(0, 100)).Equals(canSet)
		assert(array.SetInt64(-1, 100)).IsFalse()
		assert(array.SetInt64(1000, 100)).IsFalse()

		assert(array.AppendUint64(100)).Equals(canAppend)
		assert(array.SetUint64(0, 100)).Equals(canSet)
		assert(array.SetUint64(-1, 100)).IsFalse()
		assert(array.SetUint64(1000, 100)).IsFalse()

		assert(array.AppendString("hello")).Equals(canAppend)
		assert(array.SetString(0, "hello")).Equals(canSet)
		assert(array.SetString(-1, "hello")).IsFalse()
		assert(array.SetString(1000, "hello")).IsFalse()

		assert(array.AppendBytes([]byte{1, 2, 3})).Equals(canAppend)
		assert(array.SetBytes(0, []byte{1, 2, 3})).Equals(canSet)
		assert(array.SetBytes(-1, []byte{1, 2, 3})).IsFalse()
		assert(array.SetBytes(1000, []byte{1, 2, 3})).IsFalse()

		rpcArray := newRPCArray(validCtx)
		assert(array.AppendRPCArray(rpcArray)).Equals(canAppend)
		assert(array.SetRPCArray(0, rpcArray)).Equals(canSet)
		assert(array.SetRPCArray(-1, rpcArray)).IsFalse()
		assert(array.SetRPCArray(1000, rpcArray)).IsFalse()

		rpcMap := newRPCMap(validCtx)
		assert(array.AppendRPCMap(rpcMap)).Equals(canAppend)
		assert(array.SetRPCMap(0, rpcMap)).Equals(canSet)
		assert(array.SetRPCMap(-1, rpcMap)).IsFalse()
		assert(array.SetRPCMap(1000, rpcMap)).IsFalse()

		assert(array.Append("hello")).Equals(canAppend)
		assert(array.Set(0, "hello")).Equals(canSet)
		assert(array.Set(-1, "hello")).IsFalse()
		assert(array.Set(1000, "hello")).IsFalse()
	}

	fnTest(array1, true, true)
	fnTest(bugArray0, false, false)
	fnTest(bugArray1, false, false)
	fnTest(bugArray2, false, false)
}
