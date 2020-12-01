package xadapter

import (
	"sync/atomic"
)

type LoopChannel struct {
	activeConnCount int64
	poller          *Poller
	connMap         map[int]XConn
	connCH          chan XConn
}

func NewLoopChannel() (*LoopChannel, error) {
	poller, err := NewPoller()

	if err != nil {
		return nil, err
	}

	return &LoopChannel{
		activeConnCount: 0,
		poller:          poller,
		connMap:         make(map[int]XConn),
		connCH:          make(chan XConn, 4096),
	}, nil
}

func (p *LoopChannel) AddConn(conn XConn) {
	_ = p.poller.Trigger()
	p.connCH <- conn
	_ = p.poller.Trigger()
}

func (p *LoopChannel) GetActiveConnCount() int64 {
	return atomic.LoadInt64(&p.activeConnCount)
}

func (p *LoopChannel) TickRead() {
	p.poller.Polling(func(fd int, isClose bool) {
		if conn, ok := p.connMap[fd]; ok {
			if isClose {
				_ = p.poller.Delete(fd)
				delete(p.connMap, fd)
				conn.OnClose()
			} else {
				conn.OnRead()
			}
		}
	}, func() {
		for {
			select {
			case conn := <-p.connCH:
				p.connMap[conn.FD()] = conn
				p.poller.Add(conn.FD())
			default:
				return
			}
		}
	})
}

func (p *LoopChannel) Open() {
	for {
		p.TickRead()
	}
}

func (p *LoopChannel) Close() error {
	return nil
}

type LoopManager struct {
	channels []*LoopChannel

	currChannel *LoopChannel
	currRemains uint64
}

func NewLoopManager(channels int) (*LoopManager, error) {
	if channels < 1 {
		channels = 1
	}

	ret := &LoopManager{
		channels:    make([]*LoopChannel, channels),
		currChannel: nil,
		currRemains: 0,
	}

	for i := 0; i < channels; i++ {
		if channel, err := NewLoopChannel(); err != nil {
			return nil, err
		} else {
			ret.channels[i] = channel
		}

	}

	return ret, nil
}

func (p *LoopManager) Open() {
	for i := 0; i < len(p.channels); i++ {
		go func(idx int) {
			p.channels[idx].Open()
		}(i)
	}
}

func (p *LoopManager) Close() error {
	return nil
}

func (p *LoopManager) AllocChannel() *LoopChannel {
	if p.currRemains <= 0 {
		maxConn := int64(-1)
		for i := 0; i < len(p.channels); i++ {
			if connCount := p.channels[i].GetActiveConnCount(); connCount > maxConn {
				p.currChannel = p.channels[i]
				maxConn = connCount
			}
		}
		p.currRemains = 256
	}

	p.currRemains--
	return p.currChannel
}
