package adapter

import (
	"crypto/tls"
	"net"

	"github.com/rpccloud/rpc/internal/errors"
)

// SyncClientAdapter ...
type SyncClientAdapter struct {
	network   string
	addr      string
	tlsConfig *tls.Config
	rBufSize  int
	wBufSize  int
	receiver  IReceiver
	conn      *NetConn
}

// NewSyncClientAdapter ...
func NewSyncClientAdapter(
	network string,
	addr string,
	tlsConfig *tls.Config,
	rBufSize int,
	wBufSize int,
	receiver IReceiver,
) *RunnableService {
	return NewRunnableService(&SyncClientAdapter{
		network:   network,
		addr:      addr,
		tlsConfig: tlsConfig,
		rBufSize:  rBufSize,
		wBufSize:  wBufSize,
		receiver:  receiver,
	})
}

// OnRun ...
func (p *SyncClientAdapter) OnRun(service *RunnableService) {
	switch p.network {
	case "tcp":
		p.runAsTCPClient(service)
	default:
		panic("not implemented")
	}
}

func (p *SyncClientAdapter) runAsTCPClient(service *RunnableService) {
	netConn := net.Conn(nil)
	e := error(nil)

	if p.tlsConfig == nil {
		netConn, e = net.Dial(p.network, p.addr)
	} else {
		netConn, e = tls.Dial(p.network, p.addr, p.tlsConfig)
	}

	if e != nil {
		p.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(e.Error()))
		return
	}

	p.conn = NewNetConn(netConn, p.rBufSize, p.wBufSize)
	p.conn.SetNext(NewStreamConn(p.conn, p.receiver))

	p.conn.OnOpen()
	for service.IsRunning() {
		if ok := p.conn.OnReadReady(); !ok {
			break
		}
	}
	p.conn.OnClose()
}

// OnStop ...
func (p *SyncClientAdapter) OnStop(service *RunnableService) {
	// if OnStop is caused by Close(), don't close the conn again
	if p.conn != nil {
		if service.IsRunning() {
			p.conn.Close()
		}
		p.conn = nil
	}
}

// Close ...
func (p *SyncClientAdapter) Close() {
	p.conn.Close()
}
