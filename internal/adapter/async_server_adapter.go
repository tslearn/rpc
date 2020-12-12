package adapter

import (
	"github.com/rpccloud/rpc/internal/adapter/netpoll"
	"github.com/rpccloud/rpc/internal/base"
	"net"
	"runtime"
	"strings"
	"time"
)

type AsyncServerAdapter struct {
	network  string
	addr     string
	rBufSize int
	wBufSize int
	receiver IReceiver
	manager  *netpoll.Manager
}

func NewAsyncServerAdapter(
	network string,
	addr string,
	rBufSize int,
	wBufSize int,
	receiver IReceiver,
) *RunnableService {
	return NewRunnableService(&AsyncServerAdapter{
		network:  network,
		addr:     addr,
		rBufSize: rBufSize,
		wBufSize: wBufSize,
		receiver: receiver,
		manager:  nil,
	})
}

func (p *AsyncServerAdapter) OnOpen() bool {
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

func (p *AsyncServerAdapter) OnRun(service *RunnableService) {
	for service.IsRunning() {
		time.Sleep(50 * time.Millisecond)
	}
}

func (p *AsyncServerAdapter) OnWillClose() {
	// do nothing
}

func (p *AsyncServerAdapter) OnDidClose() {
	p.manager.Close()
}

func (p *AsyncServerAdapter) GetConnectFunc() func(*netpoll.Channel, int, net.Addr, net.Addr) netpoll.Conn {
	if strings.HasPrefix(p.network, "tcp") {
		return func(channel *netpoll.Channel, fd int, lAddr net.Addr, rAddr net.Addr) netpoll.Conn {
			conn := NewAsyncConn(channel, fd, lAddr, rAddr, p.rBufSize, p.wBufSize)
			conn.SetNext(NewStreamConn(conn, p.receiver))
			return conn
		}
	} else {
		panic("not implement")
	}
}
