package util

import (
	"sync"
)

// AutoLock ...
type AutoLock = *rpcAutoLock

type rpcAutoLock struct {
	sync.Mutex
}

// NewAutoLock ...
func NewAutoLock() AutoLock {
	return &rpcAutoLock{}
}

// DoWithLock ...
func (p *rpcAutoLock) DoWithLock(fn func()) {
	p.Lock()
	defer p.Unlock()
	fn()
}

// CallWithLock ...
func (p *rpcAutoLock) CallWithLock(fn func() interface{}) interface{} {
	p.Lock()
	defer p.Unlock()
	return fn()
}
