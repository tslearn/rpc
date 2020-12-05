package xadapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"sync/atomic"
)

type Channel struct {
	onError         func(err *base.Error)
	activeConnCount int64
}

func NewChannel(onError func(err *base.Error)) *Channel {
	ret := &LoopChannel{
		manager:         manager,
		activeConnCount: 0,
		poller:          nil,
		connMap:         make(map[int]*EventConn),
		connCH:          make(chan *EventConn, 4096),
	}

	ret.poller = NewPoller(
		ret.onTriggerAdd,
		ret.onTriggerExit,
		ret.onReadReady,
		ret.onClose,
		ret.onError,
	)

	if ret.poller == nil {
		return nil
	}

	return ret
}

func (p *Channel) CloseFD(fd int) {

}

func (p *LoopChannel) Close() {
	if err := p.poller.Close(); err != nil {
		p.onError(err)
	}
}

func (p *LoopChannel) GetActiveConnCount() int64 {
	return atomic.LoadInt64(&p.activeConnCount)
}
