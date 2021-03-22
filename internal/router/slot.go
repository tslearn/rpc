package router

import (
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

func (p *Slot) RunAt(
	index uint16,
	conn net.Conn,
	remoteSendSequence uint64,
	remoteReceiveSequence uint64,
) *base.Error {
	if index < numOfChannelPerSlot && conn != nil {
		channel := p.dataChannels[index]
		if err := channel.initSlaveConn(conn, remoteSendSequence, remoteReceiveSequence); err != nil {
			return err
		} else {
			go func() {
				channel.RunWithConn(conn)
			}()
			return nil
		}
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
