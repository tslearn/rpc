package adapter

import (
	"github.com/rpccloud/rpc/internal/errors"
	"net"
)

type SyncClientAdapter struct {
	network  string
	addr     string
	rBufSize int
	wBufSize int
	receiver IReceiver
	conn     *SyncConn
}

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

func (p *SyncClientAdapter) OnOpen() bool {
	if netConn, e := net.Dial(p.network, p.addr); e != nil {
		p.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(e.Error()))
		return false
	} else {
		p.conn = NewSyncConn(netConn, p.rBufSize, p.wBufSize)
		p.conn.SetNext(NewStreamConn(p.conn, p.receiver))
		p.conn.OnOpen()
		return true
	}
}

func (p *SyncClientAdapter) OnRun(service *RunnableService) {
	for service.IsRunning() {
		if err := p.conn.TriggerRead(); err != nil {
			p.conn.OnError(err)
			break
		}
	}
}

func (p *SyncClientAdapter) OnWillClose() {

}

func (p *SyncClientAdapter) OnDidClose() {
	p.conn.Close()
}
