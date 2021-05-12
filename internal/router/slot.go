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
	controlCH    chan *rpc.Stream
	dataCH       chan *rpc.Stream
	dataChannels []*Channel
}

func NewSlot(connectMeta *ConnectMeta, streamHub rpc.IStreamHub) *Slot {
	ret := &Slot{
		controlCH:    make(chan *rpc.Stream, 1024),
		dataCH:       make(chan *rpc.Stream, 8192),
		dataChannels: make([]*Channel, numOfChannelPerSlot),
	}

	for i := 0; i < numOfChannelPerSlot; i++ {
		if i == 0 {
			ret.dataChannels[i] = NewChannel(
				uint16(i), connectMeta, ret.controlCH, streamHub,
			)
		} else {
			ret.dataChannels[i] = NewChannel(
				uint16(i), connectMeta, ret.dataCH, streamHub,
			)
		}
	}

	return ret
}

func (p *Slot) AddConn(conn net.Conn, initBuffer [32]byte) *base.Error {
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
