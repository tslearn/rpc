package base

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	orcBitLock      = 1 << 8
	orcStatusClosed = int32(0)
	orcStatusReady  = int32(1)
)

type StatusORC struct {
	status int32
	mu     sync.Mutex
}

func NewStatusORC() *StatusORC {
	return &StatusORC{status: orcStatusClosed}
}

func (p *StatusORC) isRunning() bool {
	return atomic.LoadInt32(&p.status)&0xFF == orcStatusReady
}

func (p *StatusORC) Open(fn func() bool) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		switch atomic.LoadInt32(&p.status) {
		case orcStatusClosed:
			if fn() {
				atomic.StoreInt32(&p.status, orcStatusReady)
				return true
			}

			atomic.StoreInt32(&p.status, orcStatusClosed)
			return false
		case orcBitLock | orcStatusClosed:
			p.mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			p.mu.Lock()
		default:
			return false
		}
	}
}

func (p *StatusORC) Run(fn func(isRunning func() bool)) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		switch atomic.LoadInt32(&p.status) {
		case orcStatusReady:
			atomic.StoreInt32(&p.status, orcBitLock|orcStatusReady)
			p.mu.Unlock()
			fn(p.isRunning)
			p.mu.Lock()
			atomic.StoreInt32(&p.status, p.status&0xFF)
			return true
		case orcBitLock | orcStatusReady:
			p.mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			p.mu.Lock()
		default:
			return false
		}
	}
}

func (p *StatusORC) Close(fn func()) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		switch atomic.LoadInt32(&p.status) {
		case orcStatusReady:
			atomic.StoreInt32(&p.status, orcStatusClosed)
			fn()
			return true
		case orcBitLock | orcStatusReady:
			atomic.StoreInt32(&p.status, orcBitLock|orcStatusClosed)
			fn()
			for p.status&orcBitLock != 0 {
				p.mu.Unlock()
				time.Sleep(10 * time.Millisecond)
				p.mu.Lock()
			}
			return true
		default:
			return false
		}
	}
}
