package base

import (
	"sync"
	"sync/atomic"
)

const (
	orcBitLock      = 1 << 8
	orcStatusClosed = int32(0)
	orcStatusReady  = int32(1)

	orcCondNone = int32(0)
	orcCondFree = int32(1)
	orcCondBusy = int32(2)
)

type StatusORC struct {
	status     int32
	condStatus int32
	mu         sync.Mutex
	cond       sync.Cond
}

func NewStatusORC() *StatusORC {
	return &StatusORC{}
}

func (p *StatusORC) isRunning() bool {
	return atomic.LoadInt32(&p.status)&0xFF == orcStatusReady
}

func (p *StatusORC) setStatus(status int32) {
	if status != p.status {
		atomic.StoreInt32(&p.status, status)

		if p.condStatus == orcCondBusy {
			p.condStatus = orcCondFree
			p.cond.Broadcast()
		}
	}
}

func (p *StatusORC) waitStatusChange() {
	if p.condStatus == orcCondNone {
		p.cond.L = &p.mu
	}

	p.condStatus = orcCondBusy
	p.cond.Wait()
}

func (p *StatusORC) Open(fn func() bool) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		switch atomic.LoadInt32(&p.status) {
		case orcStatusClosed:
			if fn() {
				p.setStatus(orcStatusReady)
				return true
			}

			return false
		case orcBitLock | orcStatusClosed:
			p.waitStatusChange()
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
			p.setStatus(orcBitLock | orcStatusReady)
			p.mu.Unlock()
			fn(p.isRunning)
			p.mu.Lock()
			p.setStatus(p.status & 0xFF)
			return true
		case orcBitLock | orcStatusReady:
			p.waitStatusChange()
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
			p.setStatus(orcStatusClosed)
			fn()
			return true
		case orcBitLock | orcStatusReady:
			p.setStatus(orcBitLock | orcStatusClosed)
			fn()
			for p.status&orcBitLock != 0 {
				p.waitStatusChange()
			}
			return true
		default:
			return false
		}
	}
}
