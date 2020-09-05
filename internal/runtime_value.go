package internal

type posRecord uint64

func (p posRecord) getPos() int64 {
	return int64(p) & 0x7FFFFFFFFFFFFFFF
}

func (p posRecord) needBuffer() bool {
	return (p & 0x8000000000000000) != 0
}

func makePosRecord(pos int64, needBuffer bool) posRecord {
	if !needBuffer {
		return posRecord(pos)
	} else {
		return 0x8000000000000000 | posRecord(pos)
	}
}

type RTValue struct {
	rt  Runtime
	pos int64
	buf []byte
}

func makeRTValue(rt Runtime, record posRecord) RTValue {
	if !record.needBuffer() {
		return RTValue{
			rt:  rt,
			pos: record.getPos(),
			buf: nil,
		}
	} else if thread := rt.lock(); thread == nil {
		return RTValue{
			rt:  rt,
			pos: record.getPos(),
			buf: nil,
		}
	} else {
		defer rt.unlock()
		pos := record.getPos()
		return RTValue{
			rt:  rt,
			pos: record.getPos(),
			buf: thread.rtStream.tryToReadBufferCacheForStringOrBytes(int(pos)),
		}
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
	if p.buf != nil && p.buf[0] > 128 && p.buf[0] < 191 {
		return string(p.buf[1:]), true
	}
	return "", false
}
