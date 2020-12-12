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

// OnOpen ...
func (p *SyncClientAdapter) OnOpen() bool {
	netConn, e := net.Dial(p.network, p.addr)

	if e != nil {
		p.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(e.Error()))
		return false
	}

	p.conn = NewSyncConn(netConn, p.rBufSize, p.wBufSize)
	p.conn.SetNext(NewStreamConn(p.conn, p.receiver))
	p.conn.OnOpen()
	return true
}

// OnRun ...
func (p *SyncClientAdapter) OnRun(service *RunnableService) {
	for service.IsRunning() {
		if err := p.conn.TriggerRead(); err != nil {
			p.conn.OnError(err)
			break
		}
	}
}

// OnWillClose ...
func (p *SyncClientAdapter) OnWillClose() {
	// do nothing
}

// OnDidClose ...
func (p *SyncClientAdapter) OnDidClose() {
	p.conn.Close()
}
