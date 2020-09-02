package internal

// Runtime ...
type Runtime struct {
	id     uint64
	thread *rpcThread
}

// OK ...
func (p Runtime) OK(value interface{}) Return {
	ctxID := p.id
	if thread := p.thread; thread == nil {
		reportPanic(
			NewReplyPanic(
				"Runtime is illegal in current goroutine",
			).AddDebug(GetFileLine(1)),
		)
		return emptyReturn
	} else if thread.lock(ctxID) == nil {
		return thread.WriteError(
			NewReplyPanic(
				"Runtime is illegal in current goroutine",
			).AddDebug(GetFileLine(1)),
			1,
		)
	} else {
		defer thread.unlock(ctxID)
		return thread.WriteOK(value, 1)
	}
}

// Error ...
func (p Runtime) Error(value error) Return {
	ctxID := p.id
	if thread := p.thread; thread == nil {
		reportPanic(
			NewReplyPanic(
				"Runtime is illegal in current goroutine",
			).AddDebug(GetFileLine(1)),
		)
		return emptyReturn
	} else if thread.lock(ctxID) == nil {
		return thread.WriteError(
			NewReplyPanic(
				"Runtime is illegal in current goroutine",
			).AddDebug(GetFileLine(1)),
			1,
		)
	} else {
		defer thread.unlock(ctxID)

		if err, ok := value.(Error); ok && err != nil {
			return p.thread.WriteError(
				err.AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
				1,
			)
		} else if value != nil {
			return p.thread.WriteError(
				NewReplyError(
					value.Error(),
				).AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
				1,
			)
		} else {
			return p.thread.WriteError(
				NewReplyError(
					"argument should not nil",
				).AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
				1,
			)
		}
	}
}

func (p Runtime) Call(target string, args ...interface{}) (interface{}, Error) {
	ctxID := p.id
	if thread := p.thread; thread == nil || thread.lock(ctxID) == nil {
		return nil, NewReplyPanic(
			"Runtime is illegal in current goroutine",
		)
	} else {
		defer thread.unlock(ctxID)
		stream := NewStream()
		defer stream.Release()

		frame := thread.top

		stream.SetDepth(frame.depth + 1)
		// write target
		stream.WriteString(target)
		// write from
		stream.WriteString(frame.from)
		// write args
		for i := 0; i < len(args); i++ {
			if stream.Write(args[i]) != StreamWriteOK {
				return nil, NewReplyPanic(ConcatString(
					ConvertOrdinalToString(uint(i+1)),
					" argument not supported",
				))
			}
		}

		thread.pushFrame()
		defer thread.popFrame()

		thread.Eval(
			stream,
			func(stream *Stream) {},
			func(thread *rpcThread) {},
		)

		// parse the stream
		if errKind, ok := stream.ReadUint64(); !ok {
			return nil, NewProtocolError(ErrStringBadStream)
		} else if errKind == uint64(ErrorKindNone) {
			if ret, ok := stream.Read(); ok {
				return ret, nil
			}
			return nil, NewProtocolError(ErrStringBadStream)
		} else if message, ok := stream.ReadString(); !ok {
			return nil, NewProtocolError(ErrStringBadStream)
		} else if dbg, ok := stream.ReadString(); !ok {
			return nil, NewProtocolError(ErrStringBadStream)
		} else if !stream.IsReadFinish() {
			return nil, NewProtocolError(ErrStringBadStream)
		} else {
			return nil, NewError(ErrorKind(errKind), message, dbg)
		}
	}
}

func (p Runtime) GetServiceData() interface{} {
	if thread := p.thread; thread == nil {
		reportPanic(
			NewReplyPanic(
				"Runtime is illegal in current goroutine",
			).AddDebug(GetFileLine(1)),
		)
		return nil
	} else if node := thread.GetReplyNode(); node == nil {
		reportPanic(
			NewReplyPanic(
				"Runtime is illegal in current goroutine",
			).AddDebug(GetFileLine(1)),
		)
		return nil
	} else {
		return node.service.addMeta.data
	}
}
