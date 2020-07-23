package internal

import (
	"sync/atomic"
	"unsafe"
)

type Context = *ContextObject

type ContextObject struct {
	thread unsafe.Pointer
}

var nilContext = (Context)(nil)

func (p *ContextObject) getThread() *rpcThread {
	if thread := (*rpcThread)(atomic.LoadPointer(&p.thread)); thread == nil {
		ReportPanic(
			NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(GetFileLine(2)),
		)
		return nil
	} else if node := thread.execReplyNode; node == nil {
		ReportPanic(
			NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(GetFileLine(2)),
		)
		return nil
	} else if !thread.IsDebug() {
		return thread
	} else if thread.GetGoId() != CurrentGoroutineID() {
		ReportPanic(
			NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(GetFileLine(2)),
		)
		return nil
	} else {
		return thread
	}
}

func (p *ContextObject) stop() bool {
	if p == nil {
		ReportPanic(
			NewKernelError(ErrStringUnexpectedNil).AddDebug(GetFileLine(1)),
		)
		return false
	} else {
		atomic.StorePointer(&p.thread, nil)
		return true
	}
}

func (p *ContextObject) OK(value interface{}) Return {
	if p == nil {
		ReportPanic(
			NewReplyPanic(ErrStringUnexpectedNil).AddDebug(GetFileLine(1)),
		)
		return nilReturn
	} else if thread := p.getThread(); thread == nil {
		// ReportPanic has already run
		return nilReturn
	} else {
		return thread.WriteOK(value, 2)
	}
}

func (p *ContextObject) Error(value error) Return {
	if p == nil {
		ReportPanic(
			NewReplyPanic(ErrStringUnexpectedNil).AddDebug(GetFileLine(1)),
		)
		return nilReturn
	} else if thread := p.getThread(); thread == nil {
		// ReportPanic has already run
		return nilReturn
	} else if err, ok := value.(Error); ok && err != nil {
		return thread.WriteError(
			err.AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
	} else if value != nil {
		return thread.WriteError(
			NewReplyError(value.Error()).
				AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
	} else {
		ReportPanic(
			NewReplyPanic(ErrStringUnexpectedNil).AddDebug(GetFileLine(1)),
		)
		return nilReturn
	}
}
