package router

import (
	"encoding/binary"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
	"net"
)

const (
	numOfChannelPerSlot = 8
	numOfCacheBuffer    = 512
	bufferSize          = 65536
)

type Slot struct {
	dataCH       chan *rpc.Stream
	dataChannels []*Channel
}

func NewSlot(connectMeta *ConnectMeta, streamHub rpc.IStreamHub) *Slot {
	ret := &Slot{
		dataCH:       make(chan *rpc.Stream, 8192),
		dataChannels: make([]*Channel, numOfChannelPerSlot),
	}

	for i := 0; i < numOfChannelPerSlot; i++ {
		ret.dataChannels[i] = NewChannel(
			uint16(i), connectMeta, ret.dataCH, streamHub,
		)
	}

	return ret
}

func (p *Slot) AddSlaveConn(conn net.Conn, initBuffer [32]byte) *base.Error {
	index := binary.LittleEndian.Uint16(initBuffer[4:])
	if index < numOfChannelPerSlot && conn != nil {
		channel := p.dataChannels[index]
		if err := channel.runSlaveThread(conn, initBuffer); err != nil {
			return err
		}

		return nil
	} else {
		return base.ErrRouterConnProtocol
	}
}

func (p *Slot) SendStream(s *rpc.Stream) {
	if s != nil {
		p.dataCH <- s
	}
}

func (p *Slot) Close() {
	for i := 0; i < numOfChannelPerSlot; i++ {
		p.dataChannels[i].Close()
	}
}
