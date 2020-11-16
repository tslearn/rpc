package client

import (
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"sync"
	"sync/atomic"
	"time"
)

const sendItemStatusRunning = int32(1)
const sendItemStatusFinish = int32(2)

type SendItem struct {
	id         uint64
	status     int32
	startTime  time.Time
	sendTime   time.Time
	timeout    time.Duration
	returnCH   chan *core.Stream
	sendStream *core.Stream
	next       *SendItem
}

var sendItemCache = &sync.Pool{
	New: func() interface{} {
		return &SendItem{
			sendStream: core.NewStream(),
		}
	},
}

func newSendItem() *SendItem {
	ret := sendItemCache.Get().(*SendItem)
	ret.id = 0
	ret.status = sendItemStatusRunning
	ret.startTime = time.Time{}
	ret.sendTime = time.Time{}
	ret.timeout = 0
	ret.returnCH = make(chan *core.Stream, 1)
	ret.next = nil
	return ret
}

func (p *SendItem) Return(stream *core.Stream) bool {
	if stream == nil {
		return false
	} else if !atomic.CompareAndSwapInt32(
		&p.status,
		sendItemStatusRunning,
		sendItemStatusFinish,
	) {
		stream.Release()
		return false
	} else {
		p.returnCH <- stream
		return true
	}
}

func (p *SendItem) Timeout() bool {
	if atomic.CompareAndSwapInt32(
		&p.status,
		sendItemStatusRunning,
		sendItemStatusFinish,
	) {
		// return timeout stream
		stream := core.NewStream()
		stream.SetCallbackID(p.sendStream.GetCallbackID())
		stream.WriteUint64(errors.ErrClientTimeout.GetCode())
		stream.WriteString(errors.ErrClientTimeout.GetMessage())
		stream.WriteString("")
		p.returnCH <- stream
		return true
	}

	return false
}

func (p *SendItem) Release() {
	close(p.returnCH)
	p.returnCH = nil
	p.sendStream.Reset()
	sendItemCache.Put(p)
}
