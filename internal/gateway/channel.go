package gateway

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
)

// Channel ...
type Channel struct {
	sequence   uint64
	backTimeNS int64
	backStream *core.Stream
}

// In ...
func (p *Channel) In(id uint64) (canIn bool, backStream *core.Stream) {
	if id > p.sequence {
		p.Clean()
		p.sequence = id
		return true, nil
	} else if id == p.sequence {
		return false, p.backStream
	} else {
		return false, nil
	}
}

// Out ...
func (p *Channel) Out(stream *core.Stream) (canOut bool) {
	id := stream.GetCallbackID()

	if id == p.sequence {
		if p.backTimeNS == 0 {
			p.backTimeNS = base.TimeNow().UnixNano()
			p.backStream = stream
			return true
		}
		return false
	} else if id == 0 {
		return true
	} else {
		return false
	}
}

// TimeCheck ...
func (p *Channel) TimeCheck(nowNS int64, timeout int64) {
	if p.backTimeNS > 0 && nowNS-p.backTimeNS > timeout {
		p.Clean()
	}
}

// Clean ...
func (p *Channel) Clean() {
	p.backTimeNS = 0
	if p.backStream != nil {
		p.backStream.Release()
		p.backStream = nil
	}
}
