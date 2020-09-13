package internal

import (
	"sync/atomic"
	"unsafe"
)

type posRecord uint64

const sizeOfPosRecord = int(unsafe.Sizeof(posRecord(0)))

func (p posRecord) getPos() int64 {
	return int64(p) & 0x7FFFFFFFFFFFFFFF
}

func (p posRecord) isString() bool {
	return (p & 0x8000000000000000) != 0
}

func makePosRecord(pos int64, isString bool) posRecord {
	if !isString {
		return posRecord(pos)
	} else {
		return 0x8000000000000000 | posRecord(pos)
	}
}

type RTValue struct {
	rt          Runtime
	pos         int64
	cacheString string
	cacheOK     bool
}

func makeRTValue(rt Runtime, record posRecord) RTValue {
	if !record.isString() {
		return RTValue{
			rt:          rt,
			pos:         record.getPos(),
			cacheString: "",
			cacheOK:     false,
		}
	} else if thread := rt.lock(); thread == nil {
		return RTValue{
			rt:          rt,
			pos:         record.getPos(),
			cacheString: "",
			cacheOK:     false,
		}
	} else {
		defer rt.unlock()
		pos := record.getPos()
		thread.rtStream.SetReadPos(int(pos))
		ret := RTValue{
			rt:  rt,
			pos: record.getPos(),
		}
		ret.cacheString, ret.cacheOK = thread.rtStream.ReadUnsafeString()
		return ret
	}
}

func (p RTValue) ToBool() (Bool, bool) {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadBool()
	}

	return false, false
}

func (p RTValue) ToInt64() (Int64, bool) {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadInt64()
	}

	return 0, false
}

func (p RTValue) ToUint64() (Uint64, bool) {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadUint64()
	}

	return 0, false
}

func (p RTValue) ToFloat64() (Float64, bool) {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadFloat64()
	}

	return 0, false
}

func (p RTValue) ToString() (String, bool) {
	buf := []byte(p.cacheString)
	ret := string(buf)

	if p.cacheOK &&
		atomic.LoadUint64(&p.rt.thread.top.lockStatus) == p.rt.id {
		return ret, true
	} else {
		return "", false
	}
}

func (p RTValue) ToBytes() (Bytes, bool) {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadBytes()
	}

	return Bytes(nil), false
}

func (p RTValue) ToArray() (Array, bool) {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadArray()
	}

	return Array(nil), false
}

func (p RTValue) ToRTArray() (RTArray, bool) {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadRTArray(p.rt)
	}

	return RTArray{}, false
}

func (p RTValue) ToMap() (Map, bool) {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadMap()
	}

	return Map(nil), false
}

func (p RTValue) ToRTMap() (RTMap, bool) {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.SetReadPos(int(p.pos))
		return thread.rtStream.ReadRTMap(p.rt)
	}

	return RTMap{}, false
}
