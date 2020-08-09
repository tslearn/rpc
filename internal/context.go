package internal

// Context ...
type Context struct {
	id     uint64
	thread *rpcThread
}

// OK ...
func (p Context) OK(value interface{}) Return {
	if thread := p.thread; thread == nil {
		reportPanic(
			NewReplyPanic(
				"Context is illegal in current goroutine",
			).AddDebug(GetFileLine(1)),
		)
		return nilReturn
	} else if code := thread.setReturn(p.id); code == rpcThreadReturnStatusOK {
		stream := thread.execStream
		stream.SetWritePosToBodyStart()
		stream.WriteUint64(uint64(ErrorKindNone))
		if stream.Write(value) != StreamWriteOK {
			thread.processor.Panic(
				NewReplyPanic(checkValue(value, "value", 64)).
					AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
			)
			return thread.WriteError(
				NewReplyError("reply return value error").
					AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
			)
		}
		return nilReturn
	} else if code == rpcThreadReturnStatusAlreadyCalled {
		reportPanic(
			NewReplyPanic(
				"Context.OK or Context.Error has been called before",
			).AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
		return nilReturn
	} else {
		reportPanic(
			NewReplyPanic(
				"Context is illegal in current goroutine",
			).AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
		return nilReturn
	}
}

// Error ...
func (p Context) Error(value error) Return {
	if thread := p.thread; thread == nil {
		reportPanic(
			NewReplyPanic(
				"Context is illegal in current goroutine",
			).AddDebug(GetFileLine(1)),
		)
		return nilReturn
	} else if code := thread.setReturn(p.id); code == rpcThreadReturnStatusOK {
		if err, ok := value.(Error); ok && err != nil {
			return p.thread.WriteError(
				err.AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
			)
		} else if value != nil {
			return p.thread.WriteError(
				NewReplyError(
					value.Error(),
				).AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
			)
		} else {
			return p.thread.WriteError(
				NewReplyError(
					"argument should not nil",
				).AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
			)
		}
	} else if code == rpcThreadReturnStatusAlreadyCalled {
		reportPanic(
			NewReplyPanic(
				"Context.OK or Context.Error has been called before",
			).AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
		return nilReturn
	} else {
		reportPanic(
			NewReplyPanic(
				"Context is illegal in current goroutine",
			).AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
		return nilReturn
	}
}
