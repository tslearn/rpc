package core

import (
	"reflect"
	"unsafe"
)

const (
	sizeOfPosRecord = int(unsafe.Sizeof(posRecord(0)))

	vkBool    = 'B'
	vkInt64   = 'I'
	vkUint64  = 'U'
	vkFloat64 = 'F'
	vkString  = 'S'
	vkBytes   = 'X'
	vkArray   = 'A'
	vkMap     = 'M'
	vkRTValue = 'V'
	vkRTArray = 'Y'
	vkRTMap   = 'Z'
)

var (
	emptyReturn = &ReturnObject{}

	runtimeType = reflect.ValueOf(Runtime{}).Type()
	returnType  = reflect.ValueOf(emptyReturn).Type()
	boolType    = reflect.ValueOf(true).Type()
	int64Type   = reflect.ValueOf(int64(0)).Type()
	uint64Type  = reflect.ValueOf(uint64(0)).Type()
	float64Type = reflect.ValueOf(float64(0)).Type()
	stringType  = reflect.ValueOf("").Type()
	bytesType   = reflect.ValueOf(Bytes{}).Type()
	arrayType   = reflect.ValueOf(Array{}).Type()
	mapType     = reflect.ValueOf(Map{}).Type()
	rtValueType = reflect.ValueOf(RTValue{}).Type()
	rtArrayType = reflect.ValueOf(RTArray{}).Type()
	rtMapType   = reflect.ValueOf(RTMap{}).Type()
)

// ActionCache ...
type ActionCache interface {
	Get(fnString string) ActionCacheFunc
}

// ActionCacheFunc ...
type ActionCacheFunc = func(
	rt Runtime,
	stream *Stream,
	fn interface{},
) bool

// Bool ...
type Bool = bool

// Int64 ...
type Int64 = int64

// Uint64 ...
type Uint64 = uint64

// Float64 ...
type Float64 = float64

// String ...
type String = string

// Bytes ...
type Bytes = []byte

// Array ...
type Array = []interface{}

// Map common Map type
type Map = map[string]interface{}

// Any ...
type Any = interface{}

// ReturnObject ...
type ReturnObject struct{}

// Return ...
type Return = *ReturnObject

type posRecord uint64

func (p posRecord) getPos() int64 {
	return int64(p) & 0x7FFFFFFFFFFFFFFF
}

func (p posRecord) isString() bool {
	return (p & 0x8000000000000000) != 0
}

func makePosRecord(pos int64, isString bool) posRecord {
	if isString {
		return 0x8000000000000000 | posRecord(pos)
	}

	return posRecord(pos)
}
