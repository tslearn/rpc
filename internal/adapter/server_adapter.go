package adapter

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/rpccloud/rpc/internal/adapter/common"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
)

// IServer ...
type IServer interface {
	Open() bool
	Run() bool
	Close() bool
}

// ServerTCP ...
type ServerTCP struct {
	network    string
	addr       string
	tlsConfig  *tls.Config
	onConnect  func(conn net.Conn)
	onError    func(err *base.Error)
	ln         net.Listener
	orcManager *base.ORCManager
}

// NewServerTCP ...
func NewServerTCP(
	network string,
	addr string,
	tlsConfig *tls.Config,
	onConnect func(conn net.Conn),
	onError func(err *base.Error),
) IServer {
	return &ServerTCP{
		network:    network,
		addr:       addr,
		tlsConfig:  tlsConfig,
		onConnect:  onConnect,
		onError:    onError,
		ln:         nil,
		orcManager: base.NewORCManager(),
	}
}

func (p *ServerTCP) Open() bool {
	return p.orcManager.Open(func() bool {
		e := error(nil)

		if p.tlsConfig == nil {
			p.ln, e = net.Listen(p.network, p.addr)
		} else {
			p.ln, e = tls.Listen(p.network, p.addr, p.tlsConfig)
		}

		if e != nil {
			p.onError(errors.ErrTemp.AddDebug(e.Error()))
			return false
		}

		return true
	})
}

// Serve ...
func (p *ServerTCP) Run() bool {
	return p.orcManager.Run(func(isRunning func() bool) {
		for isRunning() {
			conn, e := p.ln.Accept()

			if e != nil {
				p.onError(errors.ErrTemp.AddDebug(e.Error()))
			} else {
				p.onConnect(conn)
			}
		}
	})
}

// Close ...
func (p *ServerTCP) Close() bool {
	return p.orcManager.Close(func() {
		if e := p.ln.Close(); e != nil {
			p.onError(errors.ErrTemp.AddDebug(e.Error()))
		}
		p.ln = nil
	})
}

// ServerWebSocket ...
type ServerWebSocket struct {
	addr       string
	tlsConfig  *tls.Config
	onConnect  func(conn net.Conn)
	onError    func(err *base.Error)
	ln         net.Listener
	server     *http.Server
	orcManager *base.ORCManager
}

// NewServerWebSocket ...
func NewServerWebSocket(
	addr string,
	tlsConfig *tls.Config,
	onConnect func(conn net.Conn),
	onError func(err *base.Error),
) IServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, e := ws.UpgradeHTTP(r, w)
		if e != nil {
			onError(errors.ErrTemp.AddDebug(e.Error()))
		} else {
			onConnect(conn)
		}
	})

	return &ServerWebSocket{
		addr:      addr,
		tlsConfig: tlsConfig,
		onConnect: onConnect,
		onError:   onError,
		ln:        nil,
		server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
		orcManager: base.NewORCManager(),
	}
}

// Serve ...
func (p *ServerWebSocket) Open() bool {
	return p.orcManager.Open(func() bool {
		e := error(nil)

		if p.tlsConfig == nil {
			p.ln, e = net.Listen("tcp", p.addr)
		} else {
			p.ln, e = tls.Listen("tcp", p.addr, p.tlsConfig)
		}

		if e != nil {
			p.onError(errors.ErrTemp.AddDebug(e.Error()))
			return false
		}

		return true
	})
}

// Serve ...
func (p *ServerWebSocket) Run() bool {
	return p.orcManager.Run(func(isRunning func() bool) {
		for isRunning() {
			if e := p.server.Serve(p.ln); e != nil {
				p.onError(errors.ErrTemp.AddDebug(e.Error()))
			}
		}
	})
}

// Close ...
func (p *ServerWebSocket) Close() bool {
	return p.orcManager.Close(func() {
		if e := p.server.Close(); e != nil {
			p.onError(errors.ErrTemp.AddDebug(e.Error()))
		}
		p.ln = nil
	})
}

// ServerAdapter ...
type ServerAdapter struct {
	network   string
	addr      string
	tlsConfig *tls.Config
	rBufSize  int
	wBufSize  int
	receiver  common.IReceiver
	server    IServer
}

// NewServerAdapter ...
func NewServerAdapter(
	network string,
	addr string,
	tlsConfig *tls.Config,
	rBufSize int,
	wBufSize int,
	receiver common.IReceiver,
) *ServerAdapter {
	ret := &ServerAdapter{
		network:   network,
		addr:      addr,
		tlsConfig: tlsConfig,
		rBufSize:  rBufSize,
		wBufSize:  wBufSize,
		receiver:  receiver,
		server:    nil,
	}

	switch network {
	case "tcp4":
		fallthrough
	case "tcp6":
		fallthrough
	case "tcp":
		ret.server = NewServerTCP(
			network,
			addr,
			tlsConfig,
			ret.onConnect,
			func(err *base.Error) {
				receiver.OnConnError(nil, err)
			},
		)
	case "ws":
		fallthrough
	case "wss":
		ret.server = NewServerWebSocket(
			addr,
			tlsConfig,
			ret.onConnect,
			func(err *base.Error) {
				receiver.OnConnError(nil, err)
			},
		)
	default:
		panic(fmt.Sprintf("unsupported protocol %s", network))
	}

	return ret
}

func (p *ServerAdapter) onConnect(conn net.Conn) {
	go func() {
		netConn := common.NewNetConn(conn, p.rBufSize, p.wBufSize)
		netConn.SetNext(common.NewStreamConn(netConn, p.receiver))
		netConn.OnOpen()
		for {
			if ok := netConn.OnReadReady(); !ok {
				break
			}
		}
		netConn.OnClose()
		netConn.Close()
	}()
}

func (p *ServerAdapter) Open() bool {
	return p.server.Open()
}

func (p *ServerAdapter) Run() bool {
	return p.server.Run()
}

func (p *ServerAdapter) Close() bool {
	return p.server.Close()
}
