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

// ServerTCP ...
type ServerTCP struct {
	adapter    *ServerAdapter
	ln         net.Listener
	orcManager *base.ORCManager
}

// NewServerTCP ...
func NewServerTCP(adapter *ServerAdapter) base.IORCService {
	return &ServerTCP{
		adapter:    adapter,
		ln:         nil,
		orcManager: base.NewORCManager(),
	}
}

// Open ...
func (p *ServerTCP) Open() bool {
	return p.orcManager.Open(func() bool {
		e := error(nil)
		adapter := p.adapter
		if p.adapter.tlsConfig == nil {
			p.ln, e = net.Listen(adapter.network, adapter.addr)
		} else {
			p.ln, e = tls.Listen(
				adapter.network,
				adapter.addr,
				adapter.tlsConfig,
			)
		}

		if e != nil {
			adapter.receiver.OnConnError(
				nil,
				errors.ErrTemp.AddDebug(e.Error()),
			)
			return false
		}

		return true
	})
}

// Run ...
func (p *ServerTCP) Run() bool {
	return p.orcManager.Run(func(isRunning func() bool) {
		for isRunning() {
			conn, e := p.ln.Accept()
			p.adapter.onConnect(conn, e)
		}
	})
}

// Close ...
func (p *ServerTCP) Close() bool {
	return p.orcManager.Close(func() {
		if e := p.ln.Close(); e != nil {
			p.adapter.receiver.OnConnError(
				nil,
				errors.ErrTemp.AddDebug(e.Error()),
			)
		}
	}, func() {
		p.ln = nil
	})
}

// ServerWebSocket ...
type ServerWebSocket struct {
	adapter    *ServerAdapter
	ln         net.Listener
	server     *http.Server
	orcManager *base.ORCManager
}

// NewServerWebSocket ...
func NewServerWebSocket(adapter *ServerAdapter) base.IORCService {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, e := ws.UpgradeHTTP(r, w)
		adapter.onConnect(conn, e)
	})

	return &ServerWebSocket{
		adapter: adapter,
		ln:      nil,
		server: &http.Server{
			Addr:    adapter.addr,
			Handler: mux,
		},
		orcManager: base.NewORCManager(),
	}
}

// Open ...
func (p *ServerWebSocket) Open() bool {
	return p.orcManager.Open(func() bool {
		e := error(nil)
		adapter := p.adapter
		if adapter.tlsConfig == nil {
			p.ln, e = net.Listen("tcp", adapter.addr)
		} else {
			p.ln, e = tls.Listen("tcp", adapter.addr, adapter.tlsConfig)
		}

		if e != nil {
			adapter.receiver.OnConnError(
				nil,
				errors.ErrTemp.AddDebug(e.Error()),
			)
			return false
		}

		return true
	})
}

// Run ...
func (p *ServerWebSocket) Run() bool {
	return p.orcManager.Run(func(_ func() bool) {
		if e := p.server.Serve(p.ln); e != nil {
			p.adapter.receiver.OnConnError(
				nil,
				errors.ErrTemp.AddDebug(e.Error()),
			)
		}
	})
}

// Close ...
func (p *ServerWebSocket) Close() bool {
	return p.orcManager.Close(func() {
		if e := p.server.Close(); e != nil {
			p.adapter.receiver.OnConnError(
				nil,
				errors.ErrTemp.AddDebug(e.Error()),
			)
		}
	}, func() {
		p.ln = nil
	})
}

// ServerAdapter ...
type ServerAdapter struct {
	network    string
	addr       string
	tlsConfig  *tls.Config
	rBufSize   int
	wBufSize   int
	receiver   common.IReceiver
	server     base.IORCService
	orcManager *base.ORCManager
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
	return &ServerAdapter{
		network:    network,
		addr:       addr,
		tlsConfig:  tlsConfig,
		rBufSize:   rBufSize,
		wBufSize:   wBufSize,
		receiver:   receiver,
		server:     nil,
		orcManager: base.NewORCManager(),
	}
}

func (p *ServerAdapter) onConnect(conn net.Conn, e error) {
	if e != nil {
		p.receiver.OnConnError(
			nil,
			errors.ErrTemp.AddDebug(e.Error()),
		)
	} else {
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
}

// Open ...
func (p *ServerAdapter) Open() bool {
	return p.orcManager.Open(func() bool {
		switch p.network {
		case "tcp4":
			fallthrough
		case "tcp6":
			fallthrough
		case "tcp":
			p.server = NewServerTCP(p)
			return true
		case "ws":
			fallthrough
		case "wss":
			p.server = NewServerWebSocket(p)
			return true
		default:
			p.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(
				fmt.Sprintf("unsupported protocol %s", p.network),
			))
			return false
		}
	})
}

// Run ...
func (p *ServerAdapter) Run() bool {
	return p.orcManager.Run(func(isRunning func() bool) {
		for isRunning() {
			if p.server.Open() {
				p.server.Run()
				p.server.Close()
			}
		}
	})
}

// Close ...
func (p *ServerAdapter) Close() bool {
	return p.orcManager.Close(func() {
		p.server.Close()
	}, func() {
		p.server = nil
	})
}
