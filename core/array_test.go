package core

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestRpcArrayInner_free(t *testing.T) {
	assert := NewAssert(t)
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
	assert := NewAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
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
	assert := NewAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}

	assert(nilRPCArray.ok()).IsFalse()
	assert(newRPCArray(validCtx).ok()).IsTrue()
	assert(rpcArray{ctx: validCtx}.ok()).IsFalse()
}

func TestRpcArray_release(t *testing.T) {
	assert := NewAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}

	nilRPCArray := rpcArray{}
	nilRPCArray.release()

	emptyRPCArray := newRPCArray(validCtx)
	emptyRPCArray.release()

	assert(nilRPCArray.ctx).IsNil()
	assert(nilRPCArray.in).IsNil()

	assert(emptyRPCArray.ctx).IsNil()
	assert(emptyRPCArray.in).IsNil()

	bugRPCArray1 := rpcArray{
		ctx: nil,
		in:  rpcArrayInnerCache.Get().(*rpcArrayInner),
	}
	bugRPCArray1.release()
	assert(bugRPCArray1.ctx).IsNil()
	assert(bugRPCArray1.in).IsNil()

	bugRPCArray2 := rpcArray{
		ctx: nil,
		in:  (*rpcArrayInner)(unsafe.Pointer(uintptr(1323))),
	}
	defer func() {
		if err := recover(); err != nil {
			assert(fmt.Sprint(err)).Contains("runtime error")
			fmt.Println(err)
		}
	}()
	bugRPCArray2.release()
}

func TestRpcArray_getStream(t *testing.T) {
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

func TestRpcArray_Get(t *testing.T) {
	assert := NewAssert(t)
	testArray := make([]interface{}, 16, 16)

	testCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}

	testArray[0] = nil
	testArray[1] = false
	testArray[2] = float64(3.14)
	testArray[3] = int64(30000)
	testArray[4] = uint64(30000)
	testArray[5] = ""
	testArray[6] = "hello"
	testArray[7] = []byte{}
	testArray[8] = []byte{0x53}
	testArray[9] = newRPCArray(testCtx)
	testArray10 := newRPCArray(testCtx)
	testArray10.Append("world")
	testArray[10] = testArray10
	testArray[11] = newRPCMap(testCtx)
	testSmallArray12 := newRPCMap(testCtx)
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
		rpcArray := newRPCArray(ctx)
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
		case "string":
			assert(arr.GetString(arr.Size()-1)).Equals(array[index], true)
			assert(arr.ctx.getCacheStream().GetWritePos()).Equals(arr.ctx.getCacheStream().GetReadPos())
			assert(arr.GetString(index)).Equals(array[index], true)
			assert(arr.GetString(arr.Size())).Equals("", false)
			ctx.close()
			assert(arr.GetString(arr.Size()-1)).Equals("", false)
			assert(arr.GetString(index)).Equals("", false)
		case "bytes":
			assert(arr.GetBytes(arr.Size()-1)).Equals(array[index], true)
			assert(arr.ctx.getCacheStream().GetWritePos()).Equals(arr.ctx.getCacheStream().GetReadPos())
			assert(arr.GetBytes(index)).Equals(array[index], true)
			assert(arr.GetBytes(arr.Size())).Equals(emptyBytes, false)
			ctx.close()
			assert(arr.GetBytes(arr.Size()-1)).Equals(emptyBytes, false)
			assert(arr.GetBytes(index)).Equals(emptyBytes, false)
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
	fnTestArray(testArray, 5, "string")
	fnTestArray(testArray, 6, "string")
	fnTestArray(testArray, 7, "bytes")
	fnTestArray(testArray, 8, "bytes")
	fnTestArray(testArray, 9, "rpcArray")
	fnTestArray(testArray, 10, "rpcArray")
	fnTestArray(testArray, 11, "rpcMap")
	fnTestArray(testArray, 12, "rpcMap")
}

func TestRpcArray_Set(t *testing.T) {
	assert := NewAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}
	invalidCtx := &rpcContext{
		inner: nil,
	}

	array1 := newRPCArray(validCtx)
	array2 := newRPCArray(nil)
	array3 := newRPCArray(invalidCtx)

	assert(array1.AppendNil()).IsTrue()
	assert(array1.SetNil(0)).IsTrue()
	assert(array2.AppendNil()).IsFalse()
	assert(array2.SetNil(0)).IsFalse()
	assert(array3.AppendNil()).IsFalse()
	assert(array3.SetNil(0)).IsFalse()

	assert(array1.AppendBool(false)).IsTrue()
	assert(array1.SetBool(0, false)).IsTrue()
	assert(array2.AppendBool(false)).IsFalse()
	assert(array2.SetBool(0, false)).IsFalse()
	assert(array3.AppendBool(false)).IsFalse()
	assert(array3.SetBool(0, false)).IsFalse()

	assert(array1.AppendFloat64(3.14)).IsTrue()
	assert(array1.SetFloat64(0, 3.14)).IsTrue()
	assert(array2.AppendFloat64(3.14)).IsFalse()
	assert(array2.SetFloat64(0, 3.14)).IsFalse()
	assert(array3.AppendFloat64(3.14)).IsFalse()
	assert(array3.SetFloat64(0, 3.14)).IsFalse()

	assert(array1.AppendInt64(100)).IsTrue()
	assert(array1.SetInt64(0, 100)).IsTrue()
	assert(array2.AppendInt64(100)).IsFalse()
	assert(array2.SetInt64(0, 100)).IsFalse()
	assert(array3.AppendInt64(100)).IsFalse()
	assert(array3.SetInt64(0, 100)).IsFalse()

	assert(array1.AppendUint64(100)).IsTrue()
	assert(array1.SetUint64(0, 100)).IsTrue()
	assert(array2.AppendUint64(100)).IsFalse()
	assert(array2.SetUint64(0, 100)).IsFalse()
	assert(array3.AppendUint64(100)).IsFalse()
	assert(array3.SetUint64(0, 100)).IsFalse()

	assert(array1.AppendString("hello")).IsTrue()
	assert(array1.SetString(0, "hello")).IsTrue()
	assert(array2.AppendString("hello")).IsFalse()
	assert(array2.SetString(0, "hello")).IsFalse()
	assert(array3.AppendString("hello")).IsFalse()
	assert(array3.SetString(0, "hello")).IsFalse()

	assert(array1.AppendBytes([]byte{1, 2, 3})).IsTrue()
	assert(array1.SetBytes(0, []byte{1, 2, 3})).IsTrue()
	assert(array2.AppendBytes([]byte{1, 2, 3})).IsFalse()
	assert(array2.SetBytes(0, []byte{1, 2, 3})).IsFalse()
	assert(array3.AppendBytes([]byte{1, 2, 3})).IsFalse()
	assert(array3.SetBytes(0, []byte{1, 2, 3})).IsFalse()

	rpcArray := newRPCArray(validCtx)
	assert(array1.AppendRPCArray(rpcArray)).IsTrue()
	assert(array1.SetRPCArray(0, rpcArray)).IsTrue()
	assert(array2.AppendRPCArray(rpcArray)).IsFalse()
	assert(array2.SetRPCArray(0, rpcArray)).IsFalse()
	assert(array3.AppendRPCArray(rpcArray)).IsFalse()
	assert(array3.SetRPCArray(0, rpcArray)).IsFalse()

	rpcMap := newRPCMap(validCtx)
	assert(array1.AppendRPCMap(rpcMap)).IsTrue()
	assert(array1.SetRPCMap(0, rpcMap)).IsTrue()
	assert(array2.AppendRPCMap(rpcMap)).IsFalse()
	assert(array2.SetRPCMap(0, rpcMap)).IsFalse()
	assert(array3.AppendRPCMap(rpcMap)).IsFalse()
	assert(array3.SetRPCMap(0, rpcMap)).IsFalse()

	assert(array1.Append("hello")).IsTrue()
	assert(array1.Set(0, "hello")).IsTrue()
	assert(array2.Append("hello")).IsFalse()
	assert(array2.Set(0, "hello")).IsFalse()
	assert(array3.Append("hello")).IsFalse()
	assert(array3.Set(0, "hello")).IsFalse()
}
