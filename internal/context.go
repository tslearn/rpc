package internal

import (
	"sync/atomic"
	"unsafe"
)

type Context = *ContextObject

type ContextObject struct {
	thread unsafe.Pointer
}

func (p *ContextObject) getThread() *rpcThread {
	return (*rpcThread)(atomic.LoadPointer(&p.thread))
}

func (p *ContextObject) stop() {
	atomic.StorePointer(&p.thread, nil)
}

func (p *ContextObject) OK(value interface{}) Return {
	if thread := p.getThread(); thread == nil {
		panic("rpc: running out of reply goroutine")
	} else {
		return thread.WriteOK(value, 2)
	}
}

func (p *ContextObject) Error(value error) Return {
	if thread := p.getThread(); thread == nil {
		panic("rpc: running out of reply goroutine")
	} else if err, ok := value.(Error); ok && err != nil {
		thread.WriteError(
			err.AddDebug(thread.GetExecReplyNodeDebug()),
		)
	} else if value != nil {
		thread.WriteError(
			NewError(value.Error()).
				AddDebug(GetCodePosition(thread.GetExecReplyNodePath(), 1)),
		)
	} else {
		thread.WriteError(
			NewError("value is nil").
				AddDebug(GetCodePosition(thread.GetExecReplyNodePath(), 1)),
		)
	}

	return nilReturn
}
