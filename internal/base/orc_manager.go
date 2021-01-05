package base

import (
	"sync"
	"sync/atomic"
)

const (
	orcBitLock = 1 << 8

	orcStatusClosed  = int32(0)
	orcStatusClosing = int32(1)
	orcStatusReady   = int32(2)

	orcCondNone = int32(0)
	orcCondFree = int32(1)
	orcCondBusy = int32(2)
)

// IORCService ...
type IORCService interface {
	Open() bool
	Run() bool
	Close() bool
}

func execORCOpen(fn func() bool) (ret bool) {
	defer func() {
		if v := recover(); v != nil {
			ret = false
		}
	}()

	if fn != nil {
		return fn()
	}

	return false
}

func execORCRun(fn func(isRunning func() bool), isRunning func() bool) {
	defer func() {
		_ = recover()
	}()

	if fn != nil {
		fn(isRunning)
	}
}

func execORCClose(fn func()) {
	defer func() {
		_ = recover()
	}()

	if fn != nil {
		fn()
	}
}

// ORCManager ...
type ORCManager struct {
	status     int32
	condStatus int32
	mu         sync.Mutex
	cond       sync.Cond
}

// NewORCManager ...
func NewORCManager() *ORCManager {
	return &ORCManager{}
}

func (p *ORCManager) isRunning() bool {
	switch atomic.LoadInt32(&p.status) & 0xFF {
	case orcStatusReady:
		return true
	case orcStatusClosing:
		return func() bool {
			p.mu.Lock()
			defer p.mu.Unlock()

			for p.status&0xFF == orcStatusClosing {
				p.waitStatusChange()
			}

			return p.status&0xFF == orcStatusReady
		}()
	default:
		return false
	}
}

func (p *ORCManager) setStatus(status int32) {
	if status != p.status {
		atomic.StoreInt32(&p.status, status)

		if p.condStatus == orcCondBusy {
			p.condStatus = orcCondFree
			p.cond.Broadcast()
		}
	}
}

func (p *ORCManager) waitStatusChange() {
	if p.condStatus == orcCondNone {
		p.cond.L = &p.mu
	}

	p.condStatus = orcCondBusy
	p.cond.Wait()
}

// Open ...
func (p *ORCManager) Open(fn func() bool) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		switch p.status {
		case orcStatusClosed:
			if execORCOpen(fn) {
				p.setStatus(orcStatusReady)
				return true
			}

			return false
		case orcStatusClosing:
			p.waitStatusChange()
		case orcBitLock | orcStatusClosing:
			p.waitStatusChange()
		default:
			return false
		}
	}
}

// Run ...
func (p *ORCManager) Run(fn func(isRunning func() bool)) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		switch p.status {
		case orcStatusReady:
			p.setStatus(orcBitLock | orcStatusReady)
			p.mu.Unlock()
			execORCRun(fn, p.isRunning)
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

// Close ...
func (p *ORCManager) Close(willClose func(), didClose func()) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		switch p.status {
		case orcStatusReady:
			p.setStatus(orcStatusClosing)
			execORCClose(willClose)
			execORCClose(didClose)
			p.setStatus(orcStatusClosed)
			return true
		case orcBitLock | orcStatusReady:
			p.setStatus(orcBitLock | orcStatusClosing)
			execORCClose(willClose)
			for p.status&orcBitLock != 0 {
				p.waitStatusChange()
			}
			execORCClose(didClose)
			p.setStatus(orcStatusClosed)
			return true
		default:
			return false
		}
	}
}
