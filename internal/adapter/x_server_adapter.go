package adapter

import (
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
	manager  *netpoll.Manager
}

// NewXServerAdapter ...
func NewXServerAdapter(
	network string,
	addr string,
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
		manager:  nil,
	})
}

// OnOpen ...
func (p *XServerAdapter) OnOpen() bool {
	p.manager = netpoll.NewManager(
		p.network,
		p.addr, func(err *base.Error) {
			p.receiver.OnConnError(nil, err)
		},
		p.GetConnectFunc(),
		base.MaxInt(runtime.NumCPU()/2, 1),
	)

	return p.manager != nil
}

// OnRun ...
func (p *XServerAdapter) OnRun(service *RunnableService) {
	for service.IsRunning() {
		time.Sleep(50 * time.Millisecond)
	}
}

// OnWillClose ...
func (p *XServerAdapter) OnWillClose() {
	// do nothing
}

// OnDidClose ...
func (p *XServerAdapter) OnDidClose() {
	p.manager.Close()
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
