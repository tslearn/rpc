package internal

import (
	"reflect"
	"unsafe"
)

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
