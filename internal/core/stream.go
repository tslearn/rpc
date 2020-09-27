package core

import (
	"encoding/binary"
	"github.com/rpccloud/rpc/internal/base"
	"math"
	"strconv"
)

const (
	streamBlockSize = 512

	streamPosTargetID   = 0
	streamPosSourceID   = 8
	streamPosZoneID     = 16
	streamPosIPMap      = 18
	streamPosSessionID  = 26
	streamPosCallbackID = 34
	streamPosDepth      = 42

	streamPosServerHead = 0
	streamPosClientHead = 34
	streamPosBody       = 44

	// StreamWriteOK ...
	StreamWriteOK = ""
)

var (
	initFrame   = make([]byte, streamBlockSize)
	streamCache = &base.SyncPool{
		New: func() interface{} {
			ret := &Stream{
				readSeg:    0,
				readIndex:  streamPosBody,
				writeSeg:   0,
				writeIndex: streamPosBody,
			}

			zeroFrame := ret.bufferBytes[0:]
			ret.frames = ret.bufferFrames[0:1]
			ret.frames[0] = &zeroFrame
			ret.readFrame = zeroFrame
			ret.writeFrame = zeroFrame
			ret.header = zeroFrame[:streamPosBody]
			return ret
		},
	}
	frameCache = &base.SyncPool{
		New: func() interface{} {
			ret := make([]byte, streamBlockSize)
			return &ret
		},
	}
	readSkipArray = [256]int{
		-0x01, +0x01, +0x01, +0x01, +0x01, +0x09, +0x03, +0x05,
		+0x09, +0x03, +0x05, +0x09, -0x01, -0x01, +0x01, +0x01,
		+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
		+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
		+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
		+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
		+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
		+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
		+0x01, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
		+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
		+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
		+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
		+0x01, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
		+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
		+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
		+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
		+0x01, +0x03, +0x04, +0x05, +0x06, +0x07, +0x08, +0x09,
		+0x0a, +0x0b, +0x0c, +0x0d, +0x0e, +0x0f, +0x10, +0x11,
		+0x12, +0x13, +0x14, +0x15, +0x16, +0x17, +0x18, +0x19,
		+0x1a, +0x1b, +0x1c, +0x1d, +0x1e, +0x1f, +0x20, +0x21,
		+0x22, +0x23, +0x24, +0x25, +0x26, +0x27, +0x28, +0x29,
		+0x2a, +0x2b, +0x2c, +0x2d, +0x2e, +0x2f, +0x30, +0x31,
		+0x32, +0x33, +0x34, +0x35, +0x36, +0x37, +0x38, +0x39,
		+0x3a, +0x3b, +0x3c, +0x3d, +0x3e, +0x3f, +0x40, +0x00,
		+0x01, +0x02, +0x03, +0x04, +0x05, +0x06, +0x07, +0x08,
		+0x09, +0x0a, +0x0b, +0x0c, +0x0d, +0x0e, +0x0f, +0x10,
		+0x11, +0x12, +0x13, +0x14, +0x15, +0x16, +0x17, +0x18,
		+0x19, +0x1a, +0x1b, +0x1c, +0x1d, +0x1e, +0x1f, +0x20,
		+0x21, +0x22, +0x23, +0x24, +0x25, +0x26, +0x27, +0x28,
		+0x29, +0x2a, +0x2b, +0x2c, +0x2d, +0x2e, +0x2f, +0x30,
		+0x31, +0x32, +0x33, +0x34, +0x35, +0x36, +0x37, +0x38,
		+0x39, +0x3a, +0x3b, +0x3c, +0x3d, +0x3e, +0x3f, +0x00,
	}
)

// Stream ...
type Stream struct {
	frames []*[]byte

	readSeg   int
	readIndex int
	readFrame []byte

	writeSeg   int
	writeIndex int
	writeFrame []byte

	header       []byte
	bufferBytes  [streamBlockSize]byte
	bufferFrames [4]*[]byte
}

// NewStream ...
func NewStream() *Stream {
	return streamCache.Get().(*Stream)
}

// Reset ...
// Reset must initialize all frames to prevent cross-stream attack
func (p *Stream) Reset() {
	copy(*(p.frames[0]), initFrame)

	// reset frames
	if len(p.frames) != 1 {
		for i := 1; i < len(p.frames); i++ {
			copy(*(p.frames[i]), initFrame)
			frameCache.Put(p.frames[i])
			p.frames[i] = nil
		}

		p.frames = p.bufferFrames[0:1]

		if p.readSeg != 0 {
			p.readSeg = 0
			p.readFrame = *(p.frames[0])
		}

		if p.writeSeg != 0 {
			p.writeSeg = 0
			p.writeFrame = *(p.frames[0])
		}
	}

	p.readIndex = streamPosBody
	p.writeIndex = streamPosBody
}

// Release clean the Stream
func (p *Stream) Release() {
	p.Reset()
	streamCache.Put(p)
}

// GetTargetID ...
func (p *Stream) GetTargetID() uint64 {
	return binary.LittleEndian.Uint64(p.header[streamPosTargetID:])
}

// SetTargetID ...
func (p *Stream) SetTargetID(v uint64) {
	binary.LittleEndian.PutUint64(p.header[streamPosTargetID:], v)
}

// GetSourceID ...
func (p *Stream) GetSourceID() uint64 {
	return binary.LittleEndian.Uint64(p.header[streamPosSourceID:])
}

// SetSourceID ...
func (p *Stream) SetSourceID(v uint64) {
	binary.LittleEndian.PutUint64(p.header[streamPosSourceID:], v)
}

// GetZoneID ...
func (p *Stream) GetZoneID() uint16 {
	return binary.LittleEndian.Uint16(p.header[streamPosZoneID:])
}

// SetZoneID ...
func (p *Stream) SetZoneID(v uint16) {
	binary.LittleEndian.PutUint16(p.header[streamPosZoneID:], v)
}

// GetIPMap ...
func (p *Stream) GetIPMap() uint64 {
	return binary.LittleEndian.Uint64(p.header[streamPosIPMap:])
}

// SetIPMap ...
func (p *Stream) SetIPMap(v uint64) {
	binary.LittleEndian.PutUint64(p.header[streamPosIPMap:], v)
}

// GetSessionID ...
func (p *Stream) GetSessionID() uint64 {
	return binary.LittleEndian.Uint64(p.header[streamPosSessionID:])
}

// SetSessionID ...
func (p *Stream) SetSessionID(v uint64) {
	binary.LittleEndian.PutUint64(p.header[streamPosSessionID:], v)
}

// GetCallbackID ...
func (p *Stream) GetCallbackID() uint64 {
	return binary.LittleEndian.Uint64(p.header[streamPosCallbackID:])
}

// SetCallbackID ...
func (p *Stream) SetCallbackID(v uint64) {
	binary.LittleEndian.PutUint64(p.header[streamPosCallbackID:], v)
}

// GetDepth ...
func (p *Stream) GetDepth() uint16 {
	return binary.LittleEndian.Uint16(p.header[streamPosDepth:])
}

// SetDepth ...
func (p *Stream) SetDepth(v uint16) {
	binary.LittleEndian.PutUint16(p.header[streamPosDepth:], v)
}

// GetReadPos get the current read pos of the stream
func (p *Stream) GetReadPos() int {
	return p.readSeg*streamBlockSize + p.readIndex
}

// SetReadPos set the current read pos of the stream
func (p *Stream) SetReadPos(pos int) bool {
	if pos >= streamPosBody {
		readSeg := pos / streamBlockSize
		readIndex := pos % streamBlockSize
		if (readSeg == p.writeSeg && readIndex <= p.writeIndex) ||
			(readSeg < p.writeSeg) {
			p.readIndex = readIndex
			if p.readSeg != readSeg {
				p.readSeg = readSeg
				p.readFrame = *(p.frames[readSeg])
			}
			return true
		}
	}
	return false
}

// SetReadPosToBodyStart ...
func (p *Stream) SetReadPosToBodyStart() {
	p.readIndex = streamPosBody
	if p.readSeg != 0 {
		p.readSeg = 0
		p.readFrame = *(p.frames[0])
	}
}

// GetWritePos ...
func (p *Stream) GetWritePos() int {
	return p.writeSeg*streamBlockSize + p.writeIndex
}

// SetWritePos ...
func (p *Stream) SetWritePos(pos int) bool {
	if pos >= streamPosBody {
		numToCreate := pos/streamBlockSize - len(p.frames) + 1
		for numToCreate > 0 {
			p.frames = append(p.frames, frameCache.Get().(*[]byte))
			numToCreate--
		}

		p.writeIndex = pos % streamBlockSize
		writeSeg := pos / streamBlockSize
		if p.writeSeg != writeSeg {
			p.writeSeg = writeSeg
			p.writeFrame = *(p.frames[writeSeg])
		}
		return true
	}
	return false
}

// SetWritePosToBodyStart ...
func (p *Stream) SetWritePosToBodyStart() {
	p.writeIndex = streamPosBody
	if p.writeSeg != 0 {
		p.writeSeg = 0
		p.writeFrame = *(p.frames[0])
	}
}

// CanRead return true if the stream is not finish
func (p *Stream) CanRead() bool {
	return p.readIndex < p.writeIndex || p.readSeg < p.writeSeg
}

// IsReadFinish return true if the stream is finish
func (p *Stream) IsReadFinish() bool {
	return p.readIndex == p.writeIndex && p.readSeg == p.writeSeg
}

func (p *Stream) gotoNextWriteFrame() {
	p.writeSeg++
	p.writeIndex = 0
	if p.writeSeg == len(p.frames) {
		p.frames = append(p.frames, frameCache.Get().(*[]byte))
	}
	p.writeFrame = *(p.frames[p.writeSeg])
}

func (p *Stream) gotoNextReadFrameUnsafe() {
	p.readSeg++
	p.readIndex = 0
	p.readFrame = *(p.frames[p.readSeg])
}

func (p *Stream) gotoNextReadByteUnsafe() {
	p.readIndex++
	if p.readIndex == streamBlockSize {
		p.gotoNextReadFrameUnsafe()
	}
}

func (p *Stream) hasNBytesToRead(n int) bool {
	return p.GetReadPos()+n <= p.GetWritePos()
}

func (p *Stream) isSafetyReadNBytesInCurrentFrame(n int) bool {
	lineEnd := p.readIndex + n
	return (lineEnd < streamBlockSize) &&
		(lineEnd <= p.writeIndex || p.readSeg < p.writeSeg)
}

func (p *Stream) peekNBytesCrossFrameUnsafe(n int) []byte {
	v := make([]byte, n)
	copyBytes := copy(v, p.readFrame[p.readIndex:])
	copy(v[copyBytes:], *(p.frames[p.readSeg+1]))
	return v
}

func (p *Stream) readNBytesCrossFrameUnsafe(n int) []byte {
	v := make([]byte, n)
	copyBytes := copy(v, p.readFrame[p.readIndex:])
	p.readSeg++
	p.readFrame = *(p.frames[p.readSeg])
	p.readIndex = copy(v[copyBytes:], p.readFrame)
	return v
}

// return how many bytes to skip, 0 means error
func (p *Stream) peekSkip() (int, byte) {
	op := p.readFrame[p.readIndex]
	if skip := readSkipArray[op]; skip > 0 && p.CanRead() {
		return skip, op
	} else if skip < 0 {
		return 0, 0
	} else if p.isSafetyReadNBytesInCurrentFrame(5) {
		b := p.readFrame[p.readIndex:]
		return int(uint32(b[1]) |
			(uint32(b[2]) << 8) |
			(uint32(b[3]) << 16) |
			(uint32(b[4]) << 24)), op
	} else if p.hasNBytesToRead(5) {
		b := p.peekNBytesCrossFrameUnsafe(5)
		return int(uint32(b[1]) |
			(uint32(b[2]) << 8) |
			(uint32(b[3]) << 16) |
			(uint32(b[4]) << 24)), op
	} else {
		return 0, 0
	}
}

// return the item read pos, and skip it
// end must be a valid pos
func (p *Stream) readSkipItem(end int) int {
	ret := p.GetReadPos()
	skip, _ := p.peekSkip()

	if skip > 0 && ret+skip <= end {
		if p.readIndex+skip < streamBlockSize {
			p.readIndex += skip
		} else {
			p.SetReadPos(ret + skip)
		}
		return ret
	} else {
		return -1
	}
}

func (p *Stream) writeStreamNext(s *Stream) bool {
	if skip, _ := s.peekSkip(); skip <= 0 {
		return false
	} else if s.isSafetyReadNBytesInCurrentFrame(skip) &&
		p.writeIndex+skip < streamBlockSize {
		copy(
			p.writeFrame[p.writeIndex:],
			s.readFrame[s.readIndex:s.readIndex+skip],
		)
		p.writeIndex += skip
		s.readIndex += skip
		return true
	} else if s.hasNBytesToRead(skip) {
		for skip > 0 {
			n := copy(
				p.writeFrame[p.writeIndex:],
				s.readFrame[s.readIndex:base.MinInt(s.readIndex+skip, streamBlockSize)],
			)
			skip -= n
			p.writeIndex += n
			if p.writeIndex == streamBlockSize {
				p.gotoNextWriteFrame()
			}
			s.readIndex += n
			if s.readIndex == streamBlockSize {
				s.gotoNextReadFrameUnsafe()
			}
		}
		return true
	} else {
		return false
	}
}

// GetHeader ...
func (p *Stream) GetHeader() []byte {
	return p.header
}

// GetBuffer ...
func (p *Stream) GetBuffer() []byte {
	length := p.GetWritePos()
	ret := make([]byte, length)
	for i := 0; i <= p.writeSeg; i++ {
		copy(ret[i*streamBlockSize:], *(p.frames[i]))
	}
	return ret
}

// GetBufferUnsafe ...
func (p *Stream) GetBufferUnsafe() []byte {
	if p.writeSeg == 0 {
		return p.writeFrame[0:p.writeIndex]
	}
	return p.GetBuffer()
}

// PutBytes ...
func (p *Stream) PutBytes(v []byte) {
	if p.writeIndex+len(v) < streamBlockSize {
		p.writeIndex += copy(p.writeFrame[p.writeIndex:], v)
	} else {
		for len(v) > 0 {
			n := copy(p.writeFrame[p.writeIndex:], v)
			p.writeIndex += n
			if p.writeIndex == streamBlockSize {
				p.gotoNextWriteFrame()
			}
			v = v[n:]
		}
	}
}

// PutBytesTo ...
func (p *Stream) PutBytesTo(v []byte, pos int) bool {
	if pos+len(v) < streamPosBody {
		return false
	} else if pos+len(v) < streamBlockSize {
		p.writeIndex = pos
		if p.writeSeg != 0 {
			p.writeSeg = 0
			p.writeFrame = *(p.frames[0])
		}
		p.writeIndex += copy(p.writeFrame[p.writeIndex:], v)
		return true
	} else {
		p.SetWritePos(pos)
		for len(v) > 0 {
			n := copy(p.writeFrame[p.writeIndex:], v)
			p.writeIndex += n
			if p.writeIndex == streamBlockSize {
				p.gotoNextWriteFrame()
			}
			v = v[n:]
		}
		return true
	}
}

// WriteNil put nil value to stream
func (p *Stream) WriteNil() {
	p.writeFrame[p.writeIndex] = 1
	p.writeIndex++
	if p.writeIndex == streamBlockSize {
		p.gotoNextWriteFrame()
	}
}

// WriteBool write bool value to stream
func (p *Stream) WriteBool(v bool) {
	switch v {
	case true:
		p.writeFrame[p.writeIndex] = 2
	case false:
		p.writeFrame[p.writeIndex] = 3
	}
	p.writeIndex++
	if p.writeIndex == streamBlockSize {
		p.gotoNextWriteFrame()
	}
}

// WriteFloat64 write float64 value to stream
func (p *Stream) WriteFloat64(value float64) {
	if value == 0 {
		p.writeFrame[p.writeIndex] = 4
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
	} else {
		v := math.Float64bits(value)
		if p.writeIndex < streamBlockSize-9 {
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
			p.PutBytes([]byte{
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
func (p *Stream) WriteInt64(v int64) {
	if v > -8 && v < 33 {
		p.writeFrame[p.writeIndex] = byte(v + 21)
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
		return
	}

	if v >= -32768 && v < 32768 {
		v += 32768
		if p.writeIndex < streamBlockSize-3 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 6
			b[1] = byte(v)
			b[2] = byte(v >> 8)
			p.writeIndex += 3
			return
		}
		p.PutBytes([]byte{
			6,
			byte(v),
			byte(v >> 8),
		})
	} else if v >= -2147483648 && v < 2147483648 {
		v += 2147483648
		if p.writeIndex < streamBlockSize-5 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 7
			b[1] = byte(v)
			b[2] = byte(v >> 8)
			b[3] = byte(v >> 16)
			b[4] = byte(v >> 24)
			p.writeIndex += 5
			return
		}
		p.PutBytes([]byte{
			7,
			byte(v),
			byte(v >> 8),
			byte(v >> 16),
			byte(v >> 24),
		})
	} else {
		uv := uint64(0)
		if v < 0 {
			uv = 9223372036854775808 - uint64(-v)
		} else {
			uv = 9223372036854775808 + uint64(v)
		}
		if p.writeIndex < streamBlockSize-9 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 8
			b[1] = byte(uv)
			b[2] = byte(uv >> 8)
			b[3] = byte(uv >> 16)
			b[4] = byte(uv >> 24)
			b[5] = byte(uv >> 32)
			b[6] = byte(uv >> 40)
			b[7] = byte(uv >> 48)
			b[8] = byte(uv >> 56)
			p.writeIndex += 9
			return
		}
		p.PutBytes([]byte{
			8,
			byte(uv),
			byte(uv >> 8),
			byte(uv >> 16),
			byte(uv >> 24),
			byte(uv >> 32),
			byte(uv >> 40),
			byte(uv >> 48),
			byte(uv >> 56),
		})
	}
}

// WriteUint64 write uint64 value to stream
func (p *Stream) WriteUint64(v uint64) {
	if v < 10 {
		p.writeFrame[p.writeIndex] = byte(v + 54)
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
	} else if v < 65536 {
		if p.writeIndex < streamBlockSize-3 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 9
			b[1] = byte(v)
			b[2] = byte(v >> 8)
			p.writeIndex += 3
			return
		}
		p.PutBytes([]byte{
			9,
			byte(v),
			byte(v >> 8),
		})
	} else if v < 4294967296 {
		if p.writeIndex < streamBlockSize-5 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 10
			b[1] = byte(v)
			b[2] = byte(v >> 8)
			b[3] = byte(v >> 16)
			b[4] = byte(v >> 24)
			p.writeIndex += 5
			return
		}
		p.PutBytes([]byte{
			10,
			byte(v),
			byte(v >> 8),
			byte(v >> 16),
			byte(v >> 24),
		})
	} else {
		if p.writeIndex < streamBlockSize-9 {
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
		p.PutBytes([]byte{
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
func (p *Stream) WriteString(v string) {
	length := len(v)
	if length == 0 {
		p.writeFrame[p.writeIndex] = 128
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
	} else if length < 63 {
		if p.writeIndex+length < streamBlockSize-2 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = byte(length + 128)
			b[length+1] = 0
			p.writeIndex += copy(b[1:], v) + 2
			return
		}
		// write header
		p.writeFrame[p.writeIndex] = byte(length + 128)
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
		// write body
		p.PutBytes([]byte(v))
		// write zero tail
		p.writeFrame[p.writeIndex] = 0
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
	} else {
		totalLength := length + 6
		if p.writeIndex+totalLength < streamBlockSize {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 191
			b[1] = byte(uint32(totalLength))
			b[2] = byte(uint32(totalLength) >> 8)
			b[3] = byte(uint32(totalLength) >> 16)
			b[4] = byte(uint32(totalLength) >> 24)
			b[totalLength-1] = 0
			p.writeIndex += copy(b[5:], v) + 6
			return
		} else if p.writeIndex < streamBlockSize-5 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 191
			b[1] = byte(uint32(totalLength))
			b[2] = byte(uint32(totalLength) >> 8)
			b[3] = byte(uint32(totalLength) >> 16)
			b[4] = byte(uint32(totalLength) >> 24)
			p.writeIndex += 5
		} else {
			p.PutBytes([]byte{
				191,
				byte(uint32(totalLength)),
				byte(uint32(totalLength) >> 8),
				byte(uint32(totalLength) >> 16),
				byte(uint32(totalLength) >> 24),
			})
		}
		// write body
		p.PutBytes(base.StringToBytesUnsafe(v))
		// write zero tail
		p.writeFrame[p.writeIndex] = 0
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
	}
}

// WriteBytes write Bytes value to stream
func (p *Stream) WriteBytes(v Bytes) {
	if v == nil {
		p.WriteNil()
		return
	}
	length := len(v)
	if length == 0 {
		p.writeFrame[p.writeIndex] = 192
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
	} else if length < 63 {
		if p.writeIndex+length < streamBlockSize-1 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = byte(length + 192)
			p.writeIndex += copy(b[1:], v) + 1
			return
		}
		// write header
		p.writeFrame[p.writeIndex] = byte(length + 192)
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
		// write body
		p.PutBytes(v)
	} else {
		totalLength := length + 5
		if p.writeIndex+totalLength < streamBlockSize {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 255
			b[1] = byte(uint32(totalLength))
			b[2] = byte(uint32(totalLength) >> 8)
			b[3] = byte(uint32(totalLength) >> 16)
			b[4] = byte(uint32(totalLength) >> 24)
			p.writeIndex += copy(b[5:], v) + 5
			return
		} else if p.writeIndex < streamBlockSize-5 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 255
			b[1] = byte(uint32(totalLength))
			b[2] = byte(uint32(totalLength) >> 8)
			b[3] = byte(uint32(totalLength) >> 16)
			b[4] = byte(uint32(totalLength) >> 24)
			p.writeIndex += 5
		} else {
			p.PutBytes([]byte{
				255,
				byte(uint32(totalLength)),
				byte(uint32(totalLength) >> 8),
				byte(uint32(totalLength) >> 16),
				byte(uint32(totalLength) >> 24),
			})
		}
		// write body
		p.PutBytes(v)
	}
}

// WriteArray ...
func (p *Stream) WriteArray(v Array) string {
	return p.writeArray(v, 64)
}

func (p *Stream) writeArray(v Array, depth int) string {
	if v == nil {
		p.WriteNil()
		return StreamWriteOK
	}

	length := len(v)
	if length == 0 {
		p.writeFrame[p.writeIndex] = 64
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
		return StreamWriteOK
	}

	startPos := p.GetWritePos()

	b := p.writeFrame[p.writeIndex:]
	if p.writeIndex < streamBlockSize-5 {
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
		if p.writeIndex < streamBlockSize-4 {
			l := p.writeFrame[p.writeIndex:]
			l[0] = byte(uint32(length))
			l[1] = byte(uint32(length) >> 8)
			l[2] = byte(uint32(length) >> 16)
			l[3] = byte(uint32(length) >> 24)
			p.writeIndex += 4
		} else {
			p.PutBytes([]byte{
				byte(uint32(length)),
				byte(uint32(length) >> 8),
				byte(uint32(length) >> 16),
				byte(uint32(length) >> 24),
			})
		}
	}

	for i := 0; i < length; i++ {
		if reason := p.write(v[i], depth-1); reason != StreamWriteOK {
			p.SetWritePos(startPos)
			return base.ConcatString("[", strconv.Itoa(i), "]", reason)
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
		p.SetWritePos(startPos + 1)
		p.PutBytes([]byte{
			byte(totalLength),
			byte(totalLength >> 8),
			byte(totalLength >> 16),
			byte(totalLength >> 24),
		})
		p.SetWritePos(endPos)
	}

	return StreamWriteOK
}

// WriteMap write Map value to stream
func (p *Stream) WriteMap(v Map) string {
	return p.writeMap(v, 64)
}

func (p *Stream) writeMap(v Map, depth int) string {
	if v == nil {
		p.WriteNil()
		return StreamWriteOK
	}

	length := len(v)

	if length == 0 {
		p.writeFrame[p.writeIndex] = 96
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
		return StreamWriteOK
	}

	startPos := p.GetWritePos()

	b := p.writeFrame[p.writeIndex:]
	if p.writeIndex < streamBlockSize-5 {
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
		if p.writeIndex < streamBlockSize-4 {
			l := p.writeFrame[p.writeIndex:]
			l[0] = byte(uint32(length))
			l[1] = byte(uint32(length) >> 8)
			l[2] = byte(uint32(length) >> 16)
			l[3] = byte(uint32(length) >> 24)
			p.writeIndex += 4
		} else {
			p.PutBytes([]byte{
				byte(uint32(length)),
				byte(uint32(length) >> 8),
				byte(uint32(length) >> 16),
				byte(uint32(length) >> 24),
			})
		}
	}

	for name, value := range v {
		p.WriteString(name)
		if reason := p.write(value, depth-1); reason != StreamWriteOK {
			p.SetWritePos(startPos)
			return base.ConcatString("[\"", name, "\"]", reason)
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
		p.SetWritePos(startPos + 1)
		p.PutBytes([]byte{
			byte(totalLength),
			byte(totalLength >> 8),
			byte(totalLength >> 16),
			byte(totalLength >> 24),
		})
		p.SetWritePos(endPos)
	}

	return StreamWriteOK
}

func (p *Stream) Write(v interface{}) string {
	return p.write(v, 64)
}

func (p *Stream) write(v interface{}, depth int) string {
	if depth <= 0 {
		return " write overflow"
	}

	switch v.(type) {
	case nil:
		p.WriteNil()
		return StreamWriteOK
	case bool:
		p.WriteBool(v.(bool))
		return StreamWriteOK
	case int:
		p.WriteInt64(int64(v.(int)))
		return StreamWriteOK
	case int8:
		p.WriteInt64(int64(v.(int8)))
		return StreamWriteOK
	case int16:
		p.WriteInt64(int64(v.(int16)))
		return StreamWriteOK
	case int32:
		p.WriteInt64(int64(v.(int32)))
		return StreamWriteOK
	case int64:
		p.WriteInt64(v.(int64))
		return StreamWriteOK
	case uint:
		p.WriteUint64(uint64(v.(uint)))
		return StreamWriteOK
	case uint8:
		p.WriteUint64(uint64(v.(uint8)))
		return StreamWriteOK
	case uint16:
		p.WriteUint64(uint64(v.(uint16)))
		return StreamWriteOK
	case uint32:
		p.WriteUint64(uint64(v.(uint32)))
		return StreamWriteOK
	case uint64:
		p.WriteUint64(v.(uint64))
		return StreamWriteOK
	case float32:
		p.WriteFloat64(float64(v.(float32)))
		return StreamWriteOK
	case float64:
		p.WriteFloat64(v.(float64))
		return StreamWriteOK
	case string:
		p.WriteString(v.(string))
		return StreamWriteOK
	case Bytes:
		p.WriteBytes(v.(Bytes))
		return StreamWriteOK
	case Array:
		return p.writeArray(v.(Array), depth)
	case Map:
		return p.writeMap(v.(Map), depth)
	default:
		return base.ConcatString(
			" type is not supported",
		)
	}
}

// ReadNil read a nil
func (p *Stream) ReadNil() bool {
	if p.CanRead() && p.readFrame[p.readIndex] == 1 {
		p.gotoNextReadByteUnsafe()
		return true
	}
	return false
}

// ReadBool read a bool
func (p *Stream) ReadBool() (bool, bool) {
	if p.CanRead() {
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
func (p *Stream) ReadFloat64() (float64, bool) {
	v := p.readFrame[p.readIndex]
	if v == 4 {
		if p.CanRead() {
			p.gotoNextReadByteUnsafe()
			return 0, true
		}
	} else if v == 5 {
		if p.isSafetyReadNBytesInCurrentFrame(9) {
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
			b := p.readNBytesCrossFrameUnsafe(9)
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
func (p *Stream) ReadInt64() (int64, bool) {
	v := p.readFrame[p.readIndex]
	if v > 13 && v < 54 {
		if p.CanRead() {
			p.gotoNextReadByteUnsafe()
			return int64(v) - 21, true
		}
	} else if v == 6 {
		if p.isSafetyReadNBytesInCurrentFrame(3) {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 3
			return int64(uint16(b[1])|
				(uint16(b[2])<<8),
			) - 32768, true
		}
		if p.hasNBytesToRead(3) {
			b := p.readNBytesCrossFrameUnsafe(3)
			return int64(uint16(b[1])|
				(uint16(b[2])<<8),
			) - 32768, true
		}
	} else if v == 7 {
		if p.isSafetyReadNBytesInCurrentFrame(5) {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 5
			return int64(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24),
			) - 2147483648, true
		}
		if p.hasNBytesToRead(5) {
			b := p.readNBytesCrossFrameUnsafe(5)
			return int64(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24),
			) - 2147483648, true
		}
	} else if v == 8 {
		if p.isSafetyReadNBytesInCurrentFrame(9) {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 9
			return int64(
				uint64(b[1]) |
					(uint64(b[2]) << 8) |
					(uint64(b[3]) << 16) |
					(uint64(b[4]) << 24) |
					(uint64(b[5]) << 32) |
					(uint64(b[6]) << 40) |
					(uint64(b[7]) << 48) |
					(uint64(b[8]) << 56) -
					9223372036854775808), true
		}
		if p.hasNBytesToRead(9) {
			b := p.readNBytesCrossFrameUnsafe(9)
			return int64(
				uint64(b[1]) |
					(uint64(b[2]) << 8) |
					(uint64(b[3]) << 16) |
					(uint64(b[4]) << 24) |
					(uint64(b[5]) << 32) |
					(uint64(b[6]) << 40) |
					(uint64(b[7]) << 48) |
					(uint64(b[8]) << 56) -
					9223372036854775808), true
		}
	}
	return 0, false
}

// ReadUint64 read a uint64
func (p *Stream) ReadUint64() (uint64, bool) {
	v := p.readFrame[p.readIndex]
	if v > 53 && v < 64 {
		if p.CanRead() {
			p.gotoNextReadByteUnsafe()
			return uint64(v) - 54, true
		}
	} else if v == 9 {
		if p.isSafetyReadNBytesInCurrentFrame(3) {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 3
			return uint64(b[1]) |
				(uint64(b[2]) << 8), true
		}
		if p.hasNBytesToRead(3) {
			b := p.readNBytesCrossFrameUnsafe(3)
			return uint64(b[1]) |
				(uint64(b[2]) << 8), true
		}
	} else if v == 10 {
		if p.isSafetyReadNBytesInCurrentFrame(5) {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 5
			return uint64(b[1]) |
				(uint64(b[2]) << 8) |
				(uint64(b[3]) << 16) |
				(uint64(b[4]) << 24), true
		}
		if p.hasNBytesToRead(5) {
			b := p.readNBytesCrossFrameUnsafe(5)
			return uint64(b[1]) |
				(uint64(b[2]) << 8) |
				(uint64(b[3]) << 16) |
				(uint64(b[4]) << 24), true
		}
	} else if v == 11 {
		if p.isSafetyReadNBytesInCurrentFrame(9) {
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
			b := p.readNBytesCrossFrameUnsafe(9)
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
func (p *Stream) ReadString() (string, bool) {
	// empty string
	v := p.readFrame[p.readIndex]
	if v == 128 {
		if p.CanRead() {
			p.gotoNextReadByteUnsafe()
			return "", true
		}
	} else if v > 128 && v < 191 {
		strLen := int(v - 128)
		if p.isSafetyReadNBytesInCurrentFrame(strLen + 2) {
			if p.readFrame[p.readIndex+strLen+1] == 0 {
				b := p.readFrame[p.readIndex+1 : p.readIndex+strLen+1]
				if base.IsUTF8Bytes(b) {
					p.readIndex += strLen + 2
					return string(b), true
				}
			}
		} else if p.hasNBytesToRead(strLen + 2) {
			readStart := p.GetReadPos()
			b := make([]byte, strLen)
			copyBytes := copy(b, p.readFrame[p.readIndex+1:])
			p.readIndex += copyBytes + 1
			if p.readIndex == streamBlockSize {
				p.readSeg++
				p.readFrame = *(p.frames[p.readSeg])
				p.readIndex = copy(b[copyBytes:], p.readFrame)
			}
			if p.readFrame[p.readIndex] == 0 && base.IsUTF8Bytes(b) {
				p.gotoNextReadByteUnsafe()
				return string(b), true
			}
			p.SetReadPos(readStart)
		}
	} else if v == 191 {
		readStart := p.GetReadPos()
		strLen := -1

		if p.isSafetyReadNBytesInCurrentFrame(5) {
			b := p.readFrame[p.readIndex:]
			strLen = int(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24)) - 6
			p.readIndex += 5
		} else if p.hasNBytesToRead(5) {
			b := p.readNBytesCrossFrameUnsafe(5)
			strLen = int(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24)) - 6
		}

		if strLen > 62 {
			if p.isSafetyReadNBytesInCurrentFrame(strLen + 1) {
				if p.readFrame[p.readIndex+strLen] == 0 {
					b := p.readFrame[p.readIndex : p.readIndex+strLen]
					if base.IsUTF8Bytes(b) {
						p.readIndex += strLen + 1
						return string(b), true
					}
				}
			} else if p.hasNBytesToRead(strLen + 1) {
				b := make([]byte, strLen)
				reads := 0
				for reads < strLen {
					readLen := copy(b[reads:], p.readFrame[p.readIndex:])
					reads += readLen
					p.readIndex += readLen
					if p.readIndex == streamBlockSize {
						p.gotoNextReadFrameUnsafe()
					}
				}
				if p.readFrame[p.readIndex] == 0 && base.IsUTF8Bytes(b) {
					p.gotoNextReadByteUnsafe()
					return string(b), true
				}
			}
		}
		p.SetReadPos(readStart)
	}

	return "", false
}

func (p *Stream) readUnsafeString() (ret string, safe bool, ok bool) {
	// empty string
	v := p.readFrame[p.readIndex]
	if v == 128 {
		if p.CanRead() {
			p.gotoNextReadByteUnsafe()
			return "", true, true
		}
	} else if v > 128 && v < 191 {
		strLen := int(v - 128)
		if p.isSafetyReadNBytesInCurrentFrame(strLen + 2) {
			if p.readFrame[p.readIndex+strLen+1] == 0 {
				b := p.readFrame[p.readIndex+1 : p.readIndex+strLen+1]
				if base.IsUTF8Bytes(b) {
					p.readIndex += strLen + 2
					return base.BytesToStringUnsafe(b), false, true
				}
			}
		} else if p.hasNBytesToRead(strLen + 2) {
			readStart := p.GetReadPos()
			b := make([]byte, strLen)
			copyBytes := copy(b, p.readFrame[p.readIndex+1:])
			p.readIndex += copyBytes + 1
			if p.readIndex == streamBlockSize {
				p.readSeg++
				p.readFrame = *(p.frames[p.readSeg])
				p.readIndex = copy(b[copyBytes:], p.readFrame)
			}
			if p.readFrame[p.readIndex] == 0 && base.IsUTF8Bytes(b) {
				p.gotoNextReadByteUnsafe()
				return base.BytesToStringUnsafe(b), true, true
			}
			p.SetReadPos(readStart)
		}
	} else if v == 191 {
		readStart := p.GetReadPos()
		strLen := -1

		if p.isSafetyReadNBytesInCurrentFrame(5) {
			b := p.readFrame[p.readIndex:]
			strLen = int(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24)) - 6
			p.readIndex += 5
		} else if p.hasNBytesToRead(5) {
			b := p.readNBytesCrossFrameUnsafe(5)
			strLen = int(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24)) - 6
		}

		if strLen > 62 {
			if p.isSafetyReadNBytesInCurrentFrame(strLen + 1) {
				if p.readFrame[p.readIndex+strLen] == 0 {
					b := p.readFrame[p.readIndex : p.readIndex+strLen]
					if base.IsUTF8Bytes(b) {
						p.readIndex += strLen + 1
						return base.BytesToStringUnsafe(b), false, true
					}
				}
			} else if p.hasNBytesToRead(strLen + 1) {
				b := make([]byte, strLen)
				reads := 0
				for reads < strLen {
					readLen := copy(b[reads:], p.readFrame[p.readIndex:])
					reads += readLen
					p.readIndex += readLen
					if p.readIndex == streamBlockSize {
						p.gotoNextReadFrameUnsafe()
					}
				}
				if p.readFrame[p.readIndex] == 0 && base.IsUTF8Bytes(b) {
					p.gotoNextReadByteUnsafe()
					return base.BytesToStringUnsafe(b), true, true
				}
			}
		}

		p.SetReadPos(readStart)
	}
	return "", true, false
}

// ReadBytes read a Bytes value
func (p *Stream) ReadBytes() (Bytes, bool) {
	// empty bytes
	v := p.readFrame[p.readIndex]

	if v == 1 {
		if p.CanRead() {
			p.gotoNextReadByteUnsafe()
			return Bytes(nil), true
		}
	} else if v == 192 {
		if p.CanRead() {
			p.gotoNextReadByteUnsafe()
			return Bytes{}, true
		}
	} else if v > 192 && v < 255 {
		bytesLen := int(v - 192)
		ret := make(Bytes, bytesLen)
		if p.isSafetyReadNBytesInCurrentFrame(bytesLen + 1) {
			copy(ret, p.readFrame[p.readIndex+1:])
			p.readIndex += bytesLen + 1
			return ret, true
		} else if p.hasNBytesToRead(bytesLen + 1) {
			copyBytes := copy(ret, p.readFrame[p.readIndex+1:])
			p.readIndex += copyBytes + 1
			if p.readIndex == streamBlockSize {
				p.readSeg++
				p.readFrame = *(p.frames[p.readSeg])
				p.readIndex = copy(ret[copyBytes:], p.readFrame)
			}
			return ret, true
		}
	} else if v == 255 {
		readStart := p.GetReadPos()
		bytesLen := -1
		if p.isSafetyReadNBytesInCurrentFrame(5) {
			b := p.readFrame[p.readIndex:]
			bytesLen = int(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24)) - 5
			p.readIndex += 5
		} else if p.hasNBytesToRead(5) {
			b := p.readNBytesCrossFrameUnsafe(5)
			bytesLen = int(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24)) - 5
		}

		if bytesLen > 62 {
			if p.isSafetyReadNBytesInCurrentFrame(bytesLen) {
				ret := make(Bytes, bytesLen)
				copy(ret, p.readFrame[p.readIndex:])
				p.readIndex += bytesLen
				return ret, true
			} else if p.hasNBytesToRead(bytesLen) {
				ret := make(Bytes, bytesLen)
				reads := 0
				for reads < bytesLen {
					readLen := copy(ret[reads:], p.readFrame[p.readIndex:])
					reads += readLen
					p.readIndex += readLen
					if p.readIndex == streamBlockSize {
						p.gotoNextReadFrameUnsafe()
					}
				}
				return ret, true
			}
		}
		p.SetReadPos(readStart)
	}
	return Bytes(nil), false
}

// ReadArray ...
func (p *Stream) ReadArray() (Array, bool) {
	v := p.readFrame[p.readIndex]
	if v == 1 {
		if p.CanRead() {
			p.gotoNextReadByteUnsafe()
			return Array(nil), true
		}
	} else if v >= 64 && v < 96 {
		arrLen := 0
		totalLen := 0
		readStart := p.GetReadPos()

		if v == 64 {
			if p.CanRead() {
				p.gotoNextReadByteUnsafe()
				return Array{}, true
			}
		} else if v < 95 {
			arrLen = int(v - 64)
			if p.isSafetyReadNBytesInCurrentFrame(5) {
				b := p.readFrame[p.readIndex:]
				totalLen = int(uint32(b[1]) |
					(uint32(b[2]) << 8) |
					(uint32(b[3]) << 16) |
					(uint32(b[4]) << 24))
				p.readIndex += 5
			} else if p.hasNBytesToRead(5) {
				bytes4 := p.readNBytesCrossFrameUnsafe(5)
				totalLen = int(uint32(bytes4[1]) |
					(uint32(bytes4[2]) << 8) |
					(uint32(bytes4[3]) << 16) |
					(uint32(bytes4[4]) << 24))
			}
		} else {
			if p.isSafetyReadNBytesInCurrentFrame(9) {
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
				bytes8 := p.readNBytesCrossFrameUnsafe(9)
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
			ret := make(Array, arrLen)
			for i := 0; i < arrLen; i++ {
				if rv, ok := p.Read(); ok {
					ret[i] = rv
				} else {
					p.SetReadPos(readStart)
					return Array(nil), false
				}
			}
			if p.GetReadPos() == readStart+totalLen {
				return ret, true
			}
		}
		p.SetReadPos(readStart)
	}

	return Array(nil), false
}

// ReadMap read a Map value
func (p *Stream) ReadMap() (Map, bool) {
	v := p.readFrame[p.readIndex]
	if v == 1 {
		if p.CanRead() {
			p.gotoNextReadByteUnsafe()
			return Map(nil), true
		}
	} else if v >= 96 && v < 128 {
		mapLen := 0
		totalLen := 0
		readStart := p.GetReadPos()

		if v == 96 {
			if p.CanRead() {
				p.gotoNextReadByteUnsafe()
				return Map{}, true
			}
		} else if v < 127 {
			mapLen = int(v - 96)
			if p.isSafetyReadNBytesInCurrentFrame(5) {
				b := p.readFrame[p.readIndex:]
				totalLen = int(uint32(b[1]) |
					(uint32(b[2]) << 8) |
					(uint32(b[3]) << 16) |
					(uint32(b[4]) << 24))
				p.readIndex += 5
			} else if p.hasNBytesToRead(5) {
				bytes4 := p.readNBytesCrossFrameUnsafe(5)
				totalLen = int(uint32(bytes4[1]) |
					(uint32(bytes4[2]) << 8) |
					(uint32(bytes4[3]) << 16) |
					(uint32(bytes4[4]) << 24))
			}
		} else {
			if p.isSafetyReadNBytesInCurrentFrame(9) {
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
				bytes8 := p.readNBytesCrossFrameUnsafe(9)
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
					p.SetReadPos(readStart)
					return Map(nil), false
				}
				if rv, ok := p.Read(); ok {
					ret[name] = rv
				} else {
					p.SetReadPos(readStart)
					return Map(nil), false
				}
			}
			if p.GetReadPos() == end {
				return ret, true
			}
		}
		p.SetReadPos(readStart)
	}

	return Map(nil), false
}

// Read read a generic value
func (p *Stream) Read() (Any, bool) {
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

// ReadRTArray read a RPCArray value
func (p *Stream) ReadRTArray(rt Runtime) (RTArray, bool) {
	v := p.readFrame[p.readIndex]
	if v >= 64 && v < 96 {
		thread := rt.lock()
		if thread == nil {
			return RTArray{}, false
		}
		defer rt.unlock()
		cs := thread.rtStream
		arrLen := 0
		totalLen := 0
		readStart := p.GetReadPos()

		if p != cs {
			cs.SetReadPos(cs.GetWritePos())
			if !cs.writeStreamNext(p) {
				return RTArray{}, false
			}
		}

		start := cs.GetReadPos()

		if v == 64 {
			if cs.CanRead() {
				cs.gotoNextReadByteUnsafe()
				return newRTArray(rt, 0), true
			}
		} else if v < 95 {
			arrLen = int(v - 64)
			if cs.isSafetyReadNBytesInCurrentFrame(5) {
				b := cs.readFrame[cs.readIndex:]
				totalLen = int(uint32(b[1]) |
					(uint32(b[2]) << 8) |
					(uint32(b[3]) << 16) |
					(uint32(b[4]) << 24))
				cs.readIndex += 5
			} else if cs.hasNBytesToRead(5) {
				bytes4 := cs.readNBytesCrossFrameUnsafe(5)
				totalLen = int(uint32(bytes4[1]) |
					(uint32(bytes4[2]) << 8) |
					(uint32(bytes4[3]) << 16) |
					(uint32(bytes4[4]) << 24))
			}
		} else {
			if cs.isSafetyReadNBytesInCurrentFrame(9) {
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
				bytes8 := cs.readNBytesCrossFrameUnsafe(9)
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
			ret := newRTArray(rt, arrLen)

			itemPos := 0
			for i := 0; i < arrLen; i++ {
				op := cs.readFrame[cs.readIndex]
				skip := readSkipArray[op]
				if skip > 0 && cs.isSafetyReadNBytesInCurrentFrame(skip) {
					itemPos = cs.GetReadPos()
					if itemPos < start || itemPos+skip > end {
						p.SetReadPos(readStart)
						return RTArray{}, false
					}
					cs.readIndex += skip
				} else {
					itemPos = cs.readSkipItem(end)
					if itemPos < start {
						p.SetReadPos(readStart)
						return RTArray{}, false
					}
				}

				if op > 128 && op < 191 {
					ret.items = append(ret.items, makePosRecord(int64(itemPos), true))
				} else {
					ret.items = append(ret.items, makePosRecord(int64(itemPos), false))
				}
			}
			if cs.GetReadPos() == end {
				return ret, true
			}
		}

		p.SetReadPos(readStart)
	}

	return RTArray{}, false
}

// ReadRPCMap read a RPCMap value
func (p *Stream) ReadRTMap(rt Runtime) (RTMap, bool) {
	v := p.readFrame[p.readIndex]

	if v >= 96 && v < 128 {
		thread := rt.lock()
		if thread == nil {
			return RTMap{}, false
		}
		defer rt.unlock()
		cs := thread.rtStream

		mapLen := 0
		totalLen := 0
		readStart := p.GetReadPos()

		if p != cs {
			cs.SetReadPos(cs.GetWritePos())
			if !cs.writeStreamNext(p) {
				return RTMap{}, false
			}
		}

		start := cs.GetReadPos()

		if v == 96 {
			if cs.CanRead() {
				cs.gotoNextReadByteUnsafe()
				return newRTMap(rt, 0), true
			}
		} else if v < 127 {
			mapLen = int(v - 96)
			if cs.isSafetyReadNBytesInCurrentFrame(5) {
				b := cs.readFrame[cs.readIndex:]
				totalLen = int(uint32(b[1]) |
					(uint32(b[2]) << 8) |
					(uint32(b[3]) << 16) |
					(uint32(b[4]) << 24))
				cs.readIndex += 5
			} else if cs.hasNBytesToRead(5) {
				bytes4 := cs.readNBytesCrossFrameUnsafe(5)
				totalLen = int(uint32(bytes4[1]) |
					(uint32(bytes4[2]) << 8) |
					(uint32(bytes4[3]) << 16) |
					(uint32(bytes4[4]) << 24))
			}
		} else {
			if cs.isSafetyReadNBytesInCurrentFrame(9) {
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
				bytes8 := cs.readNBytesCrossFrameUnsafe(9)
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
			ret := newRTMap(rt, mapLen)

			itemPos := 0

			for i := 0; i < mapLen; i++ {
				key, _, ok := cs.readUnsafeString()
				if !ok {
					p.SetReadPos(readStart)
					return RTMap{}, false
				}
				op := cs.readFrame[cs.readIndex]
				skip := readSkipArray[op]
				if skip > 0 && cs.isSafetyReadNBytesInCurrentFrame(skip) {
					itemPos = cs.GetReadPos()
					if itemPos < start || itemPos+skip > end {
						p.SetReadPos(readStart)
						return RTMap{}, false
					}
					cs.readIndex += skip
				} else {
					itemPos = cs.readSkipItem(end)
					if itemPos < start {
						p.SetReadPos(readStart)
						return RTMap{}, false
					}
				}

				if op>>6 == 2 {
					ret.appendValue(key, makePosRecord(int64(itemPos), true))
				} else {
					ret.appendValue(key, makePosRecord(int64(itemPos), false))
				}
			}
			if cs.GetReadPos() == end {
				return ret, true
			}
		}

		p.SetReadPos(readStart)
	}

	return RTMap{}, false
}

// WriteRTArray write RTArray to stream
func (p *Stream) WriteRTArray(v RTArray) string {
	if thread := v.rt.lock(); thread != nil {
		defer v.rt.unlock()
		readStream := thread.rtStream

		length := v.Size()
		if length == -1 {
			return "WriteRTArray: parameter is not available"
		}

		if length == 0 {
			p.writeFrame[p.writeIndex] = 64
			p.writeIndex++
			if p.writeIndex == streamBlockSize {
				p.gotoNextWriteFrame()
			}
			return ""
		}

		startPos := p.GetWritePos()

		b := p.writeFrame[p.writeIndex:]
		if p.writeIndex < streamBlockSize-5 {
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
			if p.writeIndex < streamBlockSize-4 {
				l := p.writeFrame[p.writeIndex:]
				l[0] = byte(uint32(length))
				l[1] = byte(uint32(length) >> 8)
				l[2] = byte(uint32(length) >> 16)
				l[3] = byte(uint32(length) >> 24)
				p.writeIndex += 4
			} else {
				p.PutBytes([]byte{
					byte(uint32(length)),
					byte(uint32(length) >> 8),
					byte(uint32(length) >> 16),
					byte(uint32(length) >> 24),
				})
			}
		}

		for i := 0; i < length; i++ {
			readStream.SetReadPos(int(v.items[i].getPos()))
			if !p.writeStreamNext(readStream) {
				p.SetWritePos(startPos)
				return "WriteRTArray: parameter is broken"
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
			p.SetWritePos(startPos + 1)
			p.PutBytes([]byte{
				byte(totalLength),
				byte(totalLength >> 8),
				byte(totalLength >> 16),
				byte(totalLength >> 24),
			})
			p.SetWritePos(endPos)
		}

		return ""
	} else {
		return "WriteRTArray: parameter is not available"
	}
}

// WriteRPCMap write RPCMap value to stream
func (p *Stream) WriteRTMap(v RTMap) string {
	if thread := v.rt.lock(); thread != nil {
		defer v.rt.unlock()
		readStream := thread.rtStream

		length := v.Size()
		if length == -1 {
			return "WriteRTMap: parameter is not available"
		}

		if length == 0 {
			p.writeFrame[p.writeIndex] = 96
			p.writeIndex++
			if p.writeIndex == streamBlockSize {
				p.gotoNextWriteFrame()
			}
			return ""
		}

		startPos := p.GetWritePos()

		b := p.writeFrame[p.writeIndex:]
		if p.writeIndex < streamBlockSize-5 {
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
			if p.writeIndex < streamBlockSize-4 {
				l := p.writeFrame[p.writeIndex:]
				l[0] = byte(uint32(length))
				l[1] = byte(uint32(length) >> 8)
				l[2] = byte(uint32(length) >> 16)
				l[3] = byte(uint32(length) >> 24)
				p.writeIndex += 4
			} else {
				p.PutBytes([]byte{
					byte(uint32(length)),
					byte(uint32(length) >> 8),
					byte(uint32(length) >> 16),
					byte(uint32(length) >> 24),
				})
			}
		}

		if v.items != nil {
			for i := 0; i < length; i++ {
				p.WriteString(v.items[i].key)
				readStream.SetReadPos(int(v.items[i].pos.getPos()))
				if !p.writeStreamNext(readStream) {
					p.SetWritePos(startPos)
					return "WriteRTMap: parameter is broken"
				}
			}
		} else if v.largeMap != nil {
			for name, pos := range v.largeMap {
				p.WriteString(name)
				readStream.SetReadPos(int(pos.getPos()))
				if !p.writeStreamNext(readStream) {
					p.SetWritePos(startPos)
					return "WriteRTMap: parameter is broken"
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
			p.SetWritePos(startPos + 1)
			p.PutBytes([]byte{
				byte(totalLength),
				byte(totalLength >> 8),
				byte(totalLength >> 16),
				byte(totalLength >> 24),
			})
			p.SetWritePos(endPos)
		}

		return ""
	} else {
		return "WriteRTMap: parameter is not available"
	}
}

// ReadRTValue read a RTValue value
func (p *Stream) ReadRTValue(rt Runtime) (RTValue, bool) {
	thread := rt.lock()
	if thread == nil {
		return RTValue{}, false
	}
	defer rt.unlock()

	cs := thread.rtStream
	if p != cs {
		cs.SetReadPos(cs.GetWritePos())
		if !cs.writeStreamNext(p) {
			return RTValue{}, false
		}
	}

	if op := cs.readFrame[cs.readIndex]; op > 128 && op < 191 {
		cacheString, _, ok := cs.readUnsafeString()
		return RTValue{
			rt:          rt,
			pos:         int64(cs.GetReadPos()),
			cacheString: cacheString,
			cacheOK:     ok,
		}, false
	} else {
		return RTValue{
			rt:          rt,
			pos:         int64(cs.GetReadPos()),
			cacheString: "",
			cacheOK:     false,
		}, false
	}
}
