package router

import (
	"crypto/tls"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
)

type Client struct {
	id         *base.GlobalID
	addr       string
	tlsConfig  *tls.Config
	slot       *Slot
	orcManager *base.ORCManager
	errorHub   rpc.IStreamHub
}

func NewClient(addr string, tlsConfig *tls.Config) *Client {
	id := base.NewGlobalID()
	errorHub := rpc.NewLogToScreenErrorStreamHub("router-client")
	ret := &Client{
		id:        id,
		addr:      addr,
		tlsConfig: tlsConfig,
		slot: NewSlot(&ConnectMeta{
			addr:      addr,
			tlsConfig: tlsConfig,
			id:        id,
		}, errorHub),
		orcManager: base.NewORCManager(),
		errorHub:   errorHub,
	}

	ret.orcManager.Open(func() bool {
		return true
	})

	return ret
}

func (p *Client) SendStream(s *rpc.Stream) {
	p.slot.SendStream(s)
}

func (p *Client) Close() bool {
	return p.orcManager.Close(func() bool {
		p.slot.Close()
		return true
	}, func() {
		p.id = nil
	})
}
