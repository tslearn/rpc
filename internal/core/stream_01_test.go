package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"testing"
)

func TestStream_ReadRTArray(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["array"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				rtArray, err := stream.ReadRTArray(streamTestRuntime)
				assert(err).IsNil()
				assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.writeRTArray(rtArray)
				assert(stream.ReadArray()).Equal(testData[0], nil)
				streamTestRuntime.thread.rtStream.SetWritePosToBodyStart()
				stream.Release()
			}
		}
	})

	t.Run("test readIndex overflow (outer stream)", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["array"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadRTArray(streamTestRuntime)).
						Equal(RTArray{}, errors.ErrStreamIsBroken)
					assert(stream.GetReadPos()).Equal(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test readIndex overflow (runtime stream)", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["array"] {
			for _, i := range testRange {
				stream := streamTestRuntime.thread.rtStream
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadRTArray(streamTestRuntime)).
						Equal(RTArray{}, errors.ErrStreamIsBroken)
					assert(stream.GetReadPos()).Equal(i)
				}
				stream.Reset()
			}
		}
	})

	t.Run("test type not match", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadRTArray(streamTestRuntime)).
				Equal(RTArray{}, errors.ErrStreamIsBroken)
			assert(stream.GetReadPos()).Equal(i)
			stream.Release()
		}
	})

	t.Run("error in stream", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["array"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				if len(testData[0].(Array)) > 0 {
					stream.SetWritePos(stream.GetWritePos() - 1)
					stream.PutBytes([]byte{13})
					assert(stream.ReadRTArray(streamTestRuntime)).
						Equal(RTArray{}, errors.ErrStreamIsBroken)
					assert(stream.GetReadPos()).Equal(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("error in stream", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{0x41, 0x07, 0x00, 0x00, 0x00, 0x02, 0x02})
			assert(stream.ReadRTArray(streamTestRuntime)).
				Equal(RTArray{}, errors.ErrStreamIsBroken)
			assert(stream.GetReadPos()).Equal(i)
			stream.Release()
		}
	})

	t.Run("runtime is not available", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		type R = Runtime
		s := ""
		f := base.GetFileLine
		assert(stream.ReadRTArray((func() R { s = f(0); return R{} })())).
			Equal(RTArray{}, errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(s))
	})

}

func TestStream_ReadRTMap(t *testing.T) {
	t.Run("runtime is not available", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		type R = Runtime
		s := ""
		f := base.GetFileLine
		assert(stream.ReadRTMap((func() R { s = f(0); return R{} })())).
			Equal(RTMap{}, errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(s))
	})
}

func TestStream_ReadRTValue(t *testing.T) {
	t.Run("runtime is not available", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		type R = Runtime
		s := ""
		f := base.GetFileLine
		assert(stream.ReadRTValue((func() R { s = f(0); return R{} })())).Equal(
			RTValue{err: errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(s)},
		)
	})
}
