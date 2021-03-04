package rpc

import (
	"sync/atomic"

	"github.com/rpccloud/rpc/internal/base"
)

// RTValue ...
type RTValue struct {
	err        *base.Error
	rt         Runtime
	pos        int64
	cacheBytes []byte
	cacheSafe  bool
	cacheError *base.Error
}

func makeRTValue(rt Runtime, record posRecord) RTValue {
	if !record.isString() {
		return RTValue{
			err:        nil,
			rt:         rt,
			pos:        record.getPos(),
			cacheBytes: nil,
			cacheSafe:  true,
			cacheError: base.ErrStream,
		}
	}

	rtStream := rt.thread.rtStream
	pos := record.getPos()
	rtStream.SetReadPos(int(pos))
	cacheString, cacheSafe, cacheError := rtStream.readUnsafeString()
	return RTValue{
		err:        nil,
		rt:         rt,
		pos:        pos,
		cacheBytes: base.StringToBytesUnsafe(cacheString),
		cacheSafe:  cacheSafe,
		cacheError: cacheError,
	}
}

// ToBool ...
func (p RTValue) ToBool() (Bool, *base.Error) {
	if p.err != nil {
		return false, p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadBool()
	} else {
		return false, base.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// ToInt64 ...
func (p RTValue) ToInt64() (Int64, *base.Error) {
	if p.err != nil {
		return 0, p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadInt64()
	} else {
		return 0, base.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// ToUint64 ...
func (p RTValue) ToUint64() (Uint64, *base.Error) {
	if p.err != nil {
		return 0, p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadUint64()
	} else {
		return 0, base.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// ToFloat64 ...
func (p RTValue) ToFloat64() (Float64, *base.Error) {
	if p.err != nil {
		return 0, p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadFloat64()
	} else {
		return 0, base.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// ToString ...
func (p RTValue) ToString() (String, *base.Error) {
	if p.err != nil {
		return "", p.err
	} else if p.rt.thread == nil ||
		atomic.LoadUint64(&p.rt.thread.top.lockStatus) != p.rt.id {
		return "", base.ErrRuntimeIllegalInCurrentGoroutine
	} else {
		return string(p.cacheBytes), p.cacheError
	}
}

// ToBytes ...
func (p RTValue) ToBytes() (Bytes, *base.Error) {
	if p.err != nil {
		return Bytes{}, p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadBytes()
	} else {
		return Bytes{}, base.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// ToArray ...
func (p RTValue) ToArray() (Array, *base.Error) {
	if p.err != nil {
		return Array{}, p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadArray()
	} else {
		return Array{}, base.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// ToRTArray ...
func (p RTValue) ToRTArray() (RTArray, *base.Error) {
	if p.err != nil {
		return newRTArray(p.rt, 0), p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadRTArray(p.rt)
	} else {
		return RTArray{}, base.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// ToMap ...
func (p RTValue) ToMap() (Map, *base.Error) {
	if p.err != nil {
		return Map{}, p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadMap()
	} else {
		return Map{}, base.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// ToRTMap ...
func (p RTValue) ToRTMap() (RTMap, *base.Error) {
	if p.err != nil {
		return RTMap{}, p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadRTMap(p.rt)
	} else {
		return RTMap{}, base.ErrRuntimeIllegalInCurrentGoroutine
	}
}
