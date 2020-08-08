package internal

// Context ...
type Context struct {
	id     uint64
	thread *rpcThread
}

var emptyContext = Context{}

//func (p Context) getThread() *rpcThread {
//  return p.th
//  if p.thread.lockByContext(p.id) {
//
//  }
//
//
//  if thread := (*rpcThread)(atomic.LoadPointer(&p.thread)); thread == nil {
//    reportPanic(
//      NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(GetFileLine(2)),
//    )
//    return nil
//  } else if node := thread.execReplyNode; node == nil {
//    thread.processor.Panic(
//      NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(GetFileLine(2)),
//    )
//    return nil
//  } else if !thread.processor.isDebug {
//    return thread
//  } else if thread.GetGoroutineID() != CurrentGoroutineID() {
//    thread.processor.Panic(
//      NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(GetFileLine(2)),
//    )
//    return nil
//  } else {
//    return thread
//  }
//}

//func (p *ContextObject) check() bool {
//  if thread := (*rpcThread)(atomic.LoadPointer(&p.thread)); thread == nil {
//    reportPanic(
//      NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(GetFileLine(2)),
//    )
//    return false
//  } else if !thread.processor.isDebug {
//    return true
//  }
//  return p != nil && atomic.LoadPointer(&p.thread) != nil
//}

//func (p *ContextObject) stop() bool {
//  if p == nil {
//    reportPanic(
//      NewKernelPanic("rpc: object is nil").AddDebug(string(debug.Stack())),
//    )
//    return false
//  }
//
//  atomic.StorePointer(&p.thread, nil)
//  return true
//}

// OK ...
func (p Context) OK(value interface{}) Return {
	if p.thread.lockByContext(p.id) {
		defer p.thread.unlockByContext(p.id)
		return p.thread.WriteOK(value, 2)
	}

	reportPanic(
		NewReplyPanic("race on Context").AddDebug(GetFileLine(1)),
	)
	return nilReturn
}

// Error ...
func (p Context) Error(value error) Return {
	if p.thread.lockByContext(p.id) {
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

	reportPanic(
		NewReplyPanic("race on Context").AddDebug(GetFileLine(1)),
	)
	return nilReturn
}
