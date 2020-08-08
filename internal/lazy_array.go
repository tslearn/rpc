package internal

import "runtime/debug"

type LazyArray struct {
	ctx    Context
	stream *Stream
	items  []lazyArrayItem
}

func newLazyArray(ctx Context, stream *Stream, size int) LazyArray {
	if thread := ctx.thread; thread == nil {
		reportPanic(
			NewKernelPanic("thread is nil").AddDebug(string(debug.Stack())),
		)
		return LazyArray{}
	} else if stream == nil {
		reportPanic(
			NewKernelPanic("stream is nil").AddDebug(string(debug.Stack())),
		)
		return LazyArray{}
	} else {
		return LazyArray{
			ctx:    ctx,
			stream: stream,
			items:  thread.lazyMemory.AllowArrayItem(size),
		}
	}
}

func (p LazyArray) getStream(i int, fn func(*Stream)) {
	if thread := p.ctx.thread; thread != nil && i >= 0 && i < len(p.items) {
		if ctxID := p.ctx.id; thread.lockByContext(ctxID, 3) {
			defer thread.unlockByContext(ctxID)
			p.stream.setReadPosUnsafe(p.items[i])
			fn(p.stream)
		}
	}
}

// Size ...
func (p LazyArray) Size() int {
	return len(p.items)
}

// GetBool ...
func (p LazyArray) GetBool(i int) (ret bool, ok bool) {
	p.getStream(i, func(stream *Stream) {
		ret, ok = p.stream.ReadBool()
	})
	return
}

// GetInt64 ...
func (p LazyArray) GetInt64(i int) (ret int64, ok bool) {
	p.getStream(i, func(stream *Stream) {
		ret, ok = p.stream.ReadInt64()
	})
	return
}

// GetUint64 ...
func (p LazyArray) GetUint64(i int) (ret uint64, ok bool) {
	p.getStream(i, func(stream *Stream) {
		ret, ok = p.stream.ReadUint64()
	})
	return
}

// GetFloat64 ...
func (p LazyArray) GetFloat64(i int) (ret float64, ok bool) {
	p.getStream(i, func(stream *Stream) {
		ret, ok = p.stream.ReadFloat64()
	})
	return
}

// GetString ...
func (p LazyArray) GetString(i int) (ret string, ok bool) {
	p.getStream(i, func(stream *Stream) {
		ret, ok = p.stream.ReadString()
	})
	return
}

// GetBytes ...
func (p LazyArray) GetBytes(i int) (ret []byte, ok bool) {
	p.getStream(i, func(stream *Stream) {
		ret, ok = p.stream.ReadBytes()
	})
	return
}

//
//// GetArray ...
//func (p LazyArray) GetArray(index int) (LazyArray, bool) {
//  if index < 0 || index > len(p.items) {
//    return LazyArray{}, false
//  } else if !p.ctx.thread.lockByContext(p.ctx.id, 2) {
//    return LazyArray{}, false
//  } else {
//    defer p.ctx.thread.unlockByContext(p.ctx.id)
//    p.stream.setReadPosUnsafe(p.items[index])
//    return p.stream.ReadLazyArray()
//  }
//}
//
//// GetMap ...
//func (p LazyArray) GetRPCMap(index int) (LazyMap, bool) {
//  if index < 0 || index > len(p.items) {
//    return LazyMap{}, false
//  } else if !p.ctx.thread.lockByContext(p.ctx.id, 2) {
//    return LazyMap{}, false
//  } else {
//    defer p.ctx.thread.unlockByContext(p.ctx.id)
//    p.stream.setReadPosUnsafe(p.items[index])
//    return p.stream.ReadLazyMap()
//  }
//}

// Get ...
func (p LazyArray) Get(i int) (ret interface{}, ok bool) {
	p.getStream(i, func(stream *Stream) {
		ret, ok = p.stream.Read()
	})
	return
}
