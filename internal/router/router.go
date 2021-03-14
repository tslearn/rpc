package router

import (
	"crypto/tls"
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
	"sync"
)

type ServerSlot struct {
	manager *ChannelManager
}

type Router struct {
	adapter  *adapter.Adapter
	errorHub rpc.IStreamHub
	sync.Mutex
}

func NewRouter(addr string, tlsConfig *tls.Config) *Router {
	ret := &Router{
		adapter:  nil,
		errorHub: rpc.NewLogToScreenErrorStreamHub("Router"),
	}

	ret.adapter = adapter.NewServerAdapter(
		false, "tcp", addr, tlsConfig, 32768, 32768, ret,
	)

	return ret
}

func (p *Router) GetErrorHub() rpc.IStreamHub {
	p.Lock()
	defer p.Unlock()
	return p.errorHub
}

func (p *Router) SetErrorHub(errorHub rpc.IStreamHub) {
	p.Lock()
	defer p.Unlock()
	p.errorHub = errorHub
}

// Open ...
func (p *Router) Open() {
	p.adapter.Run()
}

// Close ...
func (p *Router) Close() {
	p.adapter.Close()
}

// OnConnOpen ...
func (p *Router) OnConnOpen(_ *adapter.StreamConn) {

}

// OnConnReadStream ...
func (p *Router) OnConnReadStream(conn *adapter.StreamConn, s *rpc.Stream) {

}

// OnConnError ...
func (p *Router) OnConnError(conn *adapter.StreamConn, err *base.Error) {
	p.errorHub.OnReceiveStream(rpc.MakeSystemErrorStream(err))

	if conn != nil {
		conn.Close()
	}
}

// OnConnClose ...
func (p *Router) OnConnClose(_ *adapter.StreamConn) {

}
