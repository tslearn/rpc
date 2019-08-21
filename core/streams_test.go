package core

import (
	"math/rand"
	"testing"
)

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
