package adapter

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
)

// ClientAdapter ...
type ClientAdapter struct {
	network    string
	addr       string
	tlsConfig  *tls.Config
	rBufSize   int
	wBufSize   int
	receiver   IReceiver
	client     base.IORCService
	orcManager *base.ORCManager
}

// NewClientAdapter ...
func NewClientAdapter(
	network string,
	addr string,
	tlsConfig *tls.Config,
	rBufSize int,
	wBufSize int,
	receiver IReceiver,
) *ClientAdapter {
	return &ClientAdapter{
		network:    network,
		addr:       addr,
		tlsConfig:  tlsConfig,
		rBufSize:   rBufSize,
		wBufSize:   wBufSize,
		receiver:   receiver,
		client:     nil,
		orcManager: base.NewORCManager(),
	}
}

// CreateNetConn ...
func (p *ClientAdapter) CreateNetConn(conn net.Conn) *NetConn {
	ret := NewNetConn(false, conn, p.rBufSize, p.wBufSize)
	ret.SetNext(NewStreamConn(ret, p.receiver))
	return ret
}

// Open ...
func (p *ClientAdapter) Open() bool {
	return p.orcManager.Open(func() bool {
		switch p.network {
		case "tcp4":
			fallthrough
		case "tcp6":
			fallthrough
		case "tcp":
			p.client = NewClientTCP(p)
			return true
		case "ws":
			fallthrough
		case "wss":
			p.client = NewClientWebsocket(p)
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
func (p *ClientAdapter) Run() bool {
	return p.orcManager.Run(func(isRunning func() bool) {
		for isRunning() {
			start := base.TimeNow()
			p.client.Open()
			p.client.Run()
			p.client.Close()

			if isRunning() {
				if delta := base.TimeNow().Sub(start); delta < time.Second {
					time.Sleep(time.Second - delta)
				}
			}
		}
	})
}

// Close ...
func (p *ClientAdapter) Close() bool {
	return p.orcManager.Close(func() {
		p.client.Close()
	}, func() {
		p.client = nil
	})
}

// ServerAdapter ...
type ServerAdapter struct {
	network    string
	addr       string
	tlsConfig  *tls.Config
	rBufSize   int
	wBufSize   int
	receiver   IReceiver
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
	receiver IReceiver,
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
			netConn := NewNetConn(true, conn, p.rBufSize, p.wBufSize)
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
