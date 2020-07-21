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
		ReportFatal(
			NewReplyFatal(ErrStringRunOutOfScope).AddDebug(AddFileLine("", 2)),
		)
		return nil
	} else if node := thread.execReplyNode; node == nil {
		ReportFatal(
			NewReplyFatal(ErrStringRunOutOfScope).AddDebug(AddFileLine("", 2)),
		)
		return nil
	} else if meta := node.replyMeta; meta == nil {
		ReportFatal(
			NewKernelError(ErrStringUnexpectedNil).AddDebug(AddFileLine("", 1)),
		)
		return nil
	} else if !thread.IsDebug() {
		return thread
	} else {
		codeSource := AddFileLine("", 2)
		switch meta.GetCheck(codeSource) {
		case rpcReplyCheckStatusOK:
			return thread
		case rpcReplyCheckStatusError:
			ReportFatal(
				NewReplyFatal(ErrStringRunOutOfScope).AddDebug(codeSource),
			)
			return nil
		default:
			if thread.GetGoId() != CurrentGoroutineID() {
				meta.SetCheckError(codeSource)
				ReportFatal(
					NewReplyFatal(ErrStringRunOutOfScope).AddDebug(codeSource),
				)
				return nil
			} else {
				meta.SetCheckOK(codeSource)
				return thread
			}
		}
	}
}

func (p *ContextObject) stop() {
	if p == nil {
		ReportFatal(NewKernelError(ErrStringUnexpectedNil))
	} else {
		atomic.StorePointer(&p.thread, nil)
	}
}

func (p *ContextObject) OK(value interface{}) Return {
	if p == nil {
		ReportFatal(
			NewReplyFatal(ErrStringUnexpectedNil).AddDebug(AddFileLine("", 1)),
		)
		return nilReturn
	}

	if thread := p.getThread(); thread != nil {
		return thread.WriteOK(value, 2)
	}

	// do nothing, FatalReport has already run
	return nilReturn
}

func (p *ContextObject) Error(value error) Return {
	if p == nil {
		ReportFatal(
			NewReplyError(ErrStringUnexpectedNil).AddDebug(AddFileLine("", 1)),
		)
		return nilReturn
	}

	thread := p.getThread()
	if err, ok := value.(Error); ok && err != nil {
		return thread.WriteError(
			err.AddDebug(thread.GetExecReplyNodeDebug()),
		)
	} else if value != nil {
		return thread.WriteError(
			NewReplyError(value.Error()).
				AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
	} else {
		return thread.WriteError(
			NewReplyError(ErrStringUnexpectedNil).
				AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
	}
}
