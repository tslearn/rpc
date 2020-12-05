package xadapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
)

type ReadChannel struct {
	channel *Channel
	poller  *Poller
	connMap map[int]*ChannelConn
	addCH   chan *ChannelConn
}

func NewReadChannel(channel *Channel) *ReadChannel {
	ret := &ReadChannel{
		channel: channel,
		connMap: make(map[int]*ChannelConn),
		addCH:   make(chan *ChannelConn, 4096),
	}

	ret.poller = NewPoller(
		ret.onError,
		ret.onInvokeAdd,
		ret.onInvokeExit,
		ret.onFDRead,
		ret.onFDWrite,
		ret.onFDClose,
	)

	if ret.poller == nil {
		return nil
	}

	return ret
}

func (p *ReadChannel) AddConn(conn *ChannelConn) {
	_ = p.poller.InvokeAddTrigger()
	p.addCH <- conn
	_ = p.poller.InvokeAddTrigger()
}

func (p *ReadChannel) onError(err *base.Error) {
	p.channel.onError(err)
}

func (p *ReadChannel) onInvokeAdd() {
	for {
		select {
		case conn := <-p.addCH:
			if e := p.poller.RegisterReadFD(conn.GetFD()); e != nil {
				p.onError(errors.ErrKqueueSystem.AddDebug(e.Error()))
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
