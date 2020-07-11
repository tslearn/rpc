package util

import (
	"sync"
)

// AutoLock ...
type AutoLock struct {
	locker sync.Mutex
}

// CallWithLock ...
func (p *AutoLock) CallWithLock(fn func() interface{}) interface{} {
	p.locker.Lock()
	defer p.locker.Unlock()
	return fn()
}

// DoWithLock ...
func (p *AutoLock) DoWithLock(fn func()) {
	p.locker.Lock()
	defer p.locker.Unlock()
	fn()
}
