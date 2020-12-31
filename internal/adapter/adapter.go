package adapter

import (
	"crypto/tls"
	"github.com/rpccloud/rpc/internal/base"
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
		if p.isClient {
			p.service = NewClientService(p)
		} else {
			p.service = NewServerService(p)
		}

		if p.service == nil {
			return false
		}

		return p.service.Open()
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
