package async

import (
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"net"
	"runtime"
	"strings"
	"time"
)

type ServerAdapter struct {
	network  string
	addr     string
	rBufSize int
	wBufSize int
	receiver adapter.IReceiver
	manager  *Manager
}

func NewAsyncServerAdapter(
	network string,
	addr string,
	rBufSize int,
	wBufSize int,
	receiver adapter.IReceiver,
) *adapter.RunnableService {
	return adapter.NewRunnableService(&ServerAdapter{
		network:  network,
		addr:     addr,
		rBufSize: rBufSize,
		wBufSize: wBufSize,
		receiver: receiver,
		manager:  nil,
	})
}

func (p *ServerAdapter) OnOpen() bool {
	p.manager = NewManager(
		p.network,
		p.addr, func(err *base.Error) {
			p.receiver.OnConnError(nil, err)
		},
		p.GetConnectFunc(),
		runtime.NumCPU(),
	)

	return p.manager != nil
}

func (p *ServerAdapter) OnRun(service *adapter.RunnableService) {
	for service.IsRunning() {
		time.Sleep(50 * time.Millisecond)
	}
}

func (p *ServerAdapter) OnWillClose() {
	// do nothing
}

func (p *ServerAdapter) OnDidClose() {
	p.manager.Close()
}

func (p *ServerAdapter) GetConnectFunc() func(*Channel, int, net.Addr, net.Addr) *Conn {
	if strings.HasPrefix(p.network, "tcp") {
		return func(channel *Channel, fd int, lAddr net.Addr, rAddr net.Addr) *Conn {
			conn := NewConn(channel, fd, lAddr, rAddr, p.rBufSize, p.wBufSize)
			conn.SetNext(adapter.NewStreamConn(conn, p.receiver))
			return conn
		}
	} else {
		panic("not implement")
	}
}
