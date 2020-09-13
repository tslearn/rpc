package internal

import (
	"errors"
	"fmt"
	"github.com/rpccloud/rpc/internal/util"
	"net"
	"reflect"
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	contextType = reflect.ValueOf(Runtime{}).Type()
	returnType  = reflect.ValueOf(emptyReturn).Type()
	boolType    = reflect.ValueOf(true).Type()
	int64Type   = reflect.ValueOf(int64(0)).Type()
	uint64Type  = reflect.ValueOf(uint64(0)).Type()
	float64Type = reflect.ValueOf(float64(0)).Type()
	stringType  = reflect.ValueOf("").Type()
	bytesType   = reflect.ValueOf(Bytes{}).Type()
	arrayType   = reflect.ValueOf(Array{}).Type()
	mapType     = reflect.ValueOf(Map{}).Type()
	rtArrayType = reflect.ValueOf(RTArray{}).Type()
	rtMapType   = reflect.ValueOf(RTMap{}).Type()
)

func getFuncKind(fn reflect.Value) (string, error) {
	if fn.Kind() != reflect.Func {
		return "", errors.New("handler must be a function")
	} else if fn.Type().NumIn() < 1 ||
		fn.Type().In(0) != reflect.ValueOf(Runtime{}).Type() {
		return "", fmt.Errorf(
			"handler 1st argument type must be %s",
			convertTypeToString(contextType),
		)
	} else if fn.Type().NumOut() != 1 ||
		fn.Type().Out(0) != reflect.ValueOf(emptyReturn).Type() {
		return "", fmt.Errorf(
			"handler return type must be %s",
			convertTypeToString(returnType),
		)
	} else {
		sb := util.NewStringBuilder()
		defer sb.Release()

		for i := 1; i < fn.Type().NumIn(); i++ {
			switch fn.Type().In(i) {
			case bytesType:
				sb.AppendByte('X')
			case arrayType:
				sb.AppendByte('A')
			case rtArrayType:
				sb.AppendByte('Y')
			case mapType:
				sb.AppendByte('M')
			case rtMapType:
				sb.AppendByte('Z')
			case int64Type:
				sb.AppendByte('I')
			case uint64Type:
				sb.AppendByte('U')
			case boolType:
				sb.AppendByte('B')
			case float64Type:
				sb.AppendByte('F')
			case stringType:
				sb.AppendByte('S')
			default:
				return "", fmt.Errorf(
					"handler %s argument type %s is not supported",
					util.ConvertOrdinalToString(1+uint(i)),
					fn.Type().In(i),
				)
			}
		}

		return sb.String(), nil
	}
}

func convertTypeToString(reflectType reflect.Type) string {
	switch reflectType {
	case nil:
		return "<nil>"
	case contextType:
		return "rpc.Runtime"
	case returnType:
		return "rpc.Return"
	case bytesType:
		return "rpc.Bytes"
	case arrayType:
		return "rpc.Array"
	case rtArrayType:
		return "rpc.RTArray"
	case mapType:
		return "rpc.Map"
	case rtMapType:
		return "rpc.RTMap"
	case boolType:
		return "rpc.Bool"
	case int64Type:
		return "rpc.Int64"
	case uint64Type:
		return "rpc.Uint64"
	case float64Type:
		return "rpc.Float64"
	case stringType:
		return "rpc.String"
	default:
		return reflectType.String()
	}
}

// IStreamConn ...
type IStreamConn interface {
	ReadStream(timeout time.Duration, readLimit int64) (*Stream, Error)
	WriteStream(stream *Stream, timeout time.Duration) Error
	Close() Error
}

// IServerAdapter ...
type IServerAdapter interface {
	Open(onConnRun func(IStreamConn, net.Addr), onError func(uint64, Error))
	Close(onError func(uint64, Error))
}

// IClientAdapter ...
type IClientAdapter interface {
	Open(onConnRun func(IStreamConn), onError func(Error))
	Close(onError func(Error))
}

// ReplyCache ...
type ReplyCache interface {
	Get(fnString string) ReplyCacheFunc
}

// ReplyCacheFunc ...
type ReplyCacheFunc = func(
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

// Return ...
type returnObject struct{}
type Return = *returnObject

var emptyReturn = &returnObject{}

type RTArray struct {
	rt    Runtime
	items []posRecord
}

func newRTArray(rt Runtime, size int) (ret RTArray) {
	ret.rt = rt

	if thread := rt.thread; thread != nil {
		if data := thread.malloc(sizeOfPosRecord * size); data != nil {
			itemsHeader := (*reflect.SliceHeader)(unsafe.Pointer(&ret.items))
			itemsHeader.Len = 0
			itemsHeader.Cap = size
			itemsHeader.Data = uintptr(data)
			return
		}
	}

	ret.items = make([]posRecord, 0, size)
	return
}

func (p *RTArray) Get(index int) RTValue {
	if index >= 0 && index < len(p.items) {
		return makeRTValue(p.rt, p.items[index])
	} else {
		return RTValue{}
	}
}

func (p *RTArray) Size() int {
	if p.items != nil {
		return len(p.items)
	} else {
		return -1
	}
}

type mapItem struct {
	key string
	pos posRecord
}

const sizeOfMapItem = int(unsafe.Sizeof(mapItem{}))

type RTMap struct {
	rt       Runtime
	items    []mapItem
	largeMap map[string]posRecord
}

func newRTMap(rt Runtime, size int) (ret RTMap) {
	ret.rt = rt

	if thread := rt.thread; thread != nil && size <= 8 {
		if data := thread.malloc(sizeOfMapItem * size); data != nil {
			itemsHeader := (*reflect.SliceHeader)(unsafe.Pointer(&ret.items))
			itemsHeader.Len = 0
			itemsHeader.Cap = size
			itemsHeader.Data = uintptr(data)
			ret.largeMap = nil
			return
		}
	}

	if size <= 16 {
		ret.items = make([]mapItem, 0, size)
		ret.largeMap = nil
	} else {
		ret.items = nil
		ret.largeMap = make(map[string]posRecord)
	}

	return
}

func (p *RTMap) Get(key string) RTValue {
	if p.items != nil {
		for i := len(p.items) - 1; i >= 0; i-- {
			if key == p.items[i].key {
				return makeRTValue(p.rt, p.items[i].pos)
			}
		}
		return RTValue{}
	} else if p.largeMap != nil {
		if pos, ok := p.largeMap[key]; ok {
			return makeRTValue(p.rt, pos)
		}
		return RTValue{}
	} else {
		return RTValue{}
	}
}

func (p *RTMap) Size() int {
	if p.items != nil {
		return len(p.items)
	} else if p.largeMap != nil {
		return len(p.largeMap)
	} else {
		return -1
	}
}

func (p *RTMap) appendValue(key string, pos posRecord) {
	if p.items != nil {
		p.items = append(p.items, mapItem{key: key, pos: pos})
	} else {
		p.largeMap[key] = pos
	}
}

type posRecord uint64

const sizeOfPosRecord = int(unsafe.Sizeof(posRecord(0)))

func (p posRecord) getPos() int64 {
	return int64(p) & 0x7FFFFFFFFFFFFFFF
}

func (p posRecord) isString() bool {
	return (p & 0x8000000000000000) != 0
}

func makePosRecord(pos int64, isString bool) posRecord {
	if !isString {
		return posRecord(pos)
	} else {
		return 0x8000000000000000 | posRecord(pos)
	}
}

type RTValue struct {
	rt          Runtime
	pos         int64
	cacheString string
	cacheOK     bool
	cacheSafe   bool
}

func makeRTValue(rt Runtime, record posRecord) RTValue {
	if !record.isString() {
		return RTValue{
			rt:          rt,
			pos:         record.getPos(),
			cacheString: "",
			cacheOK:     false,
		}
	} else if thread := rt.lock(); thread == nil {
		return RTValue{
			rt:          rt,
			pos:         record.getPos(),
			cacheString: "",
			cacheOK:     false,
		}
	} else {
		defer rt.unlock()
		pos := record.getPos()
		thread.rtStream.SetReadPos(int(pos))
		ret := RTValue{
			rt:  rt,
			pos: record.getPos(),
		}
		ret.cacheString, ret.cacheSafe, ret.cacheOK = thread.rtStream.readUnsafeString()
		return ret
	}
}

func (p RTValue) ToBool() (Bool, bool) {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadBool()
	}

	return false, false
}

func (p RTValue) ToInt64() (Int64, bool) {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadInt64()
	}

	return 0, false
}

func (p RTValue) ToUint64() (Uint64, bool) {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadUint64()
	}

	return 0, false
}

func (p RTValue) ToFloat64() (Float64, bool) {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadFloat64()
	}

	return 0, false
}

func (p RTValue) ToString() (ret String, ok bool) {
	if !p.cacheSafe {
		ret = string(util.StringToBytesUnsafe(p.cacheString))
	} else {
		ret = p.cacheString
	}

	if p.cacheOK &&
		atomic.LoadUint64(&p.rt.thread.top.lockStatus) == p.rt.id {
		return ret, true
	} else {
		return "", false
	}
}

func (p RTValue) ToBytes() (Bytes, bool) {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadBytes()
	}

	return Bytes(nil), false
}

func (p RTValue) ToArray() (Array, bool) {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadArray()
	}

	return Array(nil), false
}

func (p RTValue) ToRTArray() (RTArray, bool) {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadRTArray(p.rt)
	}

	return RTArray{}, false
}

func (p RTValue) ToMap() (Map, bool) {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadMap()
	}

	return Map(nil), false
}

func (p RTValue) ToRTMap() (RTMap, bool) {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadRTMap(p.rt)
	}

	return RTMap{}, false
}
