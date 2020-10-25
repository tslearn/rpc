package core

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"reflect"
	"unsafe"
)

type mapItem struct {
	key     string
	fastKey uint32
	pos     posRecord
}

const sizeOfMapItem = int(unsafe.Sizeof(mapItem{}))

func getFastKey(s string) uint32 {
	header := (*reflect.StringHeader)(unsafe.Pointer(&s))
	if header.Len >= 4 {
		return *(*uint32)(unsafe.Pointer(header.Data))
	}
	ret := uint32(0)
	for i := uintptr(0); i < uintptr(header.Len); i++ {
		ret |= uint32(*(*uint8)(unsafe.Pointer(header.Data + i))) << (24 - 8*i)
	}
	return ret
}

func compareMapItem(m1 mapItem, m2 mapItem) int {
	if m1.fastKey > m2.fastKey {
		return 1
	} else if m1.fastKey < m2.fastKey {
		return -1
	} else {
		header1 := (*reflect.StringHeader)(unsafe.Pointer(&m1.key))
		header2 := (*reflect.StringHeader)(unsafe.Pointer(&m2.key))
		cmpLen := uintptr(base.MinInt(header1.Len, header2.Len))
		for i := uintptr(0); i < cmpLen; i++ {
			v1 := *(*byte)(unsafe.Pointer(header1.Data + i))
			v2 := *(*byte)(unsafe.Pointer(header2.Data + i))
			if v1 > v2 {
				return 1
			} else if v1 < v2 {
				return -1
			} else {
				// do nothing, compare next
			}
		}

		if header1.Len > header2.Len {
			return 1
		} else if header1.Len < header2.Len {
			return -1
		} else {
			return 0
		}
	}
}

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
		p.items = append(
			p.items,
			mapItem{key: key, fastKey: getFastKey(key), pos: pos},
		)
	}
}

func getSort4(items []mapItem, v uint32) uint32 {
	s1 := v & 0xF
	b1 := (v >> 4) & 0xF
	s2 := (v >> 8) & 0xF
	b2 := (v >> 12) & 0xF
	ret := uint32(0)

	if compareMapItem(items[s1], items[s2]) < 0 {
		if compareMapItem(items[b1], items[b2]) < 0 {
			ret |= s1
			ret |= b2 << 12
			if compareMapItem(items[s2], items[b1]) < 0 {
				ret |= s2 << 4
				ret |= b1 << 8
			} else {
				ret |= b1 << 4
				ret |= s2 << 8
			}
		} else {
			ret |= s1
			ret |= b1 << 12
			if compareMapItem(items[s2], items[b2]) < 0 {
				ret |= s2 << 4
				ret |= b2 << 8
			} else {
				ret |= b2 << 4
				ret |= s2 << 8
			}
		}
	} else {
		if compareMapItem(items[b1], items[b2]) < 0 {
			ret |= s2
			ret |= b2 << 12
			if compareMapItem(items[s1], items[b1]) < 0 {
				ret |= s1 << 4
				ret |= b1 << 8
			} else {
				ret |= b1 << 4
				ret |= s1 << 8
			}
		} else {
			ret |= s2
			ret |= b1 << 12
			if compareMapItem(items[s1], items[b2]) < 0 {
				ret |= s1 << 4
				ret |= b2 << 8
			} else {
				ret |= b2 << 4
				ret |= s1 << 8
			}
		}
	}
	return 0xFFFF0000 | ret
}

func (p *RTMap) getSort8() uint32 {
	if p.items != nil {
		mapSize := uint32(len(p.items))
		if mapSize > 0 && mapSize%8 == 0 {
			sort8 := uint32(0)
			sortItems := p.items[mapSize-8:]
			for i := uint32(0); i < 8; i += 2 {
				if compareMapItem(sortItems[i], sortItems[i+1]) < 0 {
					sort8 |= (i | (i+1)<<4) << (i * 4)
				} else {
					sort8 |= ((i + 1) | (i << 4)) << (i * 4)
				}
			}

			ret := uint32(0)
			pos := 0
			sort4LO := getSort4(sortItems, sort8&0xFFFF)
			sort4HI := getSort4(sortItems, sort8>>16)
			loV := sort4LO & 0xF
			hiV := sort4HI & 0xF

			for loV != 0xF && hiV != 0xF {
				if compareMapItem(sortItems[loV], sortItems[hiV]) < 0 {
					ret |= loV << pos
					sort4LO >>= 4
					loV = sort4LO & 0xF
				} else {
					ret |= hiV << pos
					sort4HI >>= 4
					hiV = sort4HI & 0xF
				}
				pos += 4
			}

			if loV != 0xF {
				ret |= sort4LO << pos
			} else if hiV != 0xF {
				ret |= sort4HI << pos
			} else {
				// do nothing
			}

			return ret
		}
	}

	return 0
}
