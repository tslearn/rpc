package util

import (
	"sync"
)

// AutoLock ...
type AutoLock struct {
	sync.Mutex
}

// NewAutoLock ...
func NewAutoLock() *AutoLock {
	return &AutoLock{}
}

// DoWithLock ...
func (p *AutoLock) DoWithLock(fn func()) {
	p.Lock()
	defer p.Unlock()
	fn()
}

// CallWithLock ...
func (p *AutoLock) CallWithLock(fn func() interface{}) interface{} {
	p.Lock()
	defer p.Unlock()
	return fn()
}
