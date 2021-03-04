package rpc

import (
	"fmt"

	"github.com/rpccloud/rpc/internal/base"
)

// RTArray ...
type RTArray struct {
	rt    Runtime
	items *[]posRecord
}

// Get ...
func (p RTArray) Get(index int) RTValue {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()

		if index >= 0 && index < len(*p.items) {
			return makeRTValue(p.rt, (*p.items)[index])
		}

		return RTValue{
			err: base.ErrRTArrayIndexOverflow.
				AddDebug(fmt.Sprintf("RTArray index %d out of range", index)),
		}
	}

	return RTValue{
		err: base.ErrRuntimeIllegalInCurrentGoroutine,
	}
}

// Set ...
func (p RTArray) Set(index int, value interface{}) *base.Error {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()

		pos := int64(thread.rtStream.GetWritePos())

		if reason := thread.rtStream.Write(value); reason != StreamWriteOK {
			return base.ErrUnsupportedValue.AddDebug(reason)
		}

		if index < 0 || index >= len(*p.items) {
			return base.ErrRTArrayIndexOverflow.
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

	return base.ErrRuntimeIllegalInCurrentGoroutine
}

// Append ...
func (p RTArray) Append(value interface{}) *base.Error {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()

		pos := int64(thread.rtStream.GetWritePos())

		if reason := thread.rtStream.Write(value); reason != StreamWriteOK {
			return base.ErrUnsupportedValue.AddDebug(reason)
		}

		switch value.(type) {
		case string:
			*p.items = append(*p.items, makePosRecord(pos, true))
		default:
			*p.items = append(*p.items, makePosRecord(pos, false))
		}
		return nil
	}

	return base.ErrRuntimeIllegalInCurrentGoroutine
}

// Delete ...
func (p RTArray) Delete(index int) *base.Error {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()

		items := *p.items
		size := len(items)
		if index < 0 || index >= size {
			return base.ErrRTArrayIndexOverflow.
				AddDebug(fmt.Sprintf("RTArray index %d out of range", index))
		}

		copy(items[index:], items[index+1:])
		*p.items = items[:size-1]

		return nil
	}

	return base.ErrRuntimeIllegalInCurrentGoroutine
}

// DeleteAll ...
func (p RTArray) DeleteAll() *base.Error {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		*p.items = (*p.items)[:0]
		return nil
	}

	return base.ErrRuntimeIllegalInCurrentGoroutine
}

// Size ...
func (p RTArray) Size() int {
	if p.rt.lock() != nil {
		defer p.rt.unlock()
		return len(*p.items)
	}

	return -1
}
