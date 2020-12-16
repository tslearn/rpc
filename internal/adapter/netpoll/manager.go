// +build linux darwin freebsd dragonfly

package netpoll

import (
	"github.com/rpccloud/rpc/internal/base"
)

// Manager ...
type Manager struct {
	channels   []*Channel
	onError    func(err *base.Error)
	curChannel *Channel
	curRemains uint64
}

// NewManager ...
func NewManager(
	onError func(err *base.Error),
	size int,
) *Manager {
	if size < 1 {
		size = 1
	}

	ret := &Manager{
		channels:   make([]*Channel, size),
		onError:    onError,
		curChannel: nil,
		curRemains: 0,
	}

	for i := 0; i < size; i++ {
		channel := NewChannel(onError)

		if channel != nil {
			ret.channels[i] = channel
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

// Close ...
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

// AllocChannel ...
func (p *Manager) AllocChannel() *Channel {
	if p.curRemains <= 0 {
		maxConn := int64(-1)
		for i := 0; i < len(p.channels); i++ {
			if connCount := p.channels[i].GetActiveConnCount(); connCount > maxConn {
				p.curChannel = p.channels[i]
				maxConn = connCount
			}
		}
		p.curRemains = 256
	}

	p.curRemains--
	return p.curChannel
}
