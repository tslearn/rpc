package client

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
)

// Channel ...
type Channel struct {
	sequence uint64
	item     *SendItem
}

// Use ...
func (p *Channel) Use(item *SendItem, channelSize int) bool {
	if p.item == nil {
		p.sequence += uint64(channelSize)
		item.sendStream.SetCallbackID(p.sequence)
		p.item = item
		p.item.sendTimeNS = base.TimeNow().UnixNano()
		return true
	}

	return false
}

// Free ...
func (p *Channel) Free(stream *core.Stream) bool {
	if item := p.item; item != nil {
		p.item = nil
		return item.Return(stream)
	}

	return false
}

// CheckTime ...
func (p *Channel) CheckTime(nowNS int64) bool {
	if p.item != nil && p.item.CheckTime(nowNS) {
		p.item = nil
		return true
	}

	return false
}
