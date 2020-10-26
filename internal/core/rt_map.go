package core

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/errors"
	"reflect"
	"unsafe"
)

type mapItem struct {
	key     string
	fastKey uint32
	active  bool
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
			mapItem{key: key, fastKey: getFastKey(key), active: true, pos: pos},
		)

		if len(p.items)%8 == 0 {
			p.sort()
		}
	}
}

func (p *RTMap) sort() {
	size := len(p.items)

	arrBuffer := [8]mapItem{}
	items := p.items[size-8:]
	sort8 := getSort8(items)

	if size == 8 {
		copy(arrBuffer[0:], items)
		for i := uint32(0); i < 8; i++ {
			oIndex := sort8 & 0xF
			sort8 >>= 4
			if i != oIndex {
				items[i] = arrBuffer[oIndex]
			}
		}
	} else {
		for i := uint32(0); i < 8; i++ {
			oIndex := sort8 & 0xF
			sort8 >>= 4
			arrBuffer[i] = items[oIndex]
		}
		copy(p.items[8:], p.items[0:])

		// merge
		i := 8
		j := 0
		k := 0
		for i < size && j < 8 {
			if orderLess(&p.items[i], &arrBuffer[j]) {
				p.items[k] = p.items[i]
				i++
			} else {
				p.items[k] = arrBuffer[j]
				j++
			}

			if p.items[k].active {
				k++
			}
		}

		for i < size {
			p.items[k] = p.items[i]
			i++
			if p.items[k].active {
				k++
			}
		}

		for j < 8 {
			p.items[k] = arrBuffer[j]
			j++
			if p.items[k].active {
				k++
			}
		}

		p.items = p.items[:k]
	}
}

func orderLess(m1 *mapItem, m2 *mapItem) bool {
	if m1.fastKey > m2.fastKey {
		return false
	} else if m1.fastKey < m2.fastKey {
		return true
	} else {
		return m1.key < m2.key
	}
}

func getSort4(items []mapItem, v uint32) uint32 {
	s1 := v & 0xF
	b1 := (v >> 4) & 0xF
	s2 := (v >> 8) & 0xF
	b2 := (v >> 12) & 0xF
	ret := uint32(0)

	if orderLess(&items[s1], &items[s2]) {
		if orderLess(&items[b1], &items[b2]) {
			ret |= s1
			ret |= b2 << 12
			if orderLess(&items[s2], &items[b1]) {
				ret |= s2 << 4
				ret |= b1 << 8
			} else {
				ret |= b1 << 4
				ret |= s2 << 8
			}
		} else {
			ret |= s1
			ret |= b1 << 12
			if orderLess(&items[s2], &items[b2]) {
				ret |= s2 << 4
				ret |= b2 << 8
			} else {
				ret |= b2 << 4
				ret |= s2 << 8
			}
		}
	} else {
		if orderLess(&items[b1], &items[b2]) {
			ret |= s2
			ret |= b2 << 12
			if orderLess(&items[s1], &items[b1]) {
				ret |= s1 << 4
				ret |= b1 << 8
			} else {
				ret |= b1 << 4
				ret |= s1 << 8
			}
		} else {
			ret |= s2
			ret |= b1 << 12
			if orderLess(&items[s1], &items[b2]) {
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

func getSort8(items []mapItem) uint32 {
	sort8 := uint32(0)
	for i := uint32(0); i < 8; i += 2 {
		if orderLess(&items[i], &items[i+1]) {
			sort8 |= (i | (i+1)<<4) << (i * 4)
		} else {
			sort8 |= ((i + 1) | (i << 4)) << (i * 4)
		}
	}

	ret := uint32(0)
	pos := 0
	sort4LO := getSort4(items, sort8&0xFFFF)
	sort4HI := getSort4(items, sort8>>16)
	loV := sort4LO & 0xF
	hiV := sort4HI & 0xF

	for loV != 0xF && hiV != 0xF {
		if orderLess(&items[loV], &items[hiV]) {
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
