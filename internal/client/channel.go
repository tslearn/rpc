package client

import (
	"github.com/rpccloud/rpc/internal/core"
	"time"
)

type Channel struct {
	id     int
	seq    uint64
	item   *SendItem
	client *Client
}

func (p *Channel) OnCallbackStream(stream *core.Stream) bool {
	if p.item != nil {
		if p.item.Return(stream) {
			p.free()
			return true
		}
	}

	return false
}

func (p *Channel) OnTimeout(now time.Time) bool {
	if p.item != nil {
		if p.item.CheckAndTimeout(now) {
			p.free()
			return true
		}
	}

	return false
}

func (p *Channel) free() {
	p.item = nil
	p.client.freeChannels.Push(p.id)
	p.seq += uint64(len(p.client.channels))
}
