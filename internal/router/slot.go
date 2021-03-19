package router

import (
	"github.com/rpccloud/rpc/internal/rpc"
	"net"
)

const (
	numOfChannelPerSlot = 8
	numOfCacheBuffer    = 512
	bufferSize          = 65536
)

type Slot struct {
	controlCH    chan *rpc.Stream
	dataCH       chan *rpc.Stream
	dataChannels []*Channel
}

func NewSlot(streamHub rpc.IStreamHub) *Slot {
	ret := &Slot{
		controlCH:    make(chan *rpc.Stream, 1024),
		dataCH:       make(chan *rpc.Stream, 8192),
		dataChannels: make([]*Channel, numOfChannelPerSlot),
	}

	for i := 0; i < numOfChannelPerSlot; i++ {
		ret.dataChannels[i] = NewChannel(i == 0, streamHub)
	}

	return ret
}

func (p *Slot) RunAt(index uint16, conn net.Conn) {
	if index < numOfChannelPerSlot && conn != nil {
		p.dataChannels[index].RunWithConn(conn)
	}
}

func (p *Slot) GetFreeChannels() []uint16 {
	ret := []uint16(nil)

	for i := 0; i < numOfChannelPerSlot; i++ {
		if p.dataChannels[i].IsNeedConnected() {
			ret = append(ret, uint16(i))
		}
	}

	return ret
}

func (p *Slot) SendStream(s *rpc.Stream) {
	if s != nil {
		if s.GetKind() == rpc.StreamKindRouterControl {
			p.controlCH <- s
		} else {
			p.dataCH <- s
		}
	}
}

func (p *Slot) Close() {
	for i := 0; i < numOfChannelPerSlot; i++ {
		p.dataChannels[i].Close()
	}
}
