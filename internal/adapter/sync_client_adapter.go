package adapter

import (
	"net"

	"github.com/rpccloud/rpc/internal/errors"
)

// SyncClientAdapter ...
type SyncClientAdapter struct {
	network  string
	addr     string
	rBufSize int
	wBufSize int
	receiver IReceiver
	conn     *SyncConn
}

// NewSyncClientAdapter ...
func NewSyncClientAdapter(
	network string,
	addr string,
	rBufSize int,
	wBufSize int,
	receiver IReceiver,
) *RunnableService {
	return NewRunnableService(&SyncClientAdapter{
		network:  network,
		addr:     addr,
		rBufSize: rBufSize,
		wBufSize: wBufSize,
		receiver: receiver,
	})
}

// OnRun ...
func (p *SyncClientAdapter) OnRun(service *RunnableService) {
	netConn, e := net.Dial(p.network, p.addr)

	if e != nil {
		p.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(e.Error()))
		return
	}

	p.conn = NewSyncConn(netConn, p.rBufSize, p.wBufSize)
	p.conn.SetNext(NewStreamConn(p.conn, p.receiver))
	p.conn.OnOpen()

	for service.IsRunning() {
		if err := p.conn.TriggerRead(); err != nil {
			p.conn.OnError(err)
			break
		}
	}
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
