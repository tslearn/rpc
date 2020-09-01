package internal

import (
	"math"
	"reflect"
	"strconv"
	"sync"
	"unsafe"
)

const (
	streamPosTargetID   = 0
	streamPosSourceID   = 8
	streamPosZone       = 16
	streamPosIPMap      = 18
	streamPosConnID     = 26
	streamPosCallbackID = 34
	streamPosDepth      = 42

	streamPosServerHead = 0
	streamPosClientHead = 34
	streamPosBody       = 44
)

const (
	// StreamWriteOK ...
	StreamWriteOK = int(0)
	// StreamWriteUnsupportedType ...
	StreamWriteUnsupportedType = int(1)
	// StreamWriteOverflow ...
	StreamWriteOverflow = int(2)
)

var (
	zeroHeader  = make([]byte, streamPosBody)
	streamCache = sync.Pool{
		New: func() interface{} {
			ret := Stream{
				frames:     make([]*[]byte, 1, 8),
				readSeg:    0,
				readIndex:  streamPosBody,
				writeSeg:   0,
				writeIndex: streamPosBody,
			}
			zeroFrame := make([]byte, 512)
			ret.frames[0] = &zeroFrame
			ret.readFrame = zeroFrame
			ret.writeFrame = zeroFrame
			ret.header = zeroFrame[:streamPosBody]
			return &ret
		},
	}
	frameCache = &sync.Pool{
		New: func() interface{} {
			ret := make([]byte, 512)
			return &ret
		},
	}
)

func checkArray(v Array, path string, depth int) string {
	for idx, item := range v {
		if reason := checkValue(
			item,
			ConcatString(path, "[", strconv.Itoa(idx), "]"),
			depth-1,
		); reason != "" {
			return reason
		}
	}

	return ""
}

func checkMap(v Map, path string, depth int) string {
	for key, item := range v {
		if reason := checkValue(
			item,
			ConcatString(path, "[\"", key, "\"]"),
			depth-1,
		); reason != "" {
			return reason
		}
	}

	return ""
}

func checkValue(v interface{}, path string, depth int) string {
	if depth <= 0 {
		return ConcatString(path, " is too complicated")
	}

	switch v := v.(type) {
	case nil:
		return ""
	case bool:
		return ""
	case int:
		return ""
	case int8:
		return ""
	case int16:
		return ""
	case int32:
		return ""
	case int64:
		return ""
	case uint:
		return ""
	case uint8:
		return ""
	case uint16:
		return ""
	case uint32:
		return ""
	case uint64:
		return ""
	case float32:
		return ""
	case float64:
		return ""
	case string:
		return ""
	case Bytes:
		return ""
	case Array:
		return checkArray(v, path, depth)
	case Map:
		return checkMap(v, path, depth)
	default:
		return ConcatString(
			path,
			" type (",
			convertTypeToString(reflect.ValueOf(v).Type()),
			") is not supported",
		)
	}
}

// Stream ...
type Stream struct {
	frames []*[]byte

	readSeg   int
	readIndex int
	readFrame []byte

	writeSeg   int
	writeIndex int
	writeFrame []byte

	header []byte
}

// NewStream ...
func NewStream() *Stream {
	return streamCache.Get().(*Stream)
}

// Reset ...
func (p *Stream) Reset() {
	// reset header
	copy(p.header, zeroHeader)

	// reset frames
	if len(p.frames) != 1 {
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
	}

	if p.readSeg != 0 {
		p.readSeg = 0
		p.readFrame = *p.frames[0]
	}
	p.readIndex = streamPosBody

	if p.writeSeg != 0 {
		p.writeSeg = 0
		p.writeFrame = *p.frames[0]
	}
	p.writeIndex = streamPosBody
}

// Release clean the Stream
func (p *Stream) Release() {
	p.Reset()
	streamCache.Put(p)
}

// GetCallbackID ...
func (p *Stream) GetCallbackID() uint64 {
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

// SetCallbackID ...
func (p *Stream) SetCallbackID(v uint64) {
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

// GetSessionID ...
func (p *Stream) GetSessionID() uint64 {
	b := p.header[8:16]
	return uint64(b[0]) |
		uint64(b[1])<<8 |
		uint64(b[2])<<16 |
		uint64(b[3])<<24 |
		uint64(b[4])<<32 |
		uint64(b[5])<<40 |
		uint64(b[6])<<48 |
		uint64(b[7])<<56
}

// SetSessionID ...
func (p *Stream) SetSessionID(v uint64) {
	b := p.header[8:16]
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
}

// GetSequence ...
func (p *Stream) GetSequence() uint64 {
	b := p.header[8:16]
	return uint64(b[0]) |
		uint64(b[1])<<8 |
		uint64(b[2])<<16 |
		uint64(b[3])<<24 |
		uint64(b[4])<<32 |
		uint64(b[5])<<40 |
		uint64(b[6])<<48 |
		uint64(b[7])<<56
}

// SetSequence ...
func (p *Stream) SetSequence(v uint64) {
	b := p.header[8:16]
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
}

// GetMachineID ...
func (p *Stream) GetMachineID() uint64 {
	b := p.header[16:24]
	return uint64(b[0]) |
		uint64(b[1])<<8 |
		uint64(b[2])<<16 |
		uint64(b[3])<<24 |
		uint64(b[4])<<32 |
		uint64(b[5])<<40 |
		uint64(b[6])<<48 |
		uint64(b[7])<<56
}

// SetMachineID ...
func (p *Stream) SetMachineID(v uint64) {
	b := p.header[16:24]
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
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
		copy(ret[i<<9:], *p.frames[i])
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

// GetReadPos get the current read pos of the stream
func (p *Stream) GetReadPos() int {
	return (p.readSeg << 9) | p.readIndex
}

// SetReadPos set the current read pos of the stream
func (p *Stream) SetReadPos(pos int) bool {
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

// SetReadPosToBodyStart ...
func (p *Stream) SetReadPosToBodyStart() bool {
	p.setReadPosUnsafe(streamPosBody)
	return true
}

func (p *Stream) setReadPosUnsafe(pos int) {
	p.readIndex = pos & 0x1FF
	readSeg := pos >> 9
	if p.readSeg != readSeg {
		p.readSeg = readSeg
		p.readFrame = *p.frames[readSeg]
	}
}

// GetWritePos ...
func (p *Stream) GetWritePos() int {
	return (p.writeSeg << 9) | p.writeIndex
}

// SetWritePos ...
func (p *Stream) SetWritePos(length int) {
	numToCreate := length>>9 - len(p.frames) + 1

	for numToCreate > 0 {
		p.frames = append(p.frames, frameCache.Get().(*[]byte))
		numToCreate--
	}
	p.setWritePosUnsafe(length)
}

// SetWritePosToBodyStart ...
func (p *Stream) SetWritePosToBodyStart() {
	p.setWritePosUnsafe(streamPosBody)
}

func (p *Stream) setWritePosUnsafe(pos int) {
	p.writeIndex = pos & 0x1FF
	writeSeg := pos >> 9
	if p.writeSeg != writeSeg {
		p.writeSeg = writeSeg
		p.writeFrame = *p.frames[writeSeg]
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
	p.writeFrame = *p.frames[p.writeSeg]
}

func (p *Stream) gotoNextReadFrameUnsafe() {
	p.readSeg++
	p.readIndex = 0
	p.readFrame = *p.frames[p.readSeg]
}

func (p *Stream) gotoNextReadByteUnsafe() {
	p.readIndex++
	if p.readIndex == 512 {
		p.gotoNextReadFrameUnsafe()
	}
}

func (p *Stream) hasOneByteToRead() bool {
	return p.readIndex < p.writeIndex || p.readSeg < p.writeSeg
}

func (p *Stream) hasNBytesToRead(n int) bool {
	return p.GetReadPos()+n <= p.GetWritePos()
}

func (p *Stream) isSafetyReadNBytesInCurrentFrame(n int) bool {
	lineEnd := p.readIndex + n
	return lineEnd < 512 && (lineEnd <= p.writeIndex || p.readSeg < p.writeSeg)
}

func (p *Stream) isSafetyRead3BytesInCurrentFrame() bool {
	return p.readIndex < 509 &&
		(p.readIndex+2 < p.writeIndex || p.readSeg < p.writeSeg)
}

func (p *Stream) isSafetyRead5BytesInCurrentFrame() bool {
	return p.readIndex < 507 &&
		(p.readIndex+4 < p.writeIndex || p.readSeg < p.writeSeg)
}

func (p *Stream) isSafetyRead9BytesInCurrentFrame() bool {
	return p.readIndex < 503 &&
		(p.readIndex+8 < p.writeIndex || p.readSeg < p.writeSeg)
}

// PutBytes ...
func (p *Stream) PutBytes(v []byte) {
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

// PutString ...
func (p *Stream) PutString(v string) {
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

func (p *Stream) read3BytesCrossFrameUnsafe() []byte {
	v := make([]byte, 3)
	copyBytes := copy(v, p.readFrame[p.readIndex:])
	p.readSeg++
	p.readFrame = *p.frames[p.readSeg]
	p.readIndex = copy(v[copyBytes:], p.readFrame)
	return v
}

func (p *Stream) peek5BytesCrossFrameUnsafe() []byte {
	v := make([]byte, 5)
	copyBytes := copy(v, p.readFrame[p.readIndex:])
	copy(v[copyBytes:], *p.frames[p.readSeg+1])
	return v
}

func (p *Stream) read5BytesCrossFrameUnsafe() []byte {
	v := make([]byte, 5)
	copyBytes := copy(v, p.readFrame[p.readIndex:])
	p.readSeg++
	p.readFrame = *p.frames[p.readSeg]
	p.readIndex = copy(v[copyBytes:], p.readFrame)
	return v
}

func (p *Stream) read9BytesCrossFrameUnsafe() []byte {
	v := make([]byte, 9)
	copyBytes := copy(v, p.readFrame[p.readIndex:])
	p.readSeg++
	p.readFrame = *p.frames[p.readSeg]
	p.readIndex = copy(v[copyBytes:], p.readFrame)
	return v
}

func (p *Stream) readNBytesUnsafe(n int) []byte {
	ret := make([]byte, n)
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

// WriteNil put nil value to stream
func (p *Stream) WriteNil() {
	p.writeFrame[p.writeIndex] = 1
	p.writeIndex++
	if p.writeIndex == 512 {
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
	if p.writeIndex == 512 {
		p.gotoNextWriteFrame()
	}
}

// WriteFloat64 write float64 value to stream
func (p *Stream) WriteFloat64(value float64) {
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
		if p.writeIndex == 512 {
			p.gotoNextWriteFrame()
		}
		return
	}

	if v >= -32768 && v < 32768 {
		v += 32768
		if p.writeIndex < 509 {
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
		if p.writeIndex < 503 {
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
		p.PutBytes([]byte{
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
		p.PutBytes([]byte{
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
		p.PutString(v)
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
			p.PutBytes([]byte{
				191,
				byte(uint32(length)),
				byte(uint32(length) >> 8),
				byte(uint32(length) >> 16),
				byte(uint32(length) >> 24),
			})
		}
		// write body
		p.PutString(v)
		// write zero tail
		p.writeFrame[p.writeIndex] = 0
		p.writeIndex++
		if p.writeIndex == 512 {
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
		p.PutBytes(v)
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
			p.PutBytes([]byte{
				255,
				byte(uint32(length)),
				byte(uint32(length) >> 8),
				byte(uint32(length) >> 16),
				byte(uint32(length) >> 24),
			})
		}
		// write body
		p.PutBytes(v)
	}
}

// WriteArray ...
func (p *Stream) WriteArray(v Array) int {
	return p.writeArray(v, 64)
}

func (p *Stream) writeArray(v Array, depth int) int {
	if v == nil {
		p.WriteNil()
		return StreamWriteOK
	}

	length := len(v)
	if length == 0 {
		p.writeFrame[p.writeIndex] = 64
		p.writeIndex++
		if p.writeIndex == 512 {
			p.gotoNextWriteFrame()
		}
		return StreamWriteOK
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
			p.PutBytes([]byte{
				byte(uint32(length)),
				byte(uint32(length) >> 8),
				byte(uint32(length) >> 16),
				byte(uint32(length) >> 24),
			})
		}
	}

	for i := 0; i < length; i++ {
		if errCode := p.write(v[i], depth-1); errCode != StreamWriteOK {
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
		p.PutBytes([]byte{
			byte(totalLength),
			byte(totalLength >> 8),
			byte(totalLength >> 16),
			byte(totalLength >> 24),
		})
		p.setWritePosUnsafe(endPos)
	}

	return StreamWriteOK
}

// WriteMap write Map value to stream
func (p *Stream) WriteMap(v Map) int {
	return p.writeMap(v, 64)
}

func (p *Stream) writeMap(v Map, depth int) int {
	if v == nil {
		p.WriteNil()
		return StreamWriteOK
	}

	length := len(v)

	if length == 0 {
		p.writeFrame[p.writeIndex] = 96
		p.writeIndex++
		if p.writeIndex == 512 {
			p.gotoNextWriteFrame()
		}
		return StreamWriteOK
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
		if errCode := p.write(value, depth-1); errCode != StreamWriteOK {
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
		p.PutBytes([]byte{
			byte(totalLength),
			byte(totalLength >> 8),
			byte(totalLength >> 16),
			byte(totalLength >> 24),
		})
		p.setWritePosUnsafe(endPos)
	}

	return StreamWriteOK
}

// Write write generic value to stream
func (p *Stream) Write(v interface{}) int {
	return p.write(v, 64)
}

func (p *Stream) write(v interface{}, depth int) int {
	if depth <= 0 {
		return StreamWriteOverflow
	}

	switch v := v.(type) {
	case nil:
		p.WriteNil()
		return StreamWriteOK
	case bool:
		p.WriteBool(v)
		return StreamWriteOK
	case int:
		p.WriteInt64(int64(v))
		return StreamWriteOK
	case int8:
		p.WriteInt64(int64(v))
		return StreamWriteOK
	case int16:
		p.WriteInt64(int64(v))
		return StreamWriteOK
	case int32:
		p.WriteInt64(int64(v))
		return StreamWriteOK
	case int64:
		p.WriteInt64(v)
		return StreamWriteOK
	case uint:
		p.WriteUint64(uint64(v))
		return StreamWriteOK
	case uint8:
		p.WriteUint64(uint64(v))
		return StreamWriteOK
	case uint16:
		p.WriteUint64(uint64(v))
		return StreamWriteOK
	case uint32:
		p.WriteUint64(uint64(v))
		return StreamWriteOK
	case uint64:
		p.WriteUint64(v)
		return StreamWriteOK
	case float32:
		p.WriteFloat64(float64(v))
		return StreamWriteOK
	case float64:
		p.WriteFloat64(v)
		return StreamWriteOK
	case string:
		p.WriteString(v)
		return StreamWriteOK
	case Bytes:
		p.WriteBytes(v)
		return StreamWriteOK
	case Array:
		return p.writeArray(v, depth)
	case Map:
		return p.writeMap(v, depth)
	}

	return StreamWriteUnsupportedType
}

// ReadNil read a nil
func (p *Stream) ReadNil() bool {
	if p.hasOneByteToRead() && p.readFrame[p.readIndex] == 1 {
		p.gotoNextReadByteUnsafe()
		return true
	}
	return false
}

// ReadBool read a bool
func (p *Stream) ReadBool() (bool, bool) {
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
func (p *Stream) ReadFloat64() (float64, bool) {
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
func (p *Stream) ReadInt64() (int64, bool) {
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
			return int64(uint16(b[1])|
				(uint16(b[2])<<8),
			) - 32768, true
		}
		if p.hasNBytesToRead(3) {
			b := p.read3BytesCrossFrameUnsafe()
			return int64(uint16(b[1])|
				(uint16(b[2])<<8),
			) - 32768, true
		}
	} else if v == 7 {
		if p.isSafetyRead5BytesInCurrentFrame() {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 5
			return int64(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24),
			) - 2147483648, true
		}
		if p.hasNBytesToRead(5) {
			b := p.read5BytesCrossFrameUnsafe()
			return int64(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24),
			) - 2147483648, true
		}
	} else if v == 8 {
		if p.isSafetyRead9BytesInCurrentFrame() {
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
			b := p.read9BytesCrossFrameUnsafe()
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
func (p *Stream) ReadString() (string, bool) {
	// empty string
	v := p.readFrame[p.readIndex]
	if v == 128 {
		if p.hasOneByteToRead() {
			p.gotoNextReadByteUnsafe()
			return "", true
		}
	} else if v > 128 && v < 191 {
		strLen := int(v - 128)
		if p.isSafetyReadNBytesInCurrentFrame(strLen + 2) {
			if p.readFrame[p.readIndex+strLen+1] == 0 {
				b := p.readFrame[p.readIndex+1 : p.readIndex+strLen+1]
				if isUTF8Bytes(b) {
					p.readIndex += strLen + 2
					return string(b), true
				}
			}
		} else if p.hasNBytesToRead(strLen + 2) {
			readStart := p.GetReadPos()
			b := make([]byte, strLen)
			copyBytes := copy(b, p.readFrame[p.readIndex+1:])
			p.readIndex += copyBytes + 1
			if p.readIndex == 512 {
				p.readSeg++
				p.readFrame = *p.frames[p.readSeg]
				p.readIndex = copy(b[copyBytes:], p.readFrame)
			}
			if p.readFrame[p.readIndex] == 0 && isUTF8Bytes(b) {
				p.gotoNextReadByteUnsafe()
				return string(b), true
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
					if isUTF8Bytes(b) {
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
					if p.readIndex == 512 {
						p.gotoNextReadFrameUnsafe()
					}
				}
				if p.readFrame[p.readIndex] == 0 && isUTF8Bytes(b) {
					p.gotoNextReadByteUnsafe()
					return string(b), true
				}
			}
		}
		p.setReadPosUnsafe(readStart)
	}

	return "", false
}

// ReadUnsafeString read a string value unsafe
func (p *Stream) ReadUnsafeString() (ret string, ok bool) {
	// empty string
	v := p.readFrame[p.readIndex]
	if v == 128 {
		if p.hasOneByteToRead() {
			p.gotoNextReadByteUnsafe()
			return "", true
		}
	} else if v > 128 && v < 191 {
		strLen := int(v - 128)
		if p.isSafetyReadNBytesInCurrentFrame(strLen + 2) {
			if p.readFrame[p.readIndex+strLen+1] == 0 {
				b := p.readFrame[p.readIndex+1 : p.readIndex+strLen+1]
				if isUTF8Bytes(b) {
					pString := (*reflect.StringHeader)(unsafe.Pointer(&ret))
					pBytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
					pString.Len = strLen
					pString.Data = pBytes.Data
					p.readIndex += strLen + 2
					return ret, true
				}
			}
		} else if p.hasNBytesToRead(strLen + 2) {
			readStart := p.GetReadPos()
			b := make([]byte, strLen)
			copyBytes := copy(b, p.readFrame[p.readIndex+1:])
			p.readIndex += copyBytes + 1
			if p.readIndex == 512 {
				p.readSeg++
				p.readFrame = *p.frames[p.readSeg]
				p.readIndex = copy(b[copyBytes:], p.readFrame)
			}
			if p.readFrame[p.readIndex] == 0 && isUTF8Bytes(b) {
				p.gotoNextReadByteUnsafe()
				return string(b), true
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
					if isUTF8Bytes(b) {
						pString := (*reflect.StringHeader)(unsafe.Pointer(&ret))
						pBytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
						pString.Len = strLen
						pString.Data = pBytes.Data
						p.readIndex += strLen + 1
						return ret, true
					}
				}
			} else if p.hasNBytesToRead(strLen + 1) {
				b := make([]byte, strLen)
				reads := 0
				for reads < strLen {
					readLen := copy(b[reads:], p.readFrame[p.readIndex:])
					reads += readLen
					p.readIndex += readLen
					if p.readIndex == 512 {
						p.gotoNextReadFrameUnsafe()
					}
				}
				if p.readFrame[p.readIndex] == 0 && isUTF8Bytes(b) {
					p.gotoNextReadByteUnsafe()
					return string(b), true
				}
			}
		}

		p.setReadPosUnsafe(readStart)
	}
	return "", false
}

// ReadBytes read a Bytes value
func (p *Stream) ReadBytes() (Bytes, bool) {
	// empty bytes
	v := p.readFrame[p.readIndex]

	if v == 1 {
		if p.hasOneByteToRead() {
			p.gotoNextReadByteUnsafe()
			return Bytes(nil), true
		}
	} else if v == 192 {
		if p.hasOneByteToRead() {
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
					if p.readIndex == 512 {
						p.gotoNextReadFrameUnsafe()
					}
				}
				return ret, true
			}
		}
		p.setReadPosUnsafe(readStart)
	}
	return Bytes(nil), false
}

// ReadUnsafeBytes read a Bytes value unsafe
func (p *Stream) ReadUnsafeBytes() (ret Bytes, ok bool) {
	// empty bytes
	v := p.readFrame[p.readIndex]
	if v == 1 {
		if p.hasOneByteToRead() {
			p.gotoNextReadByteUnsafe()
			return Bytes(nil), true
		}
	} else if v == 192 {
		if p.hasOneByteToRead() {
			p.gotoNextReadByteUnsafe()
			return Bytes{}, true
		}
	} else if v > 192 && v < 255 {
		bytesLen := int(v - 192)
		if p.isSafetyReadNBytesInCurrentFrame(bytesLen + 1) {
			p.readIndex += bytesLen + 1
			return p.readFrame[p.readIndex-bytesLen : p.readIndex], true
		} else if p.hasNBytesToRead(bytesLen + 1) {
			ret := make(Bytes, bytesLen)
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
	return Bytes(nil), false
}

// ReadArray ...
func (p *Stream) ReadArray() (Array, bool) {
	v := p.readFrame[p.readIndex]
	if v == 1 {
		if p.hasOneByteToRead() {
			p.gotoNextReadByteUnsafe()
			return Array(nil), true
		}
	} else if v >= 64 && v < 96 {
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
			ret := make(Array, arrLen)
			for i := 0; i < arrLen; i++ {
				if rv, ok := p.Read(); ok {
					ret[i] = rv
				} else {
					p.setReadPosUnsafe(readStart)
					return Array(nil), false
				}
			}
			if p.GetReadPos() == readStart+totalLen {
				return ret, true
			}
		}
		p.setReadPosUnsafe(readStart)
	}

	return Array(nil), false
}

// ReadMap read a Map value
func (p *Stream) ReadMap() (Map, bool) {
	v := p.readFrame[p.readIndex]
	if v == 1 {
		if p.hasOneByteToRead() {
			p.gotoNextReadByteUnsafe()
			return Map(nil), true
		}
	} else if v >= 96 && v < 128 {
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
					return Map(nil), false
				}
				if rv, ok := p.Read(); ok {
					ret[name] = rv
				} else {
					p.setReadPosUnsafe(readStart)
					return Map(nil), false
				}
			}
			if p.GetReadPos() == end {
				return ret, true
			}
		}
		p.setReadPosUnsafe(readStart)
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
