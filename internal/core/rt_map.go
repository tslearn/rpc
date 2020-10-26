package core

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"reflect"
	"unsafe"
)

const sizeOfMapItem = int(unsafe.Sizeof(mapItem{}))

type mapItem struct {
	key     string
	fastKey uint32
	pos     posRecord
}

// RTMap ...
type RTMap struct {
	rt    Runtime
	items *[]mapItem
}

func newRTMap(rt Runtime, size int) (ret RTMap) {
	if thread := rt.thread; thread != nil {
		ret.rt = rt
		size += 4

		if d1 := thread.malloc(sizeOfSlice); d1 != nil {
			ret.items = (*[]mapItem)(d1)

			if d2 := thread.malloc(sizeOfMapItem * size); d2 != nil && size <= 20 {
				itemsHeader := (*reflect.SliceHeader)(d1)
				itemsHeader.Len = 0
				itemsHeader.Cap = size
				itemsHeader.Data = uintptr(d2)
				return
			}

			*ret.items = make([]mapItem, 0, size)
		} else {
			items := make([]mapItem, 0, size)
			ret.items = &items
		}
	}

	return
}

func (p *RTMap) getPosRecord(key string, fastKey uint32) (int, posRecord) {
	if p.items != nil {
		items := *p.items
		size := len(items)
		randStart := size - size%8

		if randStart > 0 {
			bsStart := 0
			bsEnd := randStart - 1
			for bsStart <= bsEnd {
				mid := (bsStart + bsEnd) >> 1

				if v := compareItem(&items[mid], key, fastKey); v > 0 {
					bsEnd = mid - 1
				} else if v < 0 {
					bsStart = mid + 1
				} else {
					return mid, items[mid].pos
				}
			}
		}

		for i := size - 1; i >= randStart; i-- {
			if compareItem(&items[i], key, fastKey) == 0 {
				return i, items[i].pos
			}
		}
	}

	return -1, 0
}

// Get ...
func (p *RTMap) Get(key string) RTValue {
	if _, pos := p.getPosRecord(key, getFastKey(key)); pos > 0 {
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
		return len(*p.items)
	}

	return -1
}

// Set
func (p *RTMap) Set(key string, value interface{}) *base.Error {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()

		pos := int64(thread.rtStream.GetWritePos())

		if reason := thread.rtStream.Write(value); reason != StreamWriteOK {
			return errors.ErrUnsupportedValue.AddDebug(reason)
		}

		switch value.(type) {
		case string:
			p.appendValue(key, makePosRecord(pos, true))
		default:
			p.appendValue(key, makePosRecord(pos, false))
		}

		return nil
	} else {
		return errors.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// Delete ...
func (p *RTMap) Delete(key string) *base.Error {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()

		if idx, r := p.getPosRecord(key, getFastKey(key)); idx > 0 && r > 0 {
			(*p.items)[idx].pos = 0
			return nil
		}

		return errors.ErrRTMapNameNotFound.
			AddDebug(fmt.Sprintf("RTMap key %s is not exist", key))
	} else {
		return errors.ErrRuntimeIllegalInCurrentGoroutine
	}
}

func (p *RTMap) appendValue(key string, pos posRecord) {
	fastKey := getFastKey(key)

	if idx, _ := p.getPosRecord(key, fastKey); idx > 0 {
		(*p.items)[idx].pos = pos
	} else {
		*p.items = append(
			*p.items,
			mapItem{key: key, fastKey: fastKey, pos: pos},
		)
	}

	if len(*p.items)%8 == 0 {
		p.sort()
	}
}

func (p *RTMap) sort() {
	items := *p.items
	size := len(items)

	arrBuffer := [8]mapItem{}
	sortItems := items[size-8:]
	sort8 := getSort8(sortItems)

	if size == 8 {
		copy(arrBuffer[0:], sortItems)
		for i := uint32(0); i < 8; i++ {
			oIndex := sort8 & 0xF
			sort8 >>= 4
			if i != oIndex {
				sortItems[i] = arrBuffer[oIndex]
			}
		}
	} else {
		for i := uint32(0); i < 8; i++ {
			oIndex := sort8 & 0xF
			sort8 >>= 4
			arrBuffer[i] = sortItems[oIndex]
		}
		copy(items[8:], items[0:])

		// merge
		i := 8
		j := 0
		k := 0
		for i < size && j < 8 {
			if orderLess(&items[i], &arrBuffer[j]) {
				items[k] = items[i]
				i++
			} else {
				items[k] = arrBuffer[j]
				j++
			}

			if items[k].pos != 0 {
				k++
			}
		}

		for i < size {
			items[k] = items[i]
			i++
			if items[k].pos != 0 {
				k++
			}
		}

		for j < 8 {
			items[k] = arrBuffer[j]
			j++
			if items[k].pos != 0 {
				k++
			}
		}

		*p.items = items[:k]
	}
}

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

func compareItem(m1 *mapItem, key string, fastKey uint32) int {
	if m1.fastKey > fastKey {
		return 1
	} else if m1.fastKey < fastKey {
		return -1
	} else if m1.key > key {
		return 1
	} else if m1.key < key {
		return -1
	} else {
		return 0
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
