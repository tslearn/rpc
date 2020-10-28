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

func compareMapItem(m1 *mapItem, key string, fastKey uint32) int {
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

func isMapItemLess(m1 *mapItem, m2 *mapItem) bool {
	if m1.fastKey > m2.fastKey {
		return false
	} else if m1.fastKey < m2.fastKey {
		return true
	} else {
		return m1.key < m2.key
	}
}

func getSort4(items []mapItem, start uint64) uint64 {
	ret := uint64(0)
	s1 := start
	b1 := start
	s2 := start
	b2 := start

	if isMapItemLess(&items[start], &items[start+1]) {
		s1 += 0
		b1 += 1
	} else {
		s1 += 1
		b1 += 0
	}

	if isMapItemLess(&items[start+2], &items[start+3]) {
		s2 += 2
		b2 += 3
	} else {
		s2 += 3
		b2 += 2
	}

	if isMapItemLess(&items[s1], &items[s2]) {
		if isMapItemLess(&items[b1], &items[b2]) {
			ret |= s1
			ret |= b2 << 12
			if isMapItemLess(&items[s2], &items[b1]) {
				ret |= s2 << 4
				ret |= b1 << 8
			} else {
				ret |= b1 << 4
				ret |= s2 << 8
			}
		} else {
			ret |= s1
			ret |= s2 << 4
			ret |= b2 << 8
			ret |= b1 << 12
		}
	} else {
		if isMapItemLess(&items[b1], &items[b2]) {
			ret |= s2
			ret |= s1 << 4
			ret |= b1 << 8
			ret |= b2 << 12
		} else {
			ret |= s2
			ret |= b1 << 12
			if isMapItemLess(&items[s1], &items[b2]) {
				ret |= s1 << 4
				ret |= b2 << 8
			} else {
				ret |= b2 << 4
				ret |= s1 << 8
			}
		}
	}

	return 0xFFFFFFFFFFFF0000 | ret
}

func getSort8(items []mapItem, start uint64) uint64 {
	ret := uint64(0)
	pos := 0
	sort4LO := getSort4(items, start)
	sort4HI := getSort4(items, start+4)
	loV := sort4LO & 0xF
	hiV := sort4HI & 0xF

	for sort4LO != 0x0000FFFFFFFFFFFF && sort4HI != 0x0000FFFFFFFFFFFF {
		if isMapItemLess(&items[loV], &items[hiV]) {
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

	if sort4LO != 0x0000FFFFFFFFFFFF {
		ret |= sort4LO << pos
	} else {
		ret |= sort4HI << pos
	}

	return 0xFFFFFFFF00000000 | ret
}

func getSort16(items []mapItem) uint64 {
	ret := uint64(0)
	pos := 0
	sort8LO := getSort8(items, 0)
	sort8HI := getSort8(items, 8)
	loV := sort8LO & 0xF
	hiV := sort8HI & 0xF

	for sort8LO != 0x00000000FFFFFFFF && sort8HI != 0x00000000FFFFFFFF {
		if isMapItemLess(&items[loV], &items[hiV]) {
			ret |= loV << pos
			sort8LO >>= 4
			loV = sort8LO & 0xF
		} else {
			ret |= hiV << pos
			sort8HI >>= 4
			hiV = sort8HI & 0xF
		}
		pos += 4
	}

	if sort8LO != 0x00000000FFFFFFFF {
		ret |= sort8LO << pos
	} else {
		ret |= sort8HI << pos
	}

	return ret
}

// RTMap ...
type RTMap struct {
	rt    Runtime
	items *[]mapItem
}

func newRTMap(rt Runtime, size int) (ret RTMap) {
	if thread := rt.thread; thread != nil {
		ret.rt = rt

		if d1 := thread.malloc(sizeOfSlice); d1 != nil {
			ret.items = (*[]mapItem)(d1)

			if d2 := thread.malloc(sizeOfMapItem * size); d2 != nil && size <= 16 {
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

// Get ...
func (p *RTMap) Get(key string) RTValue {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()

		if _, pos := p.getPosRecord(key, getFastKey(key)); pos > 0 {
			return makeRTValue(p.rt, pos)
		}

		return RTValue{
			err: errors.ErrRTMapNameNotFound.
				AddDebug(fmt.Sprintf("RTMap key %s does not exist", key)),
		}
	}

	return RTValue{
		err: errors.ErrRuntimeIllegalInCurrentGoroutine,
	}
}

// Set ...
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
	}

	return errors.ErrRuntimeIllegalInCurrentGoroutine
}

// Delete ...
func (p *RTMap) Delete(key string) *base.Error {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()

		if idx, r := p.getPosRecord(key, getFastKey(key)); idx >= 0 && r > 0 {
			(*p.items)[idx].pos = 0
			return nil
		}

		return errors.ErrRTMapNameNotFound.
			AddDebug(fmt.Sprintf("RTMap key %s does not exist", key))
	}

	return errors.ErrRuntimeIllegalInCurrentGoroutine
}

// Size ...
func (p *RTMap) Size() int {
	if p.items != nil {
		return len(*p.items)
	}

	return -1
}

func (p *RTMap) getPosRecord(key string, fastKey uint32) (int, posRecord) {
	if p.items != nil {
		items := *p.items
		size := len(items)
		unsortedStart := size - size%16

		if unsortedStart > 0 {
			bsStart := 0
			bsEnd := unsortedStart - 1
			for bsStart <= bsEnd {
				mid := (bsStart + bsEnd) >> 1

				if v := compareMapItem(&items[mid], key, fastKey); v > 0 {
					bsEnd = mid - 1
				} else if v < 0 {
					bsStart = mid + 1
				} else {
					return mid, items[mid].pos
				}
			}
		}

		for i := size - 1; i >= unsortedStart; i-- {
			if compareMapItem(&items[i], key, fastKey) == 0 {
				return i, items[i].pos
			}
		}
	}

	return -1, 0
}

func (p *RTMap) appendValue(key string, pos posRecord) {
	fastKey := getFastKey(key)

	if idx, _ := p.getPosRecord(key, fastKey); idx >= 0 {
		(*p.items)[idx].pos = pos
	} else {
		*p.items = append(
			*p.items,
			mapItem{key: key, fastKey: fastKey, pos: pos},
		)
	}

	if len(*p.items)%16 == 0 {
		p.sort()
	}
}

func (p *RTMap) sort() {
	items := *p.items
	size := len(items)

	arrBuffer := [16]mapItem{}
	sortItems := items[size-16:]
	sort16 := getSort16(sortItems)

	if size == 16 {
		copy(arrBuffer[0:], sortItems)
		for i := uint64(0); i < 16; i++ {
			oIndex := sort16 & 0xF
			sort16 >>= 4
			if i != oIndex {
				sortItems[i] = arrBuffer[oIndex]
			}
		}
	} else {
		for i := 0; i < 16; i++ {
			oIndex := sort16 & 0xF
			sort16 >>= 4
			arrBuffer[i] = sortItems[oIndex]
		}
		copy(items[16:], items[0:])

		// merge
		i := 16
		j := 0
		k := 0
		for i < size && j < 16 {
			if isMapItemLess(&items[i], &arrBuffer[j]) {
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

		if i < size {
			for i < size {
				items[k] = items[i]
				i++
				if items[k].pos != 0 {
					k++
				}
			}
		} else {
			for j < 16 {
				items[k] = arrBuffer[j]
				j++
				if items[k].pos != 0 {
					k++
				}
			}
		}

		*p.items = items[:k]
	}
}
