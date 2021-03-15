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
	slotMap := make(map[uint64]*SlotManager)
	return &Router{
		errorHub: errorHub,
		slotMap:  unsafe.Pointer(&slotMap),
	}
}

func (p *Router) OnReceiveStream(s *rpc.Stream) {

}

func (p *Router) AddSlot(id uint64, conn net.Conn, channelID uint16) {
	p.Lock()
	defer p.Unlock()

	slotMap := *(*map[uint64]*SlotManager)(atomic.LoadPointer(&p.slotMap))
	oldSlotMgr, ok := slotMap[id]
	newSlotMgr := NewSlotManager(p)

	if ok {
		for i := 0; i < len(oldSlotMgr.channels); i++ {
			newSlotMgr.channels[i] = oldSlotMgr.channels[i]
		}
	}

	if int(channelID) < len(newSlotMgr.channels) {
		newSlotMgr.channels[channelID].Close()

		newSlotMgr.channels[channelID] = NewChannel(p)
	}

}

func (p *Router) DelSlot(id uint64) {
	p.Lock()
	defer p.Unlock()

}
