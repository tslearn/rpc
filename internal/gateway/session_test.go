package gateway

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/router"
	"net"
	"testing"
	"time"
)

type testNetConn struct {
	errCH   chan error
	closeCH chan bool
}

func newTestNetConn() *testNetConn {
	return &testNetConn{
		errCH:   make(chan error, 1024),
		closeCH: make(chan bool, 1024),
	}
}

func (p *testNetConn) Read(_ []byte) (n int, err error) {
	panic("not implemented")
}

func (p *testNetConn) Write(_ []byte) (n int, err error) {
	panic("not implemented")
}

func (p *testNetConn) Close() error {
	p.closeCH <- true
	return nil
}

func (p *testNetConn) LocalAddr() net.Addr {
	panic("not implemented")
}

func (p *testNetConn) RemoteAddr() net.Addr {
	panic("not implemented")
}

func (p *testNetConn) SetDeadline(_ time.Time) error {
	panic("not implemented")
}

func (p *testNetConn) SetReadDeadline(_ time.Time) error {
	panic("not implemented")
}

func (p *testNetConn) SetWriteDeadline(_ time.Time) error {
	panic("not implemented")
}

func prepareTestSession() {
	//netConn := newTestNetConn()
	//serverConn := adapter.NewServerNetConn()

}

func TestNewSession(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		gateway := NewGateWay(
			43,
			GetDefaultConfig(),
			router.NewDirectRouter(),
			func(sessionID uint64, err *base.Error) {},
		)
		v := NewSession(3, gateway)
		assert(v.id).Equal(uint64(3))
		assert(v.gateway).Equal(gateway)
		assert(len(v.security)).Equal(32)
		assert(v.conn).IsNil()
		assert(len(v.channels)).Equal(GetDefaultConfig().numOfChannels)
		assert(cap(v.channels)).Equal(GetDefaultConfig().numOfChannels)
		assert(base.TimeNow().UnixNano()-v.activeTimeNS < int64(time.Second)).
			IsTrue()
		assert(v.prev).IsNil()
		assert(v.next).IsNil()
	})
}

func TestSession_TimeCheck(t *testing.T) {

}
