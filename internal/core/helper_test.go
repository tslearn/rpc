package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"reflect"
	"testing"
)

func TestGetFuncKind(t *testing.T) {
	assert := base.NewAssert(t)

	t.Run("fn not func", func(t *testing.T) {
		v := 3
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrProcessIllegalHandler.AddDebug(
				"handler must be a function",
			))
	})

	t.Run("fn arguments length is zero", func(t *testing.T) {
		v := func() {}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrProcessIllegalHandler.AddDebug(
				"handler 1st argument type must be rpc.Runtime",
			))
	})

	t.Run("fn 1st argument is not rpc.Runtime", func(t *testing.T) {
		v := func(_ chan bool) {}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrProcessIllegalHandler.AddDebug(
				"handler 1st argument type must be rpc.Runtime",
			))
	})

	t.Run("fn without return", func(t *testing.T) {
		v := func(rt Runtime) {}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrProcessIllegalHandler.AddDebug(
				"handler return type must be rpc.Return",
			))
	})

	t.Run("fn return multiply value", func(t *testing.T) {
		v := func(rt Runtime) (Return, bool) { return emptyReturn, true }
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrProcessIllegalHandler.AddDebug(
				"handler return type must be rpc.Return",
			))
	})

	t.Run("fn return type not supported", func(t *testing.T) {
		v := func(rt Runtime, _ bool) bool { return true }
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrProcessIllegalHandler.AddDebug(
				"handler return type must be rpc.Return",
			))
	})

	t.Run("2nd argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ int32, _ int64, _ uint64, _ float64, _ string, _ Bytes,
			_ Array, _ Map, _ RTValue, _ RTArray, _ RTMap,
		) Return {
			return rt.OK(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrProcessIllegalHandler.AddDebug(
				"handler 2nd argument type int32 is not supported",
			))
	})

	t.Run("3rd argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int32, _ uint64, _ float64, _ string, _ Bytes,
			_ Array, _ Map, _ RTValue, _ RTArray, _ RTMap,
		) Return {
			return rt.OK(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrProcessIllegalHandler.AddDebug(
				"handler 3rd argument type int32 is not supported",
			))
	})

	t.Run("4th argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ int32, _ float64, _ string, _ Bytes,
			_ Array, _ Map, _ RTValue, _ RTArray, _ RTMap,
		) Return {
			return rt.OK(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrProcessIllegalHandler.AddDebug(
				"handler 4th argument type int32 is not supported",
			))
	})

	t.Run("5th argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ uint64, _ int32, _ string, _ Bytes,
			_ Array, _ Map, _ RTValue, _ RTArray, _ RTMap,
		) Return {
			return rt.OK(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrProcessIllegalHandler.AddDebug(
				"handler 5th argument type int32 is not supported",
			))
	})

	t.Run("6th argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ uint64, _ float64, _ int32, _ Bytes,
			_ Array, _ Map, _ RTValue, _ RTArray, _ RTMap,
		) Return {
			return rt.OK(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrProcessIllegalHandler.AddDebug(
				"handler 6th argument type int32 is not supported",
			))
	})

	t.Run("7th argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ uint64, _ float64, _ string, _ int32,
			_ Array, _ Map, _ RTValue, _ RTArray, _ RTMap,
		) Return {
			return rt.OK(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrProcessIllegalHandler.AddDebug(
				"handler 7th argument type int32 is not supported",
			))
	})

	t.Run("8th argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ uint64, _ float64, _ string, _ Bytes,
			_ int32, _ Map, _ RTValue, _ RTArray, _ RTMap,
		) Return {
			return rt.OK(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrProcessIllegalHandler.AddDebug(
				"handler 8th argument type int32 is not supported",
			))
	})

	t.Run("9th argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ uint64, _ float64, _ string, _ Bytes,
			_ Array, _ int32, _ RTValue, _ RTArray, _ RTMap,
		) Return {
			return rt.OK(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrProcessIllegalHandler.AddDebug(
				"handler 9th argument type int32 is not supported",
			))
	})

	t.Run("10th argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ uint64, _ float64, _ string, _ Bytes,
			_ Array, _ Map, _ int32, _ RTArray, _ RTMap,
		) Return {
			return rt.OK(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrProcessIllegalHandler.AddDebug(
				"handler 10th argument type int32 is not supported",
			))
	})

	t.Run("11th argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ uint64, _ float64, _ string, _ Bytes,
			_ Array, _ Map, _ RTValue, _ int32, _ RTMap,
		) Return {
			return rt.OK(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrProcessIllegalHandler.AddDebug(
				"handler 11th argument type int32 is not supported",
			))
	})

	t.Run("12th argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ uint64, _ float64, _ string, _ Bytes,
			_ Array, _ Map, _ RTValue, _ RTArray, _ int32,
		) Return {
			return rt.OK(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrProcessIllegalHandler.AddDebug(
				"handler 12th argument type int32 is not supported",
			))
	})

	t.Run("test ok", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ uint64, _ float64, _ string, _ Bytes,
			_ Array, _ Map, _ RTValue, _ RTArray, _ RTMap,
		) Return {
			return rt.OK(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal("BIUFSXAMVYZ", nil)
	})
}

func TestConvertTypeToString(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(convertTypeToString(nil)).Equal("<nil>")
		assert(convertTypeToString(runtimeType)).Equal("rpc.Runtime")
		assert(convertTypeToString(returnType)).Equal("rpc.Return")
		assert(convertTypeToString(boolType)).Equal("rpc.Bool")
		assert(convertTypeToString(int64Type)).Equal("rpc.Int64")
		assert(convertTypeToString(uint64Type)).Equal("rpc.Uint64")
		assert(convertTypeToString(float64Type)).Equal("rpc.Float64")
		assert(convertTypeToString(stringType)).Equal("rpc.String")
		assert(convertTypeToString(bytesType)).Equal("rpc.Bytes")
		assert(convertTypeToString(arrayType)).Equal("rpc.Array")
		assert(convertTypeToString(mapType)).Equal("rpc.Map")
		assert(convertTypeToString(rtValueType)).Equal("rpc.RTValue")
		assert(convertTypeToString(rtArrayType)).Equal("rpc.RTArray")
		assert(convertTypeToString(rtMapType)).Equal("rpc.RTMap")
		assert(convertTypeToString(reflect.ValueOf(make(chan bool)).Type())).
			Equal("chan bool")
	})
}

func TestMakeRequestStream(t *testing.T) {
	t.Run("write error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(MakeRequestStream("#", "", make(chan bool))).Equal(
			nil,
			errors.ErrRuntimeArgumentNotSupported.AddDebug(
				"2nd argument type(chan bool) is not supported",
			),
		)
	})

	t.Run("arguments is empty", func(t *testing.T) {
		assert := base.NewAssert(t)
		v, err := MakeRequestStream("#", "from")
		assert(v).IsNotNil()
		assert(err).IsNil()
		assert(v.ReadString()).Equal("#", true)
		assert(v.ReadString()).Equal("from", true)
		assert(v.IsReadFinish()).IsTrue()
		v.Release()
	})

	t.Run("arguments is not empty", func(t *testing.T) {
		assert := base.NewAssert(t)
		v, err := MakeRequestStream("#", "from", false)
		assert(v).IsNotNil()
		assert(err).IsNil()
		assert(v.ReadString()).Equal("#", true)
		assert(v.ReadString()).Equal("from", true)
		assert(v.ReadBool()).Equal(false, true)
		assert(v.IsReadFinish()).IsTrue()
		v.Release()
	})
}
