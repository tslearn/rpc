package router

import (
	"encoding/binary"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
	"net"
	"sync"
	"time"
)

type Router struct {
	errorHub rpc.IStreamHub
	slotMap  map[uint64]*Slot
	sync.Mutex
}

func NewRouter(errorHub rpc.IStreamHub) *Router {
	return &Router{
		errorHub: errorHub,
		slotMap:  make(map[uint64]*Slot),
	}
}

func (p *Router) OnReceiveStream(s *rpc.Stream) {
	p.errorHub.OnReceiveStream(s)
}

func (p *Router) AddConn(conn net.Conn) *base.Error {
	var buffer [32]byte
	n, err := connReadBytes(conn, time.Second, buffer[:])

	if err != nil || n != 32 {
		_ = conn.Close()
		return err
	}

	if binary.LittleEndian.Uint16(buffer[2:]) != channelActionInit {
		_ = conn.Close()
		return base.ErrRouterConnProtocol
	}

	slotID := binary.LittleEndian.Uint64(buffer[6:])

	p.Lock()
	slot, ok := p.slotMap[slotID]
	if !ok {
		slot = NewSlot(nil, p)
		p.slotMap[slotID] = slot
	}
	p.Unlock()

	return slot.addSlaveConn(conn, buffer)
}

func (p *Router) DelSlot(id uint64) {
	p.Lock()
	defer p.Unlock()

	if slot, ok := p.slotMap[id]; ok {
		delete(p.slotMap, id)
		slot.Close()
	}
}
