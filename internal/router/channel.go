package router

import (
	"github.com/rpccloud/rpc/internal/rpc"
	"net"
	"sync"
	"time"
)

type Channel struct {
	isRunning bool
	slot      *Slot
	sequence  uint64
	buffers   [numOfCacheBuffer][]byte
	stream    *rpc.Stream
	streamPos int
	conn      net.Conn
	streamHub rpc.IStreamHub
	sync.Mutex
}

func NewChannel(streamHub rpc.IStreamHub) *Channel {
	ret := &Channel{
		isRunning: true,
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

func (p *Channel) IsNeedConnected() bool {
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
