package core

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"reflect"
	"sync/atomic"
	"unsafe"
)

const (
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

var emptyReturn = &ReturnObject{}

// RTArray ...
type RTArray struct {
	rt    Runtime
	items []posRecord
}

func newRTArray(rt Runtime, size int) (ret RTArray) {
	ret.rt = rt

	if thread := rt.thread; thread != nil && size <= 32 {
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

// Get ...
func (p *RTArray) Get(index int) RTValue {
	if index >= 0 && index < len(p.items) {
		return makeRTValue(p.rt, p.items[index])
	}

	return RTValue{
		err: errors.ErrRTValueNotAvailable.
			AddDebug(fmt.Sprintf("RTArray index %d is overflow", index)),
	}
}

// Size ...
func (p *RTArray) Size() int {
	if p.items != nil {
		return len(p.items)
	}

	return -1
}

type mapItem struct {
	key string
	pos posRecord
}

const sizeOfMapItem = int(unsafe.Sizeof(mapItem{}))

// RTMap ...
type RTMap struct {
	rt       Runtime
	items    []mapItem
	largeMap map[string]posRecord
}

func newRTMap(rt Runtime, size int) (ret RTMap) {
	ret.rt = rt

	if thread := rt.thread; thread != nil && size <= 10 {
		if data := thread.malloc(sizeOfMapItem * size); data != nil {
			itemsHeader := (*reflect.SliceHeader)(unsafe.Pointer(&ret.items))
			itemsHeader.Len = 0
			itemsHeader.Cap = size
			itemsHeader.Data = uintptr(data)
			ret.largeMap = nil
			return
		}
	}

	if size <= 64 {
		ret.items = make([]mapItem, 0, size)
		ret.largeMap = nil
		return
	}

	ret.items = nil
	ret.largeMap = make(map[string]posRecord, size)
	return
}

// Get ...
func (p *RTMap) Get(key string) RTValue {
	if p.items != nil {
		for i := len(p.items) - 1; i >= 0; i-- {
			if key == p.items[i].key {
				return makeRTValue(p.rt, p.items[i].pos)
			}
		}
		return RTValue{
			err: errors.ErrRTValueNotAvailable.
				AddDebug(fmt.Sprintf("RTMap key %s is not exist", key)),
		}
	} else if p.largeMap != nil {
		if pos, ok := p.largeMap[key]; ok {
			return makeRTValue(p.rt, pos)
		}
		return RTValue{
			err: errors.ErrRTValueNotAvailable.
				AddDebug(fmt.Sprintf("RTMap key %s is not exist", key)),
		}
	} else {
		return RTValue{
			err: errors.ErrRTValueNotAvailable.
				AddDebug("RTMap is not available").
				AddDebug(base.GetFileLine(0)),
		}
	}
}

// Size ...
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
	if isString {
		return 0x8000000000000000 | posRecord(pos)
	}

	return posRecord(pos)
}

// RTValue ...
type RTValue struct {
	err         *base.Error
	rt          Runtime
	pos         int64
	cacheString string
	cacheSafe   bool
	cacheError  *base.Error
}

func makeRTValue(rt Runtime, record posRecord) RTValue {
	if !record.isString() {
		return RTValue{
			err:         nil,
			rt:          rt,
			pos:         record.getPos(),
			cacheString: "",
			cacheSafe:   true,
			cacheError:  errors.ErrStream,
		}
	} else if thread := rt.lock(); thread == nil {
		return RTValue{
			err:         errors.ErrRuntimeIllegalInCurrentGoroutine,
			rt:          rt,
			pos:         record.getPos(),
			cacheString: "",
			cacheSafe:   true,
			cacheError:  errors.ErrStream,
		}
	} else {
		defer rt.unlock()
		pos := record.getPos()
		thread.rtStream.SetReadPos(int(pos))
		ret := RTValue{
			err: nil,
			rt:  rt,
			pos: record.getPos(),
		}
		ret.cacheString, ret.cacheSafe, ret.err = thread.rtStream.readUnsafeString()
		return ret
	}
}

// ToBool ...
func (p RTValue) ToBool() (Bool, *base.Error) {
	if p.err != nil {
		return false, p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadBool()
	} else {
		return false, errors.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// ToInt64 ...
func (p RTValue) ToInt64() (Int64, *base.Error) {
	if p.err != nil {
		return 0, p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadInt64()
	} else {
		return 0, errors.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// ToUint64 ...
func (p RTValue) ToUint64() (Uint64, *base.Error) {
	if p.err != nil {
		return 0, p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadUint64()
	} else {
		return 0, errors.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// ToFloat64 ...
func (p RTValue) ToFloat64() (Float64, *base.Error) {
	if p.err != nil {
		return 0, p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadFloat64()
	} else {
		return 0, errors.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// ToString ...
func (p RTValue) ToString() (ret String, err *base.Error) {
	if !p.cacheSafe {
		ret = string(base.StringToBytesUnsafe(p.cacheString))
	} else {
		ret = p.cacheString
	}

	if p.err != nil {
		return "", p.err
	} else if p.rt.thread == nil ||
		atomic.LoadUint64(&p.rt.thread.top.lockStatus) != p.rt.id {
		return "", errors.ErrRuntimeIllegalInCurrentGoroutine
	} else {
		return
	}
}

// ToBytes ...
func (p RTValue) ToBytes() (Bytes, *base.Error) {
	if p.err != nil {
		return Bytes(nil), p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadBytes()
	} else {
		return Bytes(nil), errors.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// ToArray ...
func (p RTValue) ToArray() (Array, *base.Error) {
	if p.err != nil {
		return Array(nil), p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadArray()
	} else {
		return Array(nil), errors.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// ToRTArray ...
func (p RTValue) ToRTArray() (RTArray, *base.Error) {
	if p.err != nil {
		return RTArray{}, p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadRTArray(p.rt)
	} else {
		return RTArray{}, errors.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// ToMap ...
func (p RTValue) ToMap() (Map, *base.Error) {
	if p.err != nil {
		return Map{}, p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadMap()
	} else {
		return Map{}, errors.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// ToRTMap ...
func (p RTValue) ToRTMap() (RTMap, *base.Error) {
	if p.err != nil {
		return RTMap{}, p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadRTMap(p.rt)
	} else {
		return RTMap{}, errors.ErrRuntimeIllegalInCurrentGoroutine
	}
}
