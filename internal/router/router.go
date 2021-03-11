package router

import (
	"crypto/tls"
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
	"sync"
)

const (
	numOfChannelPerSlot = 8
	numOfCacheBuffer    = 512
	bufferSize          = 65536
)

type ClientSlot struct {
	addr      string
	tlsConfig *tls.Config
	hub       *ChannelHub
}

type ServerSlot struct {
	hub *ChannelHub
}

type ChannelHub struct {
	inputCH  chan *rpc.Stream
	channels [numOfChannelPerSlot]Channel
}

type Channel struct {
	hub       *ChannelHub
	sequence  uint64
	buffer    [numOfCacheBuffer][bufferSize]byte
	bufferPos int
	stream    *rpc.Stream
	streamPos int
}

func NewSlotConn() {

}

func (p *Channel) RunWithConn(conn *adapter.StreamConn) {
	waitCH := make(chan bool)
	go func() {
		p.runRead(conn)
		waitCH <- true
	}()

	go func() {
		p.runWrite(conn)
		waitCH <- true
	}()

	<-waitCH
	<-waitCH
}

func (p *Channel) runRead(conn *adapter.StreamConn) {

}

func (p *Channel) runWrite(conn *adapter.StreamConn) {

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
