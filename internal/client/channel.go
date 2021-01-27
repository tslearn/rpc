package client

import (
	"github.com/rpccloud/rpc/internal/base"
	"time"

	"github.com/rpccloud/rpc/internal/core"
)

// Channel ...
type Channel struct {
	sequence uint64
	item     *SendItem
}

func (p *Channel) Use(item *SendItem, channelSize int) bool {
	if p.item == nil {
		p.sequence += uint64(channelSize)
		item.id = p.sequence
		item.sendStream.SetCallbackID(p.sequence)
		p.item = item
		p.item.sendTime = base.TimeNow()
		return true
	}

	return false
}

func (p *Channel) Free(stream *core.Stream) bool {
	if item := p.item; item != nil {
		p.item = nil
		return item.Return(stream)
	}

	return false
}

func (p *Channel) OnTimeout(now time.Time) bool {
	if p.item != nil {
		if p.item.CheckAndTimeout(now) {
			p.item = nil
			return true
		}
	}

	return false
}
