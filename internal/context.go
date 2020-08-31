package internal

// Runtime ...
type Runtime struct {
	id     uint64
	thread *rpcThread
}

// OK ...
func (p Runtime) OK(value interface{}) Return {
	if thread := p.thread; thread == nil {
		reportPanic(
			NewReplyPanic(
				"Runtime is illegal in current goroutine",
			).AddDebug(GetFileLine(1)),
		)
		return emptyReturn
	} else if code := thread.setReturn(p.id); code == rpcThreadReturnStatusOK {
		stream := thread.top.stream
		stream.SetWritePosToBodyStart()
		stream.WriteUint64(uint64(ErrorKindNone))
		if stream.Write(value) != StreamWriteOK {
			return thread.WriteError(
				NewReplyPanic(checkValue(value, "value", 64)).
					AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
			)
		}
		return emptyReturn
	} else if code == rpcThreadReturnStatusAlreadyCalled {
		thread.WriteError(
			NewReplyPanic(
				"Runtime.OK or Runtime.Error has been called before",
			).AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
		return emptyReturn
	} else {
		thread.WriteError(
			NewReplyPanic(
				"Runtime is illegal in current goroutine",
			).AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
		return emptyReturn
	}
}

// Error ...
func (p Runtime) Error(value error) Return {
	if thread := p.thread; thread == nil {
		reportPanic(
			NewReplyPanic(
				"Runtime is illegal in current goroutine",
			).AddDebug(GetFileLine(1)),
		)
		return emptyReturn
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
		thread.WriteError(
			NewReplyPanic(
				"Runtime.OK or Runtime.Error has been called before",
			).AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
		return emptyReturn
	} else {
		thread.WriteError(
			NewReplyPanic(
				"Runtime is illegal in current goroutine",
			).AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
		return emptyReturn
	}
}

func (p Runtime) Call(target string, args ...interface{}) (interface{}, Error) {

	if thread := p.thread; thread == nil {
		return nil, NewReplyPanic(
			"Runtime is illegal in current goroutine",
		)
	} else {
		frame := thread.top
		stream := NewStream()
		// write target
		stream.WriteString(target)
		// write depth
		stream.WriteUint64(frame.depth + 1)
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

		if !thread.pushFrame(p.id) {
			stream.Release()
			return nil, NewReplyPanic(
				"Runtime is illegal in current goroutine",
			)
		}

		defer func() {
			thread.popFrame()
			stream.Release()
		}()

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
