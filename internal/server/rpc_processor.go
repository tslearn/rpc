package server

import (
	"github.com/rpccloud/rpc/internal"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"time"
)

type RPCProcessor struct {
	processor *core.Processor
}

func NewRPCProcessor(
	router internal.IStreamRouter,
	numOfThreads int,
	maxNodeDepth int16,
	maxCallDepth int16,
	threadBufferSize uint32,
	fnCache core.ActionCache,
	closeTimeout time.Duration,
	mountServices []*core.ServiceMeta,
) (*RPCProcessor, *base.Error) {
	ret := &RPCProcessor{}

	if slot, err := router.Plug(ret); err != nil {
		return nil, err
	} else if processor, err := core.NewProcessor(
		numOfThreads,
		maxNodeDepth,
		maxCallDepth,
		threadBufferSize,
		fnCache,
		closeTimeout,
		mountServices,
		func(stream *core.Stream) {
			_ = slot.SendStream(stream)
		},
	); err != nil {
		return nil, err
	} else {
		ret.processor = processor
		return ret, nil
	}
}

func (p *RPCProcessor) OnStream(stream *core.Stream) *base.Error {
	p.processor.PutStream(stream)
	return nil
}

func (p *RPCProcessor) Close() {
	p.processor.Close()
}
