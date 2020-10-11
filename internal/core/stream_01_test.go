package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"testing"
)

func getTestDepthArray(depth int) [2]interface{} {
	if depth <= 1 {
		return [2]interface{}{true, ""}
	} else {
		ret := getTestDepthArray(depth - 1)
		return [2]interface{}{Array{ret[0]}, "[0]" + ret[1].(string)}
	}
}

func getTestDepthMap(depth int) [2]interface{} {
	if depth <= 1 {
		return [2]interface{}{true, ""}
	} else {
		ret := getTestDepthMap(depth - 1)
		return [2]interface{}{Map{"a": ret[0]}, "[\"a\"]" + ret[1].(string)}
	}
}

var streamTestRuntime = Runtime{id: 0, thread: getFakeThread(true)}

var streamTestWriteCollections = map[string][][2]interface{}{
	"array": {
		{
			Array{true, true, true, make(chan bool), true},
			"[3] type(chan bool) is not supported",
		},
		{
			getTestDepthArray(65)[0],
			getTestDepthArray(65)[1].(string) + " overflows",
		},
		{
			getTestDepthArray(64)[0],
			StreamWriteOK,
		},
	},
	"map": {
		{
			Map{"0": 0, "1": make(chan bool)},
			"[\"1\"] type(chan bool) is not supported",
		},
		{
			getTestDepthMap(65)[0],
			getTestDepthMap(65)[1].(string) + " overflows",
		},
		{
			getTestDepthMap(64)[0],
			StreamWriteOK,
		},
	},
	"rtArray": {
		{
			RTArray{},
			" is not available",
		},
		{
			RTArray{rt: streamTestRuntime, items: nil},
			" is not available",
		},
		{
			RTArray{rt: streamTestRuntime, items: []posRecord{0}},
			" is not available",
		},
	},
	"rtMap": {
		{
			RTMap{},
			" is not available",
		},
		{
			RTMap{rt: streamTestRuntime, items: nil},
			" is not available",
		},
		{
			RTMap{rt: streamTestRuntime, items: []mapItem{mapItem{"key", 0}}},
			" is not available",
		},
		{
			RTMap{rt: streamTestRuntime, largeMap: map[string]posRecord{"key": 0}},
			" is not available",
		},
	},
	"rtValue": {
		{
			RTValue{},
			" is not available",
		},
		{
			RTValue{rt: streamTestRuntime, pos: 0},
			" is not available",
		},
		{
			RTValue{rt: streamTestRuntime, pos: streamPosBody},
			" is not available",
		},
	},
	"value": {
		{nil, "value is nil"},
		{true, StreamWriteOK},
		{0, StreamWriteOK},
		{int8(0), StreamWriteOK},
		{int16(0), StreamWriteOK},
		{int32(0), StreamWriteOK},
		{int64(0), StreamWriteOK},
		{uint(0), StreamWriteOK},
		{uint8(0), StreamWriteOK},
		{uint16(0), StreamWriteOK},
		{uint32(0), StreamWriteOK},
		{uint64(0), StreamWriteOK},
		{float32(0), StreamWriteOK},
		{float64(0), StreamWriteOK},
		{"", StreamWriteOK},
		{[]byte{}, StreamWriteOK},
		{Array{}, StreamWriteOK},
		{Map{}, StreamWriteOK},
		{RTArray{rt: streamTestRuntime, items: []posRecord{}}, StreamWriteOK},
		{RTMap{rt: streamTestRuntime, items: []mapItem{}}, StreamWriteOK},
		{RTValue{}, "value is not available"},
		{RTArray{}, "value is not available"},
		{RTMap{}, "value is not available"},
		{
			Array{true, true, true, make(chan bool), true},
			"value[3] type(chan bool) is not supported",
		},
		{
			getTestDepthArray(65)[0],
			"value" + getTestDepthArray(65)[1].(string) + " overflows",
		},
		{
			Map{"0": 0, "1": make(chan bool)},
			"value[\"1\"] type(chan bool) is not supported",
		},
		{
			getTestDepthMap(65)[0],
			"value" + getTestDepthMap(65)[1].(string) + " overflows",
		},
	},
}

func TestStream_WriteBool(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["bool"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.WriteBool(testData[0].(bool))
				assert(stream.GetBuffer()[i:]).Equal(testData[1])
				assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})
}

func TestStream_WriteFloat64(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["float64"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.WriteFloat64(testData[0].(float64))
				assert(stream.GetBuffer()[i:]).Equal(testData[1])
				assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})
}

func TestStream_WriteInt64(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["int64"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.WriteInt64(testData[0].(int64))
				assert(stream.GetBuffer()[i:]).Equal(testData[1])
				assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})
}

func TestStream_WriteUInt64(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["uint64"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.WriteUint64(testData[0].(uint64))
				assert(stream.GetBuffer()[i:]).Equal(testData[1])
				assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})
}

func TestStream_WriteString(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["string"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.WriteString(testData[0].(string))
				assert(stream.GetBuffer()[i:]).Equal(testData[1])
				assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})
}

func TestStream_WriteBytes(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["bytes"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.WriteBytes(testData[0].([]byte))
				assert(stream.GetBuffer()[i:]).Equal(testData[1])
				assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})
}

func TestStream_writeArray(t *testing.T) {
	t.Run("test write failed", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestWriteCollections["array"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				assert(stream.writeArray(testData[0].(Array), 64)).Equal(testData[1])
				if testData[1].(string) != StreamWriteOK {
					assert(stream.GetWritePos()).Equal(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test write ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["array"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				assert(stream.writeArray(testData[0].(Array), 64)).Equal(StreamWriteOK)
				assert(stream.GetBuffer()[i:]).Equal(testData[1])
				assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})
}

func TestStream_writeMap(t *testing.T) {
	t.Run("test write failed", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestWriteCollections["map"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				assert(stream.writeMap(testData[0].(Map), 64)).Equal(testData[1])
				if testData[1].(string) != StreamWriteOK {
					assert(stream.GetWritePos()).Equal(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test write ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["map"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				assert(stream.writeMap(testData[0].(Map), 64)).Equal(StreamWriteOK)
				assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
				stream.SetReadPos(i)
				assert(stream.ReadMap()).Equal(testData[0].(Map), nil)
				stream.Release()
			}
		}
	})
}

func TestStream_writeRTArray(t *testing.T) {
	t.Run("test write failed", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestWriteCollections["rtArray"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				assert(stream.writeRTArray(testData[0].(RTArray))).Equal(testData[1])
				if testData[1].(string) != StreamWriteOK {
					assert(stream.GetWritePos()).Equal(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test write ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["array"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				assert(stream.writeArray(testData[0].(Array), 64)).Equal(StreamWriteOK)
				rtArray, err := stream.ReadRTArray(streamTestRuntime)
				assert(err).IsNil()
				stream.SetWritePos(i)
				assert(stream.writeRTArray(rtArray)).Equal(StreamWriteOK)
				assert(stream.GetBuffer()[i:]).Equal(testData[1])
				assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
				streamTestRuntime.thread.rtStream.SetWritePosToBodyStart()
				stream.Release()
			}
		}
	})
}

func TestStream_writeRTMap(t *testing.T) {
	t.Run("test write failed", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestWriteCollections["rtMap"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				assert(stream.writeRTMap(testData[0].(RTMap))).Equal(testData[1])
				if testData[1].(string) != StreamWriteOK {
					assert(stream.GetWritePos()).Equal(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test write ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["map"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				assert(stream.writeMap(testData[0].(Map), 64)).Equal(StreamWriteOK)
				rtMap, err := stream.ReadRTMap(streamTestRuntime)
				assert(err).IsNil()
				stream.SetWritePos(i)
				assert(stream.writeRTMap(rtMap)).Equal(StreamWriteOK)
				assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
				stream.SetReadPos(i)
				assert(stream.ReadMap()).Equal(testData[0].(Map), nil)
				streamTestRuntime.thread.rtStream.SetWritePosToBodyStart()
				stream.Release()
			}
		}
	})
}

func TestStream_writeRTValue(t *testing.T) {
	t.Run("test write failed", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestWriteCollections["rtValue"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				assert(stream.writeRTValue(testData[0].(RTValue))).Equal(testData[1])
				if testData[1].(string) != StreamWriteOK {
					assert(stream.GetWritePos()).Equal(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test write ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for key := range streamTestSuccessCollections {
			for _, testData := range streamTestSuccessCollections[key] {
				for _, i := range testRange {
					stream := NewStream()
					stream.SetWritePos(i)
					stream.SetReadPos(i)
					assert(stream.Write(testData[0])).Equal(StreamWriteOK)
					rtValue := stream.ReadRTValue(streamTestRuntime)
					stream.SetWritePos(i)
					assert(stream.writeRTValue(rtValue)).Equal(StreamWriteOK)
					assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
					stream.SetReadPos(i)
					assert(stream.Read()).Equal(testData[0], nil)
					streamTestRuntime.thread.rtStream.SetWritePosToBodyStart()
					stream.Release()
				}
			}
		}
	})
}

func TestStream_Write(t *testing.T) {
	t.Run("test write", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestWriteCollections["value"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				assert(stream.Write(testData[0])).Equal(testData[1])
				if testData[1].(string) != StreamWriteOK {
					assert(stream.GetWritePos()).Equal(i)
				}
				stream.Release()
			}
		}
	})
}

//func TestStream_ReadNil(t *testing.T) {
//  assert := base.NewAssert(t)
//
//  for _, testData := range streamTestCollections["nil"] {
//    // ok
//    for i := streamPosBody; i < 2*streamBlockSize+40; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.Write(testData[0])
//
//      assert(stream.ReadNil()).Equal(true)
//      assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
//      stream.Release()
//    }
//
//    // overflow
//    for i := streamPosBody; i < streamBlockSize+20; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.Write(testData[0])
//      writePos := stream.GetWritePos()
//      for idx := i; idx < writePos-1; idx++ {
//        stream.SetReadPos(i)
//        stream.SetWritePos(idx)
//        assert(stream.ReadNil()).IsFalse()
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
//      assert(stream.ReadNil()).IsFalse()
//      assert(stream.GetReadPos()).Equal(i)
//      stream.Release()
//    }
//  }
//}
//
//func TestStream_ReadBool(t *testing.T) {
//  assert := base.NewAssert(t)
//
//  for _, testData := range streamTestCollections["bool"] {
//    // ok
//    for i := streamPosBody; i < 1100; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.Write(testData[0])
//      assert(stream.ReadBool()).Equal(testData[0], true)
//      assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
//      stream.Release()
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
//        assert(stream.ReadBool()).Equal(false, false)
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
//      assert(stream.ReadBool()).Equal(false, false)
//      assert(stream.GetReadPos()).Equal(i)
//      stream.Release()
//    }
//  }
//}
//
//func TestStream_ReadFloat64(t *testing.T) {
//  assert := base.NewAssert(t)
//
//  for _, testData := range streamTestCollections["float64"] {
//    // ok
//    for i := streamPosBody; i < 1100; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.Write(testData[0])
//      assert(stream.ReadFloat64()).Equal(testData[0], true)
//      assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
//      stream.Release()
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
//        assert(stream.ReadFloat64()).Equal(float64(0), false)
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
//      assert(stream.ReadFloat64()).Equal(float64(0), false)
//      assert(stream.GetReadPos()).Equal(i)
//      stream.Release()
//    }
//  }
//}
//
//func TestStream_ReadInt64(t *testing.T) {
//  assert := base.NewAssert(t)
//
//  for _, testData := range streamTestCollections["int64"] {
//    // ok
//    for i := streamPosBody; i < 1100; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.Write(testData[0])
//      assert(stream.ReadInt64()).Equal(testData[0], true)
//      assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
//      stream.Release()
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
//        assert(stream.ReadInt64()).Equal(int64(0), false)
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
//      assert(stream.ReadInt64()).Equal(int64(0), false)
//      assert(stream.GetReadPos()).Equal(i)
//      stream.Release()
//    }
//  }
//}
//
//func TestStream_ReadUint64(t *testing.T) {
//  assert := base.NewAssert(t)
//
//  for _, testData := range streamTestCollections["uint64"] {
//    // ok
//    for i := streamPosBody; i < 1100; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.Write(testData[0])
//      assert(stream.ReadUint64()).Equal(testData[0], true)
//      assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
//      stream.Release()
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
//        assert(stream.ReadUint64()).Equal(uint64(0), false)
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
//      assert(stream.ReadUint64()).Equal(uint64(0), false)
//      assert(stream.GetReadPos()).Equal(i)
//      stream.Release()
//    }
//  }
//}
//
//func TestStream_ReadString(t *testing.T) {
//  assert := base.NewAssert(t)
//
//  for _, testData := range streamTestCollections["string"] {
//    // ok
//    for i := streamPosBody; i < 1100; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.Write(testData[0])
//      assert(stream.ReadString()).Equal(testData[0], true)
//      assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
//      stream.Release()
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
//        assert(stream.ReadString()).Equal("", false)
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
//      assert(stream.ReadString()).Equal("", false)
//      assert(stream.GetReadPos()).Equal(i)
//      stream.Release()
//    }
//
//    // read tail is not zero
//    for i := streamPosBody; i < 550; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.Write(testData[0])
//      stream.SetWritePos(stream.GetWritePos() - 1)
//      stream.PutBytes([]byte{1})
//      assert(stream.ReadString()).Equal("", false)
//      assert(stream.GetReadPos()).Equal(i)
//      stream.Release()
//    }
//  }
//
//  // read string utf8 error
//  stream1 := NewStream()
//  stream1.PutBytes([]byte{
//    0x9E, 0xFF, 0x9F, 0x98, 0x80, 0xE2, 0x98, 0x98, 0xEF, 0xB8,
//    0x8F, 0xF0, 0x9F, 0x80, 0x84, 0xEF, 0xB8, 0x8F, 0xC2, 0xA9,
//    0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x8C, 0x88, 0xF0, 0x9F, 0x8E,
//    0xA9, 0x00,
//  })
//  assert(stream1.ReadString()).Equal("", false)
//  assert(stream1.GetReadPos()).Equal(streamPosBody)
//
//  // read string utf8 error
//  stream2 := NewStream()
//  stream2.PutBytes([]byte{
//    0xBF, 0x6D, 0x00, 0x00, 0x00, 0xFF, 0x9F, 0x98, 0x80, 0xE2,
//    0x98, 0x98, 0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x80, 0x84, 0xEF,
//    0xB8, 0x8F, 0xC2, 0xA9, 0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x8C,
//    0x88, 0xF0, 0x9F, 0x8E, 0xA9, 0xF0, 0x9F, 0x98, 0x9B, 0xF0,
//    0x9F, 0x91, 0xA9, 0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA9,
//    0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA6, 0xF0, 0x9F, 0x91,
//    0xA8, 0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA9, 0xE2, 0x80,
//    0x8D, 0xF0, 0x9F, 0x91, 0xA6, 0xE2, 0x80, 0x8D, 0xF0, 0x9F,
//    0x91, 0xA6, 0xF0, 0x9F, 0x91, 0xBC, 0xF0, 0x9F, 0x97, 0xA3,
//    0xF0, 0x9F, 0x91, 0x91, 0xF0, 0x9F, 0x91, 0x9A, 0xF0, 0x9F,
//    0x91, 0xB9, 0xF0, 0x9F, 0x91, 0xBA, 0xF0, 0x9F, 0x8C, 0xB3,
//    0xF0, 0x9F, 0x8D, 0x8A, 0x00,
//  })
//  assert(stream2.ReadString()).Equal("", false)
//  assert(stream2.GetReadPos()).Equal(streamPosBody)
//
//  // read string length error
//  stream3 := NewStream()
//  stream3.PutBytes([]byte{
//    0xBF, 0x2F, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
//    0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
//    0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
//    0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
//    0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
//    0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
//    0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x00,
//  })
//  assert(stream3.ReadString()).Equal("", false)
//  assert(stream3.GetReadPos()).Equal(streamPosBody)
//}
//
//func TestStream_ReadUnsafeString(t *testing.T) {
//  assert := base.NewAssert(t)
//
//  for _, testData := range streamTestCollections["string"] {
//    // ok
//    for i := streamPosBody; i < 1100; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.Write(testData[0])
//
//      notSafe := true
//      if len(testData[0].(string)) == 0 {
//        notSafe = false
//      } else if len(testData[0].(string)) < 62 {
//        notSafe = stream.readIndex+len(testData[0].(string)) < streamBlockSize-2
//      } else {
//        notSafe = (stream.readIndex+5)%streamBlockSize+len(testData[0].(string)) < streamBlockSize-1
//      }
//      assert(stream.readUnsafeString()).Equal(testData[0], !notSafe, true)
//      assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
//      stream.Release()
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
//        assert(stream.readUnsafeString()).Equal("", true, false)
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
//      assert(stream.readUnsafeString()).Equal("", true, false)
//      assert(stream.GetReadPos()).Equal(i)
//      stream.Release()
//    }
//
//    // read tail is not zero
//    for i := streamPosBody; i < 550; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.Write(testData[0])
//      stream.SetWritePos(stream.GetWritePos() - 1)
//      stream.PutBytes([]byte{1})
//      assert(stream.readUnsafeString()).Equal("", true, false)
//      assert(stream.GetReadPos()).Equal(i)
//      stream.Release()
//    }
//  }
//
//  // read string utf8 error
//  stream1 := NewStream()
//  stream1.PutBytes([]byte{
//    0x9E, 0xFF, 0x9F, 0x98, 0x80, 0xE2, 0x98, 0x98, 0xEF, 0xB8,
//    0x8F, 0xF0, 0x9F, 0x80, 0x84, 0xEF, 0xB8, 0x8F, 0xC2, 0xA9,
//    0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x8C, 0x88, 0xF0, 0x9F, 0x8E,
//    0xA9, 0x00,
//  })
//  assert(stream1.readUnsafeString()).Equal("", true, false)
//  assert(stream1.GetReadPos()).Equal(streamPosBody)
//
//  // read string utf8 error
//  stream2 := NewStream()
//  stream2.PutBytes([]byte{
//    0xBF, 0x6D, 0x00, 0x00, 0x00, 0xFF, 0x9F, 0x98, 0x80, 0xE2,
//    0x98, 0x98, 0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x80, 0x84, 0xEF,
//    0xB8, 0x8F, 0xC2, 0xA9, 0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x8C,
//    0x88, 0xF0, 0x9F, 0x8E, 0xA9, 0xF0, 0x9F, 0x98, 0x9B, 0xF0,
//    0x9F, 0x91, 0xA9, 0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA9,
//    0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA6, 0xF0, 0x9F, 0x91,
//    0xA8, 0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA9, 0xE2, 0x80,
//    0x8D, 0xF0, 0x9F, 0x91, 0xA6, 0xE2, 0x80, 0x8D, 0xF0, 0x9F,
//    0x91, 0xA6, 0xF0, 0x9F, 0x91, 0xBC, 0xF0, 0x9F, 0x97, 0xA3,
//    0xF0, 0x9F, 0x91, 0x91, 0xF0, 0x9F, 0x91, 0x9A, 0xF0, 0x9F,
//    0x91, 0xB9, 0xF0, 0x9F, 0x91, 0xBA, 0xF0, 0x9F, 0x8C, 0xB3,
//    0xF0, 0x9F, 0x8D, 0x8A, 0x00,
//  })
//  assert(stream2.readUnsafeString()).Equal("", true, false)
//  assert(stream2.GetReadPos()).Equal(streamPosBody)
//
//  // read string length error
//  stream3 := NewStream()
//  stream3.PutBytes([]byte{
//    0xBF, 0x2F, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
//    0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
//    0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
//    0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
//    0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
//    0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
//    0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x00,
//  })
//  assert(stream3.readUnsafeString()).Equal("", true, false)
//  assert(stream3.GetReadPos()).Equal(streamPosBody)
//}
//
//func TestStream_ReadBytes(t *testing.T) {
//  assert := base.NewAssert(t)
//
//  for _, testData := range streamTestCollections["bytes"] {
//    // ok
//    for i := streamPosBody; i < 1100; i++ {
//      stream := NewStream()
//      stream.SetWritePos(i)
//      stream.SetReadPos(i)
//      stream.Write(testData[0])
//      assert(stream.ReadBytes()).Equal(testData[0], true)
//      assert(stream.GetWritePos()).Equal(len(testData[1].([]byte)) + i)
//      stream.Release()
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
//        assert(stream.ReadBytes()).Equal(Bytes(nil), false)
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
//      assert(stream.ReadBytes()).Equal(Bytes(nil), false)
//      assert(stream.GetReadPos()).Equal(i)
//      stream.Release()
//    }
//  }
//
//  // read bytes length error
//  stream1 := NewStream()
//  stream1.PutBytes([]byte{
//    0xFF, 0x2F, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
//    0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
//    0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
//    0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
//    0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
//    0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
//    0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
//  })
//  assert(stream1.ReadBytes()).Equal(Bytes(nil), false)
//  assert(stream1.GetReadPos()).Equal(streamPosBody)
//}
//
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
