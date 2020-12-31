package adapter

import (
	"crypto/tls"
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
)

// Adapter ...
type Adapter struct {
	isClient   bool
	network    string
	addr       string
	tlsConfig  *tls.Config
	rBufSize   int
	wBufSize   int
	receiver   IReceiver
	service    base.IORCService
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
) *Adapter {
	return &Adapter{
		isClient:   true,
		network:    network,
		addr:       addr,
		tlsConfig:  tlsConfig,
		rBufSize:   rBufSize,
		wBufSize:   wBufSize,
		receiver:   receiver,
		service:    nil,
		orcManager: base.NewORCManager(),
	}
}

// NewServerAdapter ...
func NewServerAdapter(
	network string,
	addr string,
	tlsConfig *tls.Config,
	rBufSize int,
	wBufSize int,
	receiver IReceiver,
) *Adapter {
	return &Adapter{
		isClient:   false,
		network:    network,
		addr:       addr,
		tlsConfig:  tlsConfig,
		rBufSize:   rBufSize,
		wBufSize:   wBufSize,
		receiver:   receiver,
		service:    nil,
		orcManager: base.NewORCManager(),
	}
}

// Open ...
func (p *Adapter) Open() bool {
	return p.orcManager.Open(func() bool {
		switch p.network {
		case "tcp4":
			fallthrough
		case "tcp6":
			fallthrough
		case "tcp":
			if p.isClient {
				p.service = NewClientTCP(p)
			} else {
				p.service = NewServerTCP(p)
			}

			return p.service.Open()
		case "ws":
			fallthrough
		case "wss":
			if p.isClient {
				p.service = NewClientWebsocket(p)
			} else {
				p.service = NewServerWebSocket(p)
			}
			return p.service.Open()
		default:
			p.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(
				fmt.Sprintf("unsupported protocol %s", p.network),
			))
			return false
		}
	})
}

// Run ...
func (p *Adapter) Run() bool {
	return p.orcManager.Run(func(isRunning func() bool) {
		p.service.Run()
	})
}

// Close ...
func (p *Adapter) Close() bool {
	return p.orcManager.Close(func() {
		p.service.Close()
	}, func() {
		p.service = nil
	})
}
