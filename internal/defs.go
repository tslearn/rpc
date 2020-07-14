package internal

import (
	"reflect"
	"time"
)

var (
	nilContext  = (*RPCContext)(nil)
	nilReturn   = (*RPCReturn)(nil)
	contextType = reflect.ValueOf(nilContext).Type()
	returnType  = reflect.ValueOf(nilReturn).Type()
	boolType    = reflect.ValueOf(true).Type()
	int64Type   = reflect.ValueOf(int64(0)).Type()
	uint64Type  = reflect.ValueOf(uint64(0)).Type()
	float64Type = reflect.ValueOf(float64(0)).Type()
	stringType  = reflect.ValueOf("").Type()
	bytesType   = reflect.ValueOf(RPCBytes{}).Type()
	arrayType   = reflect.ValueOf(RPCArray{}).Type()
	mapType     = reflect.ValueOf(RPCMap{}).Type()
)

const StreamBodyPos = 33
const closeTimeOut = 15 * time.Second

// RPCReplyCache ...
type RPCReplyCache interface {
	Get(fnString string) RPCReplyCacheFunc
}

// RPCReplyCacheFunc ...
type RPCReplyCacheFunc = func(
	ctx *RPCContext,
	stream *RPCStream,
	fn interface{},
) bool

// RPCBool ...
type RPCBool = bool

// RPCInt ...
type RPCInt = int64

// RPCUint ...
type RPCUint = uint64

// RPCFloat ...
type RPCFloat = float64

// RPCString ...
type RPCString = string

// RPCBytes ...
type RPCBytes = []byte

// RPCAny ...
type RPCAny = interface{}

// RPCArray ...
type RPCArray = []RPCAny

// RPCMap common Map type
type RPCMap = map[string]RPCAny

// RPCReturn ...
type RPCReturn struct{}
