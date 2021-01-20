package adapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
	"testing"
	"time"
)

func TestNewServerNetConn(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)

		conn := &net.TCPConn{}
		v := NewServerNetConn(conn, 1024, 2048)
		assert(v.isServer).IsTrue()
		assert(v.isRunning).IsTrue()
		assert(v.conn).Equal(conn)
		assert(v.next).IsNil()
		assert(len(v.rBuf), cap(v.rBuf)).Equal(1024, 1024)
		assert(len(v.wBuf), cap(v.wBuf)).Equal(2048, 2048)
	})
}

func TestNewClientNetConn(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		conn := &net.TCPConn{}
		v := NewClientNetConn(conn, 1024, 2048)
		assert(v.isServer).IsFalse()
		assert(v.isRunning).IsTrue()
		assert(v.conn).Equal(conn)
		assert(v.next).IsNil()
		assert(len(v.rBuf), cap(v.rBuf)).Equal(1024, 1024)
		assert(len(v.wBuf), cap(v.wBuf)).Equal(2048, 2048)
	})
}

func TestNetConn_SetNext(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServerNetConn(&net.TCPConn{}, 1024, 2048)
		v.SetNext(v)
		assert(v.next).Equal(v)
		v.SetNext(nil)
		assert(v.next == nil).IsTrue()
	})
}

func TestNetConn_OnOpen(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		v := NewServerNetConn(&net.TCPConn{}, 1024, 2048)
		v.SetNext(streamConn)
		v.OnOpen()
		assert(receiver.GetOnOpenCount()).Equal(1)
		assert(receiver.GetOnCloseCount()).Equal(0)
		assert(receiver.GetOnErrorCount()).Equal(0)
		assert(receiver.GetOnStreamCount()).Equal(0)
	})
}

func TestNetConn_OnClose(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		v := NewServerNetConn(&net.TCPConn{}, 1024, 2048)
		v.SetNext(streamConn)
		v.OnOpen()
		v.OnClose()
		assert(receiver.GetOnOpenCount()).Equal(1)
		assert(receiver.GetOnCloseCount()).Equal(1)
		assert(receiver.GetOnErrorCount()).Equal(0)
		assert(receiver.GetOnStreamCount()).Equal(0)
	})
}

func TestNetConn_OnError(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		v := NewServerNetConn(&net.TCPConn{}, 1024, 2048)
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

func TestNetConn_Close(t *testing.T) {
	t.Run("not running", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		v := NewServerNetConn(&net.TCPConn{}, 1024, 2048)
		v.SetNext(streamConn)
		v.isRunning = false
		v.Close()
		assert(receiver.GetOnErrorCount()).Equal(0)
	})

	t.Run("running", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn(nil, 10)
		v := NewServerNetConn(netConn, 1024, 2048)
		assert(netConn.isRunning).IsTrue()
		v.Close()
		assert(netConn.isRunning).IsFalse()
	})

	t.Run("close error", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn(nil, 10)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		streamConn.OnOpen()
		v := NewServerNetConn(netConn, 1024, 2048)
		v.SetNext(streamConn)
		netConn.isRunning = false
		v.Close()
		assert(receiver.GetOnErrorCount()).Equal(1)
		assert(receiver.GetError()).
			Equal(errors.ErrConnClose.AddDebug("close error"))
	})
}

func TestNetConn_LocalAddr(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		expectedAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.12:8080")
		v := NewServerNetConn(newTestNetConn(nil, 10), 1024, 2048)
		assert(v.LocalAddr()).Equal(expectedAddr)
	})
}

func TestNetConn_RemoteAddr(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		expectedAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.11:8081")
		v := NewServerNetConn(newTestNetConn(nil, 10), 1024, 2048)
		assert(v.RemoteAddr()).Equal(expectedAddr)
	})
}

func TestNetConn_OnReadReady(t *testing.T) {
	t.Run("server error io.EOF", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServerNetConn(newTestNetConn(nil, 10), 1024, 2048)
		assert(v.OnReadReady()).IsFalse()
	})

	t.Run("server error closed", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn(nil, 10)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		streamConn.OnOpen()
		v := NewServerNetConn(netConn, 1024, 2048)
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
		netConn := newTestNetConn(stream.GetBuffer(), 10)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		streamConn.OnOpen()
		v := NewServerNetConn(netConn, 1024, 2048)
		v.SetNext(streamConn)
		assert(v.OnReadReady()).IsTrue()
		assert(receiver.GetOnStreamCount()).Equal(1)
		assert(receiver.GetStream().GetBuffer()).Equal(stream.GetBuffer())
	})

	t.Run("client error io.EOF", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewClientNetConn(newTestNetConn(nil, 10), 1024, 2048)
		assert(v.OnReadReady()).IsFalse()
	})

	t.Run("client error use of closed conn when running", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn(nil, 10)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		streamConn.OnOpen()
		v := NewClientNetConn(netConn, 1024, 2048)
		v.SetNext(streamConn)
		netConn.isRunning = false
		assert(v.OnReadReady()).IsFalse()
		assert(receiver.GetOnErrorCount()).Equal(1)
		assert(receiver.GetError()).
			Equal(errors.ErrConnRead.AddDebug(ErrNetClosingSuffix))
	})

	t.Run("client error use of closed conn when closed", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn(nil, 10)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		streamConn.OnOpen()
		v := NewClientNetConn(netConn, 1024, 2048)
		v.SetNext(streamConn)
		v.isRunning = false
		netConn.isRunning = false
		assert(v.OnReadReady()).IsFalse()
		assert(receiver.GetOnErrorCount()).Equal(0)
	})
}

func TestNetConn_OnWriteReady(t *testing.T) {
	t.Run("nothing to write", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		streamConn := NewStreamConn(nil, receiver)
		netConn := newTestNetConn(nil, 10)
		v := NewClientNetConn(netConn, 1024, 2048)
		v.SetNext(streamConn)
		assert(v.OnWriteReady()).IsTrue()
		assert(netConn.writePos).Equal(0)
	})

	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for write := 1; write < 200; write += 10 {
			stream := core.NewStream()
			for i := 0; i < write; i++ {
				stream.PutBytes([]byte{43})
			}
			exceptBuf := stream.GetBuffer()

			for writeBufSize := 1; writeBufSize < 300; writeBufSize += 10 {
				for maxWrite := 1; maxWrite < 400; maxWrite += 10 {
					receiver := newTestSingleReceiver()
					streamConn := NewStreamConn(nil, receiver)
					streamConn.writeCH <- stream.Clone()
					netConn := newTestNetConn(nil, maxWrite)
					v := NewClientNetConn(netConn, 1024, writeBufSize)
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
		netConn := newTestNetConn(nil, 10)
		netConn.isRunning = false
		v := NewClientNetConn(netConn, 1024, 1024)
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
		netConn := newTestNetConn(nil, 0)
		v := NewClientNetConn(netConn, 1024, 1024)
		v.SetNext(streamConn)
		assert(v.OnWriteReady()).IsFalse()
		assert(receiver.GetOnErrorCount()).Equal(0)
	})
}

func TestNetConn_OnReadBytes(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewClientNetConn(newTestNetConn(nil, 10), 1024, 1024)

		assert(base.RunWithCatchPanic(func() {
			v.OnReadBytes(nil)
		})).Equal("kernel error: it should not be called")
	})
}

func TestNetConn_OnFillWrite(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewClientNetConn(newTestNetConn(nil, 10), 1024, 1024)

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
		prev := NewServerNetConn(nil, 1024, 1024)
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

func TestStreamConn_Close(t *testing.T) {
	t.Run("close ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		prev := NewServerNetConn(newTestNetConn(nil, 10), 1024, 1024)
		v := NewStreamConn(prev, nil)
		v.Close()

		assert(base.RunWithCatchPanic(func() {
			v.writeCH <- core.NewStream()
		})).IsNotNil()
		assert(prev.isRunning).IsFalse()
	})

	t.Run("it has already been closed", func(t *testing.T) {
		assert := base.NewAssert(t)
		prev := NewServerNetConn(newTestNetConn(nil, 10), 1024, 1024)
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
		prev := NewServerNetConn(newTestNetConn(nil, 10), 1024, 1024)
		v := NewStreamConn(prev, nil)
		assert(v.LocalAddr()).Equal(expectedAddr)
	})
}

func TestStreamConn_RemoteAddr(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		expectedAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.11:8081")
		prev := NewServerNetConn(newTestNetConn(nil, 10), 1024, 1024)
		v := NewStreamConn(prev, nil)
		assert(v.RemoteAddr()).Equal(expectedAddr)
	})
}

func TestStreamConn_WriteStreamAndRelease(t *testing.T) {
	t.Run("write ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		conn := newTestNetConn(nil, 10)
		netConn := NewServerNetConn(conn, 1024, 1024)
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
		conn := newTestNetConn(nil, 10)
		netConn := NewServerNetConn(conn, 1024, 1024)
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
