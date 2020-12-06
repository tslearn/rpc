package xadapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"net"
)

type Manager struct {
	listener    *TCPListener
	channels    []*Channel
	onError     func(err *base.Error)
	currChannel *Channel
	currRemains uint64
}

func NewManager(
	network string,
	addr string,
	onError func(err *base.Error),
	onConnect func(fd int, localAddr net.Addr, remoteAddr net.Addr) *ChannelConn,
	size int,
) *Manager {
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

	ret.listener = NewTCPListener(
		network,
		addr,
		func(fd int, localAddr net.Addr, remoteAddr net.Addr) {
			channelConn := onConnect(fd, localAddr, remoteAddr)
			if channelConn != nil {
				ret.AllocChannel().AddConn(channelConn)
			}
		},
		onError,
	)

	if ret.listener == nil {
		ret.Close()
		return nil
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
