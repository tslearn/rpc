package core

import "testing"

func TestRpcStream_basic(t *testing.T) {
	assert := newAssert(t)

	// test rpcStreamCache
	for i := 0; i < 10000; i++ {
		stream := rpcStreamCache.Get().(*rpcStream)
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
