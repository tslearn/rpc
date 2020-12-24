package adapter

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"

	"github.com/rpccloud/rpc/internal/adapter/common"
	"github.com/rpccloud/rpc/internal/errors"

	"github.com/gobwas/ws"
	"github.com/rpccloud/rpc/internal/base"
)

// ClientTCP ...
type ClientTCP struct {
	adapter    *ClientAdapter
	conn       net.Conn
	orcManager *base.ORCManager
}

// NewClientTCP ...
func NewClientTCP(adapter *ClientAdapter) base.IORCService {
	return &ClientTCP{
		adapter:    adapter,
		conn:       nil,
		orcManager: base.NewORCManager(),
	}
}

// Open ...
func (p *ClientTCP) Open() bool {
	return p.orcManager.Open(func() bool {
		e := error(nil)
		adapter := p.adapter
		if adapter.tlsConfig == nil {
			p.conn, e = net.Dial(adapter.network, adapter.addr)
		} else {
			p.conn, e = tls.Dial(
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
func (p *ClientTCP) Run() bool {
	return p.orcManager.Run(func(isRunning func() bool) {
		p.adapter.onConnect(p.conn, nil)
	})
}

// Close ...
func (p *ClientTCP) Close() bool {
	return p.orcManager.Close(func() {
		if e := p.conn.Close(); e != nil {
			p.adapter.receiver.OnConnError(
				nil,
				errors.ErrTemp.AddDebug(e.Error()),
			)
		}
	}, func() {
		p.conn = nil
	})
}

// ClientWebsocket ...
type ClientWebsocket struct {
	adapter    *ClientAdapter
	conn       net.Conn
	orcManager *base.ORCManager
}

// NewClientWebsocket ...
func NewClientWebsocket(adapter *ClientAdapter) base.IORCService {
	return &ClientWebsocket{
		adapter:    adapter,
		conn:       nil,
		orcManager: base.NewORCManager(),
	}
}

// Open ...
func (p *ClientWebsocket) Open() bool {
	return p.orcManager.Open(func() bool {
		e := error(nil)
		adapter := p.adapter
		dialer := &ws.Dialer{TLSConfig: adapter.tlsConfig}
		u := url.URL{Scheme: adapter.network, Host: adapter.addr, Path: "/"}
		p.conn, _, _, e = dialer.Dial(context.Background(), u.String())

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
func (p *ClientWebsocket) Run() bool {
	return p.orcManager.Run(func(isRunning func() bool) {
		p.adapter.onConnect(p.conn, nil)
	})
}

// Close ...
func (p *ClientWebsocket) Close() bool {
	return p.orcManager.Close(func() {
		if e := p.conn.Close(); e != nil {
			p.adapter.receiver.OnConnError(
				nil,
				errors.ErrTemp.AddDebug(e.Error()),
			)
		}
	}, func() {
		p.conn = nil
	})
}

// ClientAdapter ...
type ClientAdapter struct {
	network    string
	addr       string
	tlsConfig  *tls.Config
	rBufSize   int
	wBufSize   int
	receiver   common.IReceiver
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
	receiver common.IReceiver,
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

func (p *ClientAdapter) onConnect(conn net.Conn, e error) {
	if e != nil {
		p.receiver.OnConnError(
			nil,
			errors.ErrTemp.AddDebug(e.Error()),
		)
	} else {
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
	}
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
			if p.client.Open() {
				p.client.Run()
				p.client.Close()
			}
		}
	})
}

// Close ...
func (p *ClientAdapter) Close() bool {
	return p.orcManager.Close(nil, func() {
		p.client = nil
	})
}
