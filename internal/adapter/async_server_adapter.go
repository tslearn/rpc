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

// IAsyncServer ...
type IAsyncServer interface {
	Serve(
		service *RunnableService,
		onConnect func(conn net.Conn),
		onError func(err *base.Error),
	)
	Close(onError func(err *base.Error))
}

type AsyncServerTCP struct {
	listener net.Listener
}

func NewAsyncServerTCP(
	network string,
	addr string,
	tlsConfig *tls.Config,
) (*AsyncServerTCP, *base.Error) {
	ret := &AsyncServerTCP{}
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

func (p *AsyncServerTCP) Serve(
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

func (p *AsyncServerTCP) Close(onError func(err *base.Error)) {
	if listener := p.listener; listener != nil {
		p.listener = nil
		if e := listener.Close(); e != nil {
			onError(errors.ErrTemp.AddDebug(e.Error()))
		}
	}
}

// AsyncServerAdapter ...
type AsyncServerAdapter struct {
	network   string
	addr      string
	rBufSize  int
	wBufSize  int
	tlsConfig *tls.Config
	receiver  IReceiver

	server  IAsyncServer
	manager *netpoll.Manager
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
	return NewRunnableService(&AsyncServerAdapter{
		network:   network,
		addr:      addr,
		tlsConfig: tlsConfig,
		rBufSize:  rBufSize,
		wBufSize:  wBufSize,
		receiver:  receiver,
	})
}

func (p *AsyncServerAdapter) onConnect(conn net.Conn) {
	netConn := NewNetConn(conn, p.rBufSize, p.wBufSize)
	netConn.SetNext(NewStreamConn(netConn, p.receiver))
	netConn.SetFD(parseFD(conn, p.tlsConfig != nil))
	p.manager.AllocChannel().AddConn(netConn)
}

// OnOpen ...
func (p *AsyncServerAdapter) OnOpen(service *RunnableService) {
	p.manager = netpoll.NewManager(
		func(err *base.Error) {
			p.receiver.OnConnError(nil, err)
		},
		base.MaxInt(runtime.NumCPU()/2, 1),
	)

	if p.manager == nil {
		return
	}

	err := (*base.Error)(nil)

	switch p.network {
	case "tcp":
		p.server, err = NewAsyncServerTCP(p.network, p.addr, p.tlsConfig)
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
func (p *AsyncServerAdapter) OnRun(service *RunnableService) {
	p.server.Serve(service, p.onConnect, func(err *base.Error) {
		p.receiver.OnConnError(nil, err)
	})
}

// func (p *AsyncServerAdapter) runAsWebsocketServer(
// 	manager *netpoll.Manager,
// 	service *RunnableService,
// ) {
// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		wsConn, _, _, e := ws.UpgradeHTTP(r, w)

// 		if e != nil {
// 			p.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(e.Error()))
// 		} else {

// 		}
// 	})

// 	srv := &http.Server{
// 		Addr:         p.addr,
// 		ReadTimeout:  10 * time.Second,
// 		WriteTimeout: 10 * time.Second,
// 		IdleTimeout:  20 * time.Second,
// 		Handler:      mux,
// 	}

// 	if p.network == "wss" {
// 		srv.TLSConfig = p.tlsConfig
// 	}

// 	atomic.StorePointer(&p.server, unsafe.Pointer(srv))
// 	srv.ListenAndServe()

// 	ln, e := net.Listen("tcp", p.addr)
// 	if e != nil {
// 		p.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(e.Error()))
// 		return
// 	}

// 	srv.Serve(ln)
// }

// OnStop ...
func (p *AsyncServerAdapter) OnStop(_ *RunnableService) {

}

// Close ...
func (p *AsyncServerAdapter) Close() {
	if server := p.server; server != nil {
		p.server = nil
		p.server.Close(func(err *base.Error) {
			p.receiver.OnConnError(nil, err)
		})
	}
}
