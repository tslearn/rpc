package core

//func TestGetFuncKind(t *testing.T) {
//	assert := base.NewAssert(t)
//
//	fn1 := 3
//	assert(getFuncKind(reflect.ValueOf(fn1))).
//		Equal("", errors.New("handler must be a function"))
//
//	fn2 := func() {}
//	assert(getFuncKind(reflect.ValueOf(fn2))).
//		Equal("", errors.New("handler 1st argument type must be rpc.Runtime"))
//
//	fn3 := func(_ chan bool) {}
//	assert(getFuncKind(reflect.ValueOf(fn3))).
//		Equal("", errors.New("handler 1st argument type must be rpc.Runtime"))
//
//	fn4 := func(rt Runtime, _ bool) {}
//	assert(getFuncKind(reflect.ValueOf(fn4))).
//		Equal("", errors.New("handler return type must be rpc.Return"))
//
//	fn5 := func(rt Runtime, _ bool) (Return, bool) { return emptyReturn, true }
//	assert(getFuncKind(reflect.ValueOf(fn5))).
//		Equal("", errors.New("handler return type must be rpc.Return"))
//
//	fn6 := func(rt Runtime, _ bool) bool { return true }
//	assert(getFuncKind(reflect.ValueOf(fn6))).
//		Equal("", errors.New("handler return type must be rpc.Return"))
//
//	fn7 := func(rt Runtime,
//		_ bool, _ int64, _ uint64, _ float64,
//		_ string, _ Bytes, _ Array, _ Map,
//	) Return {
//		return rt.OK(true)
//	}
//	assert(getFuncKind(reflect.ValueOf(fn7))).Equal("BIUFSXAM", nil)
//
//	fn8 := func(rt Runtime,
//		_ int32, _ int64, _ uint64, _ float64,
//		_ string, _ Bytes, _ Array, _ Map,
//	) Return {
//		return rt.OK(true)
//	}
//	assert(getFuncKind(reflect.ValueOf(fn8))).
//		Equal("", errors.New("handler 2nd argument type int32 is not supported"))
//
//	fn9 := func(rt Runtime,
//		_ bool, _ int32, _ uint64, _ float64,
//		_ string, _ Bytes, _ Array, _ Map,
//	) Return {
//		return rt.OK(true)
//	}
//	assert(getFuncKind(reflect.ValueOf(fn9))).
//		Equal("", errors.New("handler 3rd argument type int32 is not supported"))
//
//	fn10 := func(rt Runtime,
//		_ bool, _ int64, _ int32, _ float64,
//		_ string, _ Bytes, _ Array, _ Map,
//	) Return {
//		return rt.OK(true)
//	}
//	assert(getFuncKind(reflect.ValueOf(fn10))).
//		Equal("", errors.New("handler 4th argument type int32 is not supported"))
//
//	fn11 := func(rt Runtime,
//		_ bool, _ int64, _ uint64, _ int32,
//		_ string, _ Bytes, _ Array, _ Map,
//	) Return {
//		return rt.OK(true)
//	}
//	assert(getFuncKind(reflect.ValueOf(fn11))).
//		Equal("", errors.New("handler 5th argument type int32 is not supported"))
//
//	fn12 := func(rt Runtime,
//		_ bool, _ int64, _ uint64, _ float64,
//		_ int32, _ Bytes, _ Array, _ Map,
//	) Return {
//		return rt.OK(true)
//	}
//	assert(getFuncKind(reflect.ValueOf(fn12))).
//		Equal("", errors.New("handler 6th argument type int32 is not supported"))
//
//	fn13 := func(rt Runtime,
//		_ bool, _ int64, _ uint64, _ float64,
//		_ string, _ int32, _ Array, _ Map,
//	) Return {
//		return rt.OK(true)
//	}
//	assert(getFuncKind(reflect.ValueOf(fn13))).
//		Equal("", errors.New("handler 7th argument type int32 is not supported"))
//
//	fn14 := func(rt Runtime,
//		_ bool, _ int64, _ uint64, _ float64,
//		_ string, _ Bytes, _ int32, _ Map,
//	) Return {
//		return rt.OK(true)
//	}
//	assert(getFuncKind(reflect.ValueOf(fn14))).
//		Equal("", errors.New("handler 8th argument type int32 is not supported"))
//
//	fn15 := func(rt Runtime,
//		_ bool, _ int64, _ uint64, _ float64,
//		_ string, _ Bytes, _ Array, _ int32,
//	) Return {
//		return rt.OK(true)
//	}
//	assert(getFuncKind(reflect.ValueOf(fn15))).
//		Equal("", errors.New("handler 9th argument type int32 is not supported"))
//}
//
//func TestConvertTypeToString(t *testing.T) {
//	assert := base.NewAssert(t)
//	assert(convertTypeToString(nil)).Equal("<nil>")
//	assert(convertTypeToString(bytesType)).Equal("rpc.Bytes")
//	assert(convertTypeToString(arrayType)).Equal("rpc.Array")
//	assert(convertTypeToString(rtArrayType)).Equal("rpc.RTArray")
//	assert(convertTypeToString(mapType)).Equal("rpc.Map")
//	assert(convertTypeToString(rtMapType)).Equal("rpc.RTMap")
//	assert(convertTypeToString(boolType)).Equal("rpc.Bool")
//	assert(convertTypeToString(int64Type)).Equal("rpc.Int64")
//	assert(convertTypeToString(uint64Type)).Equal("rpc.Uint64")
//	assert(convertTypeToString(float64Type)).Equal("rpc.Float64")
//	assert(convertTypeToString(stringType)).Equal("rpc.String")
//	assert(convertTypeToString(contextType)).Equal("rpc.Runtime")
//	assert(convertTypeToString(returnType)).Equal("rpc.Return")
//	assert(convertTypeToString(reflect.ValueOf(make(chan bool)).Type())).
//		Equal("chan bool")
//}
//
