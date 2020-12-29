package adapter

import (
	"context"
	"crypto/tls"
	"github.com/gobwas/ws"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
	"net/url"
)

// ClientTCP ...
type ClientTCP struct {
	adapter    *ClientAdapter
	conn       *NetConn
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
		var e error
		var conn net.Conn

		adapter := p.adapter

		if adapter.tlsConfig == nil {
			conn, e = net.Dial(adapter.network, adapter.addr)
		} else {
			conn, e = tls.Dial(adapter.network, adapter.addr, adapter.tlsConfig)
		}

		if e != nil {
			adapter.receiver.OnConnError(
				nil,
				errors.ErrTemp.AddDebug(e.Error()),
			)
			return false
		}

		p.conn = adapter.CreateNetConn(conn)
		return true
	})
}

// Run ...
func (p *ClientTCP) Run() bool {
	return p.orcManager.Run(func(isRunning func() bool) {
		p.conn.OnOpen()
		for {
			if ok := p.conn.OnReadReady(); !ok {
				break
			}
		}
		p.conn.OnClose()
	})
}

// Close ...
func (p *ClientTCP) Close() bool {
	return p.orcManager.Close(func() {
		p.conn.Close()
	}, func() {
		p.conn = nil
	})
}

// ClientWebsocket ...
type ClientWebsocket struct {
	adapter    *ClientAdapter
	conn       *NetConn
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
		adapter := p.adapter
		dialer := &ws.Dialer{TLSConfig: adapter.tlsConfig}
		u := url.URL{Scheme: adapter.network, Host: adapter.addr, Path: "/"}
		conn, _, _, e := dialer.Dial(context.Background(), u.String())

		if e != nil {
			adapter.receiver.OnConnError(
				nil,
				errors.ErrTemp.AddDebug(e.Error()),
			)
			return false
		}

		p.conn = adapter.CreateNetConn(conn)

		return true
	})
}

// Run ...
func (p *ClientWebsocket) Run() bool {
	return p.orcManager.Run(func(isRunning func() bool) {
		p.conn.OnOpen()
		for {
			if ok := p.conn.OnReadReady(); !ok {
				break
			}
		}
		p.conn.OnClose()
	})
}

// Close ...
func (p *ClientWebsocket) Close() bool {
	return p.orcManager.Close(func() {
		p.conn.Close()
	}, func() {
		p.conn = nil
	})
}
