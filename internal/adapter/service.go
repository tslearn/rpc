package adapter

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
)

func NewClientService(adapter *Adapter) base.IORCService {
	switch adapter.network {
	case "tcp4":
		fallthrough
	case "tcp6":
		fallthrough
	case "tcp":
		fallthrough
	case "ws":
		fallthrough
	case "wss":
		return &syncClientService{
			adapter:    adapter,
			conn:       nil,
			orcManager: base.NewORCManager(),
		}
	default:
		adapter.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(
			fmt.Sprintf("unsupported protocol %s", adapter.network),
		))
		return nil
	}
}

func NewServerService(adapter *Adapter) base.IORCService {
	switch adapter.network {
	case "tcp4":
		fallthrough
	case "tcp6":
		fallthrough
	case "tcp":
		if adapter.tlsConfig == nil {
			return &asyncTCPServerService{
				adapter:    adapter,
				ln:         nil,
				orcManager: base.NewORCManager(),
			}
		} else {
			return &syncTCPServerService{
				adapter:    adapter,
				ln:         nil,
				orcManager: base.NewORCManager(),
			}
		}
	case "ws":
		fallthrough
	case "wss":
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			conn, _, _, e := ws.UpgradeHTTP(r, w)

			if e != nil {
				adapter.receiver.OnConnError(
					nil,
					errors.ErrTemp.AddDebug(e.Error()),
				)
			} else {
				syncConn := NewNetConn(
					true,
					conn,
					adapter.rBufSize,
					adapter.wBufSize,
				)
				syncConn.SetNext(NewStreamConn(syncConn, adapter.receiver))
				runIConnOnServer(syncConn)
			}
		})

		return &syncWSServerService{
			adapter: adapter,
			ln:      nil,
			server: &http.Server{
				Addr:    adapter.addr,
				Handler: mux,
			},
			orcManager: base.NewORCManager(),
		}
	default:
		adapter.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(
			fmt.Sprintf("unsupported protocol %s", adapter.network),
		))
		return nil
	}
}

func runIConnOnServer(conn IConn) {
	go func() {
		conn.OnOpen()
		for {
			if ok := conn.OnReadReady(); !ok {
				break
			}
		}
		conn.OnClose()
		conn.Close()
	}()
}

// -----------------------------------------------------------------------------
// syncTCPServerService
// -----------------------------------------------------------------------------
type syncTCPServerService struct {
	adapter    *Adapter
	ln         net.Listener
	orcManager *base.ORCManager
}

// Open ...
func (p *syncTCPServerService) Open() bool {
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
func (p *syncTCPServerService) Run() bool {
	return p.orcManager.Run(func(isRunning func() bool) {
		adapter := p.adapter
		for isRunning() {
			conn, e := p.ln.Accept()

			if e != nil {
				isCloseErr := !isRunning() &&
					strings.HasSuffix(e.Error(), ErrNetClosingSuffix)

				if !isCloseErr {
					adapter.receiver.OnConnError(
						nil,
						errors.ErrTemp.AddDebug(e.Error()),
					)
				}
			} else {
				syncConn := NewNetConn(
					true,
					conn,
					adapter.rBufSize,
					adapter.wBufSize,
				)
				syncConn.SetNext(NewStreamConn(syncConn, adapter.receiver))
				runIConnOnServer(syncConn)
			}
		}
	})
}

// Close ...
func (p *syncTCPServerService) Close() bool {
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

// -----------------------------------------------------------------------------
// asyncTCPServerService
// -----------------------------------------------------------------------------
type asyncTCPServerService struct {
	adapter    *Adapter
	ln         *XListener
	orcManager *base.ORCManager
}

func (p *asyncTCPServerService) Open() bool {
	return p.orcManager.Open(func() bool {
		adapter := p.adapter

		p.ln = NewXListener(
			adapter.network,
			adapter.addr,
			func(conn IConn) {
				conn.SetNext(NewStreamConn(conn, adapter.receiver))
				runIConnOnServer(conn)
			},
			func(err *base.Error) {
				adapter.receiver.OnConnError(nil, err)
			},
			adapter.rBufSize,
			adapter.wBufSize,
		)

		if p.ln == nil {
			return false
		}

		return p.ln.Open()
	})
}

func (p *asyncTCPServerService) Run() bool {
	return true
}

func (p *asyncTCPServerService) Close() bool {
	return p.orcManager.Close(func() {
		p.ln.Close()
	}, func() {
		p.ln = nil
	})
}

// -----------------------------------------------------------------------------
// syncWSServerService
// -----------------------------------------------------------------------------
type syncWSServerService struct {
	adapter    *Adapter
	ln         net.Listener
	server     *http.Server
	orcManager *base.ORCManager
}

// Open ...
func (p *syncWSServerService) Open() bool {
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
func (p *syncWSServerService) Run() bool {
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
func (p *syncWSServerService) Close() bool {
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

// -----------------------------------------------------------------------------
// syncClientService
// -----------------------------------------------------------------------------
type syncClientService struct {
	adapter    *Adapter
	conn       *NetConn
	orcManager *base.ORCManager
	sync.Mutex
}

func (p *syncClientService) openConn() bool {
	p.Lock()
	defer p.Unlock()

	var e error
	var conn net.Conn

	adapter := p.adapter
	switch adapter.network {
	case "tcp4":
		fallthrough
	case "tcp6":
		fallthrough
	case "tcp":
		if adapter.tlsConfig == nil {
			conn, e = net.Dial(adapter.network, adapter.addr)
		} else {
			conn, e = tls.Dial(adapter.network, adapter.addr, adapter.tlsConfig)
		}
	case "ws":
		fallthrough
	case "wss":
		dialer := &ws.Dialer{TLSConfig: adapter.tlsConfig}
		u := url.URL{Scheme: adapter.network, Host: adapter.addr, Path: "/"}
		conn, _, _, e = dialer.Dial(context.Background(), u.String())
	default:
		adapter.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(
			fmt.Sprintf("unsupported protocol %s", adapter.network),
		))
		return false
	}

	if e != nil {
		adapter.receiver.OnConnError(
			nil,
			errors.ErrTemp.AddDebug(e.Error()),
		)
		return false
	}

	p.conn = NewNetConn(false, conn, adapter.rBufSize, adapter.wBufSize)
	p.conn.SetNext(NewStreamConn(p.conn, p.adapter.receiver))
	return true
}

func (p *syncClientService) closeConn() {
	p.Lock()
	defer p.Unlock()

	if conn := p.conn; conn != nil {
		conn.Close()
	}
}

// Open ...
func (p *syncClientService) Open() bool {
	return p.orcManager.Open(func() bool {
		return true
	})
}

// Run ...
func (p *syncClientService) Run() bool {
	return p.orcManager.Run(func(isRunning func() bool) {
		for isRunning() {
			start := base.TimeNow()

			if p.openConn() {
				p.conn.OnOpen()
				for {
					if ok := p.conn.OnReadReady(); !ok {
						break
					}
				}
				p.conn.OnClose()

				p.closeConn()
			}

			sleepInterval := 100 * time.Millisecond
			runningTime := base.TimeNow().Sub(start)
			sleepCount := (3*time.Second - runningTime) / sleepInterval

			for isRunning() && sleepCount > 0 {
				time.Sleep(sleepInterval)
				sleepCount--
			}
		}
	})
}

// Close ...
func (p *syncClientService) Close() bool {
	return p.orcManager.Close(func() {
		p.closeConn()
	}, func() {
		p.conn = nil
	})
}
