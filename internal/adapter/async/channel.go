package async

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"sync"
	"sync/atomic"
)

type InnerChannel struct {
	isReadMode bool
	channel    *Channel
	poller     *Poller
	connMap    map[int]*AsyncConn
	addCH      chan *AsyncConn
}

func NewInnerChannel(isReadMode bool, channel *Channel) *InnerChannel {
	ret := &InnerChannel{
		isReadMode: isReadMode,
		channel:    channel,
		connMap:    make(map[int]*AsyncConn),
		addCH:      make(chan *AsyncConn, 4096),
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

func (p *InnerChannel) AddConn(conn *AsyncConn) {
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
		if e := closeFD(fd); e != nil {
			conn.OnError(errors.ErrKqueueSystem.AddDebug(e.Error()))
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
	p.poller.Close()
}

type Channel struct {
	onError         func(err *base.Error)
	activeConnCount int64
	rChannel        *InnerChannel
	wChannel        *InnerChannel
	sync.Mutex
}

func NewChannel(onError func(err *base.Error)) *Channel {
	ret := &Channel{
		onError:         onError,
		activeConnCount: 0,
		rChannel:        nil,
		wChannel:        nil,
	}

	ret.rChannel = NewInnerChannel(true, ret)
	ret.wChannel = NewInnerChannel(false, ret)

	if ret.rChannel == nil || ret.wChannel == nil {
		ret.Close()
		return nil
	}

	return ret
}

func (p *Channel) Close() {
	p.Lock()
	defer p.Unlock()

	if p.rChannel != nil {
		p.rChannel.Close()
		p.rChannel = nil
	}

	if p.wChannel != nil {
		p.wChannel.Close()
		p.wChannel = nil
	}
}

func (p *Channel) AddConn(conn *AsyncConn) {
	p.rChannel.AddConn(conn)
	p.wChannel.AddConn(conn)
}

func (p *Channel) GetActiveConnCount() int64 {
	return atomic.LoadInt64(&p.activeConnCount)
}
