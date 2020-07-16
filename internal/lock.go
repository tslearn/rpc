package internal

import (
	"sync"
)

// Lock ...
type Lock struct {
	sync.Mutex
}

// NewLock ...
func NewLock() *Lock {
	return &Lock{}
}

// DoWithLock ...
func (p *Lock) DoWithLock(fn func()) {
	p.Lock()
	defer p.Unlock()
	fn()
}

// CallWithLock ...
func (p *Lock) CallWithLock(fn func() interface{}) interface{} {
	p.Lock()
	defer p.Unlock()
	return fn()
}
