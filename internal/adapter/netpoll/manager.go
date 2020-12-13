// +build linux darwin freebsd dragonfly

package netpoll

import (
	"net"

	"github.com/rpccloud/rpc/internal/base"
)

// Manager ...
type Manager struct {
	listener    *TCPListener
	channels    []*Channel
	onError     func(err *base.Error)
	currChannel *Channel
	currRemains uint64
}

// NewManager ...
func NewManager(
	network string,
	addr string,
	onError func(err *base.Error),
	onConnect func(channel *Channel, fd int, lAddr net.Addr, rAddr net.Addr) Conn,
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
			channel := ret.AllocChannel()
			if conn := onConnect(channel, fd, localAddr, remoteAddr); conn != nil {
				channel.AddConn(conn)
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
