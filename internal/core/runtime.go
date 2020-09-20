package core

import (
	"github.com/rpccloud/rpc/internal/base"
)

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

	base.PublishPanic(
		base.ReplyFatal.AddDebug(
			"Runtime is illegal in current goroutine",
		).AddDebug(base.GetFileLine(1)),
	)
	return emptyReturn
}

// Error ...
func (p Runtime) Error(value error) Return {
	if thread := p.lock(); thread != nil {
		defer p.unlock()

		if err, ok := value.(*base.Error); ok && err != nil {
			return p.thread.WriteError(
				err.AddDebug(base.AddFileLine(thread.GetExecReplyNodePath(), 1)),
				1,
			)
		} else if value != nil {
			return p.thread.WriteError(
				base.ReplyError.AddDebug(
					value.Error(),
				).AddDebug(base.AddFileLine(thread.GetExecReplyNodePath(), 1)),
				1,
			)
		} else {
			return p.thread.WriteError(
				base.ReplyError.AddDebug(
					"argument should not nil",
				).AddDebug(base.AddFileLine(thread.GetExecReplyNodePath(), 1)),
				1,
			)
		}
	}

	base.PublishPanic(
		base.ReplyFatal.AddDebug(
			"Runtime is illegal in current goroutine",
		).AddDebug(base.GetFileLine(1)),
	)
	return emptyReturn
}

func (p Runtime) Call(target string, args ...interface{}) (interface{}, *base.Error) {
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
				return nil, base.ReplyFatal.AddDebug(base.ConcatString(
					base.ConvertOrdinalToString(uint(i+1)),
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
		if errCode, ok := stream.ReadUint64(); !ok {
			return nil, base.ProtocolErrorBadStream
		} else if errCode == 0 {
			if ret, ok := stream.Read(); ok {
				return ret, nil
			}
			return nil, base.ProtocolErrorBadStream
		} else if message, ok := stream.ReadString(); !ok {
			return nil, base.ProtocolErrorBadStream
		} else if !stream.IsReadFinish() {
			return nil, base.ProtocolErrorBadStream
		} else {
			return nil, base.NewError(errCode, message)
		}
	}

	return nil, base.ReplyFatal.AddDebug(
		"Runtime is illegal in current goroutine",
	)
}

func (p Runtime) GetServiceData() interface{} {
	if thread := p.thread; thread == nil {
		base.PublishPanic(
			base.ReplyFatal.AddDebug(
				"Runtime is illegal in current goroutine",
			).AddDebug(base.GetFileLine(1)),
		)
		return nil
	} else if node := thread.GetReplyNode(); node == nil {
		base.PublishPanic(
			base.ReplyFatal.AddDebug(
				"Runtime is illegal in current goroutine",
			).AddDebug(base.GetFileLine(1)),
		)
		return nil
	} else {
		return node.service.addMeta.data
	}
}
