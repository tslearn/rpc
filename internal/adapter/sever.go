package adapter

import (
	"crypto/tls"
	"github.com/gobwas/ws"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
	"net/http"
	"strings"
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

			if e != nil {
				if !isRunning() &&
					strings.HasSuffix(e.Error(), ErrNetClosingSuffix) {
					return
				}
			}

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
			if e != http.ErrServerClosed {
				p.adapter.receiver.OnConnError(
					nil,
					errors.ErrTemp.AddDebug(e.Error()),
				)
			}
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
