package adapter

import (
	"github.com/rpccloud/rpc/internal/base"
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
