package core

import (
	"math"
	"reflect"
	"sync"
	"unsafe"
)

var (
	rpcStreamCache = sync.Pool{
		New: func() interface{} {
			ret := rpcStream{
				frames:     make([]*[]byte, 1, 8),
				readSeg:    0,
				readIndex:  17,
				writeSeg:   0,
				writeIndex: 17,
			}
			zeroFrame := make([]byte, 512, 512)
			ret.frames[0] = &zeroFrame
			ret.readFrame = zeroFrame
			ret.writeFrame = zeroFrame
			ret.writeFrame[0] = 1
			ret.header = zeroFrame[1:17]
			return &ret
		},
	}
	frameCache = &sync.Pool{
		New: func() interface{} {
			ret := make([]byte, 512, 512)
			return &ret
		},
	}

	readSkipArray = make([]int, 256, 256)
)

func init() {
	for op := 0; op < 256; op++ {
		switch op {
		case 0:
			readSkipArray[op] = -64
		case 1:
			readSkipArray[op] = 1
		case 2:
			readSkipArray[op] = 1
		case 3:
			readSkipArray[op] = 1
		case 4:
			readSkipArray[op] = 1
		case 5:
			readSkipArray[op] = 9
		case 6:
			readSkipArray[op] = 3
		case 7:
			readSkipArray[op] = 5
		case 8:
			readSkipArray[op] = 9
		case 9:
			readSkipArray[op] = 3
		case 10:
			readSkipArray[op] = 5
		case 11:
			readSkipArray[op] = 9
		case 12:
			readSkipArray[op] = -64
		case 13:
			readSkipArray[op] = -64
		case 64:
			readSkipArray[op] = 1
		case 96:
			readSkipArray[op] = 1
		case 128:
			readSkipArray[op] = 1
		case 191:
			readSkipArray[op] = -6
		case 192:
			readSkipArray[op] = 1
		case 255:
			readSkipArray[op] = -5
		default:
			switch op >> 6 {
			case 0:
				readSkipArray[op] = 1
			case 1:
				readSkipArray[op] = 0
			case 2:
				readSkipArray[op] = int(op) - 126
			case 3:
				readSkipArray[op] = int(op) - 191
			}
		}
	}
}

// rpcStream ...
type rpcStream struct {
	frames []*[]byte

	readSeg   int
	readIndex int
	readFrame []byte

	writeSeg   int
	writeIndex int
	writeFrame []byte

	header []byte
}

// newRPCStream ...
func newRPCStream() *rpcStream {
	return rpcStreamCache.Get().(*rpcStream)
}

// Reset ...
func (p *rpcStream) Reset() {
	// reset frames
	for i := 1; i < len(p.frames); i++ {
		frameCache.Put(p.frames[i])
		p.frames[i] = nil
	}
	if cap(p.frames) > 8 {
		newFrames := make([]*[]byte, 1, 8)
		newFrames[0] = p.frames[0]
		p.frames = newFrames
	} else {
		p.frames = p.frames[:1]
	}

	p.readSeg = 0
	p.readIndex = 17
	p.readFrame = *p.frames[0]

	p.writeSeg = 0
	p.writeIndex = 17
	p.writeFrame = *p.frames[0]
}

func (p *rpcStream) getServerCallbackID() uint64 {
	b := p.header[0:8]
	return uint64(b[0]) |
		uint64(b[1])<<8 |
		uint64(b[2])<<16 |
		uint64(b[3])<<24 |
		uint64(b[4])<<32 |
		uint64(b[5])<<40 |
		uint64(b[6])<<48 |
		uint64(b[7])<<56
}

func (p *rpcStream) setServerCallbackID(v uint64) {
	b := p.header[0:8]
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
}

func (p *rpcStream) getClientCallbackID() uint32 {
	b := p.header[0:4]
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func (p *rpcStream) setClientCallbackID(v uint32) {
	b := p.header[0:4]
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}

func (p *rpcStream) getClientConnID() uint32 {
	b := p.header[4:8]
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func (p *rpcStream) setClientConnID(v uint32) {
	b := p.header[4:8]
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}

func (p *rpcStream) getMachineID() uint32 {
	b := p.header[8:12]
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func (p *rpcStream) setMachineID(v uint32) {
	b := p.header[8:12]
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}

func (p *rpcStream) getRouterID() uint32 {
	b := p.header[12:16]
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func (p *rpcStream) setRouterID(v uint32) {
	b := p.header[12:16]
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}

// Release clean the rpcStream
func (p *rpcStream) Release() {
	p.Reset()
	rpcStreamCache.Put(p)
}

// getBuffer ...
func (p *rpcStream) getBuffer() []byte {
	length := p.GetWritePos()
	ret := make([]byte, length, length)
	for i := 0; i <= p.writeSeg; i++ {
		copy(ret[i<<9:], *p.frames[i])
	}
	return ret
}

// getBufferUnsafe
func (p *rpcStream) getBufferUnsafe() []byte {
	if p.writeSeg == 0 {
		return p.writeFrame[0:p.writeIndex]
	} else {
		return p.getBuffer()
	}
}

// GetReadPos get the current read pos of the stream
func (p *rpcStream) GetReadPos() int {
	return (p.readSeg << 9) | p.readIndex
}

// SetReadPos set the current read pos of the stream
func (p *rpcStream) SetReadPos(pos int) bool {
	readSeg := pos >> 9
	readIndex := pos & 0x1FF
	if (readSeg == p.writeSeg && readIndex <= p.writeIndex) ||
		(readSeg < p.writeSeg && readSeg >= 0) {
		p.readIndex = readIndex
		if p.readSeg != readSeg {
			p.readSeg = readSeg
			p.readFrame = *p.frames[readSeg]
		}
		return true
	}
	return false
}

func (p *rpcStream) setReadPosUnsafe(pos int) {
	p.readIndex = pos & 0x1FF
	readSeg := pos >> 9
	if p.readSeg != readSeg {
		p.readSeg = readSeg
		p.readFrame = *p.frames[readSeg]
	}
}

// GetWritePos ...
func (p *rpcStream) GetWritePos() int {
	return (p.writeSeg << 9) | p.writeIndex
}

// SetWritePos ...
func (p *rpcStream) SetWritePos(length int) {
	numToCreate := length>>9 - len(p.frames) + 1

	for numToCreate > 0 {
		p.frames = append(p.frames, frameCache.Get().(*[]byte))
		numToCreate--
	}
	p.setWritePosUnsafe(length)
}

func (p *rpcStream) setWritePosUnsafe(pos int) {
	p.writeIndex = pos & 0x1FF
	writeSeg := pos >> 9
	if p.writeSeg != writeSeg {
		p.writeSeg = writeSeg
		p.writeFrame = *p.frames[writeSeg]
	}
}

// CanRead return true if the stream is not finish
func (p *rpcStream) CanRead() bool {
	return p.readIndex < p.writeIndex || p.readSeg < p.writeSeg
}

func (p *rpcStream) gotoNextWriteFrame() {
	p.writeSeg++
	p.writeIndex = 0
	if p.writeSeg == len(p.frames) {
		p.frames = append(p.frames, frameCache.Get().(*[]byte))
	}
	p.writeFrame = *p.frames[p.writeSeg]
}

func (p *rpcStream) gotoNextReadFrameUnsafe() {
	p.readSeg++
	p.readIndex = 0
	p.readFrame = *p.frames[p.readSeg]
}

func (p *rpcStream) gotoNextReadByteUnsafe() {
	p.readIndex++
	if p.readIndex == 512 {
		p.gotoNextReadFrameUnsafe()
	}
}

func (p *rpcStream) hasOneByteToRead() bool {
	return p.readIndex < p.writeIndex || p.readSeg < p.writeSeg
}

func (p *rpcStream) hasNBytesToRead(n int) bool {
	return p.GetReadPos()+n <= p.GetWritePos()
}

func (p *rpcStream) isSafetyReadNBytesInCurrentFrame(n int) bool {
	lineEnd := p.readIndex + n
	return lineEnd < 512 && (lineEnd <= p.writeIndex || p.readSeg < p.writeSeg)
}

func (p *rpcStream) isSafetyRead3BytesInCurrentFrame() bool {
	return p.readIndex < 509 &&
		(p.readIndex+2 < p.writeIndex || p.readSeg < p.writeSeg)
}

func (p *rpcStream) isSafetyRead5BytesInCurrentFrame() bool {
	return p.readIndex < 507 &&
		(p.readIndex+4 < p.writeIndex || p.readSeg < p.writeSeg)
}

func (p *rpcStream) isSafetyRead9BytesInCurrentFrame() bool {
	return p.readIndex < 503 &&
		(p.readIndex+8 < p.writeIndex || p.readSeg < p.writeSeg)
}

func (p *rpcStream) putBytes(v []byte) {
	if p.writeIndex+len(v) < 512 {
		p.writeIndex += copy(p.writeFrame[p.writeIndex:], v)
	} else {
		pBytes := (*reflect.SliceHeader)(unsafe.Pointer(&v))
		remains := pBytes.Len
		for remains > 0 {
			canWriteCount := 512 - p.writeIndex
			if canWriteCount > remains {
				canWriteCount = remains
			}
			pBytes.Len = canWriteCount
			remains -= copy(p.writeFrame[p.writeIndex:], v)
			p.writeIndex += canWriteCount
			pBytes.Data += uintptr(canWriteCount)
			if p.writeIndex == 512 {
				p.gotoNextWriteFrame()
			}
		}
	}
}

func (p *rpcStream) putString(v string) {
	if p.writeIndex+len(v) < 512 {
		p.writeIndex += copy(p.writeFrame[p.writeIndex:], v)
	} else {
		sBytes := (*reflect.StringHeader)(unsafe.Pointer(&v))
		remains := sBytes.Len
		for remains > 0 {
			canWriteCount := 512 - p.writeIndex
			if canWriteCount > remains {
				canWriteCount = remains
			}
			sBytes.Len = canWriteCount
			remains -= copy(p.writeFrame[p.writeIndex:], v)
			p.writeIndex += canWriteCount
			sBytes.Data += uintptr(canWriteCount)
			if p.writeIndex == 512 {
				p.gotoNextWriteFrame()
			}
		}
	}
}

func (p *rpcStream) read3BytesCrossFrameUnsafe() []byte {
	v := make([]byte, 3, 3)
	copyBytes := copy(v, p.readFrame[p.readIndex:])
	p.readSeg++
	p.readFrame = *p.frames[p.readSeg]
	p.readIndex = copy(v[copyBytes:], p.readFrame)
	return v
}

func (p *rpcStream) peek5BytesCrossFrameUnsafe() []byte {
	v := make([]byte, 5, 5)
	copyBytes := copy(v, p.readFrame[p.readIndex:])
	copy(v[copyBytes:], *p.frames[p.readSeg+1])
	return v
}

func (p *rpcStream) read5BytesCrossFrameUnsafe() []byte {
	v := make([]byte, 5, 5)
	copyBytes := copy(v, p.readFrame[p.readIndex:])
	p.readSeg++
	p.readFrame = *p.frames[p.readSeg]
	p.readIndex = copy(v[copyBytes:], p.readFrame)
	return v
}

func (p *rpcStream) read9BytesCrossFrameUnsafe() []byte {
	v := make([]byte, 9, 9)
	copyBytes := copy(v, p.readFrame[p.readIndex:])
	p.readSeg++
	p.readFrame = *p.frames[p.readSeg]
	p.readIndex = copy(v[copyBytes:], p.readFrame)
	return v
}

func (p *rpcStream) readNBytesUnsafe(n int) []byte {
	ret := make([]byte, n, n)
	reads := 0
	for reads < n {
		readLen := copy(ret[reads:], p.readFrame[p.readIndex:])
		reads += readLen
		p.readIndex += readLen
		if p.readIndex == 512 {
			p.gotoNextReadFrameUnsafe()
		}
	}
	return ret
}

// return how many bytes to skip
func (p *rpcStream) peekSkip() int {
	skip := readSkipArray[p.readFrame[p.readIndex]]
	if skip > 0 {
		return skip
	}

	if skip == -64 {
		return 0
	}

	if p.isSafetyRead5BytesInCurrentFrame() {
		b := p.readFrame[p.readIndex:]
		return int(uint32(b[1])|
			(uint32(b[2])<<8)|
			(uint32(b[3])<<16)|
			(uint32(b[4])<<24)) - skip
	} else if p.hasNBytesToRead(5) {
		b := p.peek5BytesCrossFrameUnsafe()
		return int(uint32(b[1])|
			(uint32(b[2])<<8)|
			(uint32(b[3])<<16)|
			(uint32(b[4])<<24)) - skip
	} else {
		return 0
	}
}

// return the item read pos, and skip it
// end must be a valid pos
func (p *rpcStream) readSkipItem(end int) int {
	ret := p.GetReadPos()
	skip := p.peekSkip()

	if skip > 0 && ret+skip <= end {
		if p.readIndex+skip < 512 {
			p.readIndex += skip
		} else {
			p.setReadPosUnsafe(ret + skip)
		}
		return ret
	} else {
		return -1
	}
}

func (p *rpcStream) writeStreamUnsafe(s *rpcStream, length int) {
	if p.writeIndex+length < 512 {
		if s.readIndex+length < 512 {
			copy(
				p.writeFrame[p.writeIndex:],
				s.readFrame[s.readIndex:s.readIndex+length],
			)
			p.writeIndex += length
			s.readIndex += length
		} else {
			copyCount := copy(p.writeFrame[p.writeIndex:], s.readFrame[s.readIndex:])
			p.writeIndex += copyCount
			length -= copyCount
			s.gotoNextReadFrameUnsafe()
			copy(p.writeFrame[p.writeIndex:p.writeIndex+length], s.readFrame)
			p.writeIndex += length
			s.readIndex = length
		}
	} else if p.writeIndex < s.readIndex {
		copyCount := copy(p.writeFrame[p.writeIndex:], s.readFrame[s.readIndex:])
		p.writeIndex += copyCount
		length -= copyCount
		s.gotoNextReadFrameUnsafe()

		for length > 511 {
			s.readIndex = copy(p.writeFrame[p.writeIndex:], s.readFrame)
			p.gotoNextWriteFrame()
			p.writeIndex = copy(p.writeFrame, s.readFrame[s.readIndex:])
			s.gotoNextReadFrameUnsafe()
			length -= 512
		}

		if p.writeIndex+length < 512 {
			copy(p.writeFrame[p.writeIndex:p.writeIndex+length], s.readFrame)
			p.writeIndex += length
			s.readIndex = length
		} else {
			copyCount = copy(p.writeFrame[p.writeIndex:], s.readFrame)
			s.readIndex = copyCount
			length -= copyCount
			p.gotoNextWriteFrame()
			copy(p.writeFrame, s.readFrame[s.readIndex:s.readIndex+length])
			p.writeIndex = length
			s.readIndex += length
		}
	} else if p.writeIndex > s.readIndex {
		copyCount := copy(p.writeFrame[p.writeIndex:], s.readFrame[s.readIndex:])
		s.readIndex += copyCount
		length -= copyCount
		p.gotoNextWriteFrame()
		if length < 512-s.readIndex {
			p.writeIndex = copy(
				p.writeFrame,
				s.readFrame[s.readIndex:s.readIndex+length],
			)
			s.readIndex += length
			return
		}
		copyCount = copy(p.writeFrame, s.readFrame[s.readIndex:])
		p.writeIndex = copyCount
		length -= copyCount
		s.gotoNextReadFrameUnsafe()

		for length > 511 {
			s.readIndex = copy(p.writeFrame[p.writeIndex:], s.readFrame)
			p.gotoNextWriteFrame()
			p.writeIndex = copy(p.writeFrame, s.readFrame[s.readIndex:])
			s.gotoNextReadFrameUnsafe()
			length -= 512
		}

		if p.writeIndex+length < 512 {
			copy(p.writeFrame[p.writeIndex:p.writeIndex+length], s.readFrame)
			p.writeIndex += length
			s.readIndex = length
		} else {
			copyCount = copy(p.writeFrame[p.writeIndex:], s.readFrame)
			s.readIndex = copyCount
			length -= copyCount
			p.gotoNextWriteFrame()
			copy(p.writeFrame, s.readFrame[s.readIndex:s.readIndex+length])
			p.writeIndex = length
			s.readIndex += length
		}
	} else {
		length -= copy(p.writeFrame[p.writeIndex:], s.readFrame[s.readIndex:])
		s.gotoNextReadFrameUnsafe()
		p.gotoNextWriteFrame()
		for length > 511 {
			length -= copy(p.writeFrame, s.readFrame)
			s.gotoNextReadFrameUnsafe()
			p.gotoNextWriteFrame()
		}
		copy(p.writeFrame, s.readFrame[:length])
		p.writeIndex = length
		s.readIndex = length
	}
}

func (p *rpcStream) writeStreamNext(s *rpcStream) bool {
	skip := s.peekSkip()
	if skip > 0 {
		if s.isSafetyReadNBytesInCurrentFrame(skip) && p.writeIndex+skip < 512 {
			copy(
				p.writeFrame[p.writeIndex:],
				s.readFrame[s.readIndex:s.readIndex+skip],
			)
			p.writeIndex += skip
			s.readIndex += skip
			return true
		} else if s.hasNBytesToRead(skip) {
			p.writeStreamUnsafe(s, skip)
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

// WriteNil put nil value to stream
func (p *rpcStream) WriteNil() {
	p.writeFrame[p.writeIndex] = 1
	p.writeIndex++
	if p.writeIndex == 512 {
		p.gotoNextWriteFrame()
	}
}

// WriteBool write bool value to stream
func (p *rpcStream) WriteBool(v bool) {
	switch v {
	case true:
		p.writeFrame[p.writeIndex] = 2
	case false:
		p.writeFrame[p.writeIndex] = 3
	}
	p.writeIndex++
	if p.writeIndex == 512 {
		p.gotoNextWriteFrame()
	}
}

// WriteFloat64 write float64 value to stream
func (p *rpcStream) WriteFloat64(value float64) {
	if value == 0 {
		p.writeFrame[p.writeIndex] = 4
		p.writeIndex++
		if p.writeIndex == 512 {
			p.gotoNextWriteFrame()
		}
	} else {
		v := math.Float64bits(value)
		if p.writeIndex < 503 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 5
			b[1] = byte(v)
			b[2] = byte(v >> 8)
			b[3] = byte(v >> 16)
			b[4] = byte(v >> 24)
			b[5] = byte(v >> 32)
			b[6] = byte(v >> 40)
			b[7] = byte(v >> 48)
			b[8] = byte(v >> 56)
			p.writeIndex += 9
		} else {
			p.putBytes([]byte{
				5,
				byte(v),
				byte(v >> 8),
				byte(v >> 16),
				byte(v >> 24),
				byte(v >> 32),
				byte(v >> 40),
				byte(v >> 48),
				byte(v >> 56),
			})
		}
	}
}

// WriteInt64 write int64 value to stream
func (p *rpcStream) WriteInt64(v int64) {
	if v > -8 && v < 33 {
		p.writeFrame[p.writeIndex] = byte(v + 21)
		p.writeIndex++
		if p.writeIndex == 512 {
			p.gotoNextWriteFrame()
		}
		return
	}

	if v >= -32768 && v < 32768 {
		if p.writeIndex < 509 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 6
			b[1] = byte(v)
			b[2] = byte(v >> 8)
			p.writeIndex += 3
			return
		}
		p.putBytes([]byte{
			6,
			byte(v),
			byte(v >> 8),
		})
	} else if v >= -2147483648 && v < 2147483648 {
		if p.writeIndex < 507 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 7
			b[1] = byte(v)
			b[2] = byte(v >> 8)
			b[3] = byte(v >> 16)
			b[4] = byte(v >> 24)
			p.writeIndex += 5
			return
		}
		p.putBytes([]byte{
			7,
			byte(v),
			byte(v >> 8),
			byte(v >> 16),
			byte(v >> 24),
		})
	} else {
		if p.writeIndex < 503 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 8
			b[1] = byte(v)
			b[2] = byte(v >> 8)
			b[3] = byte(v >> 16)
			b[4] = byte(v >> 24)
			b[5] = byte(v >> 32)
			b[6] = byte(v >> 40)
			b[7] = byte(v >> 48)
			b[8] = byte(v >> 56)
			p.writeIndex += 9
			return
		}
		p.putBytes([]byte{
			8,
			byte(v),
			byte(v >> 8),
			byte(v >> 16),
			byte(v >> 24),
			byte(v >> 32),
			byte(v >> 40),
			byte(v >> 48),
			byte(v >> 56),
		})
	}
}

// WriteUint64 write uint64 value to stream
func (p *rpcStream) WriteUint64(v uint64) {
	if v < 10 {
		p.writeFrame[p.writeIndex] = byte(v + 54)
		p.writeIndex++
		if p.writeIndex == 512 {
			p.gotoNextWriteFrame()
		}
	} else if v < 65536 {
		if p.writeIndex < 509 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 9
			b[1] = byte(v)
			b[2] = byte(v >> 8)
			p.writeIndex += 3
			return
		}
		p.putBytes([]byte{
			9,
			byte(v),
			byte(v >> 8),
		})
	} else if v < 4294967296 {
		if p.writeIndex < 507 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 10
			b[1] = byte(v)
			b[2] = byte(v >> 8)
			b[3] = byte(v >> 16)
			b[4] = byte(v >> 24)
			p.writeIndex += 5
			return
		}
		p.putBytes([]byte{
			10,
			byte(v),
			byte(v >> 8),
			byte(v >> 16),
			byte(v >> 24),
		})
	} else {
		if p.writeIndex < 503 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 11
			b[1] = byte(v)
			b[2] = byte(v >> 8)
			b[3] = byte(v >> 16)
			b[4] = byte(v >> 24)
			b[5] = byte(v >> 32)
			b[6] = byte(v >> 40)
			b[7] = byte(v >> 48)
			b[8] = byte(v >> 56)
			p.writeIndex += 9
			return
		}
		p.putBytes([]byte{
			11,
			byte(v),
			byte(v >> 8),
			byte(v >> 16),
			byte(v >> 24),
			byte(v >> 32),
			byte(v >> 40),
			byte(v >> 48),
			byte(v >> 56),
		})
	}
}

// WriteString write string value to stream
func (p *rpcStream) WriteString(v string) {
	length := len(v)
	if length == 0 {
		p.writeFrame[p.writeIndex] = 128
		p.writeIndex++
		if p.writeIndex == 512 {
			p.gotoNextWriteFrame()
		}
	} else if length < 63 {
		if p.writeIndex+length < 510 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = byte(length + 128)
			b[length+1] = 0
			p.writeIndex += copy(b[1:], v) + 2
			return
		}
		// write header
		p.writeFrame[p.writeIndex] = byte(length + 128)
		p.writeIndex++
		if p.writeIndex == 512 {
			p.gotoNextWriteFrame()
		}
		// write body
		p.putString(v)
		// write zero tail
		p.writeFrame[p.writeIndex] = 0
		p.writeIndex++
		if p.writeIndex == 512 {
			p.gotoNextWriteFrame()
		}
	} else {
		if p.writeIndex+length < 506 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 191
			b[1] = byte(uint32(length))
			b[2] = byte(uint32(length) >> 8)
			b[3] = byte(uint32(length) >> 16)
			b[4] = byte(uint32(length) >> 24)
			b[length+5] = 0
			p.writeIndex += copy(b[5:], v) + 6
			return
		} else if p.writeIndex < 507 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 191
			b[1] = byte(uint32(length))
			b[2] = byte(uint32(length) >> 8)
			b[3] = byte(uint32(length) >> 16)
			b[4] = byte(uint32(length) >> 24)
			p.writeIndex += 5
		} else {
			p.putBytes([]byte{
				191,
				byte(uint32(length)),
				byte(uint32(length) >> 8),
				byte(uint32(length) >> 16),
				byte(uint32(length) >> 24),
			})
		}
		// write body
		p.putString(v)
		// write zero tail
		p.writeFrame[p.writeIndex] = 0
		p.writeIndex++
		if p.writeIndex == 512 {
			p.gotoNextWriteFrame()
		}
	}
}

// WriteBytes write []byte value to stream
func (p *rpcStream) WriteBytes(v []byte) {
	length := len(v)
	if length == 0 {
		p.writeFrame[p.writeIndex] = 192
		p.writeIndex++
		if p.writeIndex == 512 {
			p.gotoNextWriteFrame()
		}
	} else if length < 63 {
		if p.writeIndex+length < 511 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = byte(length + 192)
			p.writeIndex += copy(b[1:], v) + 1
			return
		}
		// write header
		p.writeFrame[p.writeIndex] = byte(length + 192)
		p.writeIndex++
		if p.writeIndex == 512 {
			p.gotoNextWriteFrame()
		}
		// write body
		p.putBytes(v)
	} else {
		if p.writeIndex+length < 507 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 255
			b[1] = byte(uint32(length))
			b[2] = byte(uint32(length) >> 8)
			b[3] = byte(uint32(length) >> 16)
			b[4] = byte(uint32(length) >> 24)
			p.writeIndex += copy(b[5:], v) + 5
			return
		} else if p.writeIndex < 507 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 255
			b[1] = byte(uint32(length))
			b[2] = byte(uint32(length) >> 8)
			b[3] = byte(uint32(length) >> 16)
			b[4] = byte(uint32(length) >> 24)
			p.writeIndex += 5
		} else {
			p.putBytes([]byte{
				255,
				byte(uint32(length)),
				byte(uint32(length) >> 8),
				byte(uint32(length) >> 16),
				byte(uint32(length) >> 24),
			})
		}
		// write body
		p.putBytes(v)
	}
}

func (p *rpcStream) WriteArray(v Array) RPCStreamWriteErrorCode {
	length := len(v)
	if length == 0 {
		p.writeFrame[p.writeIndex] = 64
		p.writeIndex++
		if p.writeIndex == 512 {
			p.gotoNextWriteFrame()
		}
		return RPCStreamWriteOK
	}

	startPos := p.GetWritePos()

	b := p.writeFrame[p.writeIndex:]
	if p.writeIndex < 507 {
		p.writeIndex += 5
	} else {
		b = b[0:1]
		p.SetWritePos(startPos + 5)
	}

	if length < 31 {
		b[0] = byte(64 + length)
	} else {
		b[0] = 95
	}

	if length > 30 {
		if p.writeIndex < 508 {
			l := p.writeFrame[p.writeIndex:]
			l[0] = byte(uint32(length))
			l[1] = byte(uint32(length) >> 8)
			l[2] = byte(uint32(length) >> 16)
			l[3] = byte(uint32(length) >> 24)
			p.writeIndex += 4
		} else {
			p.putBytes([]byte{
				byte(uint32(length)),
				byte(uint32(length) >> 8),
				byte(uint32(length) >> 16),
				byte(uint32(length) >> 24),
			})
		}
	}

	for i := 0; i < length; i++ {
		errCode := p.Write(v[i])
		if errCode != RPCStreamWriteOK {
			p.setWritePosUnsafe(startPos)
			return errCode
		}
	}

	totalLength := uint32(p.GetWritePos() - startPos)
	if len(b) > 1 {
		b[1] = byte(totalLength)
		b[2] = byte(totalLength >> 8)
		b[3] = byte(totalLength >> 16)
		b[4] = byte(totalLength >> 24)
	} else {
		endPos := p.GetWritePos()
		p.setWritePosUnsafe(startPos + 1)
		p.putBytes([]byte{
			byte(totalLength),
			byte(totalLength >> 8),
			byte(totalLength >> 16),
			byte(totalLength >> 24),
		})
		p.setWritePosUnsafe(endPos)
	}

	return RPCStreamWriteOK
}

// WriteRPCArray write RPCArray value to stream
func (p *rpcStream) WriteRPCArray(v rpcArray) RPCStreamWriteErrorCode {
	in, readStream := v.getIS()

	if readStream == nil {
		return RPCStreamWriteRPCArrayIsNotAvailable
	}

	length := len(in.items)
	if length == 0 {
		p.writeFrame[p.writeIndex] = 64
		p.writeIndex++
		if p.writeIndex == 512 {
			p.gotoNextWriteFrame()
		}
		return RPCStreamWriteOK
	}

	startPos := p.GetWritePos()

	b := p.writeFrame[p.writeIndex:]
	if p.writeIndex < 507 {
		p.writeIndex += 5
	} else {
		b = b[0:1]
		p.SetWritePos(startPos + 5)
	}

	if length < 31 {
		b[0] = byte(64 + length)
	} else {
		b[0] = 95
	}

	if length > 30 {
		if p.writeIndex < 508 {
			l := p.writeFrame[p.writeIndex:]
			l[0] = byte(uint32(length))
			l[1] = byte(uint32(length) >> 8)
			l[2] = byte(uint32(length) >> 16)
			l[3] = byte(uint32(length) >> 24)
			p.writeIndex += 4
		} else {
			p.putBytes([]byte{
				byte(uint32(length)),
				byte(uint32(length) >> 8),
				byte(uint32(length) >> 16),
				byte(uint32(length) >> 24),
			})
		}
	}

	for i := 0; i < length; i++ {
		readStream.setReadPosUnsafe(in.items[i])
		if !p.writeStreamNext(readStream) {
			p.setWritePosUnsafe(startPos)
			return RPCStreamWriteRPCArrayError
		}
	}

	totalLength := uint32(p.GetWritePos() - startPos)
	if len(b) > 1 {
		b[1] = byte(totalLength)
		b[2] = byte(totalLength >> 8)
		b[3] = byte(totalLength >> 16)
		b[4] = byte(totalLength >> 24)
	} else {
		endPos := p.GetWritePos()
		p.setWritePosUnsafe(startPos + 1)
		p.putBytes([]byte{
			byte(totalLength),
			byte(totalLength >> 8),
			byte(totalLength >> 16),
			byte(totalLength >> 24),
		})
		p.setWritePosUnsafe(endPos)
	}

	return RPCStreamWriteOK
}

// WriteRPCMap write RPCMap value to stream
func (p *rpcStream) WriteMap(v Map) RPCStreamWriteErrorCode {
	length := len(v)

	if length == 0 {
		p.writeFrame[p.writeIndex] = 96
		p.writeIndex++
		if p.writeIndex == 512 {
			p.gotoNextWriteFrame()
		}
		return RPCStreamWriteOK
	}

	startPos := p.GetWritePos()

	b := p.writeFrame[p.writeIndex:]
	if p.writeIndex < 507 {
		p.writeIndex += 5
	} else {
		b = b[0:1]
		p.SetWritePos(startPos + 5)
	}

	if length < 31 {
		b[0] = byte(96 + length)
	} else {
		b[0] = 127
	}

	if length > 30 {
		if p.writeIndex < 508 {
			l := p.writeFrame[p.writeIndex:]
			l[0] = byte(uint32(length))
			l[1] = byte(uint32(length) >> 8)
			l[2] = byte(uint32(length) >> 16)
			l[3] = byte(uint32(length) >> 24)
			p.writeIndex += 4
		} else {
			p.putBytes([]byte{
				byte(uint32(length)),
				byte(uint32(length) >> 8),
				byte(uint32(length) >> 16),
				byte(uint32(length) >> 24),
			})
		}
	}

	for name, value := range v {
		p.WriteString(name)
		errCode := p.Write(value)
		if errCode != RPCStreamWriteOK {
			p.setWritePosUnsafe(startPos)
			return errCode
		}
	}

	totalLength := uint32(p.GetWritePos() - startPos)
	if len(b) > 1 {
		b[1] = byte(totalLength)
		b[2] = byte(totalLength >> 8)
		b[3] = byte(totalLength >> 16)
		b[4] = byte(totalLength >> 24)
	} else {
		endPos := p.GetWritePos()
		p.setWritePosUnsafe(startPos + 1)
		p.putBytes([]byte{
			byte(totalLength),
			byte(totalLength >> 8),
			byte(totalLength >> 16),
			byte(totalLength >> 24),
		})
		p.setWritePosUnsafe(endPos)
	}

	return RPCStreamWriteOK
}

// WriteRPCMap write RPCMap value to stream
func (p *rpcStream) WriteRPCMap(v rpcMap) RPCStreamWriteErrorCode {
	in, readStream := v.getIS()

	if readStream == nil {
		return RPCStreamWriteRPCMapIsNotAvailable
	}

	length := 0
	if in.largeMap == nil {
		length = len(in.smallMap)
	} else {
		length = len(in.largeMap)
	}

	if length == 0 {
		p.writeFrame[p.writeIndex] = 96
		p.writeIndex++
		if p.writeIndex == 512 {
			p.gotoNextWriteFrame()
		}
		return RPCStreamWriteOK
	}

	startPos := p.GetWritePos()

	b := p.writeFrame[p.writeIndex:]
	if p.writeIndex < 507 {
		p.writeIndex += 5
	} else {
		b = b[0:1]
		p.SetWritePos(startPos + 5)
	}

	if length < 31 {
		b[0] = byte(96 + length)
	} else {
		b[0] = 127
	}

	if length > 30 {
		if p.writeIndex < 508 {
			l := p.writeFrame[p.writeIndex:]
			l[0] = byte(uint32(length))
			l[1] = byte(uint32(length) >> 8)
			l[2] = byte(uint32(length) >> 16)
			l[3] = byte(uint32(length) >> 24)
			p.writeIndex += 4
		} else {
			p.putBytes([]byte{
				byte(uint32(length)),
				byte(uint32(length) >> 8),
				byte(uint32(length) >> 16),
				byte(uint32(length) >> 24),
			})
		}
	}

	if length <= 16 {
		for i := 0; i < length; i++ {
			p.WriteString(in.smallMap[i].key)
			readStream.setReadPosUnsafe(in.smallMap[i].pos)
			if !p.writeStreamNext(readStream) {
				p.setWritePosUnsafe(startPos)
				return RPCStreamWriteRPCMapError
			}
		}
	} else {
		for name, pos := range in.largeMap {
			p.WriteString(name)
			readStream.setReadPosUnsafe(pos)
			if !p.writeStreamNext(readStream) {
				p.setWritePosUnsafe(startPos)
				return RPCStreamWriteRPCMapError
			}
		}
	}

	totalLength := uint32(p.GetWritePos() - startPos)
	if len(b) > 1 {
		b[1] = byte(totalLength)
		b[2] = byte(totalLength >> 8)
		b[3] = byte(totalLength >> 16)
		b[4] = byte(totalLength >> 24)
	} else {
		endPos := p.GetWritePos()
		p.setWritePosUnsafe(startPos + 1)
		p.putBytes([]byte{
			byte(totalLength),
			byte(totalLength >> 8),
			byte(totalLength >> 16),
			byte(totalLength >> 24),
		})
		p.setWritePosUnsafe(endPos)
	}

	return RPCStreamWriteOK
}

// Write write generic value to stream
func (p *rpcStream) Write(v interface{}) RPCStreamWriteErrorCode {
	switch v.(type) {
	case nil:
		p.WriteNil()
		return RPCStreamWriteOK
	case bool:
		p.WriteBool(v.(bool))
		return RPCStreamWriteOK
	case int:
		p.WriteInt64(int64(v.(int)))
		return RPCStreamWriteOK
	case int8:
		p.WriteInt64(int64(v.(int8)))
		return RPCStreamWriteOK
	case int16:
		p.WriteInt64(int64(v.(int16)))
		return RPCStreamWriteOK
	case int32:
		p.WriteInt64(int64(v.(int32)))
		return RPCStreamWriteOK
	case int64:
		p.WriteInt64(v.(int64))
		return RPCStreamWriteOK
	case uint:
		p.WriteUint64(uint64(v.(uint)))
		return RPCStreamWriteOK
	case uint8:
		p.WriteUint64(uint64(v.(uint8)))
		return RPCStreamWriteOK
	case uint16:
		p.WriteUint64(uint64(v.(uint16)))
		return RPCStreamWriteOK
	case uint32:
		p.WriteUint64(uint64(v.(uint32)))
		return RPCStreamWriteOK
	case uint64:
		p.WriteUint64(v.(uint64))
		return RPCStreamWriteOK
	case float32:
		p.WriteFloat64(float64(v.(float32)))
		return RPCStreamWriteOK
	case float64:
		p.WriteFloat64(v.(float64))
		return RPCStreamWriteOK
	case string:
		p.WriteString(v.(string))
		return RPCStreamWriteOK
	case []byte:
		p.WriteBytes(v.([]byte))
		return RPCStreamWriteOK
	case rpcArray:
		return p.WriteRPCArray(v.(rpcArray))
	case rpcMap:
		return p.WriteRPCMap(v.(rpcMap))
	case Array:
		return p.WriteArray(v.(Array))
	case Map:
		return p.WriteMap(v.(Map))
	}

	return RPCStreamWriteUnsupportedType
}

// ReadNil read a nil
func (p *rpcStream) ReadNil() bool {
	if p.hasOneByteToRead() && p.readFrame[p.readIndex] == 1 {
		p.gotoNextReadByteUnsafe()
		return true
	}
	return false
}

// ReadBool read a bool
func (p *rpcStream) ReadBool() (bool, bool) {
	if p.hasOneByteToRead() {
		switch p.readFrame[p.readIndex] {
		case 2:
			p.gotoNextReadByteUnsafe()
			return true, true
		case 3:
			p.gotoNextReadByteUnsafe()
			return false, true
		}
	}
	return false, false
}

// ReadFloat64 read a float64
func (p *rpcStream) ReadFloat64() (float64, bool) {
	v := p.readFrame[p.readIndex]
	if v == 4 {
		if p.hasOneByteToRead() {
			p.gotoNextReadByteUnsafe()
			return 0, true
		}
	} else if v == 5 {
		if p.isSafetyRead9BytesInCurrentFrame() {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 9
			return math.Float64frombits(
				uint64(b[1]) |
					(uint64(b[2]) << 8) |
					(uint64(b[3]) << 16) |
					(uint64(b[4]) << 24) |
					(uint64(b[5]) << 32) |
					(uint64(b[6]) << 40) |
					(uint64(b[7]) << 48) |
					(uint64(b[8]) << 56),
			), true
		}
		if p.hasNBytesToRead(9) {
			b := p.read9BytesCrossFrameUnsafe()
			return math.Float64frombits(
				uint64(b[1]) |
					(uint64(b[2]) << 8) |
					(uint64(b[3]) << 16) |
					(uint64(b[4]) << 24) |
					(uint64(b[5]) << 32) |
					(uint64(b[6]) << 40) |
					(uint64(b[7]) << 48) |
					(uint64(b[8]) << 56),
			), true
		}
	}
	return 0, false
}

// ReadInt64 read a int64
func (p *rpcStream) ReadInt64() (int64, bool) {
	v := p.readFrame[p.readIndex]
	if v > 13 && v < 54 {
		if p.hasOneByteToRead() {
			p.gotoNextReadByteUnsafe()
			return int64(v) - 21, true
		}
	} else if v == 6 {
		if p.isSafetyRead3BytesInCurrentFrame() {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 3
			return int64(int16(b[1]) |
				(int16(b[2]) << 8),
			), true
		}
		if p.hasNBytesToRead(3) {
			b := p.read3BytesCrossFrameUnsafe()
			return int64(int16(b[1]) |
				(int16(b[2]) << 8),
			), true
		}
	} else if v == 7 {
		if p.isSafetyRead5BytesInCurrentFrame() {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 5
			return int64(int32(b[1]) |
				(int32(b[2]) << 8) |
				(int32(b[3]) << 16) |
				(int32(b[4]) << 24),
			), true
		}
		if p.hasNBytesToRead(5) {
			b := p.read5BytesCrossFrameUnsafe()
			return int64(int32(b[1]) |
				(int32(b[2]) << 8) |
				(int32(b[3]) << 16) |
				(int32(b[4]) << 24),
			), true
		}
	} else if v == 8 {
		if p.isSafetyRead9BytesInCurrentFrame() {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 9
			return int64(b[1]) |
				(int64(b[2]) << 8) |
				(int64(b[3]) << 16) |
				(int64(b[4]) << 24) |
				(int64(b[5]) << 32) |
				(int64(b[6]) << 40) |
				(int64(b[7]) << 48) |
				(int64(b[8]) << 56), true
		}
		if p.hasNBytesToRead(9) {
			b := p.read9BytesCrossFrameUnsafe()
			return int64(b[1]) |
				(int64(b[2]) << 8) |
				(int64(b[3]) << 16) |
				(int64(b[4]) << 24) |
				(int64(b[5]) << 32) |
				(int64(b[6]) << 40) |
				(int64(b[7]) << 48) |
				(int64(b[8]) << 56), true
		}
	}
	return 0, false
}

// ReadUint64 read a uint64
func (p *rpcStream) ReadUint64() (uint64, bool) {
	v := p.readFrame[p.readIndex]
	if v > 53 && v < 64 {
		if p.hasOneByteToRead() {
			p.gotoNextReadByteUnsafe()
			return uint64(v) - 54, true
		}
	} else if v == 9 {
		if p.isSafetyRead3BytesInCurrentFrame() {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 3
			return uint64(b[1]) |
				(uint64(b[2]) << 8), true
		}
		if p.hasNBytesToRead(3) {
			b := p.read3BytesCrossFrameUnsafe()
			return uint64(b[1]) |
				(uint64(b[2]) << 8), true
		}
	} else if v == 10 {
		if p.isSafetyRead5BytesInCurrentFrame() {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 5
			return uint64(b[1]) |
				(uint64(b[2]) << 8) |
				(uint64(b[3]) << 16) |
				(uint64(b[4]) << 24), true
		}
		if p.hasNBytesToRead(5) {
			b := p.read5BytesCrossFrameUnsafe()
			return uint64(b[1]) |
				(uint64(b[2]) << 8) |
				(uint64(b[3]) << 16) |
				(uint64(b[4]) << 24), true
		}
	} else if v == 11 {
		if p.isSafetyRead9BytesInCurrentFrame() {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 9
			return uint64(b[1]) |
				(uint64(b[2]) << 8) |
				(uint64(b[3]) << 16) |
				(uint64(b[4]) << 24) |
				(uint64(b[5]) << 32) |
				(uint64(b[6]) << 40) |
				(uint64(b[7]) << 48) |
				(uint64(b[8]) << 56), true
		}
		if p.hasNBytesToRead(9) {
			b := p.read9BytesCrossFrameUnsafe()
			return uint64(b[1]) |
				(uint64(b[2]) << 8) |
				(uint64(b[3]) << 16) |
				(uint64(b[4]) << 24) |
				(uint64(b[5]) << 32) |
				(uint64(b[6]) << 40) |
				(uint64(b[7]) << 48) |
				(uint64(b[8]) << 56), true
		}
	}
	return 0, false
}

// ReadString read a string value
func (p *rpcStream) ReadString() (string, bool) {
	// empty string
	v := p.readFrame[p.readIndex]
	if v == 128 {
		if p.hasOneByteToRead() {
			p.gotoNextReadByteUnsafe()
			return emptyString, true
		}
	} else if v > 128 && v < 191 {
		strLen := int(v - 128)
		if p.isSafetyReadNBytesInCurrentFrame(strLen + 2) {
			if p.readFrame[p.readIndex+strLen+1] == 0 {
				b := p.readFrame[p.readIndex+1 : p.readIndex+strLen+1]
				p.readIndex += strLen + 2
				return string(b), true
			}
		} else if p.hasNBytesToRead(strLen + 2) {
			readStart := p.GetReadPos()
			strBuffer := make([]byte, strLen, strLen)
			copyBytes := copy(strBuffer, p.readFrame[p.readIndex+1:])
			p.readIndex += copyBytes + 1
			if p.readIndex == 512 {
				p.readSeg++
				p.readFrame = *p.frames[p.readSeg]
				p.readIndex = copy(strBuffer[copyBytes:], p.readFrame)
			}
			if p.readFrame[p.readIndex] == 0 {
				p.gotoNextReadByteUnsafe()
				return string(strBuffer), true
			}
			p.setReadPosUnsafe(readStart)
		}
	} else if v == 191 {
		readStart := p.GetReadPos()
		strLen := -1

		if p.isSafetyRead5BytesInCurrentFrame() {
			b := p.readFrame[p.readIndex:]
			strLen = int(uint32(b[1]) |
				(uint32(b[2]) << 8) |
				(uint32(b[3]) << 16) |
				(uint32(b[4]) << 24))
			p.readIndex += 5
		} else if p.hasNBytesToRead(5) {
			b := p.read5BytesCrossFrameUnsafe()
			strLen = int(uint32(b[1]) |
				(uint32(b[2]) << 8) |
				(uint32(b[3]) << 16) |
				(uint32(b[4]) << 24))
		}

		if strLen > 62 {
			if p.isSafetyReadNBytesInCurrentFrame(strLen + 1) {
				if p.readFrame[p.readIndex+strLen] == 0 {
					b := p.readFrame[p.readIndex : p.readIndex+strLen]
					p.readIndex += strLen + 1
					return string(b), true
				}
			} else if p.hasNBytesToRead(strLen + 1) {
				b := make([]byte, strLen, strLen)
				reads := 0
				for reads < strLen {
					readLen := copy(b[reads:], p.readFrame[p.readIndex:])
					reads += readLen
					p.readIndex += readLen
					if p.readIndex == 512 {
						p.gotoNextReadFrameUnsafe()
					}
				}
				if p.readFrame[p.readIndex] == 0 {
					p.gotoNextReadByteUnsafe()
					return string(b), true
				}
			}
		}
		p.setReadPosUnsafe(readStart)
	}

	return emptyString, false
}

// ReadUnsafeString read a string value unsafe
func (p *rpcStream) ReadUnsafeString() (ret string, ok bool) {
	// empty string
	v := p.readFrame[p.readIndex]
	if v == 128 {
		if p.hasOneByteToRead() {
			p.gotoNextReadByteUnsafe()
			return emptyString, true
		}
	} else if v > 128 && v < 191 {
		strLen := int(v - 128)
		if p.isSafetyReadNBytesInCurrentFrame(strLen + 2) {
			if p.readFrame[p.readIndex+strLen+1] == 0 {
				b := p.readFrame[p.readIndex+1:]
				pString := (*reflect.StringHeader)(unsafe.Pointer(&ret))
				pBytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
				pString.Len = strLen
				pString.Data = pBytes.Data
				p.readIndex += strLen + 2
				return ret, true
			}
		} else if p.hasNBytesToRead(strLen + 2) {
			readStart := p.GetReadPos()
			strBuffer := make([]byte, strLen, strLen)
			copyBytes := copy(strBuffer, p.readFrame[p.readIndex+1:])
			p.readIndex += copyBytes + 1
			if p.readIndex == 512 {
				p.readSeg++
				p.readFrame = *p.frames[p.readSeg]
				p.readIndex = copy(strBuffer[copyBytes:], p.readFrame)
			}
			if p.readFrame[p.readIndex] == 0 {
				p.gotoNextReadByteUnsafe()
				return string(strBuffer), true
			}
			p.setReadPosUnsafe(readStart)
		}
	} else if v == 191 {
		readStart := p.GetReadPos()
		strLen := -1

		if p.isSafetyRead5BytesInCurrentFrame() {
			b := p.readFrame[p.readIndex:]
			strLen = int(uint32(b[1]) |
				(uint32(b[2]) << 8) |
				(uint32(b[3]) << 16) |
				(uint32(b[4]) << 24))
			p.readIndex += 5
		} else if p.hasNBytesToRead(5) {
			b := p.read5BytesCrossFrameUnsafe()
			strLen = int(uint32(b[1]) |
				(uint32(b[2]) << 8) |
				(uint32(b[3]) << 16) |
				(uint32(b[4]) << 24))
		}

		if strLen > 62 {
			if p.isSafetyReadNBytesInCurrentFrame(strLen + 1) {
				if p.readFrame[p.readIndex+strLen] == 0 {
					b := p.readFrame[p.readIndex:]
					pString := (*reflect.StringHeader)(unsafe.Pointer(&ret))
					pBytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
					pString.Len = strLen
					pString.Data = pBytes.Data
					p.readIndex += strLen + 1
					return ret, true
				}
			} else if p.hasNBytesToRead(strLen + 1) {
				b := make([]byte, strLen, strLen)
				reads := 0
				for reads < strLen {
					readLen := copy(b[reads:], p.readFrame[p.readIndex:])
					reads += readLen
					p.readIndex += readLen
					if p.readIndex == 512 {
						p.gotoNextReadFrameUnsafe()
					}
				}
				if p.readFrame[p.readIndex] == 0 {
					p.gotoNextReadByteUnsafe()
					return string(b), true
				}
			}
		}

		p.setReadPosUnsafe(readStart)
	}
	return emptyString, false
}

// ReadRPCBytes read a rpcBytes value
func (p *rpcStream) ReadBytes() ([]byte, bool) {
	// empty bytes
	v := p.readFrame[p.readIndex]
	if v == 192 {
		if p.hasOneByteToRead() {
			p.gotoNextReadByteUnsafe()
			return emptyBytes, true
		}
	} else if v > 192 && v < 255 {
		bytesLen := int(v - 192)
		ret := make([]byte, bytesLen, bytesLen)
		if p.isSafetyReadNBytesInCurrentFrame(bytesLen + 1) {
			copy(ret, p.readFrame[p.readIndex+1:])
			p.readIndex += bytesLen + 1
			return ret, true
		} else if p.hasNBytesToRead(bytesLen + 1) {
			copyBytes := copy(ret, p.readFrame[p.readIndex+1:])
			p.readIndex += copyBytes + 1
			if p.readIndex == 512 {
				p.readSeg++
				p.readFrame = *p.frames[p.readSeg]
				p.readIndex = copy(ret[copyBytes:], p.readFrame)
			}
			return ret, true
		}
	} else if v == 255 {
		readStart := p.GetReadPos()
		bytesLen := -1
		if p.isSafetyRead5BytesInCurrentFrame() {
			b := p.readFrame[p.readIndex:]
			bytesLen = int(uint32(b[1]) |
				(uint32(b[2]) << 8) |
				(uint32(b[3]) << 16) |
				(uint32(b[4]) << 24))
			p.readIndex += 5
		} else if p.hasNBytesToRead(5) {
			b := p.read5BytesCrossFrameUnsafe()
			bytesLen = int(uint32(b[1]) |
				(uint32(b[2]) << 8) |
				(uint32(b[3]) << 16) |
				(uint32(b[4]) << 24))
		}

		if bytesLen > 62 {
			if p.isSafetyReadNBytesInCurrentFrame(bytesLen) {
				ret := make([]byte, bytesLen, bytesLen)
				copy(ret, p.readFrame[p.readIndex:])
				p.readIndex += bytesLen
				return ret, true
			} else if p.hasNBytesToRead(bytesLen) {
				ret := make([]byte, bytesLen, bytesLen)
				reads := 0
				for reads < bytesLen {
					readLen := copy(ret[reads:], p.readFrame[p.readIndex:])
					reads += readLen
					p.readIndex += readLen
					if p.readIndex == 512 {
						p.gotoNextReadFrameUnsafe()
					}
				}
				return ret, true
			}
		}
		p.setReadPosUnsafe(readStart)
	}
	return emptyBytes, false
}

// ReadUnsafeBytes read a []byte value unsafe
func (p *rpcStream) ReadUnsafeBytes() (ret []byte, ok bool) {
	// empty bytes
	v := p.readFrame[p.readIndex]
	if v == 192 {
		if p.hasOneByteToRead() {
			p.gotoNextReadByteUnsafe()
			return emptyBytes, true
		}
	} else if v > 192 && v < 255 {
		bytesLen := int(v - 192)
		if p.isSafetyReadNBytesInCurrentFrame(bytesLen + 1) {
			p.readIndex += bytesLen + 1
			return p.readFrame[p.readIndex-bytesLen : p.readIndex], true
		} else if p.hasNBytesToRead(bytesLen + 1) {
			ret := make([]byte, bytesLen, bytesLen)
			copyBytes := copy(ret, p.readFrame[p.readIndex+1:])
			p.readIndex += copyBytes + 1
			if p.readIndex == 512 {
				p.readSeg++
				p.readFrame = *p.frames[p.readSeg]
				p.readIndex = copy(ret[copyBytes:], p.readFrame)
			}
			return ret, true
		}
	} else if v == 255 {
		readStart := p.GetReadPos()
		bytesLen := -1

		if p.isSafetyRead5BytesInCurrentFrame() {
			b := p.readFrame[p.readIndex:]
			bytesLen = int(uint32(b[1]) |
				(uint32(b[2]) << 8) |
				(uint32(b[3]) << 16) |
				(uint32(b[4]) << 24))
			p.readIndex += 5
		} else if p.hasNBytesToRead(5) {
			b := p.read5BytesCrossFrameUnsafe()
			bytesLen = int(uint32(b[1]) |
				(uint32(b[2]) << 8) |
				(uint32(b[3]) << 16) |
				(uint32(b[4]) << 24))
		}

		if bytesLen > 62 {
			if p.isSafetyReadNBytesInCurrentFrame(bytesLen) {
				p.readIndex += bytesLen
				return p.readFrame[p.readIndex-bytesLen : p.readIndex], true
			} else if p.hasNBytesToRead(bytesLen) {
				return p.readNBytesUnsafe(bytesLen), true
			}
		}
		p.setReadPosUnsafe(readStart)
	}
	return emptyBytes, false
}

func (p *rpcStream) ReadArray() (Array, bool) {
	v := p.readFrame[p.readIndex]
	if v >= 64 && v < 96 {
		arrLen := 0
		totalLen := 0
		readStart := p.GetReadPos()

		if v == 64 {
			if p.hasOneByteToRead() {
				p.gotoNextReadByteUnsafe()
				return Array{}, true
			}
		} else if v < 95 {
			arrLen = int(v - 64)
			if p.isSafetyRead5BytesInCurrentFrame() {
				b := p.readFrame[p.readIndex:]
				totalLen = int(uint32(b[1]) |
					(uint32(b[2]) << 8) |
					(uint32(b[3]) << 16) |
					(uint32(b[4]) << 24))
				p.readIndex += 5
			} else if p.hasNBytesToRead(5) {
				bytes4 := p.read5BytesCrossFrameUnsafe()
				totalLen = int(uint32(bytes4[1]) |
					(uint32(bytes4[2]) << 8) |
					(uint32(bytes4[3]) << 16) |
					(uint32(bytes4[4]) << 24))
			}
		} else {
			if p.isSafetyRead9BytesInCurrentFrame() {
				b := p.readFrame[p.readIndex:]
				totalLen = int(uint32(b[1]) |
					(uint32(b[2]) << 8) |
					(uint32(b[3]) << 16) |
					(uint32(b[4]) << 24))
				arrLen = int(uint32(b[5]) |
					(uint32(b[6]) << 8) |
					(uint32(b[7]) << 16) |
					(uint32(b[8]) << 24))
				p.readIndex += 9
			} else if p.hasNBytesToRead(9) {
				bytes8 := p.read9BytesCrossFrameUnsafe()
				totalLen = int(uint32(bytes8[1]) |
					(uint32(bytes8[2]) << 8) |
					(uint32(bytes8[3]) << 16) |
					(uint32(bytes8[4]) << 24))
				arrLen = int(uint32(bytes8[5]) |
					(uint32(bytes8[6]) << 8) |
					(uint32(bytes8[7]) << 16) |
					(uint32(bytes8[8]) << 24))
			}
		}

		if arrLen > 0 && totalLen > 4 {
			ret := make(Array, arrLen, arrLen)
			for i := 0; i < arrLen; i++ {
				if rv, ok := p.Read(); ok {
					ret[i] = rv
				} else {
					p.setReadPosUnsafe(readStart)
					return nil, false
				}
			}
			if p.GetReadPos() == readStart+totalLen {
				return ret, true
			}
		}
		p.setReadPosUnsafe(readStart)
	}

	return nil, false
}

// ReadRPCArray read a RPCArray value
func (p *rpcStream) ReadRPCArray(ctx *rpcContext) (rpcArray, bool) {
	v := p.readFrame[p.readIndex]
	if cs := ctx.getCacheStream(); cs != nil && v >= 64 && v < 96 {
		ret := newRPCArray(ctx)
		in := ret.in
		arrLen := 0
		totalLen := 0
		readStart := p.GetReadPos()

		if p != cs {
			wPos := cs.GetWritePos()
			if !cs.writeStreamNext(p) {
				return nilRPCArray, false
			}
			cs.SetReadPos(wPos)
		}

		start := cs.GetReadPos()

		if v == 64 {
			if cs.hasOneByteToRead() {
				cs.gotoNextReadByteUnsafe()
				return ret, true
			}
		} else if v < 95 {
			arrLen = int(v - 64)
			if cs.isSafetyRead5BytesInCurrentFrame() {
				b := cs.readFrame[cs.readIndex:]
				totalLen = int(uint32(b[1]) |
					(uint32(b[2]) << 8) |
					(uint32(b[3]) << 16) |
					(uint32(b[4]) << 24))
				cs.readIndex += 5
			} else if cs.hasNBytesToRead(5) {
				bytes4 := cs.read5BytesCrossFrameUnsafe()
				totalLen = int(uint32(bytes4[1]) |
					(uint32(bytes4[2]) << 8) |
					(uint32(bytes4[3]) << 16) |
					(uint32(bytes4[4]) << 24))
			}
		} else {
			if cs.isSafetyRead9BytesInCurrentFrame() {
				b := cs.readFrame[cs.readIndex:]
				totalLen = int(uint32(b[1]) |
					(uint32(b[2]) << 8) |
					(uint32(b[3]) << 16) |
					(uint32(b[4]) << 24))
				arrLen = int(uint32(b[5]) |
					(uint32(b[6]) << 8) |
					(uint32(b[7]) << 16) |
					(uint32(b[8]) << 24))
				cs.readIndex += 9
			} else if cs.hasNBytesToRead(9) {
				bytes8 := cs.read9BytesCrossFrameUnsafe()
				totalLen = int(uint32(bytes8[1]) |
					(uint32(bytes8[2]) << 8) |
					(uint32(bytes8[3]) << 16) |
					(uint32(bytes8[4]) << 24))
				arrLen = int(uint32(bytes8[5]) |
					(uint32(bytes8[6]) << 8) |
					(uint32(bytes8[7]) << 16) |
					(uint32(bytes8[8]) << 24))
			}
		}

		end := start + totalLen
		if arrLen > 0 && totalLen > 4 {
			itemPos := 0
			for i := 0; i < arrLen; i++ {
				skip := readSkipArray[cs.readFrame[cs.readIndex]]
				if skip > 0 && cs.isSafetyReadNBytesInCurrentFrame(skip) {
					itemPos = cs.GetReadPos()
					cs.readIndex += skip
				} else {
					itemPos = cs.readSkipItem(end)
					if itemPos < start || itemPos >= end {
						ret.release()
						p.setReadPosUnsafe(readStart)
						return nilRPCArray, false
					}
				}
				in.items = append(in.items, itemPos)
			}
			if cs.GetReadPos() == end {
				return ret, true
			}
		}

		ret.release()
		p.setReadPosUnsafe(readStart)
	}

	return nilRPCArray, false
}

// ReadRPCMap read a RPCMap value
func (p *rpcStream) ReadMap() (Map, bool) {
	v := p.readFrame[p.readIndex]
	if v >= 96 && v < 128 {
		mapLen := 0
		totalLen := 0
		readStart := p.GetReadPos()

		if v == 96 {
			if p.hasOneByteToRead() {
				p.gotoNextReadByteUnsafe()
				return Map{}, true
			}
		} else if v < 127 {
			mapLen = int(v - 96)
			if p.isSafetyRead5BytesInCurrentFrame() {
				b := p.readFrame[p.readIndex:]
				totalLen = int(uint32(b[1]) |
					(uint32(b[2]) << 8) |
					(uint32(b[3]) << 16) |
					(uint32(b[4]) << 24))
				p.readIndex += 5
			} else if p.hasNBytesToRead(5) {
				bytes4 := p.read5BytesCrossFrameUnsafe()
				totalLen = int(uint32(bytes4[1]) |
					(uint32(bytes4[2]) << 8) |
					(uint32(bytes4[3]) << 16) |
					(uint32(bytes4[4]) << 24))
			}
		} else {
			if p.isSafetyRead9BytesInCurrentFrame() {
				b := p.readFrame[p.readIndex:]
				totalLen = int(uint32(b[1]) |
					(uint32(b[2]) << 8) |
					(uint32(b[3]) << 16) |
					(uint32(b[4]) << 24))
				mapLen = int(uint32(b[5]) |
					(uint32(b[6]) << 8) |
					(uint32(b[7]) << 16) |
					(uint32(b[8]) << 24))
				p.readIndex += 9
			} else if p.hasNBytesToRead(9) {
				bytes8 := p.read9BytesCrossFrameUnsafe()
				totalLen = int(uint32(bytes8[1]) |
					(uint32(bytes8[2]) << 8) |
					(uint32(bytes8[3]) << 16) |
					(uint32(bytes8[4]) << 24))
				mapLen = int(uint32(bytes8[5]) |
					(uint32(bytes8[6]) << 8) |
					(uint32(bytes8[7]) << 16) |
					(uint32(bytes8[8]) << 24))
			}
		}

		end := readStart + totalLen
		if mapLen > 0 && totalLen > 4 {
			ret := Map{}
			for i := 0; i < mapLen; i++ {
				name, ok := p.ReadString()
				if !ok {
					p.setReadPosUnsafe(readStart)
					return nil, false
				}
				if rv, ok := p.Read(); ok {
					ret[name] = rv
				} else {
					p.setReadPosUnsafe(readStart)
					return nil, false
				}
			}
			if p.GetReadPos() == end {
				return ret, true
			}
		}
		p.setReadPosUnsafe(readStart)
		return nil, false
	}

	return nil, false
}

// ReadRPCMap read a RPCMap value
func (p *rpcStream) ReadRPCMap(ctx *rpcContext) (rpcMap, bool) {
	v := p.readFrame[p.readIndex]
	if cs := ctx.getCacheStream(); cs != nil && v >= 96 && v < 128 {
		ret := newRPCMap(ctx)
		in := ret.in
		mapLen := 0
		totalLen := 0
		readStart := p.GetReadPos()

		if p != cs {
			wPos := cs.GetWritePos()
			if !cs.writeStreamNext(p) {
				return nilRPCMap, false
			}
			cs.SetReadPos(wPos)
		}

		start := cs.GetReadPos()

		if v == 96 {
			if cs.hasOneByteToRead() {
				cs.gotoNextReadByteUnsafe()
				return ret, true
			}
		} else if v < 127 {
			mapLen = int(v - 96)
			if cs.isSafetyRead5BytesInCurrentFrame() {
				b := cs.readFrame[cs.readIndex:]
				totalLen = int(uint32(b[1]) |
					(uint32(b[2]) << 8) |
					(uint32(b[3]) << 16) |
					(uint32(b[4]) << 24))
				cs.readIndex += 5
			} else if cs.hasNBytesToRead(5) {
				bytes4 := cs.read5BytesCrossFrameUnsafe()
				totalLen = int(uint32(bytes4[1]) |
					(uint32(bytes4[2]) << 8) |
					(uint32(bytes4[3]) << 16) |
					(uint32(bytes4[4]) << 24))
			}
		} else {
			if cs.isSafetyRead9BytesInCurrentFrame() {
				b := cs.readFrame[cs.readIndex:]
				totalLen = int(uint32(b[1]) |
					(uint32(b[2]) << 8) |
					(uint32(b[3]) << 16) |
					(uint32(b[4]) << 24))
				mapLen = int(uint32(b[5]) |
					(uint32(b[6]) << 8) |
					(uint32(b[7]) << 16) |
					(uint32(b[8]) << 24))
				cs.readIndex += 9
			} else if cs.hasNBytesToRead(9) {
				bytes8 := cs.read9BytesCrossFrameUnsafe()
				totalLen = int(uint32(bytes8[1]) |
					(uint32(bytes8[2]) << 8) |
					(uint32(bytes8[3]) << 16) |
					(uint32(bytes8[4]) << 24))
				mapLen = int(uint32(bytes8[5]) |
					(uint32(bytes8[6]) << 8) |
					(uint32(bytes8[7]) << 16) |
					(uint32(bytes8[8]) << 24))
			}
		}

		end := start + totalLen
		if mapLen > 0 && totalLen > 4 {
			itemPos := 0
			if mapLen <= 16 {
				for i := 0; i < mapLen; i++ {
					key, ok := cs.ReadUnsafeString()
					if !ok || in.hasKey(key) {
						ret.release()
						p.setReadPosUnsafe(readStart)
						return nilRPCMap, false
					}
					skip := readSkipArray[cs.readFrame[cs.readIndex]]
					if skip > 0 && cs.isSafetyReadNBytesInCurrentFrame(skip) {
						itemPos = cs.GetReadPos()
						cs.readIndex += skip
					} else {
						itemPos = cs.readSkipItem(end)
						if itemPos < start || itemPos >= end {
							ret.release()
							p.setReadPosUnsafe(readStart)
							return nilRPCMap, false
						}
					}
					in.smallMap = append(in.smallMap, rpcMapItem{
						key: key,
						pos: itemPos,
					})
				}
			} else {
				in.largeMap = make(map[string]int)
				for i := 0; i < mapLen; i++ {
					name, ok := cs.ReadUnsafeString()
					if !ok {
						ret.release()
						p.setReadPosUnsafe(readStart)
						return nilRPCMap, false
					}
					skip := readSkipArray[cs.readFrame[cs.readIndex]]
					if skip > 0 && cs.isSafetyReadNBytesInCurrentFrame(skip) {
						itemPos = cs.GetReadPos()
						cs.readIndex += skip
					} else {
						itemPos = cs.readSkipItem(end)
						if itemPos < 0 {
							ret.release()
							p.setReadPosUnsafe(readStart)
							return nilRPCMap, false
						}
					}
					in.largeMap[name] = itemPos
				}
			}
			if cs.GetReadPos() == end {
				return ret, true
			}
		}

		ret.release()
		p.setReadPosUnsafe(readStart)
		return nilRPCMap, false
	}

	return nilRPCMap, false
}

// ReadByContext read a generic value
func (p *rpcStream) ReadByContext(ctx *rpcContext) (interface{}, bool) {
	op := p.readFrame[p.readIndex]
	switch op {
	case byte(1):
		return nil, p.ReadNil()
	case byte(2):
		fallthrough
	case byte(3):
		return p.ReadBool()
	case byte(4):
		fallthrough
	case byte(5):
		return p.ReadFloat64()
	case byte(6):
		fallthrough
	case byte(7):
		fallthrough
	case byte(8):
		return p.ReadInt64()
	case byte(9):
		fallthrough
	case byte(10):
		fallthrough
	case byte(11):
		return p.ReadUint64()
	case byte(12):
		return nil, false
	case byte(13):
		return nil, false
	}

	switch op >> 6 {
	case 0:
		if op < 54 {
			return p.ReadInt64()
		}
		return p.ReadUint64()
	case 1:
		if op < 96 {
			return p.ReadRPCArray(ctx)
		}
		return p.ReadRPCMap(ctx)
	case 2:
		return p.ReadString()
	default:
		return p.ReadBytes()
	}
}

// Read read a generic value
func (p *rpcStream) Read() (interface{}, bool) {
	op := p.readFrame[p.readIndex]
	switch op {
	case byte(1):
		return nil, p.ReadNil()
	case byte(2):
		fallthrough
	case byte(3):
		return p.ReadBool()
	case byte(4):
		fallthrough
	case byte(5):
		return p.ReadFloat64()
	case byte(6):
		fallthrough
	case byte(7):
		fallthrough
	case byte(8):
		return p.ReadInt64()
	case byte(9):
		fallthrough
	case byte(10):
		fallthrough
	case byte(11):
		return p.ReadUint64()
	case byte(12):
		return nil, false
	case byte(13):
		return nil, false
	}

	switch op >> 6 {
	case 0:
		if op < 54 {
			return p.ReadInt64()
		}
		return p.ReadUint64()
	case 1:
		if op < 96 {
			return p.ReadArray()
		}
		return p.ReadMap()
	case 2:
		return p.ReadString()
	default:
		return p.ReadBytes()
	}
}
