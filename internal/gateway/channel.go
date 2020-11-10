package gateway

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"sync/atomic"
	"unsafe"
)

const u63Mask = 0x8000000000000000

type Channel struct {
	seq       uint64
	retTimeNS int64
	retStream unsafe.Pointer
}

func (p *Channel) In(id uint64, gap uint64) (*core.Stream, *base.Error) {
	if atomic.CompareAndSwapUint64(&p.seq, id, id|u63Mask) {
		if stream := (*core.Stream)(p.retStream); stream != nil {
			stream.Release()
			p.retStream = nil
		}
		p.retTimeNS = int64(id + gap)
		return nil, nil
	} else if atomic.CompareAndSwapUint64(&p.seq, id|u63Mask, id|u63Mask) {
		return (*core.Stream)(atomic.LoadPointer(&p.retStream)), nil
	} else {
		return nil, errors.ErrGatewaySequenceError
	}
}

func (p *Channel) Out(stream *core.Stream) *base.Error {
	id := stream.GetCallbackID()
	if atomic.CompareAndSwapUint64(&p.seq, id|u63Mask, uint64(p.retTimeNS)) {
		p.retTimeNS = base.TimeNow().UnixNano()
		atomic.StorePointer(&p.retStream, unsafe.Pointer(stream))
		return nil
	} else if id == 0 {
		return nil
	} else {
		return errors.ErrGatewaySequenceError
	}
}
