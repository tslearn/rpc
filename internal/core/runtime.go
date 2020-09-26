package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
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
		errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(base.GetFileLine(1)),
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
				errors.ErrGeneralCustomError.
					AddDebug(value.Error()).
					AddDebug(base.AddFileLine(thread.GetExecReplyNodePath(), 1)),
				1,
			)
		} else {
			return p.thread.WriteError(
				errors.ErrRuntimeErrorArgumentIsNil.AddDebug(
					base.AddFileLine(thread.GetExecReplyNodePath(), 1),
				),
				1,
			)
		}
	}

	base.PublishPanic(
		errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(base.GetFileLine(1)),
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
			if reason := stream.Write(args[i]); reason != StreamWriteOK {
				return nil, errors.ErrRuntimeArgumentNotSupported.
					AddDebug(base.ConcatString("value", reason)).
					AddDebug(base.AddFileLine(thread.GetExecReplyNodePath(), 1))
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
			return nil, errors.ErrBadStream
		} else if errCode == 0 {
			if ret, ok := stream.Read(); ok {
				return ret, nil
			}
			return nil, errors.ErrBadStream
		} else if message, ok := stream.ReadString(); !ok {
			return nil, errors.ErrBadStream
		} else if !stream.IsReadFinish() {
			return nil, errors.ErrBadStream
		} else {
			return nil, base.NewError(errCode, message)
		}
	}

	return nil, errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(base.GetFileLine(1))
}

func (p Runtime) GetServiceData() interface{} {
	if thread := p.thread; thread == nil {
		base.PublishPanic(
			errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(base.GetFileLine(1)),
		)
		return nil
	} else if node := thread.GetReplyNode(); node == nil {
		base.PublishPanic(
			errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(base.GetFileLine(1)),
		)
		return nil
	} else {
		return node.service.addMeta.data
	}
}
