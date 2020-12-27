package server

import (
	"time"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/router"
)

// RPCProcessor ...
type RPCProcessor struct {
	processor *core.Processor
}

// NewRPCProcessor ...
func NewRPCProcessor(
	router router.IRouter,
	numOfThreads int,
	maxNodeDepth int16,
	maxCallDepth int16,
	threadBufferSize uint32,
	fnCache core.ActionCache,
	closeTimeout time.Duration,
	mountServices []*core.ServiceMeta,
) (*RPCProcessor, *base.Error) {
	ret := &RPCProcessor{}
	routeSender := router.Plug(ret)
	processor, err := core.NewProcessor(
		numOfThreads,
		maxNodeDepth,
		maxCallDepth,
		threadBufferSize,
		fnCache,
		closeTimeout,
		mountServices,
		func(stream *core.Stream) {
			stream.SetDirectionOut()
			_ = routeSender.SendStreamToRouter(stream)
		},
	)

	if err != nil {
		return nil, err
	}

	ret.processor = processor
	return ret, nil
}

// ReceiveStreamFromRouter ...
func (p *RPCProcessor) ReceiveStreamFromRouter(stream *core.Stream) *base.Error {
	p.processor.PutStream(stream)
	return nil
}

// Close ...
func (p *RPCProcessor) Close() {
	p.processor.Close()
}
