package gateway

import (
	"github.com/rpccloud/rpc/internal/base"
	"sync/atomic"
)

type SessionIDGenerator interface {
	GetID() (uint64, *base.Error)
}

type SingleGenerator struct {
	id uint64
}

func NewSingleGenerator() *SingleGenerator {
	return &SingleGenerator{
		id: 10000,
	}
}

func (p *SingleGenerator) GetID() (uint64, *base.Error) {
	return atomic.AddUint64(&p.id, 1), nil
}
