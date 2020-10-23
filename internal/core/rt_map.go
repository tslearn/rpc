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
	rt     Runtime
	sItems []mapItem
	lItems map[string]posRecord
}

func newRTMap(rt Runtime, size int) (ret RTMap) {
	ret.rt = rt

	if thread := rt.thread; thread != nil && size <= 16 {
		if data := thread.malloc(sizeOfMapItem * size); data != nil {
			itemsHeader := (*reflect.SliceHeader)(unsafe.Pointer(&ret.sItems))
			itemsHeader.Len = 0
			itemsHeader.Cap = size
			itemsHeader.Data = uintptr(data)
			ret.lItems = nil
			return
		}
	}

	if size <= 32 {
		ret.sItems = make([]mapItem, 0, size)
		ret.lItems = nil
		return
	}

	ret.sItems = nil
	ret.lItems = make(map[string]posRecord, size+6)
	return
}

func (p *RTMap) getPosRecord(key string) posRecord {
	if p.sItems != nil {
		for i := len(p.sItems) - 1; i >= 0; i-- {
			if key == p.sItems[i].key {
				return p.sItems[i].pos
			}
		}
		return 0
	} else if p.lItems != nil {
		if pos, ok := p.lItems[key]; ok {
			return pos
		}
		return 0
	} else {
		return 0
	}
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
	if p.sItems != nil {
		return len(p.sItems)
	} else if p.lItems != nil {
		return len(p.lItems)
	} else {
		return -1
	}
}

func (p *RTMap) appendValue(key string, pos posRecord) {
	if p.sItems != nil {
		p.sItems = append(p.sItems, mapItem{key: key, pos: pos})
	} else if p.lItems != nil {
		p.lItems[key] = pos
	} else {
		// do nothing
	}
}

func (p *RTMap) toLarge() {
	if p.sItems != nil {
		p.lItems = make(map[string]posRecord, len(p.sItems)+6)
		for _, v := range p.sItems {
			p.lItems[v.key] = v.pos
		}
		p.sItems = nil
	}
}
