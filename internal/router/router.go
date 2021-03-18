package router

import (
	"github.com/rpccloud/rpc/internal/rpc"
	"net"
	"sync"
	"sync/atomic"
	"unsafe"
)

type Router struct {
	errorHub rpc.IStreamHub
	slotMap  unsafe.Pointer
	sync.Mutex
}

func NewRouter(errorHub rpc.IStreamHub) *Router {
	slotMap := make(map[uint64]*Slot)
	return &Router{
		errorHub: errorHub,
		slotMap:  unsafe.Pointer(&slotMap),
	}
}

func (p *Router) OnReceiveStream(s *rpc.Stream) {

}

func (p *Router) AddSlot(slotID uint64, conn net.Conn, channelID uint16) {
	p.Lock()
	defer p.Unlock()

	slotMap := *(*map[uint64]*Slot)(atomic.LoadPointer(&p.slotMap))
	oldSlot, ok := slotMap[slotID]
	newSlot := NewSlot(p)

	if ok {
		for i := 0; i < len(oldSlot.dataChannels); i++ {
			newSlot.dataChannels[i] = oldSlot.dataChannels[i]
		}
	}

	if int(channelID) < len(newSlot.dataChannels) {
		newSlot.dataChannels[channelID].setConn(conn)
		slotMap[slotID] = newSlot
	}
}

func (p *Router) DelSlot(id uint64) {
	p.Lock()
	defer p.Unlock()

	slotMap := *(*map[uint64]*Slot)(atomic.LoadPointer(&p.slotMap))
	slot, ok := slotMap[id]
	if ok {
		delete(slotMap, id)
		slot.Close()
	}
}
