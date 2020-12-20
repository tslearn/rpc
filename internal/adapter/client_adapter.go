package adapter

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"

	"github.com/gobwas/ws"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
)

// NewClientTCP ...
func NewClientTCP(
	network string,
	addr string,
	tlsConfig *tls.Config,
	onError func(err *base.Error),
) net.Conn {
	ret := net.Conn(nil)
	e := error(nil)

	if tlsConfig == nil {
		ret, e = net.Dial(network, addr)
	} else {
		ret, e = tls.Dial(network, addr, tlsConfig)
	}

	if e != nil {
		onError(errors.ErrTemp.AddDebug(e.Error()))
		return nil
	}

	return ret
}

// NewClientWebsocket ...
func NewClientWebsocket(
	network string,
	addr string,
	tlsConfig *tls.Config,
	onError func(err *base.Error),
) net.Conn {
	dialer := &ws.Dialer{}

	if network == "wss" {
		dialer.TLSConfig = tlsConfig
	}

	u := url.URL{Scheme: network, Host: addr, Path: "/"}
	conn, _, _, e := dialer.Dial(context.Background(), u.String())

	fmt.Println("DDDD", e)
	if e != nil {
		onError(errors.ErrTemp.AddDebug(e.Error()))
		return nil
	}

	return conn
}

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

	switch p.network {
	case "tcp":
		netConn = NewClientTCP(p.network, p.addr, p.tlsConfig, func(err *base.Error) {
			p.receiver.OnConnError(nil, err)
		})
	case "ws":
		fallthrough
	case "wss":
		netConn = NewClientWebsocket(p.network, p.addr, p.tlsConfig, func(err *base.Error) {
			p.receiver.OnConnError(nil, err)
		})
	default:
		panic("not implemented")
	}

	if netConn == nil {
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
	// if OnStop is caused by Close(), don't close again
	if service.IsRunning() {
		p.Close()
	}
}

// Close ...
func (p *ClientAdapter) Close() {
	if conn := p.conn; conn != nil {
		conn.Close()
	}
}
