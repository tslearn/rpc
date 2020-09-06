package internal

// Runtime ...
type Runtime struct {
	id     uint64
	thread *rpcThread
}

func (p Runtime) lock() *rpcThread {
	if thread := p.thread; thread != nil {
		return thread.lock(p.id)
	}

	return nil
}

func (p Runtime) unlock() {
	if thread := p.thread; thread != nil {
		thread.unlock(p.id)
	}
}

// OK ...
func (p Runtime) OK(value interface{}) Return {
	if thread := p.lock(); thread != nil {
		defer p.unlock()
		return thread.WriteOK(value, 1)
	}

	reportPanic(
		NewReplyPanic(
			"Runtime is illegal in current goroutine",
		).AddDebug(GetFileLine(1)),
	)
	return emptyReturn
}

// Error ...
func (p Runtime) Error(value error) Return {
	if thread := p.lock(); thread != nil {
		defer p.unlock()

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

	reportPanic(
		NewReplyPanic(
			"Runtime is illegal in current goroutine",
		).AddDebug(GetFileLine(1)),
	)
	return emptyReturn
}

func (p Runtime) Call(target string, args ...interface{}) (interface{}, Error) {
	if thread := p.lock(); thread != nil {
		defer p.unlock()

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

	return nil, NewReplyPanic(
		"Runtime is illegal in current goroutine",
	)
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
