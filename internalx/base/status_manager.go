package base

import (
	"sync"
	"sync/atomic"
)

const statusManagerClosed = int32(0)
const statusManagerRunning = int32(1)
const statusManagerClosing = int32(2)

// StatusManager ...
type StatusManager struct {
	status  int32
	closeCH chan bool
	mutex   sync.Mutex
}

// SetRunning ...
func (p *StatusManager) SetRunning(onSuccess func()) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if atomic.CompareAndSwapInt32(
		&p.status,
		statusManagerClosed,
		statusManagerRunning,
	) {
		if onSuccess != nil {
			onSuccess()
		}
		return true
	}

	return false
}

// SetClosing ...
func (p *StatusManager) SetClosing(onSuccess func(ch chan bool)) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if atomic.CompareAndSwapInt32(
		&p.status,
		statusManagerRunning,
		statusManagerClosing,
	) {
		p.closeCH = make(chan bool, 1)
		if onSuccess != nil {
			onSuccess(p.closeCH)
		}
		return true
	}

	return false
}

// SetClosed ...
func (p *StatusManager) SetClosed(onSuccess func()) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if atomic.CompareAndSwapInt32(
		&p.status,
		statusManagerClosing,
		statusManagerClosed,
	) {
		if onSuccess != nil {
			onSuccess()
		}
		p.closeCH <- true
		close(p.closeCH)
		p.closeCH = nil
		return true
	}

	return false
}

// IsRunning ...
func (p *StatusManager) IsRunning() bool {
	return atomic.CompareAndSwapInt32(
		&p.status,
		statusManagerRunning,
		statusManagerRunning,
	)
}
