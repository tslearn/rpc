package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"testing"
)

func TestStream_ReadUnsafeString(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["string"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])

				notSafe := true
				sLen := len(testData[0].(string))
				if sLen == 0 {
					notSafe = false
				} else if sLen < 62 {
					notSafe = stream.readIndex+sLen < streamBlockSize-2
				} else {
					notSafe =
						(stream.readIndex+5)%streamBlockSize+sLen < streamBlockSize-1
				}
				assert(stream.readUnsafeString()).Equal(testData[0], !notSafe, nil)
				assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})

	t.Run("test readIndex overflow", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["string"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.readUnsafeString()).
						Equal("", true, errors.ErrStreamIsBroken)
					assert(stream.GetReadPos()).Equal(i)
				}
				stream.Release()
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
			assert(stream.readUnsafeString()).
				Equal("", true, errors.ErrStreamIsBroken)
			assert(stream.GetReadPos()).Equal(i)
			stream.Release()
		}
	})

	t.Run("read tail is not zero", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["string"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				stream.SetWritePos(stream.GetWritePos() - 1)
				stream.PutBytes([]byte{1})
				assert(stream.readUnsafeString()).
					Equal("", true, errors.ErrStreamIsBroken)
				assert(stream.GetReadPos()).Equal(i)
				stream.Release()
			}
		}
	})

	t.Run("read string utf8 error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.PutBytes([]byte{
			0x9E, 0xFF, 0x9F, 0x98, 0x80, 0xE2, 0x98, 0x98, 0xEF, 0xB8,
			0x8F, 0xF0, 0x9F, 0x80, 0x84, 0xEF, 0xB8, 0x8F, 0xC2, 0xA9,
			0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x8C, 0x88, 0xF0, 0x9F, 0x8E,
			0xA9, 0x00,
		})
		assert(stream.readUnsafeString()).Equal("", true, errors.ErrStreamIsBroken)
		assert(stream.GetReadPos()).Equal(streamPosBody)
		stream.Release()
	})

	t.Run("read string utf8 error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.PutBytes([]byte{
			0xBF, 0x6D, 0x00, 0x00, 0x00, 0xFF, 0x9F, 0x98, 0x80, 0xE2,
			0x98, 0x98, 0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x80, 0x84, 0xEF,
			0xB8, 0x8F, 0xC2, 0xA9, 0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x8C,
			0x88, 0xF0, 0x9F, 0x8E, 0xA9, 0xF0, 0x9F, 0x98, 0x9B, 0xF0,
			0x9F, 0x91, 0xA9, 0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA9,
			0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA6, 0xF0, 0x9F, 0x91,
			0xA8, 0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA9, 0xE2, 0x80,
			0x8D, 0xF0, 0x9F, 0x91, 0xA6, 0xE2, 0x80, 0x8D, 0xF0, 0x9F,
			0x91, 0xA6, 0xF0, 0x9F, 0x91, 0xBC, 0xF0, 0x9F, 0x97, 0xA3,
			0xF0, 0x9F, 0x91, 0x91, 0xF0, 0x9F, 0x91, 0x9A, 0xF0, 0x9F,
			0x91, 0xB9, 0xF0, 0x9F, 0x91, 0xBA, 0xF0, 0x9F, 0x8C, 0xB3,
			0xF0, 0x9F, 0x8D, 0x8A, 0x00,
		})
		assert(stream.readUnsafeString()).Equal("", true, errors.ErrStreamIsBroken)
		assert(stream.GetReadPos()).Equal(streamPosBody)
		stream.Release()
	})

	t.Run("read string length error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.PutBytes([]byte{
			0xBF, 0x2F, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x00,
		})
		assert(stream.readUnsafeString()).Equal("", true, errors.ErrStreamIsBroken)
		assert(stream.GetReadPos()).Equal(streamPosBody)
		stream.Release()
	})
}

func TestStream_ReadBytes(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["bytes"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				assert(stream.ReadBytes()).Equal(testData[0], nil)
				assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})

	t.Run("test readIndex overflow", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["bytes"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadBytes()).Equal(Bytes{}, errors.ErrStreamIsBroken)
					assert(stream.GetReadPos()).Equal(i)
				}
				stream.Release()
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
			assert(stream.ReadBytes()).Equal(Bytes{}, errors.ErrStreamIsBroken)
			assert(stream.GetReadPos()).Equal(i)
			stream.Release()
		}
	})

	t.Run("read string length error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.PutBytes([]byte{
			0xFF, 0x2F, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		})
		assert(stream.ReadBytes()).Equal(Bytes{}, errors.ErrStreamIsBroken)
		assert(stream.GetReadPos()).Equal(streamPosBody)
		stream.Release()
	})
}

//func TestStream_ReadArray(t *testing.T) {
//  assert := base.NewAssert(t)
//
//  for _, testData := range streamTestCollections["array"] {
//    // ok
//    for i := streamPosBody; i < 530; i++ {
//      for j := 1; j < 530; j++ {
//        // skip for performance
//        if j > 10 && j < 500 {
//          continue
//        }
//
//        stream := NewStream()
//        stream.SetWritePos(i)
//        stream.SetReadPos(i)
//        stream.Write(testData[0])
//        assert(stream.ReadArray()).Equal(testData[0].(Array), true)
//        assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
//        stream.Release()
//      }
//    }
//
//    // overflow
//    for i := streamPosBody; i < 550; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.Write(testData[0])
//      writePos := stream.GetWritePos()
//      for idx := i; idx < writePos-1; idx++ {
//        stream.SetReadPos(i)
//        stream.SetWritePos(idx)
//        assert(stream.ReadArray()).Equal(Array(nil), false)
//        assert(stream.GetReadPos()).Equal(i)
//      }
//      stream.Release()
//    }
//
//    // type not match
//    for i := streamPosBody; i < 550; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.PutBytes([]byte{13})
//      assert(stream.ReadArray()).Equal(Array(nil), false)
//      assert(stream.GetReadPos()).Equal(i)
//      stream.Release()
//    }
//
//    // error in stream
//    for i := streamPosBody; i < 550; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.Write(testData[0])
//      if len(testData[0].(Array)) > 0 {
//        stream.SetWritePos(stream.GetWritePos() - 1)
//        stream.PutBytes([]byte{13})
//        assert(stream.ReadArray()).Equal(Array(nil), false)
//        assert(stream.GetReadPos()).Equal(i)
//      }
//      stream.Release()
//    }
//
//    // error in stream
//    for i := streamPosBody; i < 550; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.PutBytes([]byte{0x41, 0x07, 0x00, 0x00, 0x00, 0x02, 0x02})
//      assert(stream.ReadArray()).Equal(Array(nil), false)
//      assert(stream.GetReadPos()).Equal(i)
//      stream.Release()
//    }
//  }
//}
//
//func TestStream_ReadMap(t *testing.T) {
//  assert := base.NewAssert(t)
//
//  for _, testData := range streamTestCollections["map"] {
//    // ok
//    for i := streamPosBody; i < 530; i++ {
//      for j := 1; j < 530; j++ {
//        // skip for performance
//        if j > 10 && j < 500 {
//          continue
//        }
//        stream := NewStream()
//        stream.SetWritePos(i)
//        stream.SetReadPos(i)
//        stream.Write(testData[0])
//        assert(stream.ReadMap()).Equal(testData[0], true)
//        assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
//        stream.Release()
//      }
//    }
//
//    // overflow
//    for i := streamPosBody; i < 530; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.Write(testData[0])
//      writePos := stream.GetWritePos()
//      for idx := i; idx < writePos-1; idx++ {
//        stream.SetReadPos(i)
//        stream.SetWritePos(idx)
//        assert(stream.ReadMap()).Equal(Map(nil), false)
//        assert(stream.GetReadPos()).Equal(i)
//      }
//      stream.Release()
//    }
//
//    // type not match
//    for i := streamPosBody; i < 550; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.PutBytes([]byte{13})
//      assert(stream.ReadMap()).Equal(Map(nil), false)
//      assert(stream.GetReadPos()).Equal(i)
//      stream.Release()
//    }
//
//    // error in stream
//    for i := streamPosBody; i < 550; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.Write(testData[0])
//      if len(testData[0].(Map)) > 0 {
//        stream.SetWritePos(stream.GetWritePos() - 1)
//        stream.PutBytes([]byte{13})
//        assert(stream.ReadMap()).Equal(Map(nil), false)
//        assert(stream.GetReadPos()).Equal(i)
//      }
//      stream.Release()
//    }
//
//    // error in stream, length error
//    for i := streamPosBody; i < 550; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.PutBytes([]byte{
//        0x61, 0x0A, 0x00, 0x00, 0x00, 0x81, 0x31, 0x00, 0x02, 0x02,
//      })
//      assert(stream.ReadMap()).Equal(Map(nil), false)
//      assert(stream.GetReadPos()).Equal(i)
//      stream.Release()
//    }
//
//    // error in stream, key error
//    for i := streamPosBody; i < 550; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.Write(testData[0])
//      wPos := stream.GetWritePos()
//      mapSize := len(testData[0].(Map))
//
//      if mapSize > 30 {
//        stream.SetWritePos(i + 9)
//        stream.PutBytes([]byte{13})
//        stream.SetWritePos(wPos)
//        assert(stream.ReadMap()).Equal(Map(nil), false)
//        assert(stream.GetReadPos()).Equal(i)
//      } else if mapSize > 0 {
//        stream.SetWritePos(i + 5)
//        stream.PutBytes([]byte{13})
//        stream.SetWritePos(wPos)
//        assert(stream.ReadMap()).Equal(Map(nil), false)
//        assert(stream.GetReadPos()).Equal(i)
//      }
//      stream.Release()
//    }
//  }
//}
//
//func TestStream_Read(t *testing.T) {
//  assert := base.NewAssert(t)
//
//  testCollections := make([][2]interface{}, 0)
//
//  for key := range streamTestCollections {
//    testCollections = append(testCollections, streamTestCollections[key]...)
//  }
//
//  for _, item := range testCollections {
//    stream := NewStream()
//    stream.PutBytes(item[1].([]byte))
//    if base.IsNil(item[0]) {
//      assert(stream.Read()).Equal(nil, true)
//    } else {
//      assert(stream.Read()).Equal(item[0], true)
//    }
//  }
//
//  stream := NewStream()
//  stream.PutBytes([]byte{12})
//  assert(stream.Read()).Equal(nil, false)
//
//  stream = NewStream()
//  stream.PutBytes([]byte{13})
//  assert(stream.Read()).Equal(nil, false)
//}
//
//func BenchmarkRPCStream_ReadString(b *testing.B) {
//  stream := NewStream()
//  stream.WriteString("#.user.login:isUserARight")
//
//  for i := 0; i < b.N; i++ {
//    stream.SetReadPos(streamPosBody)
//    stream.ReadString()
//  }
//}
