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

func (p Runtime) Call(
	target string,
	args ...interface{},
) (RTValue, *base.Error) {
	if thread := p.lock(); thread != nil {
		defer p.unlock()
		frame := thread.top

		// make stream
		stream, err := MakeRequestStream(target, frame.from, args...)
		if err != nil {
			return RTValue{}, err.
				AddDebug(base.AddFileLine(thread.GetExecReplyNodePath(), 1))
		}
		defer stream.Release()
		stream.SetDepth(frame.depth + 1)

		// switch thread frame
		thread.pushFrame()
		defer thread.popFrame()

		// eval
		thread.Eval(stream, func(stream *Stream) {})

		// return
		return p.ParseResponseStream(stream)
	}

	return RTValue{}, errors.
		ErrRuntimeIllegalInCurrentGoroutine.AddDebug(base.GetFileLine(1))
}

func (p Runtime) GetServiceConfig() interface{} {
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
		return node.service.addMeta.config
	}
}

func (p Runtime) ParseResponseStream(stream *Stream) (RTValue, *base.Error) {
	if errCode, ok := stream.ReadUint64(); !ok {
		return RTValue{}, errors.ErrBadStream
	} else if errCode == 0 {
		if ret, ok := stream.ReadRTValue(p); ok {
			return ret, nil
		}
		return RTValue{}, errors.ErrBadStream
	} else if message, ok := stream.ReadString(); !ok {
		return RTValue{}, errors.ErrBadStream
	} else if !stream.IsReadFinish() {
		return RTValue{}, errors.ErrBadStream
	} else {
		return RTValue{}, base.NewError(errCode, message)
	}
}
