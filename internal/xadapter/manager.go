package xadapter

import "github.com/rpccloud/rpc/internal/base"

type Manager struct {
	channels    []*Channel
	onError     func(err *base.Error)
	currChannel *Channel
	currRemains uint64
}

func NewManager(size int, onError func(err *base.Error)) *Manager {
	if size < 1 {
		size = 1
	}

	ret := &Manager{
		channels:    make([]*Channel, size),
		onError:     onError,
		currChannel: nil,
		currRemains: 0,
	}

	for i := 0; i < size; i++ {
		poll := NewChannel(onError)

		if poll != nil {
			ret.channels[i] = poll
		} else {
			// clean up and return nil
			for j := 0; j < i; j++ {
				ret.channels[j].Close()
				ret.channels[j] = nil
			}
			return nil
		}
	}

	return ret
}

func (p *Manager) Close() {
	waitCH := make(chan bool)
	channelSize := len(p.channels)

	for i := 0; i < channelSize; i++ {
		go func(idx int) {
			p.channels[idx].Close()
			waitCH <- true
		}(i)
	}

	for i := 0; i < channelSize; i++ {
		<-waitCH
	}
}

func (p *Manager) AllocChannel() *Channel {
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
