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

// Reply ...
func (p Runtime) Reply(value interface{}) Return {
	thread := p.lock()

	if thread == nil {
		base.PublishPanic(
			errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(base.GetFileLine(1)),
		)
		return emptyReturn
	}

	defer p.unlock()
	return thread.Write(value, 1, true)
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
			err: err.AddDebug(base.AddFileLine(thread.GetExecActionNodePath(), 1)),
		}
	}
	defer stream.Release()
	stream.SetDepth(frame.depth + 1)

	// switch thread frame
	thread.pushFrame()
	thread.Eval(stream, func(stream *Stream) {})
	thread.popFrame()

	// return
	ret := p.ParseResponseStream(stream)
	if ret.err != nil {
		ret.err.AddDebug(base.AddFileLine(thread.GetExecActionNodePath(), 1))
	}

	return ret
}

// GetServiceData ...
func (p Runtime) GetServiceData(key string) (Any, *base.Error) {
	if thread := p.thread; thread == nil {
		return nil, errors.ErrRuntimeIllegalInCurrentGoroutine.
			AddDebug(base.GetFileLine(1))
	} else if actionNode := thread.GetActionNode(); actionNode == nil {
		return nil, errors.ErrRuntimeIllegalInCurrentGoroutine.
			AddDebug(base.GetFileLine(1))
	} else if serviceNode := actionNode.service; serviceNode == nil {
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
	} else if actionNode := thread.GetActionNode(); actionNode == nil {
		return errors.ErrRuntimeIllegalInCurrentGoroutine.
			AddDebug(base.GetFileLine(1))
	} else if serviceNode := actionNode.service; serviceNode == nil {
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
		ret, _ := stream.ReadRTValue(p)
		return ret
	} else if message, err := stream.ReadString(); err != nil {
		return RTValue{err: err}
	} else if !stream.IsReadFinish() {
		return RTValue{err: errors.ErrStream}
	} else {
		return RTValue{err: base.NewError(errCode, message)}
	}
}

func (p Runtime) NewRTArray() RTArray {
	return newRTArray(p, 16)
}

func (p Runtime) NewRTMap() RTMap {
	return newRTMap(p, 12)
}
