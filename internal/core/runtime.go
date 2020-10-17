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
	thread := p.lock()

	if thread == nil {
		base.PublishPanic(
			errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(base.GetFileLine(1)),
		)
		return emptyReturn
	}

	defer p.unlock()
	return thread.WriteOK(value, 1)
}

// Error ...
func (p Runtime) Error(value error) Return {
	thread := p.lock()
	if thread == nil {
		base.PublishPanic(
			errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(base.GetFileLine(1)),
		)
		return emptyReturn
	}
	defer p.unlock()

	if err, ok := value.(*base.Error); ok && err != nil {
		return p.thread.WriteError(
			err.AddDebug(base.AddFileLine(thread.GetExecReplyNodePath(), 1)),
			1,
		)
	} else if value != nil {
		return p.thread.WriteError(
			errors.ErrReplyCustom.
				AddDebug(value.Error()).
				AddDebug(base.AddFileLine(thread.GetExecReplyNodePath(), 1)),
			1,
		)
	} else {
		return p.thread.WriteError(
			errors.ErrUnsupportedValue.
				AddDebug("value type(nil) is not supported").
				AddDebug(base.AddFileLine(thread.GetExecReplyNodePath(), 1)),
			1,
		)
	}
}

// Call ...
func (p Runtime) Call(target string, args ...interface{}) RTValue {
	thread := p.lock()
	if thread == nil {
		return RTValue{
			err: errors.ErrRuntimeIllegalInCurrentGoroutine.
				AddDebug(base.GetFileLine(1)),
		}
	}
	defer p.unlock()

	frame := thread.top

	// make stream
	stream, err := MakeRequestStream(target, frame.from, args...)
	if err != nil {
		return RTValue{
			err: err.AddDebug(base.AddFileLine(thread.GetExecReplyNodePath(), 1)),
		}
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

// GetServiceData ...
func (p Runtime) GetServiceData(key string) (Any, *base.Error) {
	if thread := p.thread; thread == nil {
		return nil, errors.ErrRuntimeIllegalInCurrentGoroutine.
			AddDebug(base.GetFileLine(1))
	} else if replyNode := thread.GetReplyNode(); replyNode == nil {
		return nil, errors.ErrRuntimeIllegalInCurrentGoroutine.
			AddDebug(base.GetFileLine(1))
	} else if serviceNode := replyNode.service; serviceNode == nil {
		return nil, errors.ErrGetServiceDataServiceNodeIsNil
	} else {
		return serviceNode.GetData(key), nil
	}
}

// SetServiceData ...
func (p Runtime) SetServiceData(key string, value Any) *base.Error {
	if thread := p.thread; thread == nil {
		return errors.ErrRuntimeIllegalInCurrentGoroutine.
			AddDebug(base.GetFileLine(1))
	} else if replyNode := thread.GetReplyNode(); replyNode == nil {
		return errors.ErrRuntimeIllegalInCurrentGoroutine.
			AddDebug(base.GetFileLine(1))
	} else if serviceNode := replyNode.service; serviceNode == nil {
		return errors.ErrSetServiceDataServiceNodeIsNil
	} else {
		serviceNode.SetData(key, value)
		return nil
	}
}

// ParseResponseStream ...
func (p Runtime) ParseResponseStream(stream *Stream) RTValue {
	if errCode, err := stream.ReadUint64(); err != nil {
		return RTValue{err: err}
	} else if errCode == 0 {
		return stream.ReadRTValue(p)
	} else if message, err := stream.ReadString(); err != nil {
		return RTValue{err: err}
	} else if !stream.IsReadFinish() {
		return RTValue{err: errors.ErrStream}
	} else {
		return RTValue{err: base.NewError(errCode, message)}
	}
}
