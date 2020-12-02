package adapter

import (
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

func NewLoopChannel(manager *LoopManager) (*LoopChannel, error) {
	poller, err := NewPoller()

	if err != nil {
		return nil, err
	}

	return &LoopChannel{
		manager:         manager,
		activeConnCount: 0,
		poller:          poller,
		connMap:         make(map[int]*EventConn),
		connCH:          make(chan *EventConn, 4096),
	}, nil
}

func (p *LoopChannel) AddConn(conn *EventConn) {
	_ = p.poller.InvokeAddTrigger()
	p.connCH <- conn
	_ = p.poller.InvokeAddTrigger()
}

func (p *LoopChannel) GetActiveConnCount() int64 {
	return atomic.LoadInt64(&p.activeConnCount)
}

func (p *LoopChannel) onAddConn() {
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

func (p *LoopChannel) onRead(fd int) {
	if conn, ok := p.connMap[fd]; ok {
		if err := conn.reOnRead(); err != nil {
			p.onError(err)
		}
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

func (p *LoopChannel) onExit() {

}

func (p *LoopChannel) Open() {
	if err := p.poller.Polling(
		p.onAddConn,
		p.onRead,
		p.onClose,
		p.onExit,
	); err != nil {
		p.manager.receiver.OnEventConnError(nil, err)
	}
}

func (p *LoopChannel) Close() error {
	return nil
}
