package gateway

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
)

const u63Mask = 0x8000000000000000

type Channel struct {
	seq       uint64
	retTimeNS int64
	retStream *core.Stream
}

func (p *Channel) In(id uint64) (bool, *core.Stream) {
	if id > p.seq {
		p.seq = id
		return true, nil
	} else if id == p.seq {
		return false, p.retStream
	} else {
		return false, nil
	}
}

func (p *Channel) Out(stream *core.Stream) bool {
	id := stream.GetCallbackID()

	if id == p.seq {
		if p.retTimeNS != 0 {
			p.retTimeNS = base.TimeNow().UnixNano()
			p.retStream = stream
		}
		return true
	} else if id == 0 {
		return true
	} else {
		return false
	}
}

func (p *Channel) Timeout(nowNS int64, timeout int64) {
	if p.retTimeNS > 0 && nowNS-p.retTimeNS > timeout {
		p.Clean()
	}
}

func (p *Channel) Clean() {
	p.retTimeNS = 0
	if p.retStream != nil {
		p.retStream.Release()
		p.retStream = nil
	}
}
