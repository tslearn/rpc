package adapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"sync/atomic"
)

type LoopChannel struct {
	manager         *LoopManager
	activeConnCount int64
	poller          *Poller
	connMap         map[int]*EventConn
	connCH          chan *EventConn
}

func NewLoopChannel(manager *LoopManager) *LoopChannel {
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

	return ret
}

func (p *LoopChannel) Close() {
	if err := p.poller.Close(); err != nil {
		p.onError(err)
	}
}

func (p *LoopChannel) AddConn(conn *EventConn) {
	_ = p.poller.InvokeAddTrigger()
	p.connCH <- conn
	_ = p.poller.InvokeAddTrigger()
}

func (p *LoopChannel) GetActiveConnCount() int64 {
	return atomic.LoadInt64(&p.activeConnCount)
}

func (p *LoopChannel) onTriggerAdd() {
	for {
		select {
		case conn := <-p.connCH:
			p.connMap[conn.GetFD()] = conn
			if e := p.poller.RegisterFD(conn.GetFD()); e != nil {
				conn.receiver.OnEventConnError(
					conn,
					errors.ErrKqueueSystem.AddDebug(e.Error()),
				)
			}
		default:
			return
		}
	}
}

func (p *LoopChannel) onTriggerExit() {

}

func (p *LoopChannel) onReadReady(fd int) {
	if conn, ok := p.connMap[fd]; ok {
		conn.OnReadyReady()
	}
}

func (p *LoopChannel) onClose(fd int) {
	if e := p.poller.UnregisterFD(fd); e != nil {
		p.manager.receiver.OnEventConnError(
			nil,
			errors.ErrKqueueSystem.AddDebug(e.Error()),
		)
	}

	if conn, ok := p.connMap[fd]; ok {
		delete(p.connMap, fd)
		conn.receiver.OnEventConnClose(conn)
	}
}

func (p *LoopChannel) onError(err *base.Error) {
	p.manager.receiver.OnEventConnError(nil, err)
}
