package base

import (
	"sync"
	"sync/atomic"
)

const (
	orcLockBit    = 1 << 2
	orcStatusMask = orcLockBit - 1

	orcStatusClosed  = 0
	orcStatusReady   = 1
	orcStatusClosing = 2
)

// IORCService ...
type IORCService interface {
	Open() bool
	Run() bool
	Close() bool
}

// ORCManager ...
type ORCManager struct {
	sequence     uint64
	isWaitChange bool
	mu           sync.Mutex
	cond         sync.Cond
}

// NewORCManager ...
func NewORCManager() *ORCManager {
	return &ORCManager{}
}

func (p *ORCManager) getBaseSequence() uint64 {
	return p.sequence - p.sequence%8
}

func (p *ORCManager) getStatus() uint64 {
	return p.sequence % 8
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

func (p *ORCManager) setStatus(status uint64) {
	if curStatus := p.getStatus(); curStatus != status {
		baseSequence := p.getBaseSequence()

		if status == orcStatusClosed {
			atomic.StoreUint64(&p.sequence, baseSequence+status+8)
		} else {
			atomic.StoreUint64(&p.sequence, baseSequence+status)
		}

		if p.isWaitChange {
			p.isWaitChange = false
			p.cond.Broadcast()
		}
	}
}

func (p *ORCManager) waitStatusChange() {
	if p.cond.L == nil {
		p.cond.L = &p.mu
	}

	p.isWaitChange = true
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
				if didOpen != nil {
					didOpen()
				}
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

				if didRun != nil {
					func() {
						// open the lock and then call didRun.
						// at last lock again
						p.mu.Unlock()
						defer p.mu.Lock()
						didRun(isRunningFn)
					}()
				}

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
				if didClose != nil {
					didClose()
				}
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
				if didClose != nil {
					didClose()
				}
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
