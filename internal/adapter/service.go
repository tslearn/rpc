package adapter

import (
	"sync/atomic"
	"time"
)

type IRunnable interface {
	OnOpen() bool
	OnRun(service *RunnableService) bool
	OnClose()
}

const serviceLoading = int32(1)
const serviceRunning = int32(2)
const serviceClosing = int32(3)
const serviceClosed = int32(0)

type RunnableService struct {
	status   int32
	runnable IRunnable
}

func NewRunnableService(runnable IRunnable) *RunnableService {
	return &RunnableService{
		status:   serviceClosed,
		runnable: runnable,
	}
}

// Open ...
func (p *RunnableService) Open() bool {
	if atomic.CompareAndSwapInt32(&p.status, serviceClosed, serviceLoading) {
		if p.runnable.OnOpen() {
			atomic.StoreInt32(&p.status, serviceRunning)
			p.runnable.OnRun(p)
			atomic.StoreInt32(&p.status, serviceClosed)
		} else {
			atomic.StoreInt32(&p.status, serviceClosed)
		}

		return true
	}

	return false
}

// Close ...
func (p *RunnableService) Close() bool {
	if atomic.CompareAndSwapInt32(&p.status, serviceRunning, serviceClosing) {
		p.runnable.OnClose()

		for atomic.LoadInt32(&p.status) != serviceClosed {
			time.Sleep(50 * time.Millisecond)
		}

		return true
	}

	return false
}

func (p *RunnableService) IsRunning() bool {
	return atomic.LoadInt32(&p.status) == serviceRunning
}
