package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"sync/atomic"
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

func makeRTValue(rt Runtime, record posRecord) (ret RTValue) {
	if !record.isString() {
		return RTValue{
			err:        nil,
			rt:         rt,
			pos:        record.getPos(),
			cacheBytes: []byte{},
			cacheSafe:  true,
			cacheError: errors.ErrStream,
		}
	} else {
		pos := record.getPos()
		rtStream := rt.thread.rtStream
		ret.rt = rt
		ret.pos = pos
		rtStream.SetReadPos(int(pos))
		cacheString, cacheSafe, cacheError := rtStream.readUnsafeString()
		ret.cacheBytes = base.StringToBytesUnsafe(cacheString)
		ret.cacheSafe = cacheSafe
		ret.cacheError = cacheError
		return
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
		return false, errors.ErrRuntimeIllegalInCurrentGoroutine
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
		return 0, errors.ErrRuntimeIllegalInCurrentGoroutine
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
		return 0, errors.ErrRuntimeIllegalInCurrentGoroutine
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
		return 0, errors.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// ToString ...
func (p RTValue) ToString() (String, *base.Error) {
	if p.err != nil {
		return "", p.err
	} else if p.rt.thread == nil ||
		atomic.LoadUint64(&p.rt.thread.top.lockStatus) != p.rt.id {
		return "", errors.ErrRuntimeIllegalInCurrentGoroutine
	} else {
		return string(p.cacheBytes), p.cacheError
	}
}

// ToBytes ...
func (p RTValue) ToBytes() (Bytes, *base.Error) {
	if p.err != nil {
		return Bytes(nil), p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadBytes()
	} else {
		return Bytes(nil), errors.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// ToArray ...
func (p RTValue) ToArray() (Array, *base.Error) {
	if p.err != nil {
		return Array(nil), p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadArray()
	} else {
		return Array(nil), errors.ErrRuntimeIllegalInCurrentGoroutine
	}
}

// ToRTArray ...
func (p RTValue) ToRTArray() (RTArray, *base.Error) {
	if p.err != nil {
		return RTArray{}, p.err
	} else if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadRTArray(p.rt)
	} else {
		return RTArray{}, errors.ErrRuntimeIllegalInCurrentGoroutine
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
		return Map{}, errors.ErrRuntimeIllegalInCurrentGoroutine
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
		return RTMap{}, errors.ErrRuntimeIllegalInCurrentGoroutine
	}
}
