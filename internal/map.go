package internal

import (
	"reflect"
	"unsafe"
)

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

func (p RTMap) Get(key string) RTValue {
	if p.items != nil {
		for i := 0; i < len(p.items); i++ {
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

func (p RTMap) Size() int {
	if p.items != nil {
		return len(p.items)
	} else if p.largeMap != nil {
		return len(p.largeMap)
	} else {
		return -1
	}
}

func (p RTMap) appendValue(key string, pos posRecord) {
	if p.items != nil {
		p.items = append(p.items, mapItem{key: key, pos: pos})
	} else {
		p.largeMap[key] = pos
	}
}
