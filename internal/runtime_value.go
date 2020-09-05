package internal

import "sync/atomic"

type posRecord uint64

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
		thread.rtStream.setReadPosUnsafe(int(pos))
		ret := RTValue{
			rt:  rt,
			pos: record.getPos(),
		}
		ret.cacheString, ret.cacheOK = thread.rtStream.ReadUnsafeString()
		return ret
	}
}

func (p RTValue) ToUint64() (uint64, bool) {
	if thread := p.rt.lock(); thread != nil {
		defer p.rt.unlock()
		thread.rtStream.setReadPosUnsafe(int(p.pos))
		return thread.rtStream.ReadUint64()
	}

	return 0, false
}

func (p RTValue) ToString() (string, bool) {
	ret, ok := p.cacheString, p.cacheOK
	if p.rt.thread != nil &&
		atomic.LoadUint64(&p.rt.thread.top.lockStatus) == p.rt.id {
		return ret, ok
	} else {
		return "", false
	}
}
