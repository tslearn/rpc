package core

import (
	"math/rand"
	"testing"
)

var rpcStreamTestCollections = map[string][][2]interface{}{
	"nil": {
		{nil, []byte{0x01}},
	},
	"bool": {
		{true, []byte{0x02}},
		{false, []byte{0x03}},
	},
	"float64": {
		{float64(0), []byte{0x04}},
		{float64(3.1415926), []byte{
			0x05, 0x4a, 0xd8, 0x12, 0x4d, 0xfb, 0x21, 0x09, 0x40,
		}},
		{float64(-3.1415926), []byte{
			0x05, 0x4a, 0xd8, 0x12, 0x4d, 0xfb, 0x21, 0x09, 0xc0,
		}},
	},
	"int64": {
		{int64(-9223372036854775808), []byte{
			0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80,
		}},
		{int64(-2147483649), []byte{
			0x08, 0xff, 0xff, 0xff, 0x7f, 0xff, 0xff, 0xff, 0xff,
		}},
		{int64(-2147483648), []byte{0x07, 0x00, 0x00, 0x00, 0x80}},
		{int64(-32769), []byte{0x07, 0xff, 0x7f, 0xff, 0xff}},
		{int64(-32768), []byte{0x06, 0x00, 0x80}},
		{int64(-8), []byte{0x06, 0xf8, 0xff}},
		{int64(-7), []byte{0x0e}},
		{int64(-1), []byte{0x14}},
		{int64(0), []byte{0x15}},
		{int64(1), []byte{0x16}},
		{int64(32), []byte{0x35}},
		{int64(33), []byte{0x06, 0x21, 0x00}},
		{int64(32767), []byte{0x06, 0xff, 0x7f}},
		{int64(32768), []byte{0x07, 0x00, 0x80, 0x00, 0x00}},
		{int64(2147483647), []byte{0x07, 0xff, 0xff, 0xff, 0x7f}},
		{int64(2147483648), []byte{
			0x08, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00,
		}},
		{int64(9223372036854775807), []byte{
			0x08, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f,
		}},
	},
	"uint64": {
		{uint64(0), []byte{0x36}},
		{uint64(9), []byte{0x3f}},
		{uint64(10), []byte{0x09, 0x0a, 0x00}},
		{uint64(65535), []byte{0x09, 0xff, 0xff}},
		{uint64(65536), []byte{0x0a, 0x00, 0x00, 0x01, 0x00}},
		{uint64(4294967295), []byte{0x0a, 0xff, 0xff, 0xff, 0xff}},
		{uint64(4294967296), []byte{
			0x0b, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
		}},
		{uint64(18446744073709551615), []byte{
			0x0b, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		}},
	},
	"string": {
		{"", []byte{0x80}},
		{"a", []byte{0x81, 0x61, 0x00}},
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", []byte{
			0xbf, 0x3f, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x00,
		}},
	},
	"bytes": {
		{[]byte{}, []byte{0xC0}},
		{[]byte{0xDA}, []byte{0xC1, 0xDA}},
		{[]byte{
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61,
		}, []byte{
			0xFF, 0x3F, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		}},
	},
}

func TestRpcStream_basic(t *testing.T) {
	assert := newAssert(t)

	// test rpcStreamCache
	stream := rpcStreamCache.Get().(*rpcStream)
	assert(len(stream.frames)).Equals(1)
	assert(cap(stream.frames)).Equals(8)
	assert(stream.readSeg).Equals(0)
	assert(stream.readIndex).Equals(1)
	assert(stream.readFrame).Equals(*stream.frames[0])
	assert(stream.writeSeg).Equals(0)
	assert(stream.writeIndex).Equals(1)
	assert(stream.writeFrame).Equals(*stream.frames[0])

	// test frameCache
	frame := frameCache.Get().(*[]byte)
	assert(frame).IsNotNil()
	assert(len(*frame)).Equals(512)
	assert(cap(*frame)).Equals(512)

	// test readSkipArray
	assert(readSkipArray).Equals([]int{
		-64, 1, 1, 1, 1, 9, 3, 5, 9, 3, 5, 9, -64, -64, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,

		1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,

		1, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17,
		18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33,
		34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49,
		50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, -6,

		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
		33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48,
		49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, -5,
	})
}

func TestRpcStream_newRPCStream_Release_Reset(t *testing.T) {
	assert := newAssert(t)

	// test rpcStreamCache
	for i := 0; i < 5000; i++ {
		stream := newRPCStream()
		assert(len(stream.frames)).Equals(1)
		assert(cap(stream.frames)).Equals(8)
		assert(stream.readSeg).Equals(0)
		assert(stream.readIndex).Equals(1)
		assert(stream.readFrame).Equals(*stream.frames[0])
		assert(stream.writeSeg).Equals(0)
		assert(stream.writeIndex).Equals(1)
		assert(stream.writeFrame).Equals(*stream.frames[0])

		for n := 0; n < i; n++ {
			stream.putBytes([]byte{9})
		}
		stream.Release()
	}
}

func TestRpcStream_getBuffer(t *testing.T) {
	assert := newAssert(t)

	for i := 0; i < 5000; i++ {
		bytes := make([]byte, i, i)
		for n := 0; n < i; n++ {
			bytes[n] = byte(n)
		}

		stream := newRPCStream()
		stream.putBytes(bytes)
		assert(stream.getBuffer()[0]).Equals(byte(1))
		assert(stream.getBuffer()[1:]).Equals(bytes)
		stream.Release()
	}
}

func TestRpcStream_GetReadPos(t *testing.T) {
	assert := newAssert(t)

	for i := 0; i < 5000; i++ {
		stream := newRPCStream()
		stream.SetWritePos(i)
		stream.SetReadPos(i)
		assert(stream.GetReadPos()).Equals(i)
		stream.Release()
	}
}

func TestRpcStream_SetReadPos(t *testing.T) {
	assert := newAssert(t)

	for i := 1; i < 5000; i++ {
		stream := newRPCStream()
		stream.SetWritePos(i)
		assert(stream.SetReadPos(-1)).IsFalse()
		assert(stream.SetReadPos(i - 1)).IsTrue()
		assert(stream.SetReadPos(i)).IsTrue()
		assert(stream.SetReadPos(i + 1)).IsFalse()
		stream.Release()
	}
}

func TestRpcStream_setReadPosUnsafe(t *testing.T) {
	assert := newAssert(t)
	stream := newRPCStream()
	stream.SetWritePos(10000)

	for i := 0; i < 10000; i++ {
		stream.setReadPosUnsafe(i)
		assert(stream.GetReadPos()).Equals(i)
	}

	stream.Release()
}

func TestRpcStream_saveReadPos_restoreReadPos(t *testing.T) {
	assert := newAssert(t)
	stream := newRPCStream()
	stream.SetWritePos(10000)

	for i := 0; i < 10000; i++ {
		stream.SetReadPos(i)
		stream.saveReadPos()
		assert(stream.SetReadPos(rand.Int() % 10000)).IsTrue()
		stream.restoreReadPos()
		assert(stream.GetReadPos()).Equals(i)
	}

	stream.Release()
}

func TestRpcStream_GetWritePos(t *testing.T) {
	assert := newAssert(t)
	stream := newRPCStream()
	stream.SetWritePos(10000)

	for i := 0; i < 10000; i++ {
		stream.setWritePosUnsafe(i)
		assert(stream.GetWritePos()).Equals(i)
	}
}

func TestRpcStream_SetWritePos(t *testing.T) {
	assert := newAssert(t)

	for i := 0; i < 10000; i++ {
		stream := newRPCStream()
		stream.SetWritePos(i)
		assert(stream.GetWritePos()).Equals(i)
		stream.Release()
	}
}

func TestRpcStream_setWritePosUnsafe(t *testing.T) {
	assert := newAssert(t)
	stream := newRPCStream()
	stream.SetWritePos(10000)

	for i := 0; i < 10000; i++ {
		stream.setWritePosUnsafe(i)
		assert(stream.GetWritePos()).Equals(i)
	}
}

func TestRpcStream_CanRead(t *testing.T) {
	assert := newAssert(t)

	for i := 1; i < 10000; i++ {
		stream := newRPCStream()
		stream.SetWritePos(i)

		stream.setReadPosUnsafe(i - 1)
		assert(stream.CanRead()).IsTrue()

		stream.setReadPosUnsafe(i)
		assert(stream.CanRead()).IsFalse()

		if (i+1)%512 != 0 {
			stream.setReadPosUnsafe(i + 1)
			assert(stream.CanRead()).IsFalse()
		}
	}
}

func TestRpcStream_gotoNextReadFrameUnsafe(t *testing.T) {
	assert := newAssert(t)
	stream := newRPCStream()
	stream.SetWritePos(10000)

	for i := 0; i < 8000; i++ {
		assert(stream.SetReadPos(i)).IsTrue()
		stream.gotoNextReadFrameUnsafe()
		assert(stream.GetReadPos()).Equals((i/512 + 1) * 512)
	}
}

func TestRpcStream_gotoNextReadByteUnsafe(t *testing.T) {
	assert := newAssert(t)
	stream := newRPCStream()
	stream.SetWritePos(10000)

	for i := 0; i < 8000; i++ {
		assert(stream.SetReadPos(i)).IsTrue()
		stream.gotoNextReadByteUnsafe()
		assert(stream.GetReadPos()).Equals(i + 1)
	}
}

func TestRpcStream_hasOneByteToRead(t *testing.T) {
	assert := newAssert(t)

	for i := 1; i < 2000; i++ {
		stream := newRPCStream()
		stream.SetWritePos(i)

		for n := 0; n < i; n++ {
			assert(stream.SetReadPos(n))
			assert(stream.hasOneByteToRead()).IsTrue()
		}

		assert(stream.SetReadPos(i))
		assert(stream.hasOneByteToRead()).IsFalse()
		stream.Release()
	}
}

func TestRpcStream_hasNBytesToRead(t *testing.T) {
	assert := newAssert(t)
	stream := newRPCStream()
	stream.SetWritePos(1100)

	for i := 0; i < 1000; i++ {
		assert(stream.SetReadPos(i)).IsTrue()
		for n := 0; n < 1600; n++ {
			assert(stream.hasNBytesToRead(n)).Equals(i+n <= 1100)
		}
	}
}

func TestRpcStream_isSafetyReadNBytesInCurrentFrame(t *testing.T) {
	assert := newAssert(t)
	stream := newRPCStream()
	stream.SetWritePos(1100)

	for i := 0; i < 800; i++ {
		assert(stream.SetReadPos(i)).IsTrue()
		for n := 0; n < 800; n++ {
			assert(stream.isSafetyReadNBytesInCurrentFrame(n)).
				Equals(512-i%512 > n)
		}
	}
}

func TestRpcStream_isSafetyRead3BytesInCurrentFrame(t *testing.T) {
	assert := newAssert(t)
	stream := newRPCStream()
	stream.SetWritePos(1100)

	for i := 0; i < 800; i++ {
		assert(stream.SetReadPos(i)).IsTrue()
		for n := 0; n < 800; n++ {
			assert(stream.isSafetyRead3BytesInCurrentFrame()).
				Equals(512-i%512 > 3)
		}
	}
}

func TestRpcStream_isSafetyRead5BytesInCurrentFrame(t *testing.T) {
	assert := newAssert(t)
	stream := newRPCStream()
	stream.SetWritePos(1100)

	for i := 0; i < 800; i++ {
		assert(stream.SetReadPos(i)).IsTrue()
		for n := 0; n < 800; n++ {
			assert(stream.isSafetyRead5BytesInCurrentFrame()).
				Equals(512-i%512 > 5)
		}
	}
}

func TestRpcStream_isSafetyRead9BytesInCurrentFrame(t *testing.T) {
	assert := newAssert(t)
	stream := newRPCStream()
	stream.SetWritePos(1100)

	for i := 0; i < 800; i++ {
		assert(stream.SetReadPos(i)).IsTrue()
		for n := 0; n < 800; n++ {
			assert(stream.isSafetyRead9BytesInCurrentFrame()).
				Equals(512-i%512 > 9)
		}
	}
}

func TestRpcStream_putBytes(t *testing.T) {
	assert := newAssert(t)

	for i := 0; i < 600; i++ {
		for n := 0; n < 600; n++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			bytes := make([]byte, n, n)
			for z := 0; z < n; z++ {
				bytes[z] = byte(z)
			}
			stream.putBytes(bytes)
			assert(stream.getBuffer()[i:]).Equals(bytes)
			stream.Release()
		}
	}
}

func TestRpcStream_putString(t *testing.T) {
	assert := newAssert(t)

	for i := 0; i < 600; i++ {
		for n := 0; n < 600; n++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			bytes := make([]byte, n, n)
			for z := 0; z < n; z++ {
				bytes[z] = byte(z)
			}
			strVal := string(bytes)
			stream.putString(strVal)
			assert(stream.getBuffer()[i:]).Equals(bytes)
			stream.Release()
		}
	}
}

func TestRpcStream_read3BytesCrossFrameUnsafe(t *testing.T) {
	assert := newAssert(t)

	stream0 := newRPCStream()
	stream0.SetWritePos(508)
	stream0.putBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9})
	stream0.SetReadPos(509)
	assert(stream0.read3BytesCrossFrameUnsafe()).Equals([]byte{2, 3, 4})
	assert(stream0.GetReadPos()).Equals(512)
	stream0.SetReadPos(510)
	assert(stream0.read3BytesCrossFrameUnsafe()).Equals([]byte{3, 4, 5})
	assert(stream0.GetReadPos()).Equals(513)
	stream0.SetReadPos(511)
	assert(stream0.read3BytesCrossFrameUnsafe()).Equals([]byte{4, 5, 6})
	assert(stream0.GetReadPos()).Equals(514)
}

func TestRpcStream_peek5BytesCrossFrameUnsafe(t *testing.T) {
	assert := newAssert(t)

	stream0 := newRPCStream()
	stream0.SetWritePos(506)
	stream0.putBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13})
	stream0.SetReadPos(507)
	assert(stream0.peek5BytesCrossFrameUnsafe()).Equals([]byte{2, 3, 4, 5, 6})
	assert(stream0.GetReadPos()).Equals(507)
	stream0.SetReadPos(508)
	assert(stream0.peek5BytesCrossFrameUnsafe()).Equals([]byte{3, 4, 5, 6, 7})
	assert(stream0.GetReadPos()).Equals(508)
	stream0.SetReadPos(509)
	assert(stream0.peek5BytesCrossFrameUnsafe()).Equals([]byte{4, 5, 6, 7, 8})
	assert(stream0.GetReadPos()).Equals(509)
	stream0.SetReadPos(510)
	assert(stream0.peek5BytesCrossFrameUnsafe()).Equals([]byte{5, 6, 7, 8, 9})
	assert(stream0.GetReadPos()).Equals(510)
	stream0.SetReadPos(511)
	assert(stream0.peek5BytesCrossFrameUnsafe()).Equals([]byte{6, 7, 8, 9, 10})
	assert(stream0.GetReadPos()).Equals(511)
}

func TestRpcStream_read5BytesCrossFrameUnsafe(t *testing.T) {
	assert := newAssert(t)

	stream0 := newRPCStream()
	stream0.SetWritePos(506)
	stream0.putBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13})
	stream0.SetReadPos(507)
	assert(stream0.read5BytesCrossFrameUnsafe()).Equals([]byte{2, 3, 4, 5, 6})
	assert(stream0.GetReadPos()).Equals(512)
	stream0.SetReadPos(508)
	assert(stream0.read5BytesCrossFrameUnsafe()).Equals([]byte{3, 4, 5, 6, 7})
	assert(stream0.GetReadPos()).Equals(513)
	stream0.SetReadPos(509)
	assert(stream0.read5BytesCrossFrameUnsafe()).Equals([]byte{4, 5, 6, 7, 8})
	assert(stream0.GetReadPos()).Equals(514)
	stream0.SetReadPos(510)
	assert(stream0.read5BytesCrossFrameUnsafe()).Equals([]byte{5, 6, 7, 8, 9})
	assert(stream0.GetReadPos()).Equals(515)
	stream0.SetReadPos(511)
	assert(stream0.read5BytesCrossFrameUnsafe()).Equals([]byte{6, 7, 8, 9, 10})
	assert(stream0.GetReadPos()).Equals(516)
}

func TestRpcStream_read9BytesCrossFrameUnsafe(t *testing.T) {
	assert := newAssert(t)

	stream0 := newRPCStream()
	stream0.SetWritePos(502)
	stream0.putBytes([]byte{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
	})
	stream0.SetReadPos(503)
	assert(stream0.read9BytesCrossFrameUnsafe()).
		Equals([]byte{2, 3, 4, 5, 6, 7, 8, 9, 10})
	assert(stream0.GetReadPos()).Equals(512)
	stream0.SetReadPos(504)
	assert(stream0.read9BytesCrossFrameUnsafe()).
		Equals([]byte{3, 4, 5, 6, 7, 8, 9, 10, 11})
	assert(stream0.GetReadPos()).Equals(513)
	stream0.SetReadPos(505)
	assert(stream0.read9BytesCrossFrameUnsafe()).
		Equals([]byte{4, 5, 6, 7, 8, 9, 10, 11, 12})
	assert(stream0.GetReadPos()).Equals(514)
	stream0.SetReadPos(506)
	assert(stream0.read9BytesCrossFrameUnsafe()).
		Equals([]byte{5, 6, 7, 8, 9, 10, 11, 12, 13})
	assert(stream0.GetReadPos()).Equals(515)
	stream0.SetReadPos(507)
	assert(stream0.read9BytesCrossFrameUnsafe()).
		Equals([]byte{6, 7, 8, 9, 10, 11, 12, 13, 14})
	assert(stream0.GetReadPos()).Equals(516)
	stream0.SetReadPos(508)
	assert(stream0.read9BytesCrossFrameUnsafe()).
		Equals([]byte{7, 8, 9, 10, 11, 12, 13, 14, 15})
	assert(stream0.GetReadPos()).Equals(517)
	stream0.SetReadPos(509)
	assert(stream0.read9BytesCrossFrameUnsafe()).
		Equals([]byte{8, 9, 10, 11, 12, 13, 14, 15, 16})
	assert(stream0.GetReadPos()).Equals(518)
	stream0.SetReadPos(510)
	assert(stream0.read9BytesCrossFrameUnsafe()).
		Equals([]byte{9, 10, 11, 12, 13, 14, 15, 16, 17})
	assert(stream0.GetReadPos()).Equals(519)
	stream0.SetReadPos(511)
	assert(stream0.read9BytesCrossFrameUnsafe()).
		Equals([]byte{10, 11, 12, 13, 14, 15, 16, 17, 18})
	assert(stream0.GetReadPos()).Equals(520)
}

func TestRpcStream_readNBytesUnsafe(t *testing.T) {
	assert := newAssert(t)
	stream := newRPCStream()
	for i := 0; i < 2000; i++ {
		stream.putBytes([]byte{byte(i)})
	}
	streamBuf := stream.getBuffer()

	for i := 1; i < 600; i++ {
		for n := 0; n < 1100; n++ {
			stream.SetReadPos(i)
			assert(stream.readNBytesUnsafe(n)).
				Equals(streamBuf[i : i+n])
		}
	}
}

func TestRpcStream_peekSkip(t *testing.T) {
	assert := newAssert(t)

	testCollection := Array{
		Array{[]byte{0}, 0},
		Array{[]byte{1}, 1},
		Array{[]byte{2}, 1},
		Array{[]byte{3}, 1},
		Array{[]byte{4}, 1},
		Array{[]byte{5}, 9},
		Array{[]byte{6}, 3},
		Array{[]byte{7}, 5},
		Array{[]byte{8}, 9},
		Array{[]byte{9}, 3},
		Array{[]byte{10}, 5},
		Array{[]byte{11}, 9},
		Array{[]byte{12}, 0},
		Array{[]byte{13}, 0},
		Array{[]byte{14}, 1},
		Array{[]byte{63}, 1},
		Array{[]byte{64}, 1},
		Array{[]byte{65, 6, 0, 0, 0}, 6},
		Array{[]byte{94, 6, 0, 0, 0}, 6},
		Array{[]byte{95, 6, 0, 0, 0}, 6},
		Array{[]byte{96, 6, 0, 0, 0}, 1},
		Array{[]byte{97, 6, 0, 0, 0}, 6},
		Array{[]byte{126, 6, 0, 0, 0}, 6},
		Array{[]byte{127, 6, 0, 0, 0}, 6},
		Array{[]byte{128, 6, 0, 0, 0}, 1},
		Array{[]byte{129, 6, 0, 0, 0}, 3},
		Array{[]byte{190, 6, 0, 0, 0}, 64},
		Array{[]byte{191, 80, 0, 0, 0}, 86},
		Array{[]byte{192, 6, 0, 0, 0}, 1},
		Array{[]byte{193, 6, 0, 0, 0}, 2},
		Array{[]byte{254, 6, 0, 0, 0}, 63},
		Array{[]byte{255, 80, 0, 0, 0}, 85},
		Array{[]byte{255, 80, 0}, 0},
	}

	for i := 1; i < 600; i++ {
		for _, item := range testCollection {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.putBytes(item.(Array)[0].([]byte))
			assert(stream.peekSkip()).Equals(item.(Array)[1])
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}
}

func TestRpcStream_readSkipItem(t *testing.T) {
	assert := newAssert(t)

	for i := 1; i < 600; i++ {
		for j := 0; j < 600; j++ {
			// skip > 0
			bytes := make([]byte, j, j)
			for n := 0; n < j; n++ {
				bytes[n] = byte(n)
			}
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.WriteBytes(bytes)
			assert(stream.readSkipItem(stream.GetWritePos() - 1)).Equals(-1)
			assert(stream.GetReadPos()).Equals(i)
			assert(stream.readSkipItem(stream.GetWritePos())).Equals(i)
			assert(stream.GetReadPos()).Equals(stream.GetWritePos())

			// skip == 0
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.putBytes([]byte{13})
			assert(stream.readSkipItem(stream.GetWritePos())).Equals(-1)

			stream.Release()
		}
	}
}

func TestRpcStream_writeStreamUnsafe(t *testing.T) {
	assert := newAssert(t)

	dataStream := newRPCStream()
	for i := 0; i < 2000; i++ {
		dataStream.putBytes([]byte{byte(i)})
	}
	dataStreamBuf := dataStream.getBuffer()

	fnTest := func(length int) {
		for i := 0; i < 550; i++ {
			// skip for performance
			if i > 50 && i < 480 {
				continue
			}
			for j := 0; j < 550; j++ {
				bytes := make([]byte, j, j)
				for n := 0; n < j; n++ {
					bytes[n] = byte(n)
				}
				stream := newRPCStream()
				stream.putBytes(bytes)
				dataStream.SetReadPos(i)
				stream.writeStreamUnsafe(dataStream, length)
				streamBuf := stream.getBuffer()
				assert(streamBuf[0]).Equals(byte(1))
				assert(streamBuf[1 : 1+j]).Equals(bytes)
				assert(streamBuf[1+j:]).Equals(dataStreamBuf[i : i+length])
				assert(dataStream.GetReadPos()).Equals(i + length)
				assert(stream.GetWritePos()).Equals(j + 1 + length)
				stream.Release()
			}
		}
	}

	fnTest(0)
	fnTest(1)
	fnTest(2)
	fnTest(3)
	fnTest(12)
	fnTest(500)
	fnTest(511)
	fnTest(512)
	fnTest(513)
	fnTest(1100)
}

func TestRpcStream_writeStreamNext(t *testing.T) {
	assert := newAssert(t)

	for i := 0; i < 550; i++ {
		bytes := make([]byte, i, i)
		dataStream := newRPCStream()
		for n := 0; n < i; n++ {
			bytes[n] = byte(n)
		}
		dataStream.WriteBytes(bytes)

		// invalid code
		bugStream0 := newRPCStream()
		bugStream0.putBytes([]byte{13})

		// length overflow
		bugStream1 := newRPCStream()
		bugStream1.putBytes([]byte{65, 6, 0, 0, 0})

		for j := 0; j < 550; j++ {
			stream := newRPCStream()
			stream.SetWritePos(j)
			dataStream.SetReadPos(1)

			// dataStream
			assert(stream.writeStreamNext(dataStream)).IsTrue()
			assert(dataStream.GetReadPos()).Equals(dataStream.GetWritePos())
			assert(stream.GetWritePos()).
				Equals(dataStream.GetWritePos() + j - 1)
			// bugStream0
			assert(stream.writeStreamNext(bugStream0)).IsFalse()
			assert(bugStream0.GetReadPos()).Equals(1)
			assert(stream.GetWritePos()).
				Equals(dataStream.GetWritePos() + j - 1)
			// bugStream1
			assert(stream.writeStreamNext(bugStream1)).IsFalse()
			assert(bugStream1.GetReadPos()).Equals(1)
			assert(stream.GetWritePos()).
				Equals(dataStream.GetWritePos() + j - 1)

			stream.Release()
		}
	}
}

func TestRpcStream_WriteNil(t *testing.T) {
	assert := newAssert(t)

	for _, testData := range rpcStreamTestCollections["nil"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.WriteNil()
			assert(stream.getBuffer()[i:]).Equals(testData[1])
			assert(stream.GetWritePos()).Equals(i + 1)
			stream.Release()
		}
	}
}

func TestRpcStream_WriteBool(t *testing.T) {
	assert := newAssert(t)

	for _, testData := range rpcStreamTestCollections["bool"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.WriteBool(testData[0].(bool))
			assert(stream.getBuffer()[i:]).Equals(testData[1])
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}
	}
}

func TestRpcStream_WriteFloat64(t *testing.T) {
	assert := newAssert(t)

	for _, testData := range rpcStreamTestCollections["float64"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.WriteFloat64(testData[0].(float64))
			assert(stream.getBuffer()[i:]).Equals(testData[1])
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}
	}
}

func TestRpcStream_WriteInt64(t *testing.T) {
	assert := newAssert(t)

	for _, testData := range rpcStreamTestCollections["int64"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.WriteInt64(testData[0].(int64))
			assert(stream.getBuffer()[i:]).Equals(testData[1])
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}
	}
}

func TestRpcStream_WriteUInt64(t *testing.T) {
	assert := newAssert(t)

	for _, testData := range rpcStreamTestCollections["uint64"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.WriteUint64(testData[0].(uint64))
			assert(stream.getBuffer()[i:]).Equals(testData[1])
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}
	}
}

func TestRpcStream_WriteString(t *testing.T) {
	assert := newAssert(t)

	for _, testData := range rpcStreamTestCollections["string"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.WriteString(testData[0].(string))
			assert(stream.getBuffer()[i:]).Equals(testData[1])
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}
	}
}

func TestRpcStream_WriteBytes(t *testing.T) {
	assert := newAssert(t)

	for _, testData := range rpcStreamTestCollections["bytes"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.WriteBytes(testData[0].([]byte))
			assert(stream.getBuffer()[i:]).Equals(testData[1])
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}
	}
}

func TestRpcStream_ReadNil(t *testing.T) {
	assert := newAssert(t)

	for _, testData := range rpcStreamTestCollections["nil"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])

			assert(stream.ReadNil()).Equals(true)
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}

		// overflow
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadNil()).IsFalse()
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.putBytes([]byte{13})
			assert(stream.ReadNil()).IsFalse()
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}
}

func TestRpcStream_ReadBool(t *testing.T) {
	assert := newAssert(t)

	for _, testData := range rpcStreamTestCollections["bool"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			assert(stream.ReadBool()).Equals(testData[0], true)
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}

		// overflow
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadBool()).Equals(false, false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.putBytes([]byte{13})
			assert(stream.ReadBool()).Equals(false, false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}
}

func TestRpcStream_ReadFloat64(t *testing.T) {
	assert := newAssert(t)

	for _, testData := range rpcStreamTestCollections["float64"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			assert(stream.ReadFloat64()).Equals(testData[0], true)
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}

		// overflow
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadFloat64()).Equals(float64(0), false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.putBytes([]byte{13})
			assert(stream.ReadFloat64()).Equals(float64(0), false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}
}

func TestRpcStream_ReadInt64(t *testing.T) {
	assert := newAssert(t)

	for _, testData := range rpcStreamTestCollections["int64"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			assert(stream.ReadInt64()).Equals(testData[0], true)
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}

		// overflow
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadInt64()).Equals(int64(0), false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.putBytes([]byte{13})
			assert(stream.ReadInt64()).Equals(int64(0), false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}
}

func TestRpcStream_ReadUint64(t *testing.T) {
	assert := newAssert(t)

	for _, testData := range rpcStreamTestCollections["uint64"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			assert(stream.ReadUint64()).Equals(testData[0], true)
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}

		// overflow
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadUint64()).Equals(uint64(0), false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.putBytes([]byte{13})
			assert(stream.ReadUint64()).Equals(uint64(0), false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}
}

func TestRpcStream_ReadString(t *testing.T) {
	assert := newAssert(t)

	for _, testData := range rpcStreamTestCollections["string"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			assert(stream.ReadString()).Equals(testData[0], true)
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}

		// overflow
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadString()).Equals("", false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.putBytes([]byte{13})
			assert(stream.ReadString()).Equals("", false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}

		// read tail is not zero
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			stream.SetWritePos(stream.GetWritePos() - 1)
			stream.putBytes([]byte{1})
			assert(stream.ReadString()).Equals("", false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}
}

func TestRpcStream_ReadUnsafeString(t *testing.T) {
	assert := newAssert(t)

	for _, testData := range rpcStreamTestCollections["string"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			assert(stream.ReadUnsafeString()).Equals(testData[0], true)
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}

		// overflow
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadUnsafeString()).Equals(emptyString, false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.putBytes([]byte{13})
			assert(stream.ReadUnsafeString()).Equals(emptyString, false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}

		// read tail is not zero
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			stream.SetWritePos(stream.GetWritePos() - 1)
			stream.putBytes([]byte{1})
			assert(stream.ReadUnsafeString()).Equals(emptyString, false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}
}

func TestRpcStream_ReadBytes(t *testing.T) {
	assert := newAssert(t)

	for _, testData := range rpcStreamTestCollections["bytes"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			assert(stream.ReadBytes()).Equals(testData[0], true)
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}

		// overflow
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadBytes()).Equals(emptyBytes, false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.putBytes([]byte{13})
			assert(stream.ReadBytes()).Equals(emptyBytes, false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}
}

func TestRpcStream_ReadUnsafeBytes(t *testing.T) {
	assert := newAssert(t)

	for _, testData := range rpcStreamTestCollections["bytes"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			assert(stream.ReadUnsafeBytes()).Equals(testData[0], true)
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}

		// overflow
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadUnsafeBytes()).Equals(emptyBytes, false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := newRPCStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.putBytes([]byte{13})
			assert(stream.ReadUnsafeBytes()).Equals(emptyBytes, false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}
}
