package client

import (
	"github.com/rpccloud/rpc/internal/core"
)

type Channel struct {
	seq  uint64
	item *SendItem
}

func (p *Channel) onCallbackStream(stream *core.Stream) {
	if p.item != nil {
		p.item.Return(stream)
	}
}
