package xadapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"golang.org/x/sys/unix"
	"sync/atomic"
)

type InnerChannel struct {
	isReadMode bool
	channel    *Channel
	poller     *Poller
	connMap    map[int]*ChannelConn
	addCH      chan *ChannelConn
}

func NewInnerChannel(isReadMode bool, channel *Channel) *InnerChannel {
	ret := &InnerChannel{
		isReadMode: isReadMode,
		channel:    channel,
		connMap:    make(map[int]*ChannelConn),
		addCH:      make(chan *ChannelConn, 4096),
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

func (p *InnerChannel) AddConn(conn *ChannelConn) {
	_ = p.poller.InvokeAddTrigger()
	p.addCH <- conn
	_ = p.poller.InvokeAddTrigger()
}

func (p *InnerChannel) onError(err *base.Error) {
	p.channel.onError(err)
}

func (p *InnerChannel) onInvokeAdd() {
	for {
		select {
		case conn := <-p.addCH:
			if e := p.poller.RegisterReadFD(conn.GetFD()); e != nil {
				p.onError(errors.ErrKqueueSystem.AddDebug(e.Error()))
			} else {
				p.connMap[conn.GetFD()] = conn
				if p.isReadMode {
					conn.OnReadOpen()
				} else {
					conn.OnWriteOpen()
				}
			}
		default:
			return
		}
	}
}

func (p *InnerChannel) onInvokeExit() {
	panic("not implement")
}

func (p *InnerChannel) onFDRead(fd int) {
	if conn, ok := p.connMap[fd]; ok {
		conn.OnReadReady()
	}
}

func (p *InnerChannel) onFDWrite(fd int) {
	if conn, ok := p.connMap[fd]; ok {
		conn.OnWriteReady()
	}
}

func (p *InnerChannel) onFDClose(fd int) {
	if conn, ok := p.connMap[fd]; ok {
		if err := p.channel.CloseFD(fd); err != nil {
			conn.OnError(err)
		} else {
			delete(p.connMap, fd)
			if p.isReadMode {
				conn.OnReadClose()
			} else {
				conn.OnWriteClose()
			}
		}
	}
}

func (p *InnerChannel) Close() {
	if err := p.poller.Close(); err != nil {
		p.onError(err)
	}
}

type Channel struct {
	onError         func(err *base.Error)
	activeConnCount int64
	rChannel        *InnerChannel
	wChannel        *InnerChannel
}

func NewChannel(onError func(err *base.Error)) *Channel {
	ret := &Channel{
		onError:         onError,
		activeConnCount: 0,
		rChannel:        nil,
		wChannel:        nil,
	}

	ret.rChannel = NewInnerChannel(true, ret)

	if ret.rChannel == nil {
		return nil
	}

	ret.wChannel = NewInnerChannel(false, ret)
	if ret.wChannel == nil {
		ret.rChannel.Close()
		return nil
	}

	return ret
}

func (p *Channel) CloseFD(fd int) *base.Error {
	if e := unix.Close(fd); e != nil {
		return errors.ErrKqueueSystem.AddDebug(e.Error())
	}

	return nil
}

func (p *Channel) GetActiveConnCount() int64 {
	return atomic.LoadInt64(&p.activeConnCount)
}
