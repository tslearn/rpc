package internal

// Context ...
type Context struct {
	id     uint64
	thread *rpcThread
}

var emptyContext = Context{}

// OK ...
func (p Context) OK(value interface{}) Return {
	if thread := p.thread; thread == nil {
		reportPanic(
			NewReplyPanic("bad Context").AddDebug(GetFileLine(1)),
		)
		return nilReturn
	} else if !thread.lockByContext(p.id, 2) {
		reportPanic(
			NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(GetFileLine(1)),
		)
		return nilReturn
	} else {
		defer thread.unlockByContext(p.id)
		return p.thread.WriteOK(value, 2)
	}
}

// Error ...
func (p Context) Error(value error) Return {
	if thread := p.thread; thread == nil {
		reportPanic(
			NewReplyPanic("bad Context").AddDebug(GetFileLine(1)),
		)
		return nilReturn
	} else if !thread.lockByContext(p.id, 2) {
		reportPanic(
			NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(GetFileLine(1)),
		)
		return nilReturn
	} else {
		defer p.thread.unlockByContext(p.id)

		if err, ok := value.(Error); ok && err != nil {
			return p.thread.WriteError(
				err.AddDebug(AddFileLine(p.thread.GetExecReplyNodePath(), 1)),
			)
		} else if value != nil {
			return p.thread.WriteError(
				NewReplyError(value.Error()).
					AddDebug(AddFileLine(p.thread.GetExecReplyNodePath(), 1)),
			)
		} else {
			return p.thread.WriteError(
				NewReplyError("rpc: Context.Error() argument should not nil").
					AddDebug(AddFileLine(p.thread.GetExecReplyNodePath(), 1)),
			)
		}
	}
}
