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

// OK ...
// this
func (p *ContextObject) OK(value interface{}) Return {
	if thread := p.getThread(); thread != nil {
		stream := thread.outStream
		stream.SetWritePos(streamBodyPos)
		stream.WriteBool(true)

		if stream.Write(value) != StreamWriteOK {
			thread.WriteError("return type is error", GetStackString(1))
		} else {
			thread.execSuccessful = true
		}
	}
	return nilReturn
}

func (p *ContextObject) Error(err Error) Return {
	if thread := p.getThread(); thread != nil {
		if err == nil {
			thread.WriteError("parameter is nil", GetStackString(1))
		} else {
			if thread.execReplyNode != nil && thread.execReplyNode.debugString != "" {
				_ = err.AddDebug(thread.execReplyNode.debugString)
			}
			thread.WriteError(err.GetMessage(), err.GetDebug())
		}
	}
	return nilReturn
}
