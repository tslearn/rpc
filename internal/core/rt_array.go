package core

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"reflect"
)

// RTArray ...
type RTArray struct {
	rt    Runtime
	items *[]posRecord
}

func newRTArray(rt Runtime, size int) (ret RTArray) {
	if thread := rt.thread; thread != nil && size >= 0 {
		ret.rt = rt

		if d1 := thread.malloc(sizeOfSlice); d1 != nil {
			ret.items = (*[]posRecord)(d1)

			if d2 := thread.malloc(sizeOfPosRecord * size); d2 != nil && size <= 64 {
				itemsHeader := (*reflect.SliceHeader)(d1)
				itemsHeader.Len = 0
				itemsHeader.Cap = size
				itemsHeader.Data = uintptr(d2)
				return
			}

			*ret.items = make([]posRecord, 0, size)
		} else {
			items := make([]posRecord, 0, size)
			ret.items = &items
		}
	}

	return
}

// Get ...
func (p *RTArray) Get(index int) RTValue {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()

		if index >= 0 && index < len(*p.items) {
			return makeRTValue(p.rt, (*p.items)[index])
		}

		return RTValue{
			err: errors.ErrRTArrayIndexOverflow.
				AddDebug(fmt.Sprintf("RTArray index %d out of range", index)),
		}
	}

	return RTValue{
		err: errors.ErrRuntimeIllegalInCurrentGoroutine,
	}
}

func (p *RTArray) Set(index int, value interface{}) *base.Error {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()

		pos := int64(thread.rtStream.GetWritePos())

		if reason := thread.rtStream.Write(value); reason != StreamWriteOK {
			return errors.ErrUnsupportedValue.AddDebug(reason)
		}

		if index < 0 || index >= len(*p.items) {
			return errors.ErrRTArrayIndexOverflow.
				AddDebug(fmt.Sprintf("RTArray index %d out of range", index))
		}

		switch value.(type) {
		case string:
			(*p.items)[index] = makePosRecord(pos, true)
		default:
			(*p.items)[index] = makePosRecord(pos, false)
		}
		return nil
	}

	return errors.ErrRuntimeIllegalInCurrentGoroutine
}

func (p *RTArray) Append(value interface{}) *base.Error {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()

		pos := int64(thread.rtStream.GetWritePos())

		if reason := thread.rtStream.Write(value); reason != StreamWriteOK {
			return errors.ErrUnsupportedValue.AddDebug(reason)
		}

		switch value.(type) {
		case string:
			*p.items = append(*p.items, makePosRecord(pos, true))
		default:
			*p.items = append(*p.items, makePosRecord(pos, false))
		}
		return nil
	}

	return errors.ErrRuntimeIllegalInCurrentGoroutine
}

func (p *RTArray) Delete(index int) *base.Error {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()

		items := *p.items
		size := len(items)
		if index < 0 || index >= size {
			return errors.ErrRTArrayIndexOverflow.
				AddDebug(fmt.Sprintf("RTArray index %d out of range", index))
		}

		copy(items[index:], items[index+1:])
		*p.items = items[:size-1]

		return nil
	}

	return errors.ErrRuntimeIllegalInCurrentGoroutine
}

// Size ...
func (p *RTArray) Size() int {
	if p.rt.lock() != nil {
		defer p.rt.unlock()
		return len(*p.items)
	}

	return -1
}
