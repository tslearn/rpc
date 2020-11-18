package client

import (
	"github.com/rpccloud/rpc/internal/core"
)

type Channel struct {
	id     int
	seq    uint64
	item   *SendItem
	client *Client
}

func (p *Channel) onCallbackStream(stream *core.Stream) bool {
	if p.item != nil {
		if p.item.Return(stream) {
			p.item = nil
			p.client.freeChannels <- p.id
			p.seq += uint64(len(p.client.channels))
			return true
		}
	}

	return false
}
