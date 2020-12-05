package xadapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
)

type ReadChannel struct {
	channel *Channel
	poller  *Poller
	connMap map[int]*PollConn
	addCH   chan *PollConn
}

func NewReadChannel(channel *Channel) *ReadChannel {
	ret := &ReadChannel{
		channel: channel,
		connMap: make(map[int]*PollConn),
		addCH:   make(chan *PollConn, 4096),
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

func (p *ReadChannel) AddConn(conn *PollConn) {
	_ = p.poller.InvokeAddTrigger()
	p.addCH <- conn
	_ = p.poller.InvokeAddTrigger()
}

func (p *ReadChannel) onTriggerAdd() {
	for {
		select {
		case pollConn := <-p.addCH:
			if e := p.poller.RegisterFD(pollConn.conn.GetFD()); e != nil {
				conn.receiver.OnEventConnError(
					conn,
					errors.ErrKqueueSystem.AddDebug(e.Error()),
				)
			} else {
				p.connMap[conn.GetFD()] = conn
				conn.receiver.OnEventConnOpen(conn)
			}
		default:
			return
		}
	}
}

func (p *ReadChannel) onTriggerExit() {

}

func (p *ReadChannel) onReadReady(fd int) {
	if conn, ok := p.connMap[fd]; ok {
		conn.OnReadReady()
	}
}

func (p *ReadChannel) onClose(fd int) {
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

func (p *ReadChannel) onError(err *base.Error) {
	p.manager.receiver.OnEventConnError(nil, err)
}
