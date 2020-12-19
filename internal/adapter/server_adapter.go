package adapter

import (
	"crypto/tls"
	"net"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/gobwas/ws"
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
	Serve(service *RunnableService)
	Close(onError func(err *base.Error))
}

// ServerTCP ...
type ServerTCP struct {
	listener  net.Listener
	onConnect func(conn net.Conn)
	onError   func(err *base.Error)
}

// NewServerTCP ...
func NewServerTCP(
	network string,
	addr string,
	tlsConfig *tls.Config,
	onConnect func(conn net.Conn),
	onError func(err *base.Error),
) IServer {
	e := error(nil)

	ret := &ServerTCP{
		onConnect: onConnect,
		onError:   onError,
	}

	if tlsConfig == nil {
		ret.listener, e = net.Listen(network, addr)
	} else {
		ret.listener, e = tls.Listen(network, addr, tlsConfig)
	}

	if e != nil {
		onError(errors.ErrTemp.AddDebug(e.Error()))
		return nil
	}

	return ret
}

// Serve ...
func (p *ServerTCP) Serve(service *RunnableService) {
	for service.IsRunning() {
		tcpConn, e := p.listener.Accept()

		if e != nil {
			p.onError(errors.ErrTemp.AddDebug(e.Error()))
		} else {
			p.onConnect(tcpConn)
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

// ServerWebSocket ...
type ServerWebSocket struct {
	listener  net.Listener
	server    *http.Server
	onConnect func(conn net.Conn)
	onError   func(err *base.Error)
}

// NewServerWebSocket ...
func NewServerWebSocket(
	network string,
	addr string,
	tlsConfig *tls.Config,
	onConnect func(conn net.Conn),
	onError func(err *base.Error),
) IServer {
	ret := &ServerWebSocket{
		onConnect: onConnect,
		onError:   onError,
	}
	e := error(nil)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, e := ws.UpgradeHTTP(r, w)

		if e != nil {
			onError(errors.ErrTemp.AddDebug(e.Error()))
		} else {
			onConnect(conn)
		}
	})

	ret.server = &http.Server{
		Addr:         addr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  20 * time.Second,
		Handler:      mux,
	}

	if network == "wss" {
		ret.server.TLSConfig = tlsConfig
	}

	if tlsConfig == nil {
		ret.listener, e = net.Listen("tcp", addr)
	} else {
		ret.listener, e = tls.Listen("tcp", addr, tlsConfig)
	}

	if e != nil {
		onError(errors.ErrTemp.AddDebug(e.Error()))
		return nil
	}

	return ret
}

// Serve ...
func (p *ServerWebSocket) Serve(service *RunnableService) {
	if server := p.server; server != nil {
		if e := server.Serve(p.listener); e != nil {
			p.onError(errors.ErrTemp.AddDebug(e.Error()))
		}
	}
}

// Close ...
func (p *ServerWebSocket) Close(onError func(err *base.Error)) {
	if server := p.server; server != nil {
		p.server = nil
		p.listener = nil
		if e := server.Close(); e != nil {
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

		if p.network == "tcp" {
			netConn.SetFD(parseFD(conn, p.tlsConfig != nil))
		} else if p.network == "wss" {
			netConn.SetFD(parseFD(conn, true))
		} else {
			netConn.SetFD(parseFD(conn, false))
		}

		p.manager.AllocChannel().AddConn(netConn)
	}
}

// OnOpen ...
func (p *ServerAdapter) OnOpen(service *RunnableService) {
	switch p.network {
	case "tcp":
		p.server = NewServerTCP(p.network, p.addr, p.tlsConfig, p.onConnect, func(err *base.Error) {
			p.receiver.OnConnError(nil, err)
		})
	case "ws":
		fallthrough
	case "wss":
		p.server = NewServerWebSocket(p.network, p.addr, p.tlsConfig, p.onConnect, func(err *base.Error) {
			p.receiver.OnConnError(nil, err)
		})
	default:
		panic("not implemented")
	}
}

// OnRun ...
func (p *ServerAdapter) OnRun(service *RunnableService) {
	if server := p.server; server != (IServer)(nil) {
		server.Serve(service)
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
