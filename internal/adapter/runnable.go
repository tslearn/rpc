package adapter

import (
	"sync"
	"sync/atomic"
	"time"
)

// IRunnable ...
type IRunnable interface {
	OnOpen(*RunnableService)
	OnRun(*RunnableService)
	OnStop(*RunnableService)

	Close()
}

const serviceLoading = int32(1)
const serviceRunning = int32(2)
const serviceClosing = int32(3)
const serviceClosed = int32(0)

// RunnableService ...
type RunnableService struct {
	status   int32
	runnable IRunnable
	sync.Mutex
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
	if atomic.CompareAndSwapInt32(&p.status, serviceClosed, serviceLoading) {
		func() {
			p.Lock()
			defer p.Unlock()
			p.runnable.OnOpen(p)
		}()

		atomic.StoreInt32(&p.status, serviceRunning)
		p.runnable.OnRun(p)
		atomic.StoreInt32(&p.status, serviceClosing)

		func() {
			p.Lock()
			defer p.Unlock()
			p.runnable.OnStop(p)
		}()

		atomic.StoreInt32(&p.status, serviceClosed)
		return true
	}

	return false
}

// Close ...
func (p *RunnableService) Close() {
	for {
		if atomic.CompareAndSwapInt32(&p.status,
			serviceRunning,
			serviceClosing,
		) {
			func() {
				p.Lock()
				defer p.Unlock()
				p.runnable.Close()
			}()
		} else {
			switch atomic.LoadInt32(&p.status) {
			case serviceClosed:
				return
			case serviceRunning:
				continue
			default:
				time.Sleep(50 * time.Millisecond)
			}
		}
	}
}

// IsLoading ...
func (p *RunnableService) IsLoading() bool {
	return atomic.LoadInt32(&p.status) == serviceLoading
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
