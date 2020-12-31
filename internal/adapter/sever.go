package adapter

import (
	"crypto/tls"
	"net"
	"net/http"
	"strings"

	"github.com/gobwas/ws"
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
func NewServerTCP(adapter *ServerAdapter) *ServerTCP {
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

func runSyncConn(
	adapter *ServerAdapter,
	conn net.Conn,
) {
	netConn := NewNetConn(
		true,
		conn,
		adapter.rBufSize,
		adapter.wBufSize,
	)
	netConn.SetNext(NewStreamConn(netConn, adapter.receiver))

	go func() {
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

// Run ...
func (p *ServerTCP) Run() bool {
	return p.orcManager.Run(func(isRunning func() bool) {
		for isRunning() {
			conn, e := p.ln.Accept()

			if e != nil {
				isCloseErr := !isRunning() &&
					strings.HasSuffix(e.Error(), ErrNetClosingSuffix)

				if !isCloseErr {
					p.adapter.receiver.OnConnError(
						nil,
						errors.ErrTemp.AddDebug(e.Error()),
					)
				}
			} else {
				runSyncConn(p.adapter, conn)
			}
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
func NewServerWebSocket(adapter *ServerAdapter) *ServerWebSocket {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, e := ws.UpgradeHTTP(r, w)

		if e != nil {
			adapter.receiver.OnConnError(
				nil,
				errors.ErrTemp.AddDebug(e.Error()),
			)
		} else {
			runSyncConn(adapter, conn)
		}
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
	return p.orcManager.Run(func(isRunning func() bool) {
		for isRunning() {
			if e := p.server.Serve(p.ln); e != nil {
				if e != http.ErrServerClosed {
					p.adapter.receiver.OnConnError(
						nil,
						errors.ErrTemp.AddDebug(e.Error()),
					)
				}
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
