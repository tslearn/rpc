package adapter

import (
	"context"
	"crypto/tls"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
)

// ClientTCP ...
type ClientTCP struct {
	adapter    *ClientAdapter
	conn       *NetConn
	orcManager *base.ORCManager

	sync.Mutex
}

// NewClientTCP ...
func NewClientTCP(adapter *ClientAdapter) *ClientTCP {
	return &ClientTCP{
		adapter:    adapter,
		conn:       nil,
		orcManager: base.NewORCManager(),
	}
}

func (p *ClientTCP) openConn() bool {
	p.Lock()
	defer p.Unlock()

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

	p.conn = NewNetConn(false, conn, adapter.rBufSize, adapter.wBufSize)
	p.conn.SetNext(NewStreamConn(p.conn, p.adapter.receiver))
	return true
}

func (p *ClientTCP) closeConn() {
	p.Lock()
	defer p.Unlock()

	if conn := p.conn; conn != nil {
		conn.Close()
	}
}

// Open ...
func (p *ClientTCP) Open() bool {
	return p.orcManager.Open(func() bool {
		return true
	})
}

// Run ...
func (p *ClientTCP) Run() bool {
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
func (p *ClientTCP) Close() bool {
	return p.orcManager.Close(func() {
		p.closeConn()
	}, func() {
		p.conn = nil
	})
}

// ClientWebsocket ...
type ClientWebsocket struct {
	adapter    *ClientAdapter
	conn       *NetConn
	orcManager *base.ORCManager
	sync.Mutex
}

// NewClientWebsocket ...
func NewClientWebsocket(adapter *ClientAdapter) *ClientWebsocket {
	return &ClientWebsocket{
		adapter:    adapter,
		conn:       nil,
		orcManager: base.NewORCManager(),
	}
}

func (p *ClientWebsocket) openConn() bool {
	p.Lock()
	defer p.Unlock()

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

	p.conn = NewNetConn(false, conn, adapter.rBufSize, adapter.wBufSize)
	p.conn.SetNext(NewStreamConn(p.conn, p.adapter.receiver))

	return true
}

func (p *ClientWebsocket) closeConn() {
	p.Lock()
	defer p.Unlock()

	if conn := p.conn; conn != nil {
		conn.Close()
	}
}

// Open ...
func (p *ClientWebsocket) Open() bool {
	return p.orcManager.Open(func() bool {
		return true
	})
}

// Run ...
func (p *ClientWebsocket) Run() bool {
	return p.orcManager.Run(func(isRunning func() bool) {
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
	})
}

// Close ...
func (p *ClientWebsocket) Close() bool {
	return p.orcManager.Close(func() {
		p.closeConn()
	}, func() {
		p.conn = nil
	})
}
