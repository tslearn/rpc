package adapter

import (
	"sync"
	"sync/atomic"
)

const (
	orcLockBit    = 1 << 2
	orcStatusMask = 0x03

	orcStatusClosed  = 0
	orcStatusReady   = 1
	orcStatusClosing = 2

	orcCondNone = 0
	orcCondFree = 1
	orcCondBusy = 2
)

// ORCManager ...
type ORCManager struct {
	sequence   uint64
	condStatus int32
	mu         sync.Mutex
	cond       sync.Cond
}

// NewORCManager ...
func NewORCManager() *ORCManager {
	return &ORCManager{}
}

func (p *ORCManager) getRunningFn() func() bool {
	baseSequence := p.getBaseSequence()

	return func() bool {
		status := atomic.LoadUint64(&p.sequence) - baseSequence

		if status == orcStatusReady || status == orcStatusReady|orcLockBit {
			return true
		}

		return func() bool {
			p.mu.Lock()
			defer p.mu.Unlock()

			for {
				switch p.sequence - baseSequence {
				case orcStatusReady:
					return true
				case orcStatusReady | orcLockBit:
					return true
				case orcStatusClosing:
					p.waitStatusChange()
				case orcStatusClosing | orcLockBit:
					p.waitStatusChange()
				default:
					return false
				}
			}
		}()
	}
}

func (p *ORCManager) getBaseSequence() uint64 {
	return p.sequence - p.sequence%8
}

func (p *ORCManager) getStatus() uint64 {
	return p.sequence % 8
}

func (p *ORCManager) setStatus(status uint64) {
	if curStatus := p.getStatus(); curStatus != status {
		baseSequence := p.getBaseSequence()

		if status == orcStatusClosed {
			atomic.StoreUint64(&p.sequence, baseSequence+8)
		} else {
			atomic.StoreUint64(&p.sequence, status)
		}

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
func (p *ORCManager) Open(willOpen func() bool, didOpen func()) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		switch p.getStatus() {
		case orcStatusClosing:
			p.waitStatusChange()
		case orcStatusClosing | orcLockBit:
			p.waitStatusChange()
		case orcStatusClosed:
			if willOpen() {
				p.setStatus(orcStatusReady)
				didOpen()
				return true
			}
			return false
		default:
			return false
		}
	}
}

// Run ...
func (p *ORCManager) Run(
	willRun func() bool,
	didRun func(isRunning func() bool),
) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		switch p.getStatus() {
		case orcStatusReady:
			if willRun() {
				// add lock bit
				p.setStatus(orcStatusReady | orcLockBit)
				isRunningFn := p.getRunningFn()

				p.mu.Unlock()
				didRun(isRunningFn)
				p.mu.Lock()

				// delete lock bit
				p.setStatus(p.getStatus() & orcStatusMask)

				return true
			}

			return false
		case orcStatusReady | orcLockBit:
			p.waitStatusChange()
		case orcStatusClosing:
			p.waitStatusChange()
		case orcStatusClosing | orcLockBit:
			p.waitStatusChange()
		default:
			return false
		}
	}
}

// Close ...
func (p *ORCManager) Close(willClose func() bool, didClose func()) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		switch p.getStatus() {
		case orcStatusReady:
			p.setStatus(orcStatusClosing)
			if willClose() {
				didClose()
				p.setStatus(orcStatusClosed)
				return true
			}

			p.setStatus(orcStatusReady)
			return false
		case orcStatusReady | orcLockBit:
			p.setStatus(orcStatusClosing | orcLockBit)
			if willClose() {
				for p.getStatus()&orcLockBit != 0 {
					p.waitStatusChange()
				}
				didClose()
				p.setStatus(orcStatusClosed)
				return true
			}

			p.setStatus(orcStatusReady | orcLockBit)
			return false
		case orcStatusClosing:
			p.waitStatusChange()
		case orcStatusClosing | orcLockBit:
			p.waitStatusChange()
		default:
			return false
		}
	}
}
