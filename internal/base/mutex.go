package base

import "sync"

// Mutex ...
type Mutex struct {
	synchronizeMutex sync.Mutex
}

// Synchronized ...
func (p *Mutex) Synchronized(fn func()) {
	p.synchronizeMutex.Lock()
	defer p.synchronizeMutex.Unlock()
	fn()
}
