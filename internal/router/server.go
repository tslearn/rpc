package router

import (
	"crypto/tls"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	addr       string
	tlsConfig  *tls.Config
	orcManager *base.ORCManager
	errorHub   rpc.IStreamHub
	id         *base.GlobalID
	slotMap    map[uint64]*ChannelManager
	ln         net.Listener
	sync.Mutex
}

func NewServer(addr string, tlsConfig *tls.Config) *Server {
	return &Server{
		addr:       addr,
		tlsConfig:  tlsConfig,
		orcManager: base.NewORCManager(),
		errorHub:   rpc.NewLogToScreenErrorStreamHub("Router"),
		id:         nil,
		slotMap:    map[uint64]*ChannelManager{},
		ln:         nil,
	}
}

func (p *Server) GetErrorHub() rpc.IStreamHub {
	p.Lock()
	defer p.Unlock()
	return p.errorHub
}

func (p *Server) SetErrorHub(errorHub rpc.IStreamHub) {
	p.Lock()
	defer p.Unlock()
	p.errorHub = errorHub
}

// Open ...
func (p *Server) Open() bool {
	return p.orcManager.Open(func() bool {
		e := error(nil)
		if p.tlsConfig == nil {
			p.ln, e = net.Listen("tcp", p.addr)
		} else {
			p.ln, e = tls.Listen("tcp", p.addr, p.tlsConfig)
		}
		if e != nil {
			p.errorHub.OnReceiveStream(rpc.MakeSystemErrorStream(
				base.ErrRouteServerListen.AddDebug(e.Error()),
			))
			return false
		}

		return true
	})
}

// Run ...
func (p *Server) Run() bool {
	return p.orcManager.Run(func(isRunning func() bool) bool {
		for isRunning() {
			conn, e := p.ln.Accept()
			if e != nil {
				isCloseErr := !isRunning() &&
					strings.HasSuffix(e.Error(), base.ErrNetClosingSuffix)

				if !isCloseErr {
					p.errorHub.OnReceiveStream(rpc.MakeSystemErrorStream(
						base.ErrRouteServerAccept.AddDebug(e.Error()),
					))

					base.WaitAtLeastDurationWhenRunning(
						base.TimeNow().UnixNano(),
						isRunning,
						500*time.Millisecond,
					)
				}
			} else {
				go p.onConnect(conn)
			}
		}

		return true
	})
}

func (p *Server) onConnect(conn net.Conn) {

}

// Close ...
func (p *Server) Close() bool {
	return p.orcManager.Close(func() bool {
		if e := p.ln.Close(); e != nil {
			p.errorHub.OnReceiveStream(rpc.MakeSystemErrorStream(
				base.ErrRouteServerClose.AddDebug(e.Error()),
			))
		}

		return true
	}, func() {
		p.ln = nil
	})
}
