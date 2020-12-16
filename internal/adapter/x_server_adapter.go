package adapter

import (
	"crypto/tls"
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/rpccloud/rpc/internal/adapter/netpoll"
	"github.com/rpccloud/rpc/internal/base"
)

// XServerAdapter ...
type XServerAdapter struct {
	network  string
	addr     string
	rBufSize int
	wBufSize int
	receiver IReceiver
}

// NewXServerAdapter ...
func NewXServerAdapter(
	network string,
	addr string,
	tlsConfig *tls.Config,
	rBufSize int,
	wBufSize int,
	receiver IReceiver,
) *RunnableService {
	return NewRunnableService(&XServerAdapter{
		network:  network,
		addr:     addr,
		rBufSize: rBufSize,
		wBufSize: wBufSize,
		receiver: receiver,
	})
}

// OnRun ...
func (p *XServerAdapter) OnRun(service *RunnableService) {
	if manager := netpoll.NewManager(
		func(err *base.Error) {
			p.receiver.OnConnError(nil, err)
		},
		base.MaxInt(runtime.NumCPU()/2, 1),
	); manager == nil {
		return
	} else if listener := netpoll.NewTCPListener(
		p.network,
		p.addr,
		func(fd int, localAddr net.Addr, remoteAddr net.Addr) {
			channel := manager.AllocChannel()
			fnConnect := p.GetConnectFunc()
			if conn := fnConnect(channel, fd, localAddr, remoteAddr); conn != nil {
				channel.AddConn(conn)
			}
		},
		func(err *base.Error) {
			p.receiver.OnConnError(nil, err)
		},
	); listener == nil {
		manager.Close()
		return
	} else {
		for service.IsRunning() {
			time.Sleep(50 * time.Millisecond)
		}
		listener.Close()
		manager.Close()
	}
}

// OnStop ...
func (p *XServerAdapter) OnStop(_ *RunnableService) {
	// do nothing
}

// Close ...
func (p *XServerAdapter) Close() {
	// do nothing
}

// GetConnectFunc ...
func (p *XServerAdapter) GetConnectFunc() func(*netpoll.Channel, int, net.Addr, net.Addr) netpoll.Conn {
	if strings.HasPrefix(p.network, "tcp") {
		return func(channel *netpoll.Channel, fd int, lAddr net.Addr, rAddr net.Addr) netpoll.Conn {
			conn := NewXConn(channel, fd, lAddr, rAddr, p.rBufSize, p.wBufSize)
			conn.SetNext(NewStreamConn(conn, p.receiver))
			return conn
		}
	}

	panic("unsupported protocol")
}
