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
	if thread := (*rpcThread)(atomic.LoadPointer(&p.thread)); thread == nil {
		panic(ErrStringRunningOutOfScope)
	} else if node := thread.execReplyNode; node == nil {
		panic(ErrStringRunningOutOfScope)
	} else if meta := node.replyMeta; meta == nil {
		panic(ErrInternalErrorReplyMetaIsNil)
	} else if !thread.IsDebug() {
		return thread
	} else {
		codeSource := GetCodePosition("", 2)
		switch meta.GetCheck(codeSource) {
		case rpcReplyCheckStatusOK:
			return thread
		case rpcReplyCheckStatusError:
			panic(ErrStringRunningOutOfScope)
		default:
			if thread.GetGoId() != CurrentGoroutineID() {
				meta.SetCheckError(codeSource)
				panic(ErrStringRunningOutOfScope)
			} else {
				meta.SetCheckOK(codeSource)
				return thread
			}
		}
	}
}

func (p *ContextObject) stop() {
	atomic.StorePointer(&p.thread, nil)
}

func (p *ContextObject) OK(value interface{}) Return {
	return p.getThread().WriteOK(value, 2)
}

func (p *ContextObject) Error(value error) Return {
	thread := p.getThread()
	if err, ok := value.(Error); ok && err != nil {
		return thread.WriteError(
			err.AddDebug(thread.GetExecReplyNodeDebug()),
		)
	} else if value != nil {
		return thread.WriteError(
			NewError(value.Error()).
				AddDebug(GetCodePosition(thread.GetExecReplyNodePath(), 1)),
		)
	} else {
		return thread.WriteError(
			NewError("value is nil").
				AddDebug(GetCodePosition(thread.GetExecReplyNodePath(), 1)),
		)
	}
}
