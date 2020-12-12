package netpoll

import (
	"sync"
	"sync/atomic"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
)

type Channel struct {
	onError         func(err *base.Error)
	activeConnCount int64
	poller          *Poller
	connMap         map[int]Conn
	addCH           chan Conn
	sync.Mutex
}

func NewChannel(onError func(err *base.Error)) *Channel {
	ret := &Channel{
		onError:         onError,
		activeConnCount: 0,
		poller:          nil,
		connMap:         make(map[int]Conn),
		addCH:           make(chan Conn, 4096),
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

func (p *Channel) Close() {
	p.Lock()
	defer p.Unlock()

	p.poller.Close()
}

func (p *Channel) AddConn(conn Conn) {
	p.addCH <- conn
	_ = p.poller.TriggerAddConn()
}

func (p *Channel) onInvokeAdd() {
	for {
		select {
		case conn := <-p.addCH:
			if e := p.poller.RegisterFD(conn.GetFD()); e != nil {
				p.onError(errors.ErrKqueueSystem.AddDebug(e.Error()))
			} else {
				p.connMap[conn.GetFD()] = conn
				conn.OnOpen()
			}
		default:
			return
		}
	}
}

func (p *Channel) onInvokeExit() {
	panic("not implement")
}

func (p *Channel) onFDRead(fd int) {
	if conn, ok := p.connMap[fd]; ok {
		conn.OnReadReady()
	}
}

func (p *Channel) onFDWrite(fd int) {
	if conn, ok := p.connMap[fd]; ok {
		conn.OnWriteReady()
	}
}

func (p *Channel) onFDClose(fd int) {
	if conn, ok := p.connMap[fd]; ok {
		if e := CloseFD(fd); e != nil {
			conn.OnError(errors.ErrKqueueSystem.AddDebug(e.Error()))
		} else {
			delete(p.connMap, fd)
			conn.OnClose()
		}
	}
}

func (p *Channel) GetActiveConnCount() int64 {
	return atomic.LoadInt64(&p.activeConnCount)
}
