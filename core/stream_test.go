package core

import (
	"encoding/binary"
	"math/rand"
	"strconv"
	"testing"
)

var unusedBytes = make([]byte, 40960, 40960)

func fillTestStream(
	stream *rpcStream,
	outerBuffer []byte,
	startPos int,
	data interface{},
) {
	if startPos > 0 {
		stream.Reset()
		buffer := outerBuffer[0 : startPos-1]
		copy(buffer, unusedBytes)
		stream.PutBytes(buffer)
		stream.Write(data)
		stream.WriteNil()
		stream.SetReadPos(startPos)
	}
}

func fillTestStreamByBuffer(
	stream *rpcStream,
	startPos int,
	buffer []byte,
) {
	stream.Reset()
	stream.writeIndex = 0
	stream.PutBytes(buffer)
	stream.SetReadPos(startPos)
}

func getTestBuffer(
	outerBuffer []byte,
	startPos int,
	rawBytes []byte,
) []byte {
	if startPos > 0 {
		ret := outerBuffer[0:startPos]
		ret[0] = 1
		copy(ret[1:], unusedBytes)
		ret = append(ret, rawBytes...)
		ret = append(ret, 1)
		return ret
	}
	return nil
}

func getStringTestData(n uint32) [2]interface{} {
	str := GetRandString(int(n))

	if n == 0 {
		return [2]interface{}{"", []byte{128}}
	} else if n < 63 {
		bytes := make([]byte, n+2, n+2)
		bytes[0] = byte(128 + n)
		copy(bytes[1:], str)
		bytes[n+1] = 0
		return [2]interface{}{str, bytes}
	} else {
		bytes := make([]byte, n+6, n+6)
		bytes[0] = 191
		binary.LittleEndian.PutUint32(bytes[1:], uint32(n))
		copy(bytes[5:], str)
		bytes[n+5] = 0
		return [2]interface{}{str, bytes}
	}
}

func getBytesTestData(n uint32) [2]interface{} {
	if n == 0 {
		return [2]interface{}{[]byte{}, []byte{192}}
	} else if n < 63 {
		bytes := make([]byte, n+1, n+1)
		bytes[0] = byte(192 + n)
		for i := 1; i < int(n+1); i++ {
			bytes[i] = byte(i)
		}
		return [2]interface{}{bytes[1:], bytes}
	} else {
		bytes := make([]byte, n+5, n+5)
		bytes[0] = byte(255)
		binary.LittleEndian.PutUint32(bytes[1:], n)
		for i := uint32(5); i < n+5; i++ {
			bytes[i] = byte(i)
		}
		return [2]interface{}{bytes[5:], bytes}
	}
}

func getRandomStringArrayTestData(n uint32) rpcArray {
	ret := newRPCArray(&rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	})
	for i := uint32(0); i < n; i++ {
		ret.Append(GetRandString(int(GetRandUint32() % 512)))
	}
	return ret
}

func getRandomStringMapTestData(n uint32) rpcMap {
	ret := newRPCMap(&rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	})
	for i := uint32(0); i < n; i++ {
		ret.Set(GetRandString(64), GetRandString(int(GetRandUint32()%512)))
	}
	return ret
}

func Test_RPCStream_Release(t *testing.T) {
	assert := NewAssert(t)
	stream := NewRPCStream()
	for i := 0; i < 1000; i++ {
		stream.WriteInt64(int64(i))
	}

	stream.Release()

	assert(len(stream.frames)).Equals(1)
	assert(cap(stream.frames)).Equals(8)
	assert(stream.readSeg, stream.readIndex).Equals(0, 1)
	assert(stream.readFrame).Equals(*stream.frames[0])
	assert(stream.writeSeg, stream.writeIndex).Equals(0, 1)
	assert(stream.writeFrame).Equals(*stream.frames[0])
	assert(stream.saveSeg, stream.saveIndex).Equals(0, 1)
}

func Test_RPCStream_GetTotalFrames(t *testing.T) {
	assert := NewAssert(t)
	stream := NewRPCStream()
	assert(stream.GetTotalFrames()).Equals(1)
	for i := 0; i < 511; i++ {
		stream.WriteNil()
	}
	assert(stream.GetTotalFrames()).Equals(2)
}

func Test_RPCStream_SetReadPos(t *testing.T) {
	assert := NewAssert(t)
	stream := NewRPCStream()

	assert(stream.SetReadPos(-1)).IsFalse()
	assert(stream.SetReadPos(0)).IsTrue()
	assert(stream.SetReadPos(1)).IsTrue()
	assert(stream.SetReadPos(2)).IsFalse()

	for i := 0; i < 511; i++ {
		stream.WriteNil()
	}

	assert(stream.SetReadPos(-1)).IsFalse()
	assert(stream.SetReadPos(0)).IsTrue()
	assert(stream.SetReadPos(512)).IsTrue()
	assert(stream.SetReadPos(513)).IsFalse()
}

func Test_RPCStream_readByte(t *testing.T) {
	assert := NewAssert(t)
	stream := NewRPCStream()

	assert(stream.readByte(-1)).Equals(byte(0), false)
	assert(stream.readByte(0)).Equals(byte(1), true)
	assert(stream.readByte(1)).Equals(byte(0), false)
}

func Test_RPCStream_writeStreamUnsafe(t *testing.T) {
	assert := NewAssert(t)

	oBytes := make([]byte, 3000, 3000)
	wBytes := make([]byte, 3000, 3000)
	exp := make([]byte, 3000, 3000)
	exp[0] = 1

	for i := 0; i < 3000; i++ {
		oBytes[i] = 0x77
		wBytes[i] = 0x88
	}

	wStream := NewRPCStream()
	wStream.PutBytes(wBytes)

	for i := 0; i < 200000; i++ {
		oWritePos := int(GetRandUint32()) % 550
		wsReadPos := int(GetRandUint32()) % 1500
		length := int(GetRandUint32()) % 1500

		s := NewRPCStream()
		s.PutBytes(oBytes[0:oWritePos])
		wStream.SetReadPos(1 + wsReadPos)

		s.writeStreamUnsafe(wStream, length)
		copy(exp[1:], oBytes[:oWritePos])
		copy(exp[1+oWritePos:], wBytes[:length])
		assert(s.GetBuffer()).Equals(exp[:1+oWritePos+length])

		assert(s.GetReadPos()).Equals(1)
		assert(s.GetWritePos()).Equals(1 + oWritePos + length)
		assert(wStream.GetReadPos()).Equals(1 + wsReadPos + length)
		assert(wStream.GetWritePos()).Equals(3001)
		s.Release()
	}
}

func Test_RPCStream_writeStream(t *testing.T) {
	assert := NewAssert(t)
	stream := NewRPCStream()

	// writeStreamNext param stream bytecode error
	writeStream := NewRPCStream()
	copy((*writeStream.frames[0])[1:], []byte{13})
	assert(stream.writeStreamNext(writeStream)).IsFalse()

	// writeStreamUnsafe, unsafe do not check the bytecode error
	writeStream = NewRPCStream()
	copy((*writeStream.frames[0])[1:], []byte{13})
	stream.writeStreamUnsafe(writeStream, 1)
	assert(stream.GetBuffer()).Equals([]byte{1, 13})
	assert(stream.readSkipItem(stream.GetWritePos())).Equals(-1)
}

func Test_RPCStream_Nil(t *testing.T) {
	assert := NewAssert(t)
	pubBytes := make([]byte, 40960, 40960)
	stream := NewRPCStream()

	testCollection := [][2]interface{}{
		{nil, []byte{0x01}},
	}

	for _, testData := range testCollection {
		// ok
		for i := 1; i < 1100; i++ {
			fillTestStream(stream, pubBytes, i, testData[0])
			targetBuffer := getTestBuffer(pubBytes, i, testData[1].([]byte))

			assert(stream.GetBuffer(), stream.writeIndex, stream.writeSeg).
				Equals(targetBuffer, len(targetBuffer)%512, len(targetBuffer)/512)
			assert(stream.ReadNil()).IsTrue()
			assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
		}

		// error: read overflow
		for i := 1; i < 550; i++ {
			fillTestStream(stream, pubBytes, i, testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)

				stream.setWritePosUnsafe(idx)
				assert(stream.ReadNil()).IsFalse()
				assert(stream.GetReadPos()).Equals(i)

				stream.setWritePosUnsafe(writePos)
				assert(stream.ReadNil()).IsTrue()
				assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
			}
		}

		// ok: have random tail bytes
		for i := 1; i < 1100; i++ {
			targetBuffer := getTestBuffer(pubBytes, i, testData[1].([]byte))
			targetBuffer = append(targetBuffer, unusedBytes[0:rand.Uint32()%1024]...)
			fillTestStreamByBuffer(stream, i, targetBuffer)
			assert(stream.ReadNil()).IsTrue()
		}
	}
}

func Test_RPCStream_Bool(t *testing.T) {
	assert := NewAssert(t)
	pubBytes := make([]byte, 40960, 40960)
	stream := NewRPCStream()

	testCollection := [][2]interface{}{
		{true, []byte{0x02}},
		{false, []byte{0x03}},
	}

	for _, testData := range testCollection {
		// ok
		for i := 1; i < 1100; i++ {
			fillTestStream(stream, pubBytes, i, testData[0])
			targetBuffer := getTestBuffer(pubBytes, i, testData[1].([]byte))

			assert(stream.GetBuffer(), stream.writeIndex, stream.writeSeg).
				Equals(targetBuffer, len(targetBuffer)%512, len(targetBuffer)/512)
			assert(stream.ReadBool()).Equals(testData[0], true)
			assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
		}

		// skip test
		for i := 1; i < 1100; i++ {
			fillTestStream(stream, pubBytes, i, testData[0])
			endPos := stream.GetWritePos() - 1
			stream.SetReadPos(i)
			assert(stream.readSkipItem(endPos - 1)).Equals(-1)
			assert(stream.GetReadPos()).Equals(i)
			assert(stream.readSkipItem(endPos)).Equals(i)
			assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
		}

		// error: read overflow
		for i := 1; i < 550; i++ {
			fillTestStream(stream, pubBytes, i, testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)

				assert(stream.ReadBool()).Equals(false, false)
				assert(stream.GetReadPos()).Equals(i)

				stream.setWritePosUnsafe(writePos)
				assert(stream.ReadBool()).Equals(testData[0], true)
				assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
			}
		}

		// ok: have random tail bytes
		for i := 1; i < 1100; i++ {
			targetBuffer := getTestBuffer(pubBytes, i, testData[1].([]byte))
			targetBuffer = append(targetBuffer, unusedBytes[0:rand.Uint32()%1024]...)
			fillTestStreamByBuffer(stream, i, targetBuffer)
			assert(stream.ReadBool()).Equals(testData[0], true)
		}
	}
}

func Test_RPCStream_Float64(t *testing.T) {
	assert := NewAssert(t)
	pubBytes := make([]byte, 40960, 40960)
	stream := NewRPCStream()

	testCollection := [][2]interface{}{
		{float64(0), []byte{0x04}},
		{float64(3.1415926), []byte{0x05, 0x4a, 0xd8, 0x12, 0x4d, 0xfb, 0x21, 0x09, 0x40}},
		{float64(-3.1415926), []byte{0x05, 0x4a, 0xd8, 0x12, 0x4d, 0xfb, 0x21, 0x09, 0xc0}},
	}

	for _, testData := range testCollection {
		// ok
		for i := 1; i < 1100; i++ {
			fillTestStream(stream, pubBytes, i, testData[0])
			targetBuffer := getTestBuffer(pubBytes, i, testData[1].([]byte))

			assert(stream.GetBuffer(), stream.writeIndex, stream.writeSeg).
				Equals(targetBuffer, len(targetBuffer)%512, len(targetBuffer)/512)
			assert(stream.ReadFloat64()).Equals(testData[0], true)
			assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
		}

		// skip test
		for i := 1; i < 1100; i++ {
			fillTestStream(stream, pubBytes, i, testData[0])
			endPos := stream.GetWritePos() - 1
			stream.SetReadPos(i)
			assert(stream.readSkipItem(endPos - 1)).Equals(-1)
			assert(stream.GetReadPos()).Equals(i)
			assert(stream.readSkipItem(endPos)).Equals(i)
			assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
		}

		// error: read overflow
		for i := 1; i < 550; i++ {
			if i%512 < 10 || i%512 > 500 {
				fillTestStream(stream, pubBytes, i, testData[0])
				writePos := stream.GetWritePos()
				for idx := writePos - 2; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.setWritePosUnsafe(idx)

					assert(stream.ReadFloat64()).Equals(float64(0), false)
					assert(stream.GetReadPos()).Equals(i)

					stream.setWritePosUnsafe(writePos)
					assert(stream.ReadFloat64()).Equals(testData[0], true)
					assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
				}
			}
		}

		// ok: have random tail bytes
		for i := 1; i < 1100; i++ {
			targetBuffer := getTestBuffer(pubBytes, i, testData[1].([]byte))
			targetBuffer = append(targetBuffer, unusedBytes[0:rand.Uint32()%1024]...)
			fillTestStreamByBuffer(stream, i, targetBuffer)
			assert(stream.ReadFloat64()).Equals(testData[0], true)
		}
	}
}

func Test_RPCStream_Int64(t *testing.T) {
	assert := NewAssert(t)
	pubBytes := make([]byte, 40960, 40960)
	stream := NewRPCStream()

	testCollection := [][2]interface{}{
		{int64(-9223372036854775808), []byte{0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80}},
		{int64(-2147483649), []byte{0x08, 0xff, 0xff, 0xff, 0x7f, 0xff, 0xff, 0xff, 0xff}},
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
		{int64(2147483648), []byte{0x08, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00}},
		{int64(9223372036854775807), []byte{0x08, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
	}

	for _, testData := range testCollection {
		// ok
		for i := 1; i < 1100; i++ {
			fillTestStream(stream, pubBytes, i, testData[0])
			targetBuffer := getTestBuffer(pubBytes, i, testData[1].([]byte))

			assert(stream.GetBuffer(), stream.writeIndex, stream.writeSeg).
				Equals(targetBuffer, len(targetBuffer)%512, len(targetBuffer)/512)
			assert(stream.ReadInt64()).Equals(testData[0], true)
			assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
		}

		// skip test
		for i := 1; i < 1100; i++ {
			fillTestStream(stream, pubBytes, i, testData[0])
			endPos := stream.GetWritePos() - 1
			stream.SetReadPos(i)
			assert(stream.readSkipItem(endPos - 1)).Equals(-1)
			assert(stream.GetReadPos()).Equals(i)
			assert(stream.readSkipItem(endPos)).Equals(i)
			assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
		}

		// error: read overflow
		for i := 1; i < 550; i++ {
			if i%512 < 10 || i%512 > 500 {
				fillTestStream(stream, pubBytes, i, testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.setWritePosUnsafe(idx)

					assert(stream.ReadInt64()).Equals(int64(0), false)
					assert(stream.GetReadPos()).Equals(i)

					stream.setWritePosUnsafe(writePos)
					assert(stream.ReadInt64()).Equals(testData[0], true)
					assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
				}
			}
		}

		// ok: have random tail bytes
		for i := 1; i < 1100; i++ {
			targetBuffer := getTestBuffer(pubBytes, i, testData[1].([]byte))
			targetBuffer = append(targetBuffer, unusedBytes[0:rand.Uint32()%1024]...)
			fillTestStreamByBuffer(stream, i, targetBuffer)
			assert(stream.ReadInt64()).Equals(testData[0], true)
		}
	}
}

func Test_RPCStream_Uint64(t *testing.T) {
	assert := NewAssert(t)
	pubBytes := make([]byte, 40960, 40960)
	stream := NewRPCStream()
	testCollection := [][2]interface{}{
		{uint64(0), []byte{0x36}},
		{uint64(9), []byte{0x3f}},
		{uint64(10), []byte{0x09, 0x0a, 0x00}},
		{uint64(65535), []byte{0x09, 0xff, 0xff}},
		{uint64(65536), []byte{0x0a, 0x00, 0x00, 0x01, 0x00}},
		{uint64(4294967295), []byte{0x0a, 0xff, 0xff, 0xff, 0xff}},
		{uint64(4294967296), []byte{0x0b, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}},
		{uint64(18446744073709551615), []byte{0x0b, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	}

	for _, testData := range testCollection {
		// ok
		for i := 1; i < 1100; i++ {
			fillTestStream(stream, pubBytes, i, testData[0])
			targetBuffer := getTestBuffer(pubBytes, i, testData[1].([]byte))

			assert(stream.GetBuffer(), stream.writeIndex, stream.writeSeg).
				Equals(targetBuffer, len(targetBuffer)%512, len(targetBuffer)/512)
			assert(stream.ReadUint64()).Equals(testData[0], true)
			assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
		}

		// skip test
		for i := 1; i < 1100; i++ {
			fillTestStream(stream, pubBytes, i, testData[0])
			endPos := stream.GetWritePos() - 1
			stream.SetReadPos(i)
			assert(stream.readSkipItem(endPos - 1)).Equals(-1)
			assert(stream.GetReadPos()).Equals(i)
			assert(stream.readSkipItem(endPos)).Equals(i)
			assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
		}

		// error: read overflow
		for i := 1; i < 550; i++ {
			if i%512 < 10 || i%512 > 500 {
				fillTestStream(stream, pubBytes, i, testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.setWritePosUnsafe(idx)

					assert(stream.ReadUint64()).Equals(uint64(0), false)
					assert(stream.GetReadPos()).Equals(i)

					stream.setWritePosUnsafe(writePos)
					assert(stream.ReadUint64()).Equals(testData[0], true)
					assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
				}
			}
		}

		// ok: have random tail bytes
		for i := 1; i < 1100; i++ {
			targetBuffer := getTestBuffer(pubBytes, i, testData[1].([]byte))
			targetBuffer = append(targetBuffer, unusedBytes[0:rand.Uint32()%1024]...)
			fillTestStreamByBuffer(stream, i, targetBuffer)
			assert(stream.ReadUint64()).Equals(testData[0], true)
		}
	}
}

func Test_RPCStream_String(t *testing.T) {
	assert := NewAssert(t)
	pubBytes := make([]byte, 40960, 40960)
	stream := NewRPCStream()

	testCollection := [][2]interface{}{
		getStringTestData(0),
		getStringTestData(1),
		getStringTestData(61),
		getStringTestData(62),
		getStringTestData(63),
		getStringTestData(131),
		getStringTestData(691),
	}

	for _, testData := range testCollection {
		// ok
		for i := 1; i < 1100; i++ {
			fillTestStream(stream, pubBytes, i, testData[0])
			targetBuffer := getTestBuffer(pubBytes, i, testData[1].([]byte))

			assert(stream.GetBuffer(), stream.writeIndex, stream.writeSeg).
				Equals(targetBuffer, len(targetBuffer)%512, len(targetBuffer)/512)

			stream.SetReadPos(i)
			assert(stream.ReadUnsafeString()).Equals(testData[0], true)
			assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()

			stream.SetReadPos(i)
			assert(stream.ReadString()).Equals(testData[0], true)
			assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
		}

		// skip test
		for i := 1; i < 1100; i++ {
			fillTestStream(stream, pubBytes, i, testData[0])
			endPos := stream.GetWritePos() - 1
			stream.SetReadPos(i)
			assert(stream.readSkipItem(endPos - 1)).Equals(-1)
			assert(stream.GetReadPos()).Equals(i)
			assert(stream.readSkipItem(endPos)).Equals(i)
			assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
		}

		// error: string tail is not zero
		if testData[0].(string) != "" {
			for i := 1; i < 1100; i++ {
				targetBuffer := getTestBuffer(pubBytes, i, testData[1].([]byte))

				// ReadUnsafeString
				targetBuffer[len(targetBuffer)-2] = 1
				fillTestStreamByBuffer(stream, i, targetBuffer)
				assert(stream.ReadUnsafeString()).Equals("", false)
				targetBuffer[len(targetBuffer)-2] = 0
				fillTestStreamByBuffer(stream, i, targetBuffer)
				assert(stream.ReadUnsafeString()).Equals(testData[0], true)
				assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()

				// ReadRPCString
				targetBuffer[len(targetBuffer)-2] = 1
				fillTestStreamByBuffer(stream, i, targetBuffer)
				assert(stream.ReadString()).Equals(emptyString, false)
				assert(stream.GetReadPos()).Equals(i)
				targetBuffer[len(targetBuffer)-2] = 0
				fillTestStreamByBuffer(stream, i, targetBuffer)
				assert(stream.ReadString()).Equals(testData[0], true)
				assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
			}
		}

		// error: read overflow
		for i := 1; i < 550; i++ {
			if i%512 < 10 || i%512 > 500 {
				fillTestStream(stream, pubBytes, i, testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					// ReadUnsafeString
					stream.SetReadPos(i)
					stream.setWritePosUnsafe(idx)
					assert(stream.ReadUnsafeString()).Equals(emptyString, false)
					assert(stream.GetReadPos()).Equals(i)
					stream.setWritePosUnsafe(writePos)
					assert(stream.ReadUnsafeString()).Equals(testData[0], true)
					assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()

					// ReadRPCString
					stream.SetReadPos(i)
					stream.setWritePosUnsafe(idx)
					assert(stream.ReadString()).Equals(emptyString, false)
					assert(stream.GetReadPos()).Equals(i)
					stream.setWritePosUnsafe(writePos)
					assert(stream.ReadString()).Equals(testData[0], true)
					assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
				}
			}
		}

		// ok: have random tail bytes
		for i := 1; i < 1100; i++ {
			targetBuffer := getTestBuffer(pubBytes, i, testData[1].([]byte))
			targetBuffer = append(targetBuffer, unusedBytes[0:rand.Uint32()%1024]...)
			fillTestStreamByBuffer(stream, i, targetBuffer)

			// ReadUnsafeString
			stream.SetReadPos(i)
			assert(stream.ReadUnsafeString()).Equals(testData[0], true)

			// ReadRPCString
			stream.SetReadPos(i)
			assert(stream.ReadString()).Equals(testData[0], true)
		}
	}
}

func Test_RPCStream_Bytes(t *testing.T) {
	assert := NewAssert(t)
	pubBytes := make([]byte, 40960, 40960)
	stream := NewRPCStream()

	testCollection := [][2]interface{}{
		getBytesTestData(0),
		getBytesTestData(1),
		getBytesTestData(61),
		getBytesTestData(62),
		getBytesTestData(63),
		getBytesTestData(131),
		getBytesTestData(691),
	}

	for _, testData := range testCollection {
		// ok
		for i := 1; i < 1100; i++ {
			fillTestStream(stream, pubBytes, i, testData[0])
			targetBuffer := getTestBuffer(pubBytes, i, testData[1].([]byte))

			assert(stream.GetBuffer(), stream.writeIndex, stream.writeSeg).
				Equals(targetBuffer, len(targetBuffer)%512, len(targetBuffer)/512)

			// ReadUnsafeBytes
			stream.SetReadPos(i)
			assert(stream.ReadUnsafeBytes()).Equals(testData[0], true)
			assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()

			// ReadRPCBytes
			stream.SetReadPos(i)
			assert(stream.ReadBytes()).Equals(testData[0], true)
			assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
		}

		// skip test
		for i := 1; i < 1100; i++ {
			fillTestStream(stream, pubBytes, i, testData[0])
			endPos := stream.GetWritePos() - 1
			stream.SetReadPos(i)
			assert(stream.readSkipItem(endPos - 1)).Equals(-1)
			assert(stream.GetReadPos()).Equals(i)
			assert(stream.readSkipItem(endPos)).Equals(i)
			assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
		}

		// error: read overflow
		for i := 1; i < 550; i++ {
			if i%512 < 10 || i%512 > 500 {
				fillTestStream(stream, pubBytes, i, testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					// ReadUnsafeBytes
					stream.SetReadPos(i)
					stream.setWritePosUnsafe(idx)
					assert(stream.ReadUnsafeBytes()).Equals(nil, false)
					assert(stream.GetReadPos()).Equals(i)
					stream.setWritePosUnsafe(writePos)
					assert(stream.ReadUnsafeBytes()).Equals(testData[0], true)
					assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()

					// ReadRPCBytes
					stream.SetReadPos(i)
					stream.setWritePosUnsafe(idx)
					assert(stream.ReadBytes()).Equals(emptyBytes, false)
					assert(stream.GetReadPos()).Equals(i)
					stream.setWritePosUnsafe(writePos)
					assert(stream.ReadBytes()).Equals(testData[0], true)
					assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
				}
			}
		}

		// ok: have random tail bytes
		for i := 1; i < 1100; i++ {
			targetBuffer := getTestBuffer(pubBytes, i, testData[1].([]byte))
			targetBuffer = append(targetBuffer, unusedBytes[0:rand.Uint32()%1024]...)
			fillTestStreamByBuffer(stream, i, targetBuffer)

			// ReadUnsafeBytes
			stream.SetReadPos(i)
			assert(stream.ReadUnsafeBytes()).Equals(testData[0], true)

			// ReadRPCBytes
			stream.SetReadPos(i)
			assert(stream.ReadBytes()).Equals(testData[0], true)
		}
	}
}

func Test_RPCStream_RPCArray(t *testing.T) {
	assert := NewAssert(t)
	ctx := &rpcContext{
		inner: &rpcInnerContext{
			stream:       NewRPCStream(),
			serverThread: nil,
			clientThread: nil,
		},
	}
	pubBytes := make([]byte, 40960, 40960)
	stream := NewRPCStream()

	testRandomCollection := []rpcArray{
		getRandomStringArrayTestData(0),
		getRandomStringArrayTestData(1),
		getRandomStringArrayTestData(255),
		getRandomStringArrayTestData(256),
		getRandomStringArrayTestData(523),
	}

	for _, testData := range testRandomCollection {
		// ok
		for i := 1; i < 1100; i++ {
			if i%512 < 10 || i%512 > 500 {
				bytes := pubBytes[1:i]
				copy(bytes, unusedBytes)
				stream.Reset()
				stream.PutBytes(bytes)

				assert(stream.WriteRPCArray(testData)).Equals(RPCStreamWriteOK)
				stream.SetReadPos(i)
				assert(stream.ReadRPCArray(ctx)).Equals(testData, true)
				assert(stream.IsReadFinish()).IsTrue()
			}
		}
	}

	testCollection := [][2]interface{}{
		{toRPCArray([]interface{}{}, ctx), []byte{64}},
		{toRPCArray([]interface{}{true}, ctx), []byte{65, 6, 0, 0, 0, 2}},
		{toRPCArray([]interface{}{true, false}, ctx), []byte{66, 7, 0, 0, 0, 2, 3}},
		{toRPCArray([]interface{}{
			true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true, true, true, true, true, true, true,
		}, ctx), []byte{
			94, 35, 0, 0, 0,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
		}},
		{toRPCArray([]interface{}{
			true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
		}, ctx), []byte{
			95, 40, 0, 0, 0, 31, 0, 0, 0,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
		}},
		{toRPCArray([]interface{}{
			true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
		}, ctx), []byte{
			95, 41, 0, 0, 0, 32, 0, 0, 0,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
		}},
	}

	for _, testData := range testCollection {
		// ok
		for i := 1; i < 1100; i++ {
			fillTestStream(stream, pubBytes, i, testData[0])
			targetBuffer := getTestBuffer(pubBytes, i, testData[1].([]byte))

			assert(stream.GetBuffer(), stream.writeIndex, stream.writeSeg).
				Equals(targetBuffer, len(targetBuffer)%512, len(targetBuffer)/512)

			stream.SetReadPos(i)
			assert(stream.ReadRPCArray(nil)).Equals(nilRPCArray, false)

			stream.SetReadPos(i)
			assert(stream.ReadRPCArray(ctx)).Equals(testData[0], true)
			assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
		}

		// skip test
		for i := 1; i < 1100; i++ {
			fillTestStream(stream, pubBytes, i, testData[0])
			endPos := stream.GetWritePos() - 1
			stream.SetReadPos(i)
			assert(stream.readSkipItem(endPos - 1)).Equals(-1)
			assert(stream.GetReadPos()).Equals(i)
			assert(stream.readSkipItem(endPos)).Equals(i)
			assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
			stream.SetReadPos(i)
			assert(stream.ReadRPCArray(ctx)).Equals(testData[0], true)
		}

		// error: read overflow
		for i := 1; i < 550; i++ {
			if i%512 < 10 || i%512 > 500 {
				fillTestStream(stream, pubBytes, i, testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadRPCArray(nil)).Equals(nilRPCArray, false)
					assert(stream.GetReadPos()).Equals(i)

					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadRPCArray(ctx)).Equals(nilRPCArray, false)
					assert(stream.GetReadPos()).Equals(i)
					stream.SetWritePos(writePos)
					assert(stream.ReadRPCArray(ctx)).Equals(testData[0], true)
					assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
				}
			}
		}

		// ok: have random tail bytes
		for i := 1; i < 1100; i++ {
			targetBuffer := getTestBuffer(pubBytes, i, testData[1].([]byte))
			targetBuffer = append(targetBuffer, unusedBytes[0:rand.Uint32()%1024]...)
			fillTestStreamByBuffer(stream, i, targetBuffer)

			stream.SetReadPos(i)
			assert(stream.ReadRPCArray(nil)).Equals(nilRPCArray, false)

			stream.SetReadPos(i)
			assert(stream.ReadRPCArray(ctx)).Equals(testData[0], true)
		}
	}

	// ReadRPCArray error
	rpcArray := newRPCArray(nil)
	rpcArray.Append(true)
	stream = NewRPCStream()
	stream.Write(rpcArray)
	stream.WriteNil()
	invalidCtx := &rpcContext{inner: nil}
	assert(stream.ReadRPCArray(invalidCtx)).Equals(nilRPCArray, false)
	(*stream.frames[0])[2] = 5
	assert(stream.ReadRPCArray(nil)).Equals(nilRPCArray, false)
	(*stream.frames[0])[2] = 7
	assert(stream.ReadRPCArray(nil)).Equals(nilRPCArray, false)

	// WriteRPCArray error
	stream = NewRPCStream()
	assert(stream.WriteRPCArray(nilRPCArray)).Equals(RPCStreamWriteRPCArrayIsNotAvailable)

	ctx.inner.stream = NewRPCStream()
	rpcArray = newRPCArray(ctx)
	rpcArray.Append(true)
	(*rpcArray.ctx.getCacheStream().frames[0])[1] = 13
	assert(stream.WriteRPCArray(rpcArray)).Equals(RPCStreamWriteRPCArrayError)
}

func Test_RPCStream_RPCMap(t *testing.T) {
	assert := NewAssert(t)
	ctx := &rpcContext{
		inner: &rpcInnerContext{
			stream:       NewRPCStream(),
			serverThread: nil,
			clientThread: nil,
		},
	}
	pubBytes := make([]byte, 40960, 40960)
	stream := NewRPCStream()

	testRandomCollection := []rpcMap{
		getRandomStringMapTestData(0),
		getRandomStringMapTestData(1),
		getRandomStringMapTestData(255),
		getRandomStringMapTestData(256),
		getRandomStringMapTestData(600),
	}

	for _, testData := range testRandomCollection {
		// ok
		for i := 1; i < 2; i++ {
			if i%512 < 10 || i%512 > 500 {
				bytes := pubBytes[1:i]
				copy(bytes, unusedBytes)
				stream.Reset()
				stream.PutBytes(bytes)
				assert(stream.WriteRPCMap(testData)).Equals(RPCStreamWriteOK)
				stream.SetReadPos(i)
				assert(stream.ReadRPCMap(nil)).Equals(nilRPCMap, false)

				stream.SetReadPos(i)
				assert(stream.ReadRPCMap(ctx)).Equals(testData, true)
				assert(stream.IsReadFinish()).IsTrue()
			}
		}
	}

	testCollection := [][2]interface{}{
		{toRPCMap(map[string]interface{}{}, ctx), []byte{0x60}},
		{toRPCMap(map[string]interface{}{"1": true}, ctx), []byte{0x61, 0x09, 0x00, 0x00, 0x00, 0x81, 0x31, 0x00, 0x02}},
		{toRPCMap(map[string]interface{}{
			"1": true, "2": true, "3": true, "4": true,
			"5": true, "6": true, "7": true, "8": true,
			"9": true, "a": true, "b": true, "c": true,
			"d": true, "e": true, "f": true, "g": true,
			"h": true, "i": true, "j": true, "k": true,
			"l": true, "m": true, "n": true, "o": true,
			"p": true, "q": true, "r": true, "s": true,
			"t": true, "u": true,
		}, ctx), []byte{
			0x7e, 0x7d, 0x00, 0x00, 0x00,
			0x81, 0x31, 0x00, 0x02, 0x81, 0x32, 0x00, 0x02, 0x81, 0x33, 0x00, 0x02, 0x81, 0x34, 0x00, 0x02,
			0x81, 0x35, 0x00, 0x02, 0x81, 0x36, 0x00, 0x02, 0x81, 0x37, 0x00, 0x02, 0x81, 0x38, 0x00, 0x02,
			0x81, 0x39, 0x00, 0x02, 0x81, 0x61, 0x00, 0x02, 0x81, 0x62, 0x00, 0x02, 0x81, 0x63, 0x00, 0x02,
			0x81, 0x64, 0x00, 0x02, 0x81, 0x65, 0x00, 0x02, 0x81, 0x66, 0x00, 0x02, 0x81, 0x67, 0x00, 0x02,
			0x81, 0x68, 0x00, 0x02, 0x81, 0x69, 0x00, 0x02, 0x81, 0x6a, 0x00, 0x02, 0x81, 0x6b, 0x00, 0x02,
			0x81, 0x6c, 0x00, 0x02, 0x81, 0x6d, 0x00, 0x02, 0x81, 0x6e, 0x00, 0x02, 0x81, 0x6f, 0x00, 0x02,
			0x81, 0x70, 0x00, 0x02, 0x81, 0x71, 0x00, 0x02, 0x81, 0x72, 0x00, 0x02, 0x81, 0x73, 0x00, 0x02,
			0x81, 0x74, 0x00, 0x02, 0x81, 0x75, 0x00, 0x02,
		}},
		{toRPCMap(map[string]interface{}{
			"1": true, "2": true, "3": true, "4": true,
			"5": true, "6": true, "7": true, "8": true,
			"9": true, "a": true, "b": true, "c": true,
			"d": true, "e": true, "f": true, "g": true,
			"h": true, "i": true, "j": true, "k": true,
			"l": true, "m": true, "n": true, "o": true,
			"p": true, "q": true, "r": true, "s": true,
			"t": true, "u": true, "v": true,
		}, ctx), []byte{
			0x7f, 0x85, 0x00, 0x00, 0x00, 0x1f, 0x00, 0x00, 0x00,
			0x81, 0x31, 0x00, 0x02, 0x81, 0x32, 0x00, 0x02, 0x81, 0x33, 0x00, 0x02, 0x81, 0x34, 0x00, 0x02,
			0x81, 0x35, 0x00, 0x02, 0x81, 0x36, 0x00, 0x02, 0x81, 0x37, 0x00, 0x02, 0x81, 0x38, 0x00, 0x02,
			0x81, 0x39, 0x00, 0x02, 0x81, 0x61, 0x00, 0x02, 0x81, 0x62, 0x00, 0x02, 0x81, 0x63, 0x00, 0x02,
			0x81, 0x64, 0x00, 0x02, 0x81, 0x65, 0x00, 0x02, 0x81, 0x66, 0x00, 0x02, 0x81, 0x67, 0x00, 0x02,
			0x81, 0x68, 0x00, 0x02, 0x81, 0x69, 0x00, 0x02, 0x81, 0x6a, 0x00, 0x02, 0x81, 0x6b, 0x00, 0x02,
			0x81, 0x6c, 0x00, 0x02, 0x81, 0x6d, 0x00, 0x02, 0x81, 0x6e, 0x00, 0x02, 0x81, 0x6f, 0x00, 0x02,
			0x81, 0x70, 0x00, 0x02, 0x81, 0x71, 0x00, 0x02, 0x81, 0x72, 0x00, 0x02, 0x81, 0x73, 0x00, 0x02,
			0x81, 0x74, 0x00, 0x02, 0x81, 0x75, 0x00, 0x02, 0x81, 0x76, 0x00, 0x02,
		}},
		{toRPCMap(map[string]interface{}{
			"1": true, "2": true, "3": true, "4": true,
			"5": true, "6": true, "7": true, "8": true,
			"9": true, "a": true, "b": true, "c": true,
			"d": true, "e": true, "f": true, "g": true,
			"h": true, "i": true, "j": true, "k": true,
			"l": true, "m": true, "n": true, "o": true,
			"p": true, "q": true, "r": true, "s": true,
			"t": true, "u": true, "v": true, "w": true,
		}, ctx), []byte{
			0x7f, 0x89, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00,
			0x81, 0x31, 0x00, 0x02, 0x81, 0x32, 0x00, 0x02, 0x81, 0x33, 0x00, 0x02, 0x81, 0x34, 0x00, 0x02,
			0x81, 0x35, 0x00, 0x02, 0x81, 0x36, 0x00, 0x02, 0x81, 0x37, 0x00, 0x02, 0x81, 0x38, 0x00, 0x02,
			0x81, 0x39, 0x00, 0x02, 0x81, 0x61, 0x00, 0x02, 0x81, 0x62, 0x00, 0x02, 0x81, 0x63, 0x00, 0x02,
			0x81, 0x64, 0x00, 0x02, 0x81, 0x65, 0x00, 0x02, 0x81, 0x66, 0x00, 0x02, 0x81, 0x67, 0x00, 0x02,
			0x81, 0x68, 0x00, 0x02, 0x81, 0x69, 0x00, 0x02, 0x81, 0x6a, 0x00, 0x02, 0x81, 0x6b, 0x00, 0x02,
			0x81, 0x6c, 0x00, 0x02, 0x81, 0x6d, 0x00, 0x02, 0x81, 0x6e, 0x00, 0x02, 0x81, 0x6f, 0x00, 0x02,
			0x81, 0x70, 0x00, 0x02, 0x81, 0x71, 0x00, 0x02, 0x81, 0x72, 0x00, 0x02, 0x81, 0x73, 0x00, 0x02,
			0x81, 0x74, 0x00, 0x02, 0x81, 0x75, 0x00, 0x02, 0x81, 0x76, 0x00, 0x02, 0x81, 0x77, 0x00, 0x02,
		}},
	}

	for _, testData := range testCollection {
		// ok
		for i := 1; i < 1100; i++ {
			fillTestStream(stream, pubBytes, i, testData[0])
			targetBuffer := getTestBuffer(pubBytes, i, testData[1].([]byte))

			rpcMap := testData[0].(rpcMap)
			if rpcMap.Size() <= 1 {
				assert(stream.GetBuffer(), stream.writeIndex, stream.writeSeg).
					Equals(targetBuffer, len(targetBuffer)%512, len(targetBuffer)/512)
			}

			stream.SetReadPos(i)
			assert(stream.ReadRPCMap(nil)).Equals(nilRPCMap, false)
			assert(stream.ReadNil(), stream.IsReadFinish()).IsFalse()

			stream.SetReadPos(i)
			assert(stream.ReadRPCMap(ctx)).Equals(testData[0], true)
			assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
		}

		// skip test
		for i := 1; i < 1100; i++ {
			fillTestStream(stream, pubBytes, i, testData[0])
			endPos := stream.GetWritePos() - 1
			stream.SetReadPos(i)
			assert(stream.readSkipItem(endPos - 1)).Equals(-1)
			assert(stream.GetReadPos()).Equals(i)
			assert(stream.readSkipItem(endPos)).Equals(i)
			assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
		}

		// error: read overflow
		for i := 1; i < 550; i++ {
			if i%512 < 10 || i%512 > 500 {
				fillTestStream(stream, pubBytes, i, testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadRPCMap(nil)).Equals(nilRPCMap, false)
					assert(stream.GetReadPos()).Equals(i)

					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadRPCMap(ctx)).Equals(nilRPCMap, false)
					assert(stream.GetReadPos()).Equals(i)
					stream.SetWritePos(writePos)
					assert(stream.ReadRPCMap(ctx)).Equals(testData[0], true)
					assert(stream.ReadNil(), stream.IsReadFinish()).IsTrue()
				}
			}
		}

		// ok: have random tail bytes
		for i := 1; i < 1100; i++ {
			targetBuffer := getTestBuffer(pubBytes, i, testData[1].([]byte))
			targetBuffer = append(targetBuffer, unusedBytes[0:rand.Uint32()%1024]...)
			fillTestStreamByBuffer(stream, i, targetBuffer)

			stream.SetReadPos(i)
			assert(stream.ReadRPCMap(nil)).Equals(nilRPCMap, false)

			stream.SetReadPos(i)
			assert(stream.ReadRPCMap(ctx)).Equals(testData[0], true)
		}
	}

	// ReadRPCMap error
	invalidCtx := &rpcContext{inner: nil}
	rpcMap := newRPCMap(nil)
	rpcMap.Set("a", true)

	stream = NewRPCStream()
	stream.Write(rpcMap)
	stream.WriteNil()
	assert(stream.ReadRPCMap(invalidCtx)).Equals(nilRPCMap, false)
	(*stream.frames[0])[2] = 8
	assert(stream.ReadRPCMap(nil)).Equals(nilRPCMap, false)
	(*stream.frames[0])[2] = 10
	assert(stream.ReadRPCMap(nil)).Equals(nilRPCMap, false)
	(*stream.frames[0])[2] = 9
	(*stream.frames[0])[6] = 13
	assert(stream.ReadRPCMap(nil)).Equals(nilRPCMap, false)
	for i := 0; i < 16; i++ {
		rpcMap.Set(strconv.Itoa(i), true)
	}
	stream = NewRPCStream()
	stream.Write(rpcMap)
	stream.WriteNil()
	assert(stream.ReadRPCMap(invalidCtx)).Equals(nilRPCMap, false)
	(*stream.frames[0])[2] = 8
	assert(stream.ReadRPCMap(nil)).Equals(nilRPCMap, false)
	(*stream.frames[0])[2] = 10
	assert(stream.ReadRPCMap(nil)).Equals(nilRPCMap, false)
	(*stream.frames[0])[2] = 9
	(*stream.frames[0])[6] = 13
	assert(stream.ReadRPCMap(nil)).Equals(nilRPCMap, false)

	// WriteRPCMap error
	stream = NewRPCStream()
	assert(stream.WriteRPCMap(nilRPCMap)).Equals(RPCStreamWriteRPCMapIsNotAvailable)

	ctx.inner.stream = NewRPCStream()
	rpcMap = newRPCMap(ctx)
	rpcMap.Set("t", true)
	(*rpcMap.ctx.getCacheStream().frames[0])[1] = 13
	assert(stream.WriteRPCMap(rpcMap)).Equals(RPCStreamWriteRPCMapError)

	for i := 0; i < 16; i++ {
		rpcMap.Set(strconv.Itoa(i), true)
	}
	assert(stream.WriteRPCMap(rpcMap)).Equals(RPCStreamWriteRPCMapError)
}

func Test_RPCStream_ReadAndWrite(t *testing.T) {
	assert := NewAssert(t)

	stream := NewRPCStream()
	stream.PutBytes([]byte{4})
	assert(stream.Read(nil)).Equals(float64(0), true)

	stream = NewRPCStream()
	stream.PutBytes([]byte{12})
	assert(stream.Read(nil)).Equals(nil, false)

	stream = NewRPCStream()
	stream.PutBytes([]byte{13})
	assert(stream.Read(nil)).Equals(nil, false)

	stream = NewRPCStream()
	stream.PutBytes([]byte{54})
	assert(stream.Read(nil)).Equals(uint64(0), true)

	stream = NewRPCStream()

	assert(stream.Write(int8(0))).Equals(RPCStreamWriteOK)
	assert(stream.Write(int16(0))).Equals(RPCStreamWriteOK)
	assert(stream.Write(int32(0))).Equals(RPCStreamWriteOK)
	assert(stream.Write(uint(0))).Equals(RPCStreamWriteOK)
	assert(stream.Write(uint8(0))).Equals(RPCStreamWriteOK)
	assert(stream.Write(uint16(0))).Equals(RPCStreamWriteOK)
	assert(stream.Write(uint32(0))).Equals(RPCStreamWriteOK)
	assert(stream.Write(uint64(0))).Equals(RPCStreamWriteOK)
	assert(stream.Write(float32(0))).Equals(RPCStreamWriteOK)
	assert(stream.Write(make(chan bool))).Equals(RPCStreamWriteUnsupportedType)
}
