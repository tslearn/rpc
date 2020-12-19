package adapter

import (
	"crypto/tls"
	"net"
	"reflect"
	"runtime"

	"github.com/rpccloud/rpc/internal/adapter/netpoll"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
)

func parseFD(conn net.Conn, isTLS bool) int {
	if !isTLS {
		c := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
		fdVal := c.FieldByName("fd")
		pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
		return int(pfdVal.FieldByName("Sysfd").Int())
	} else {
		tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
		c := reflect.Indirect(tcpConn.Elem()).FieldByName("conn")
		fdVal := c.FieldByName("fd")
		pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
		return int(pfdVal.FieldByName("Sysfd").Int())
	}
}

// IServer ...
type IServer interface {
	Serve(
		service *RunnableService,
		onConnect func(conn net.Conn),
		onError func(err *base.Error),
	)
	Close(onError func(err *base.Error))
}

// ServerTCP ...
type ServerTCP struct {
	listener net.Listener
}

// NewServerTCP ...
func NewServerTCP(
	network string,
	addr string,
	tlsConfig *tls.Config,
) (*ServerTCP, *base.Error) {
	ret := &ServerTCP{}
	e := error(nil)

	if tlsConfig == nil {
		ret.listener, e = net.Listen(network, addr)
	} else {
		ret.listener, e = tls.Listen(network, addr, tlsConfig)
	}

	if e != nil {
		return nil, errors.ErrTemp.AddDebug(e.Error())
	}

	return ret, nil
}

// Serve ...
func (p *ServerTCP) Serve(
	service *RunnableService,
	onConnect func(conn net.Conn),
	onError func(err *base.Error),
) {
	for service.IsRunning() {
		tcpConn, e := p.listener.Accept()

		if e != nil {
			onError(errors.ErrTemp.AddDebug(e.Error()))
		} else {
			onConnect(tcpConn)
		}
	}
}

// Close ...
func (p *ServerTCP) Close(onError func(err *base.Error)) {
	if listener := p.listener; listener != nil {
		p.listener = nil
		if e := listener.Close(); e != nil {
			onError(errors.ErrTemp.AddDebug(e.Error()))
		}
	}
}

// NewSyncServerAdapter ...
func NewSyncServerAdapter(
	network string,
	addr string,
	tlsConfig *tls.Config,
	rBufSize int,
	wBufSize int,
	receiver IReceiver,
) *RunnableService {
	return NewRunnableService(&ServerAdapter{
		network:   network,
		addr:      addr,
		tlsConfig: tlsConfig,
		rBufSize:  rBufSize,
		wBufSize:  wBufSize,
		receiver:  receiver,
		manager:   nil,
		server:    nil,
	})
}

// NewAsyncServerAdapter ...
func NewAsyncServerAdapter(
	network string,
	addr string,
	tlsConfig *tls.Config,
	rBufSize int,
	wBufSize int,
	receiver IReceiver,
) *RunnableService {
	manager := netpoll.NewManager(
		func(err *base.Error) {
			receiver.OnConnError(nil, err)
		},
		base.MaxInt(runtime.NumCPU()/2, 1),
	)

	if manager == nil {
		return nil
	}

	return NewRunnableService(&ServerAdapter{
		network:   network,
		addr:      addr,
		tlsConfig: tlsConfig,
		rBufSize:  rBufSize,
		wBufSize:  wBufSize,
		receiver:  receiver,
		manager:   manager,
		server:    nil,
	})
}

// ServerAdapter ...
type ServerAdapter struct {
	network   string
	addr      string
	rBufSize  int
	wBufSize  int
	tlsConfig *tls.Config
	receiver  IReceiver
	manager   *netpoll.Manager
	server    IServer
}

func (p *ServerAdapter) onConnect(conn net.Conn) {
	if p.manager == nil { // Sync
		go func() {
			netConn := NewNetConn(conn, p.rBufSize, p.wBufSize)
			netConn.SetNext(NewStreamConn(netConn, p.receiver))
			netConn.OnOpen()
			for {
				if ok := netConn.OnReadReady(); !ok {
					break
				}
			}
			netConn.OnClose()
			netConn.Close()
		}()
	} else { // Async
		netConn := NewNetConn(conn, p.rBufSize, p.wBufSize)
		netConn.SetNext(NewStreamConn(netConn, p.receiver))
		netConn.SetFD(parseFD(conn, p.tlsConfig != nil))
		p.manager.AllocChannel().AddConn(netConn)
	}
}

// OnOpen ...
func (p *ServerAdapter) OnOpen(service *RunnableService) {
	err := (*base.Error)(nil)

	switch p.network {
	case "tcp":
		p.server, err = NewServerTCP(p.network, p.addr, p.tlsConfig)
	// case "ws":
	// 	fallthrough
	// case "wss":
	// 	p.server, err = NewAsyncServerWebsocket(p.network, p.addr, p.tlsConfig)
	default:
		panic("not implemented")
	}

	if err != nil {
		p.receiver.OnConnError(nil, err)
	}
}

// OnRun ...
func (p *ServerAdapter) OnRun(service *RunnableService) {
	if server := p.server; server != nil {
		server.Serve(service, p.onConnect, func(err *base.Error) {
			p.receiver.OnConnError(nil, err)
		})
	}
}

// OnStop ...
func (p *ServerAdapter) OnStop(_ *RunnableService) {

}

// Close ...
func (p *ServerAdapter) Close() {
	if server := p.server; server != nil {
		p.server = nil
		p.server.Close(func(err *base.Error) {
			p.receiver.OnConnError(nil, err)
		})
	}
}
