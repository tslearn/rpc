package adapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"math/rand"
	"net"
	"testing"
	"time"
)

func TestNewServerSyncConn(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)

		conn := &net.TCPConn{}
		v := NewServerSyncConn(conn, 1024, 2048)
		assert(v.isServer).IsTrue()
		assert(v.isRunning).IsTrue()
		assert(v.conn).Equal(conn)
		assert(v.next).IsNil()
		assert(len(v.rBuf), cap(v.rBuf)).Equal(1024, 1024)
		assert(len(v.wBuf), cap(v.wBuf)).Equal(2048, 2048)
	})
}

func TestNewClientSyncConn(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		conn := &net.TCPConn{}
		v := NewClientSyncConn(conn, 1024, 2048)
		assert(v.isServer).IsFalse()
		assert(v.isRunning).IsTrue()
		assert(v.conn).Equal(conn)
		assert(v.next).IsNil()
		assert(len(v.rBuf), cap(v.rBuf)).Equal(1024, 1024)
		assert(len(v.wBuf), cap(v.wBuf)).Equal(2048, 2048)
	})
}

func TestSyncConn_SetNext(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServerSyncConn(&net.TCPConn{}, 1024, 2048)
		v.SetNext(v)
		assert(v.next).Equal(v)
		v.SetNext(nil)
		assert(v.next == nil).IsTrue()
	})
}

func TestSyncConn_OnOpen(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		v := NewServerSyncConn(&net.TCPConn{}, 1024, 2048)
		v.SetNext(streamConn)
		v.OnOpen()
		assert(receiver.GetOnOpenCount()).Equal(1)
		assert(receiver.GetOnCloseCount()).Equal(0)
		assert(receiver.GetOnErrorCount()).Equal(0)
		assert(receiver.GetOnStreamCount()).Equal(0)
	})
}

func TestSyncConn_OnClose(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		v := NewServerSyncConn(&net.TCPConn{}, 1024, 2048)
		v.SetNext(streamConn)
		v.OnOpen()
		v.OnClose()
		assert(receiver.GetOnOpenCount()).Equal(1)
		assert(receiver.GetOnCloseCount()).Equal(1)
		assert(receiver.GetOnErrorCount()).Equal(0)
		assert(receiver.GetOnStreamCount()).Equal(0)
	})
}

func TestSyncConn_OnError(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		v := NewServerSyncConn(&net.TCPConn{}, 1024, 2048)
		v.SetNext(streamConn)
		v.OnOpen()
		v.OnError(errors.ErrStream.AddDebug("error"))
		v.OnClose()
		assert(receiver.GetOnOpenCount()).Equal(1)
		assert(receiver.GetOnCloseCount()).Equal(1)
		assert(receiver.GetOnErrorCount()).Equal(1)
		assert(receiver.GetOnStreamCount()).Equal(0)
		assert(receiver.GetError()).Equal(errors.ErrStream.AddDebug("error"))
	})
}

func TestSyncConn_Close(t *testing.T) {
	t.Run("not running", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		v := NewServerSyncConn(&net.TCPConn{}, 1024, 2048)
		v.SetNext(streamConn)
		v.isRunning = false
		v.Close()
		assert(receiver.GetOnErrorCount()).Equal(0)
	})

	t.Run("running", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn(nil, 10, 10)
		v := NewServerSyncConn(netConn, 1024, 2048)
		assert(netConn.isRunning).IsTrue()
		v.Close()
		assert(netConn.isRunning).IsFalse()
	})

	t.Run("close error", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn(nil, 10, 10)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		streamConn.OnOpen()
		v := NewServerSyncConn(netConn, 1024, 2048)
		v.SetNext(streamConn)
		netConn.isRunning = false
		v.Close()
		assert(receiver.GetOnErrorCount()).Equal(1)
		assert(receiver.GetError()).
			Equal(errors.ErrConnClose.AddDebug("close error"))
	})
}

func TestSyncConn_LocalAddr(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		expectedAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.12:8080")
		v := NewServerSyncConn(newTestNetConn(nil, 10, 10), 1024, 2048)
		assert(v.LocalAddr()).Equal(expectedAddr)
	})
}

func TestSyncConn_RemoteAddr(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		expectedAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.11:8081")
		v := NewServerSyncConn(newTestNetConn(nil, 10, 10), 1024, 2048)
		assert(v.RemoteAddr()).Equal(expectedAddr)
	})
}

func TestSyncConn_OnReadReady(t *testing.T) {
	t.Run("server error io.EOF", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServerSyncConn(newTestNetConn(nil, 10, 10), 1024, 2048)
		assert(v.OnReadReady()).IsFalse()
	})

	t.Run("server error closed", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn(nil, 10, 10)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		streamConn.OnOpen()
		v := NewServerSyncConn(netConn, 1024, 2048)
		v.SetNext(streamConn)
		netConn.isRunning = false
		assert(v.OnReadReady()).IsFalse()
		assert(receiver.GetOnErrorCount()).Equal(1)
		assert(receiver.GetError()).
			Equal(errors.ErrConnRead.AddDebug(ErrNetClosingSuffix))
	})

	t.Run("server ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := core.NewStream()
		stream.BuildStreamCheck()
		netConn := newTestNetConn(stream.GetBuffer(), 1024, 10)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		streamConn.OnOpen()
		v := NewServerSyncConn(netConn, 1024, 2048)
		v.SetNext(streamConn)
		assert(v.OnReadReady()).IsTrue()
		assert(receiver.GetOnStreamCount()).Equal(1)
		assert(receiver.GetStream().GetBuffer()).Equal(stream.GetBuffer())
	})

	t.Run("client error io.EOF", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewClientSyncConn(newTestNetConn(nil, 10, 10), 1024, 2048)
		assert(v.OnReadReady()).IsFalse()
	})

	t.Run("client error use of closed conn when running", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn(nil, 10, 10)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		streamConn.OnOpen()
		v := NewClientSyncConn(netConn, 1024, 2048)
		v.SetNext(streamConn)
		netConn.isRunning = false
		assert(v.OnReadReady()).IsFalse()
		assert(receiver.GetOnErrorCount()).Equal(1)
		assert(receiver.GetError()).
			Equal(errors.ErrConnRead.AddDebug(ErrNetClosingSuffix))
	})

	t.Run("client error use of closed conn when closed", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn(nil, 10, 10)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		streamConn.OnOpen()
		v := NewClientSyncConn(netConn, 1024, 2048)
		v.SetNext(streamConn)
		v.isRunning = false
		netConn.isRunning = false
		assert(v.OnReadReady()).IsFalse()
		assert(receiver.GetOnErrorCount()).Equal(0)
	})
}

func TestSyncConn_OnWriteReady(t *testing.T) {
	t.Run("nothing to write", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		netConn := newTestNetConn(nil, 10, 10)
		v := NewClientSyncConn(netConn, 1024, 2048)
		v.SetNext(streamConn)
		assert(v.OnWriteReady()).IsTrue()
		assert(netConn.writePos).Equal(0)
	})

	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)

		for write := 1; write < 200; write += 10 {
			rand.Seed(base.TimeNow().UnixNano())
			numOfStream := rand.Int()%5 + 1
			stream := core.NewStream()
			for i := 0; i < write; i++ {
				stream.PutBytes([]byte{43})
			}
			exceptBuf := make([]byte, 0)
			for i := 0; i < numOfStream; i++ {
				exceptBuf = append(exceptBuf, stream.GetBuffer()...)
			}

			for writeBufSize := 1; writeBufSize < 200; writeBufSize += 10 {
				for maxWrite := 1; maxWrite < 200; maxWrite += 10 {
					receiver := newTestSingleReceiver()
					streamConn := NewStreamConn(nil, receiver)
					for i := 0; i < numOfStream; i++ {
						streamConn.writeCH <- stream.Clone()
					}
					netConn := newTestNetConn(nil, 10, maxWrite)
					v := NewClientSyncConn(netConn, 1024, writeBufSize)
					v.SetNext(streamConn)
					assert(v.OnWriteReady()).IsTrue()
					assert(netConn.writeBuf[:netConn.writePos]).Equal(exceptBuf)
				}
			}

			stream.Release()
		}
	})

	t.Run("write error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := core.NewStream()
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		streamConn.OnOpen()
		streamConn.writeCH <- stream.Clone()
		netConn := newTestNetConn(nil, 10, 10)
		netConn.isRunning = false
		v := NewClientSyncConn(netConn, 1024, 1024)
		v.SetNext(streamConn)
		assert(v.OnWriteReady()).IsFalse()
		assert(receiver.GetOnErrorCount()).Equal(1)
		assert(receiver.GetError()).
			Equal(errors.ErrConnWrite.AddDebug(ErrNetClosingSuffix))
	})

	t.Run("write zero", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := core.NewStream()
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		streamConn.OnOpen()
		streamConn.writeCH <- stream.Clone()
		netConn := newTestNetConn(nil, 10, 0)
		v := NewClientSyncConn(netConn, 1024, 1024)
		v.SetNext(streamConn)
		assert(v.OnWriteReady()).IsFalse()
		assert(receiver.GetOnErrorCount()).Equal(0)
	})
}

func TestSyncConn_OnReadBytes(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewClientSyncConn(newTestNetConn(nil, 10, 10), 1024, 1024)

		assert(base.RunWithCatchPanic(func() {
			v.OnReadBytes(nil)
		})).Equal("kernel error: it should not be called")
	})
}

func TestSyncConn_OnFillWrite(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewClientSyncConn(newTestNetConn(nil, 10, 10), 1024, 1024)

		assert(base.RunWithCatchPanic(func() {
			v.OnFillWrite(nil)
		})).Equal("kernel error: it should not be called")
	})
}

func TestStreamConnBasic(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(streamConnStatusRunning).Equal(int32(1))
		assert(streamConnStatusClosed).Equal(int32(0))
	})
}

func TestNewStreamConn(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		prev := NewServerSyncConn(nil, 1024, 1024)
		receiver := newTestSingleReceiver()
		v := NewStreamConn(prev, receiver)
		assert(v.status).Equal(streamConnStatusRunning)
		assert(v.prev).Equal(prev)
		assert(v.receiver).Equal(receiver)
		assert(len(v.writeCH)).Equal(0)
		assert(cap(v.writeCH)).Equal(16)
		assert(v.readHeadPos).Equal(0)
		assert(len(v.readHeadBuf)).Equal(core.StreamHeadSize)
		assert(cap(v.readHeadBuf)).Equal(core.StreamHeadSize)
		assert(v.readStream).IsNil()
		assert(v.writeStream).IsNil()
		assert(v.writePos).Equal(0)
	})
}

func TestStreamConn_SetReceiver(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		initReceiver := newTestSingleReceiver()
		changeReceiver := newTestSingleReceiver()
		v := NewStreamConn(nil, initReceiver)
		v.SetReceiver(changeReceiver)
		assert(v.receiver).Equal(changeReceiver)
	})
}

func TestStreamConn_OnOpen(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		v := NewStreamConn(nil, receiver)
		v.OnOpen()
		assert(receiver.GetOnOpenCount()).Equal(1)
		assert(receiver.GetOnCloseCount()).Equal(0)
		assert(receiver.GetOnErrorCount()).Equal(0)
	})
}

func TestStreamConn_OnClose(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		v := NewStreamConn(nil, receiver)
		v.OnOpen()
		v.OnClose()
		assert(receiver.GetOnOpenCount()).Equal(1)
		assert(receiver.GetOnCloseCount()).Equal(1)
		assert(receiver.GetOnErrorCount()).Equal(0)
	})
}

func TestStreamConn_OnError(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		v := NewStreamConn(nil, receiver)
		v.OnOpen()
		v.OnError(errors.ErrStream)
		v.OnClose()
		assert(receiver.GetOnOpenCount()).Equal(1)
		assert(receiver.GetOnCloseCount()).Equal(1)
		assert(receiver.GetOnErrorCount()).Equal(1)
		assert(receiver.GetError()).Equal(errors.ErrStream)
	})
}

func TestStreamConn_OnReadBytes(t *testing.T) {
	t.Run("stream length error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		streamConn.OnOpen()
		streamConn.OnReadBytes(core.NewStream().GetBuffer())
		assert(receiver.GetOnErrorCount()).Equal(1)
		assert(receiver.GetError()).Equal(errors.ErrStream)
	})

	t.Run("stream check error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		streamConn.OnOpen()
		stream := core.NewStream()
		stream.PutBytes([]byte{12})
		stream.BuildStreamCheck()
		errBuffer := stream.GetBuffer()
		errBuffer[len(errBuffer)-1] = 11
		streamConn.OnReadBytes(errBuffer)
		assert(receiver.GetOnErrorCount()).Equal(1)
		assert(receiver.GetError()).Equal(errors.ErrStream)
	})

	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for streams := 1; streams < 5; streams++ {
			buffer := make([]byte, 0)

			for i := 0; i < streams; i++ {
				stream := core.NewStream()
				for j := 0; j < i; j++ {
					stream.PutBytes([]byte{12})
				}
				stream.BuildStreamCheck()
				buffer = append(buffer, stream.GetBuffer()...)
			}

			for readBufSize := 1; readBufSize < 200; readBufSize += 10 {
				for maxRead := 1; maxRead < 200; maxRead += 10 {
					receiver := newTestSingleReceiver()
					conn := newTestNetConn(buffer, maxRead, 1024)
					netConn := NewServerSyncConn(conn, readBufSize, 1024)
					streamConn := NewStreamConn(netConn, receiver)
					netConn.SetNext(streamConn)
					runIConn(netConn)
					assert(receiver.GetOnStreamCount()).Equal(streams)
					expectBuffer := make([]byte, 0)

					for i := 0; i < streams; i++ {
						expectBuffer = append(
							expectBuffer,
							receiver.GetStream().GetBuffer()...,
						)
					}
					assert(expectBuffer).Equal(buffer)
				}
			}
		}
	})
}

func TestStreamConn_OnFillWrite(t *testing.T) {
	t.Run("no stream", func(t *testing.T) {
		assert := base.NewAssert(t)
		buf := make([]byte, 1024)
		receiver := newTestSingleReceiver()
		v := NewStreamConn(nil, receiver)
		v.OnOpen()
		assert(v.OnFillWrite(buf)).Equal(0)
		assert(receiver.GetOnErrorCount()).Equal(0)
	})

	t.Run("streamConn is closed", func(t *testing.T) {
		assert := base.NewAssert(t)
		buf := make([]byte, 1024)
		receiver := newTestSingleReceiver()
		v := NewStreamConn(
			NewServerSyncConn(newTestNetConn(nil, 10, 10), 10, 10),
			receiver,
		)
		v.OnOpen()
		v.Close()
		assert(v.OnFillWrite(buf)).Equal(0)
		assert(receiver.GetOnErrorCount()).Equal(0)
	})

	t.Run("emulate peek kernel error", func(t *testing.T) {
		assert := base.NewAssert(t)
		buf := make([]byte, 1024)
		receiver := newTestSingleReceiver()
		v := NewStreamConn(
			NewServerSyncConn(newTestNetConn(nil, 10, 10), 10, 10),
			receiver,
		)
		v.OnOpen()
		v.writeStream = core.NewStream()
		v.writePos = v.writeStream.GetWritePos()
		assert(v.OnFillWrite(buf)).Equal(0)
		assert(receiver.GetOnErrorCount()).Equal(1)
		assert(receiver.GetError()).Equal(errors.ErrOnFillWriteFatal)
	})

	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for streams := 1; streams < 2; streams++ {
			for bufSize := 1; bufSize < 2; bufSize += 10 {
				receiver := newTestSingleReceiver()
				v := NewStreamConn(
					NewServerSyncConn(newTestNetConn(nil, 10, 10), 10, 10),
					receiver,
				)
				v.OnOpen()

				buffer := make([]byte, 0)
				for i := 0; i < streams; i++ {
					stream := core.NewStream()
					rand.Seed(base.TimeNow().UnixNano())
					for j := 0; j < rand.Int()%20; j++ {
						stream.PutBytes([]byte{12})
					}
					buffer = append(buffer, stream.GetBuffer()...)
					v.writeCH <- stream
				}

				fillBytes := 1
				start := 0

				for start < len(buffer) {
					if start+fillBytes < len(buffer) {
						b := make([]byte, fillBytes)
						assert(v.OnFillWrite(b)).Equal(fillBytes)
						assert(b).Equal(buffer[start : start+fillBytes])
					} else {
						b := make([]byte, fillBytes)
						assert(v.OnFillWrite(b)).Equal(len(buffer) - start)
						assert(b[:len(buffer)-start]).Equal(buffer[start:])
						break
					}

					start += fillBytes
					fillBytes++
				}

				assert(receiver.GetOnErrorCount()).Equal(0)
			}
		}
	})
}

func TestStreamConn_Close(t *testing.T) {
	t.Run("close ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		prev := NewServerSyncConn(newTestNetConn(nil, 10, 10), 1024, 1024)
		v := NewStreamConn(prev, nil)
		v.Close()

		assert(base.RunWithCatchPanic(func() {
			v.writeCH <- core.NewStream()
		})).IsNotNil()
		assert(prev.isRunning).IsFalse()
	})

	t.Run("it has already been closed", func(t *testing.T) {
		assert := base.NewAssert(t)
		prev := NewServerSyncConn(newTestNetConn(nil, 10, 10), 1024, 1024)
		v := NewStreamConn(prev, nil)
		v.status = streamConnStatusClosed
		v.Close()
		assert(base.RunWithCatchPanic(func() {
			v.writeCH <- core.NewStream()
		})).IsNil()
		assert(prev.isRunning).IsTrue()
	})
}

func TestStreamConn_LocalAddr(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		expectedAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.12:8080")
		prev := NewServerSyncConn(newTestNetConn(nil, 10, 10), 1024, 1024)
		v := NewStreamConn(prev, nil)
		assert(v.LocalAddr()).Equal(expectedAddr)
	})
}

func TestStreamConn_RemoteAddr(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		expectedAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.11:8081")
		prev := NewServerSyncConn(newTestNetConn(nil, 10, 10), 1024, 1024)
		v := NewStreamConn(prev, nil)
		assert(v.RemoteAddr()).Equal(expectedAddr)
	})
}

func TestStreamConn_WriteStreamAndRelease(t *testing.T) {
	t.Run("write ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		conn := newTestNetConn(nil, 10, 10)
		netConn := NewServerSyncConn(conn, 1024, 1024)
		v := NewStreamConn(netConn, newTestSingleReceiver())
		netConn.SetNext(v)
		v.OnOpen()
		v.WriteStreamAndRelease(core.NewStream())
		retStream := core.NewStream()
		retStream.BuildStreamCheck()
		assert(conn.writeBuf[:conn.writePos]).
			Equal(retStream.GetBuffer())
	})

	t.Run("write error", func(t *testing.T) {
		assert := base.NewAssert(t)
		conn := newTestNetConn(nil, 10, 10)
		netConn := NewServerSyncConn(conn, 1024, 1024)
		v := NewStreamConn(netConn, newTestSingleReceiver())
		netConn.SetNext(v)
		v.OnOpen()
		v.Close()
		v.WriteStreamAndRelease(core.NewStream())
		retStream := core.NewStream()
		retStream.BuildStreamCheck()
		assert(conn.writeBuf[:conn.writePos]).Equal([]byte{})
	})
}

func TestStreamConn_IsActive(t *testing.T) {
	t.Run("active true", func(t *testing.T) {
		assert := base.NewAssert(t)
		nowNS := base.TimeNow().UnixNano()
		v := NewStreamConn(nil, nil)
		v.activeTimeNS = nowNS
		assert(v.IsActive(nowNS, time.Second)).IsTrue()
	})

	t.Run("active true", func(t *testing.T) {
		assert := base.NewAssert(t)
		now := base.TimeNow()
		v := NewStreamConn(nil, nil)
		v.activeTimeNS = now.Add(-time.Second).UnixNano()
		assert(v.IsActive(now.UnixNano(), time.Second)).IsFalse()
	})
}

func TestStreamConn_SetNext(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewStreamConn(nil, nil)
		assert(base.RunWithCatchPanic(func() {
			v.SetNext(nil)
		})).Equal("kernel error: it should not be called")
	})
}

func TestStreamConn_OnReadReady(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewStreamConn(nil, nil)
		assert(base.RunWithCatchPanic(func() {
			v.OnReadReady()
		})).Equal("kernel error: it should not be called")
	})
}

func TestStreamConn_OnWriteReady(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewStreamConn(nil, nil)
		assert(base.RunWithCatchPanic(func() {
			v.OnWriteReady()
		})).Equal("kernel error: it should not be called")
	})
}
