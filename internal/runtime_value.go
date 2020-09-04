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
