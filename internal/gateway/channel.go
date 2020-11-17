package gateway

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
)

const u63Mask = 0x8000000000000000

type Channel struct {
	seq       uint64
	cacheSeq  uint64
	retTimeNS int64
	retStream *core.Stream
}

func (p *Channel) In(id uint64, gap uint64) (*core.Stream, *base.Error) {
	if p.seq == id {
		p.Clean()
		p.cacheSeq = p.seq
		p.seq = id + gap
		return nil, nil
	} else if p.seq == p.cacheSeq {
		return p.retStream, nil
	} else {
		return nil, errors.ErrGatewaySequenceError
	}
}

func (p *Channel) Out(stream *core.Stream) *base.Error {
	id := stream.GetCallbackID()

	if id == p.cacheSeq {
		if p.retTimeNS != 0 {
			p.retTimeNS = base.TimeNow().UnixNano()
		}
		return nil
	} else if id == 0 {
		return nil
	} else {
		return errors.ErrGatewaySequenceError
	}
}

func (p *Channel) Timeout(nowNS int64, timeout int64) {
	if p.retTimeNS > 0 && nowNS-p.retTimeNS > timeout {
		p.Clean()
	}
}

func (p *Channel) Clean() {
	p.cacheSeq = 0
	p.retTimeNS = 0
	if p.retStream != nil {
		p.retStream.Release()
		p.retStream = nil
	}
}
