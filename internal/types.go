package internal

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"unsafe"
)

var (
	nilContext  = (*RPCContext)(nil)
	nilReturn   = (*Return)(nil)
	contextType = reflect.ValueOf(nilContext).Type()
	returnType  = reflect.ValueOf(nilReturn).Type()
	boolType    = reflect.ValueOf(true).Type()
	int64Type   = reflect.ValueOf(int64(0)).Type()
	uint64Type  = reflect.ValueOf(uint64(0)).Type()
	float64Type = reflect.ValueOf(float64(0)).Type()
	stringType  = reflect.ValueOf("").Type()
	bytesType   = reflect.ValueOf(Bytes{}).Type()
	arrayType   = reflect.ValueOf(Array{}).Type()
	mapType     = reflect.ValueOf(Map{}).Type()
)

const StreamBodyPos = 33

// RPCReplyCache ...
type RPCReplyCache interface {
	Get(fnString string) RPCReplyCacheFunc
}

// RPCReplyCacheFunc ...
type RPCReplyCacheFunc = func(
	ctx *RPCContext,
	stream *Stream,
	fn interface{},
) bool

type RPCContext struct {
	thread unsafe.Pointer
}

func (p *RPCContext) stop() {
	atomic.StorePointer(&p.thread, nil)
}

func (p *RPCContext) writeError(message string, debug string) *Return {
	if thread := (*rpcThread)(p.thread); thread != nil {
		execStream := thread.outStream
		execStream.SetWritePos(StreamBodyPos)
		execStream.WriteBool(false)
		execStream.WriteString(message)
		execStream.WriteString(debug)
		thread.execSuccessful = false
	}
	return nilReturn
}

// OK get success Return  by value
func (p *RPCContext) OK(value interface{}) *Return {
	if thread := (*rpcThread)(p.thread); thread != nil {
		stream := thread.outStream
		stream.SetWritePos(StreamBodyPos)
		stream.WriteBool(true)

		if stream.Write(value) != StreamWriteOK {
			return p.writeError(
				"return type is error",
				GetStackString(1),
			)
		}

		thread.execSuccessful = true
	}
	return nilReturn
}

func (p *RPCContext) Error(err Error) *Return {
	if err == nil {
		return nilReturn
	}

	if thread := (*rpcThread)(p.thread); thread != nil &&
		thread.execReplyNode != nil &&
		thread.execReplyNode.debugString != "" {
		err.AddDebug(thread.execReplyNode.debugString)
	}

	return p.writeError(err.GetMessage(), err.GetDebug())
}

func (p *RPCContext) Errorf(format string, a ...interface{}) *Return {
	return p.Error(NewErrorByDebug(
		fmt.Sprintf(format, a...),
		GetStackString(1),
	))
}

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

// Any ...
type Any = interface{}

// Array ...
type Array = []interface{}

// Map common Map type
type Map = map[string]interface{}

// Return ...
type Return struct{}
