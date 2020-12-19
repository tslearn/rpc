package adapter

import (
	"crypto/tls"
	"net"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
)

// ClientAdapter ...
type ClientAdapter struct {
	network   string
	addr      string
	tlsConfig *tls.Config
	rBufSize  int
	wBufSize  int
	receiver  IReceiver
	conn      *NetConn
}

// NewSyncClientTCP ...
func NewSyncClientTCP(
	network string,
	addr string,
	tlsConfig *tls.Config,
) (net.Conn, *base.Error) {
	ret := net.Conn(nil)
	e := error(nil)

	if tlsConfig == nil {
		ret, e = net.Dial(network, addr)
	} else {
		ret, e = tls.Dial(network, addr, tlsConfig)
	}

	if e != nil {
		return nil, errors.ErrTemp.AddDebug(e.Error())
	}

	return ret, nil
}

// NewClientAdapter ...
func NewClientAdapter(
	network string,
	addr string,
	tlsConfig *tls.Config,
	rBufSize int,
	wBufSize int,
	receiver IReceiver,
) *RunnableService {
	return NewRunnableService(&ClientAdapter{
		network:   network,
		addr:      addr,
		tlsConfig: tlsConfig,
		rBufSize:  rBufSize,
		wBufSize:  wBufSize,
		receiver:  receiver,
		conn:      nil,
	})
}

// OnOpen ...
func (p *ClientAdapter) OnOpen(service *RunnableService) {
	netConn := (net.Conn)(nil)
	err := (*base.Error)(nil)

	switch p.network {
	case "tcp":
		netConn, err = NewSyncClientTCP(p.network, p.addr, p.tlsConfig)
	default:
		panic("not implemented")
	}

	if err != nil {
		p.receiver.OnConnError(nil, err)
		return
	}

	p.conn = NewNetConn(netConn, p.rBufSize, p.wBufSize)
	p.conn.SetNext(NewStreamConn(p.conn, p.receiver))
	p.conn.OnOpen()
}

// OnRun ...
func (p *ClientAdapter) OnRun(service *RunnableService) {
	if conn := p.conn; conn != nil {
		for service.IsRunning() {
			if ok := p.conn.OnReadReady(); !ok {
				break
			}
		}
	}
}

// OnStop ...
func (p *ClientAdapter) OnStop(service *RunnableService) {
	// if OnStop is caused by Close(), don't close the conn again
	if p.conn != nil {
		if service.IsRunning() {
			p.conn.Close()
		}
		p.conn = nil
	}
}

// Close ...
func (p *ClientAdapter) Close() {
	if conn := p.conn; conn != nil {
		conn.Close()
	}
}
