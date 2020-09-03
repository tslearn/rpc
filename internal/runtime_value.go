package internal

type rtPosRecord uint64

func (p rtPosRecord) getPos() uint64 {
	return uint64(p) & 0x7FFFFFFFFFFFFFFF
}

func (p rtPosRecord) needBuffer() bool {
	return (p & 0x8000000000000000) != 0
}

func makePosRecord(pos uint64, needBuffer bool) rtPosRecord {
	if !needBuffer {
		return rtPosRecord(pos)
	} else {
		return rtPosRecord(0x8000000000000000 | pos)
	}
}

type RTValue struct {
	rt  Runtime
	pos int
	buf []byte
}
