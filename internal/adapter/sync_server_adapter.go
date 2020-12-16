package adapter

import (
	"crypto/tls"
	"net"

	"github.com/rpccloud/rpc/internal/errors"
)

// SyncServerAdapter ...
type SyncServerAdapter struct {
	network   string
	addr      string
	rBufSize  int
	wBufSize  int
	tlsConfig *tls.Config
	receiver  IReceiver
}

// NewSyncServerAdapter ...
func NewSyncServerAdapter(
	network string,
	addr string,
	tlsConfig *tls.Config,
	rBufSize int,
	wBufSize int,
	receiver IReceiver,
) *RunnableService {
	return NewRunnableService(&SyncServerAdapter{
		network:   network,
		addr:      addr,
		tlsConfig: tlsConfig,
		rBufSize:  rBufSize,
		wBufSize:  wBufSize,
		receiver:  receiver,
	})
}

// OnRun ...
func (p *SyncServerAdapter) OnRun(service *RunnableService) {
	switch p.network {
	case "tcp":
		p.runAsTCPServer(service)
	}
}

func (p *SyncServerAdapter) runAsTCPServer(service *RunnableService) {
	listener, e := net.Listen(p.network, p.addr)

	if e != nil {
		p.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(e.Error()))
		return
	}

	for service.IsRunning() {
		tcpConn, e := listener.Accept()

		if e != nil {
			p.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(e.Error()))
		} else {
			go func(netConn net.Conn) {
				conn := NewSyncConn(netConn, p.rBufSize, p.wBufSize)
				conn.SetNext(NewStreamConn(conn, p.receiver))
				conn.OnOpen()
				for {
					if ok := conn.OnReadReady(); !ok {
						break
					}
				}
				conn.Close()
				conn.OnClose()
			}(tcpConn)
		}
	}

	if e = listener.Close(); e != nil {
		p.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(e.Error()))
	}
}

// OnStop ...
func (p *SyncServerAdapter) OnStop(_ *RunnableService) {

}

// Close ...
func (p *SyncServerAdapter) Close() {
	// do nothing
}
