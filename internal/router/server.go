package router

import (
	"crypto/tls"
	"encoding/binary"
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
	router     *Router
	ln         net.Listener
	sync.Mutex
}

func NewServer(addr string, tlsConfig *tls.Config, errorHub rpc.IStreamHub) *Server {
	ret := &Server{
		addr:       addr,
		tlsConfig:  tlsConfig,
		orcManager: base.NewORCManager(),
		errorHub:   errorHub,
		id:         nil,
		router:     NewRouter(errorHub),
		ln:         nil,
	}

	return ret
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
				base.ErrRouterConnListen.AddDebug(e.Error()),
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
						base.ErrRouterConnConnect.AddDebug(e.Error()),
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
	buffer := make([]byte, 32)

	if err := connReadBytes(conn, time.Second, buffer); err != nil {
		p.errorHub.OnReceiveStream(rpc.MakeSystemErrorStream(err))
		_ = conn.Close()
		return
	}

	if binary.LittleEndian.Uint16(buffer[2:]) != rpc.StreamKindConnectRequest {
		p.errorHub.OnReceiveStream(rpc.MakeSystemErrorStream(
			base.ErrRouterConnProtocol,
		))
		_ = conn.Close()
		return
	}

	channelIndex := binary.LittleEndian.Uint16(buffer[4:])
	slotID := binary.LittleEndian.Uint64(buffer[8:])
	remoteSendSequence := binary.LittleEndian.Uint64(buffer[16:])
	remoteReceiveSequence := binary.LittleEndian.Uint64(buffer[24:])
	p.router.AddSlot(slotID, conn, channelIndex, remoteSendSequence, remoteReceiveSequence)
}

// Close ...
func (p *Server) Close() bool {
	return p.orcManager.Close(func() bool {
		if e := p.ln.Close(); e != nil {
			p.errorHub.OnReceiveStream(rpc.MakeSystemErrorStream(
				base.ErrRouterConnClose.AddDebug(e.Error()),
			))
		}

		return true
	}, func() {
		p.ln = nil
	})
}
