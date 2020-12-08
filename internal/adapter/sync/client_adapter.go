package sync

import (
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
)

type ClientAdapter struct {
	network  string
	addr     string
	rBufSize int
	wBufSize int
	receiver adapter.IReceiver
	conn     *Conn
}

func NewSyncClientAdapter(
	network string,
	addr string,
	rBufSize int,
	wBufSize int,
	receiver adapter.IReceiver,
) *adapter.RunnableService {
	return adapter.NewRunnableService(&ClientAdapter{
		network:  network,
		addr:     addr,
		rBufSize: rBufSize,
		wBufSize: wBufSize,
		receiver: receiver,
	})
}

func (p *ClientAdapter) OnOpen() bool {
	if netConn, e := net.Dial(p.network, p.addr); e != nil {
		p.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(e.Error()))
		return false
	} else {
		p.conn = NewConn(netConn, p.rBufSize, p.wBufSize)
		p.conn.SetNext(adapter.NewStreamConn(p.conn, p.receiver))
		p.conn.OnOpen()
		return true
	}
}

func (p *ClientAdapter) OnRun(service *adapter.RunnableService) {
	for service.IsRunning() {
		if err := p.conn.TriggerRead(); err != nil {
			p.conn.OnError(err)
			break
		}
	}
}

func (p *ClientAdapter) OnWillClose() {

}

func (p *ClientAdapter) OnDidClose() {
	p.conn.Close()
}
