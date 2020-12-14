package adapter

import (
	"sync/atomic"
	"time"
)

// IRunnable ...
type IRunnable interface {
	OnRun(*RunnableService)
	OnStop(*RunnableService)

	Close()
}

const serviceRunning = int32(1)
const serviceClosing = int32(2)
const serviceClosed = int32(0)

// RunnableService ...
type RunnableService struct {
	status   int32
	runnable IRunnable
}

// NewRunnableService ...
func NewRunnableService(runnable IRunnable) *RunnableService {
	return &RunnableService{
		status:   serviceClosed,
		runnable: runnable,
	}
}

// Open ...
func (p *RunnableService) Open() bool {
	if atomic.CompareAndSwapInt32(&p.status, serviceClosed, serviceRunning) {
		p.runnable.OnRun(p)
		p.runnable.OnStop(p)
		atomic.StoreInt32(&p.status, serviceClosed)
		return true
	}

	return false
}

// Close ...
func (p *RunnableService) Close() bool {
	if atomic.CompareAndSwapInt32(&p.status, serviceRunning, serviceClosing) {
		p.runnable.Close()

		for p.IsClosing() {
			time.Sleep(50 * time.Millisecond)
		}

		return true
	}

	return false
}

// IsRunning ...
func (p *RunnableService) IsRunning() bool {
	return atomic.LoadInt32(&p.status) == serviceRunning
}

// IsClosing ...
func (p *RunnableService) IsClosing() bool {
	return atomic.LoadInt32(&p.status) == serviceClosing
}

// IsClosed ...
func (p *RunnableService) IsClosed() bool {
	return atomic.LoadInt32(&p.status) == serviceClosed
}
