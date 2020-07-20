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

// Return ...
func (p *ContextObject) Return(value interface{}) Return {
	if thread := p.getThread(); thread == nil {
		panic("rpc: running out of reply goroutine")
	} else {
		if value == nil {
			return thread.WriteOK(value, 2)
		} else if rpcErr, ok := value.(Error); ok {
			if thread.execReplyNode != nil && thread.execReplyNode.debugString != "" {
				_ = rpcErr.AddDebug(thread.execReplyNode.debugString)
			}
			return thread.WriteError(rpcErr)
		} else if sysErr, ok := value.(error); ok {
			err := NewServiceError(sysErr.Error())
			if thread.execReplyNode != nil && thread.execReplyNode.debugString != "" {
				_ = err.AddDebug(thread.execReplyNode.debugString)
			}
			return thread.WriteError(err)
		} else {
			return thread.WriteOK(value, 2)
		}
	}
}

//
//// Error ...
//func (p *ContextObject) Error(err Error) Return {
//  if thread := p.getThread(); thread == nil {
//    panic(NewError(ErrStringContextErrorOutsideScope))
//  } else if err == nil {
//    return thread.WriteOK(nil, GetStackString(1))
//  } else {
//    if thread.execReplyNode != nil && thread.execReplyNode.debugString != "" {
//      _ = err.AddDebug(thread.execReplyNode.debugString)
//    }
//    thread.WriteError(err.GetMessage(), err.GetDebug())
//  }
//}
