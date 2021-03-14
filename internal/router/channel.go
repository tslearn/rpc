package router

import (
	"github.com/rpccloud/rpc/internal/rpc"
	"net"
	"sync"
	"time"
)

const (
	numOfChannelPerSlot = 8
	numOfCacheBuffer    = 512
	bufferSize          = 65536
)

type ChannelManager struct {
	inputCH  chan *rpc.Stream
	channels []*Channel
}

func NewChannelManager(streamHub rpc.IStreamHub) *ChannelManager {
	ret := &ChannelManager{
		inputCH:  make(chan *rpc.Stream, 8192),
		channels: make([]*Channel, numOfChannelPerSlot),
	}

	for i := 0; i < numOfChannelPerSlot; i++ {
		ret.channels[i] = NewChannel(ret, streamHub)
	}

	return ret
}

func (p *ChannelManager) RunAt(index uint16, conn net.Conn) bool {
	if index < numOfChannelPerSlot && conn != nil {
		return p.channels[index].RunWithConn(conn)
	}

	return false
}

func (p *ChannelManager) GetFreeChannels() []uint16 {
	ret := []uint16(nil)

	for i := 0; i < numOfChannelPerSlot; i++ {
		if p.channels[i].IsNotConnected() {
			ret = append(ret, uint16(i))
		}
	}

	return ret
}

func (p *ChannelManager) Close() {
	for i := 0; i < numOfChannelPerSlot; i++ {
		p.channels[i].Close()
	}
}

type Channel struct {
	isRunning bool
	manager   *ChannelManager
	sequence  uint64
	buffers   [numOfCacheBuffer][]byte
	stream    *rpc.Stream
	streamPos int
	conn      net.Conn
	streamHub rpc.IStreamHub
	sync.Mutex
}

func NewChannel(manager *ChannelManager, streamHub rpc.IStreamHub) *Channel {
	ret := &Channel{
		isRunning: true,
		manager:   manager,
		sequence:  0,
		stream:    nil,
		streamPos: -1,
		conn:      nil,
		streamHub: streamHub,
	}

	for i := 0; i < numOfCacheBuffer; i++ {
		ret.buffers[i] = make([]byte, bufferSize)
	}

	return ret
}

func (p *Channel) IsNotConnected() bool {
	p.Lock()
	defer p.Unlock()

	return p.isRunning && p.conn == nil
}

func (p *Channel) getConn() net.Conn {
	p.Lock()
	defer p.Unlock()
	return p.conn
}

func (p *Channel) setConn(conn net.Conn) bool {
	p.Lock()
	defer p.Unlock()

	if !p.isRunning {
		return false
	}

	p.conn = conn
	return true
}

func (p *Channel) waitUntilConnIsNil() {
	for p.getConn() != nil {
		time.Sleep(50 * time.Millisecond)
	}
}

func (p *Channel) RunWithConn(conn net.Conn) bool {
	p.waitUntilConnIsNil()

	if !p.setConn(conn) {
		return false
	}

	waitCH := make(chan bool)

	go func() {
		p.runRead(conn)
		_ = conn.Close()
		waitCH <- true
	}()

	go func() {
		p.runWrite(conn)
		_ = conn.Close()
		waitCH <- true
	}()

	<-waitCH
	<-waitCH

	p.setConn(nil)
	return true
}

func (p *Channel) runRead(conn net.Conn) {

}

func (p *Channel) runWrite(conn net.Conn) {

}

func (p *Channel) Close() {
	func() {
		p.Lock()
		defer p.Unlock()
		p.isRunning = false
		if p.conn != nil {
			_ = p.conn.Close()
		}
	}()

	p.waitUntilConnIsNil()
}
