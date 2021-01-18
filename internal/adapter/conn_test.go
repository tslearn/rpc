package adapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
	"testing"
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
