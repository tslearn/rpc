package internal

import (
	"sync"
)

// RPCLock ...
type RPCLock struct {
	sync.Mutex
}

// NewRPCLock ...
func NewRPCLock() *RPCLock {
	return &RPCLock{}
}

// DoWithLock ...
func (p *RPCLock) DoWithLock(fn func()) {
	p.Lock()
	defer p.Unlock()
	fn()
}

// CallWithLock ...
func (p *RPCLock) CallWithLock(fn func() interface{}) interface{} {
	p.Lock()
	defer p.Unlock()
	return fn()
}
