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
	p.errorHub.OnReceiveStream(s)
}

func (p *Router) AddSlot(
	slotID uint64,
	conn net.Conn,
	channelID uint16,
	remoteSendSequence uint64,
	remoteReceiveSequence uint64,
) {
	p.Lock()
	defer p.Unlock()

	slotMap := *(*map[uint64]*Slot)(atomic.LoadPointer(&p.slotMap))
	slot, ok := slotMap[slotID]

	if !ok {
		slot = NewSlot(nil, p)
		slotMap[slotID] = slot
	}

	if int(channelID) < len(slot.dataChannels) {
		slot.RunAt(channelID, conn, remoteSendSequence, remoteReceiveSequence)
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
