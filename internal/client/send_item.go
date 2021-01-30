package client

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"sync"
)

var sendItemCache = &sync.Pool{
	New: func() interface{} {
		return &SendItem{
			returnCH:   make(chan *core.Stream, 1),
			sendStream: core.NewStream(),
		}
	},
}

// SendItem ...
type SendItem struct {
	isRunning   bool
	startTimeNS int64
	sendTimeNS  int64
	timeoutNS   int64
	returnCH    chan *core.Stream
	sendStream  *core.Stream
	next        *SendItem
}

func NewSendItem(timeoutNS int64) *SendItem {
	ret := sendItemCache.Get().(*SendItem)
	ret.isRunning = true
	ret.startTimeNS = base.TimeNow().UnixNano()
	ret.sendTimeNS = 0
	ret.timeoutNS = timeoutNS
	ret.next = nil
	return ret
}

// Return ...
func (p *SendItem) Return(stream *core.Stream) bool {
	if stream == nil || !p.isRunning {
		return false
	}

	p.returnCH <- stream
	return true
}

// CheckTime ...
func (p *SendItem) CheckTime(nowNS int64) bool {
	if nowNS-p.startTimeNS > p.timeoutNS && p.isRunning {
		p.isRunning = false

		// return timeout stream
		stream := core.NewStream()
		stream.SetCallbackID(p.sendStream.GetCallbackID())
		stream.WriteUint64(errors.ErrClientTimeout.GetCode())
		stream.WriteString(errors.ErrClientTimeout.GetMessage())
		p.returnCH <- stream
		return true
	}

	return false
}

// Release ...
func (p *SendItem) Release() {
	p.sendStream.Reset()
	sendItemCache.Put(p)
}
