package base

import "sync"

type Mutex struct {
	synchronizeMutex sync.Mutex
}

func (p *Mutex) Synchronized(fn func()) {
	p.synchronizeMutex.Lock()
	defer p.synchronizeMutex.Unlock()
	fn()
}
