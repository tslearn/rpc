package xadapter

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

type LoopChannel struct {
	activeConnCount int64
	poller          *Poller
	connMap         map[int]XConn
	sync.Mutex
}

func NewLoopChannel() (*LoopChannel, error) {
	poller, err := OpenPoller()

	if err != nil {
		return nil, err
	}

	return &LoopChannel{
		activeConnCount: 0,
		poller:          poller,
		connMap:         make(map[int]XConn),
	}, nil
}

func (p *LoopChannel) GetActiveConnCount() int64 {
	return atomic.LoadInt64(&p.activeConnCount)
}

func (p *LoopChannel) AddRead(conn XConn) error {
	p.Lock()
	defer p.Unlock()
	p.connMap[conn.FD()] = conn
	return p.poller.Add(conn.FD())
}

func (p *LoopChannel) GetConnByFD(fd int) (XConn, bool) {
	p.Lock()
	defer p.Unlock()

	conn, ok := p.connMap[fd]
	return conn, ok
}

func (p *LoopChannel) Delete(fd int) error {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.connMap[fd]; ok {
		if err := p.poller.Delete(fd); err != nil {
			return err
		}
		delete(p.connMap, fd)
		return nil
	} else {
		return errors.New("conn not exist")
	}
}

func (p *LoopChannel) TickRead() error {
	return p.poller.Polling(func(fd int, filter int16) error {
		if conn, ok := p.GetConnByFD(fd); ok {
			if filter == EVFilterSock {

				if err := p.Delete(fd); err != nil {
					fmt.Println(err)
					return err
				}

				return conn.OnClose()
				// closed
			} else {
				return conn.OnRead()
			}
		}

		return nil
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
