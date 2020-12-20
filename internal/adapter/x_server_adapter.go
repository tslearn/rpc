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
	network   string
	addr      string
	tlsConfig *tls.Config
	rBufSize  int
	wBufSize  int
	receiver  IReceiver
	manager   *netpoll.Manager
	listener  *netpoll.TCPListener
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
		network:   network,
		addr:      addr,
		tlsConfig: tlsConfig,
		rBufSize:  rBufSize,
		wBufSize:  wBufSize,
		receiver:  receiver,
		manager:   nil,
		listener:  nil,
	})
}

func (p *XServerAdapter) onConnect(fd int, lAddr net.Addr, rAddr net.Addr) {
	channel := p.manager.AllocChannel()
	conn := NewXConn(channel, fd, lAddr, rAddr, p.rBufSize, p.wBufSize)

	switch p.network {
	case "tcp":
		if p.tlsConfig == nil {
			conn.SetNext(NewStreamConn(conn, p.receiver))
		} else {
			panic("not implemented")
		}
	default:
		panic("not implemented")
	}

	channel.AddConn(conn)
}

// OnOpen ...
func (p *XServerAdapter) OnOpen(service *RunnableService) {
	p.manager = netpoll.NewManager(
		func(err *base.Error) {
			p.receiver.OnConnError(nil, err)
		},
		base.MaxInt(runtime.NumCPU()/2, 1),
	)

	if p.manager == nil {
		return
	}

	p.listener = netpoll.NewTCPListener(
		p.network,
		p.addr,
		p.onConnect,
		func(err *base.Error) {
			p.receiver.OnConnError(nil, err)
		},
	)
}

// OnRun ...
func (p *XServerAdapter) OnRun(service *RunnableService) {
	if listener := p.listener; listener != nil {
		for service.IsRunning() {
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// OnStop ...
func (p *XServerAdapter) OnStop(_ *RunnableService) {
	if listener := p.listener; listener != nil {
		p.listener = nil
		listener.Close()
	}

	if manager := p.manager; manager != nil {
		p.manager = nil
		manager.Close()
	}
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
