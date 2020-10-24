package core

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/errors"
	"reflect"
	"unsafe"
)

type mapItem struct {
	key string
	pos posRecord
}

const sizeOfMapItem = int(unsafe.Sizeof(mapItem{}))

// RTMap ...
type RTMap struct {
	rt    Runtime
	items []mapItem
}

func newRTMap(rt Runtime, size int) (ret RTMap) {
	ret.rt = rt
	size += 4

	if thread := rt.thread; thread != nil && size <= 20 {
		if data := thread.malloc(sizeOfMapItem * size); data != nil {
			itemsHeader := (*reflect.SliceHeader)(unsafe.Pointer(&ret.items))
			itemsHeader.Len = 0
			itemsHeader.Cap = size
			itemsHeader.Data = uintptr(data)
			return
		}
	}

	ret.items = make([]mapItem, 0, size)
	return
}

func (p *RTMap) getPosRecord(key string) posRecord {
	if p.items != nil {
		for i := len(p.items) - 1; i >= 0; i-- {
			if key == p.items[i].key {
				return p.items[i].pos
			}
		}
	}

	return 0
}

// Get ...
func (p *RTMap) Get(key string) RTValue {
	if pos := p.getPosRecord(key); pos > 0 {
		return makeRTValue(p.rt, pos)
	}

	return RTValue{
		err: errors.ErrRTMapNameNotFound.
			AddDebug(fmt.Sprintf("RTMap key %s is not exist", key)),
	}
}

// Size ...
func (p *RTMap) Size() int {
	if p.items != nil {
		return len(p.items)
	}

	return -1
}

func (p *RTMap) appendValue(key string, pos posRecord) {
	if p.items != nil {
		p.items = append(p.items, mapItem{key: key, pos: pos})
	}
}
