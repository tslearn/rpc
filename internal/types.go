package internal

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"time"
	"unsafe"
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

type RPCContext struct {
	thread unsafe.Pointer
}

func (p *RPCContext) stop() {
	atomic.StorePointer(&p.thread, nil)
}

func (p *RPCContext) writeError(message string, debug string) *RPCReturn {
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

// OK get success RPCReturn  by value
func (p *RPCContext) OK(value interface{}) *RPCReturn {
	if thread := (*rpcThread)(p.thread); thread != nil {
		stream := thread.outStream
		stream.SetWritePos(StreamBodyPos)
		stream.WriteBool(true)

		if stream.Write(value) != RPCStreamWriteOK {
			return p.writeError(
				"return type is error",
				GetStackString(1),
			)
		}

		thread.execSuccessful = true
	}
	return nilReturn
}

func (p *RPCContext) Error(err RPCError) *RPCReturn {
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

func (p *RPCContext) Errorf(format string, a ...interface{}) *RPCReturn {
	return p.Error(NewRPCErrorByDebug(
		fmt.Sprintf(format, a...),
		GetStackString(1),
	))
}

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
