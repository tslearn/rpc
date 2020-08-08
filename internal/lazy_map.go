package internal

import "runtime/debug"

type LazyMap struct {
	ctx    Context
	stream *Stream
	items  []lazyMapItem
}

func newLazyMap(ctx Context, stream *Stream, size int) LazyMap {
	if thread := ctx.thread; thread == nil {
		reportPanic(
			NewKernelPanic("thread is nil").AddDebug(string(debug.Stack())),
		)
		return LazyMap{}
	} else if stream == nil {
		reportPanic(
			NewKernelPanic("stream is nil").AddDebug(string(debug.Stack())),
		)
		return LazyMap{}
	} else {
		return LazyMap{
			ctx:    ctx,
			stream: stream,
			items:  thread.lazyMemory.AllowMapItem(size),
		}
	}
}

func (p LazyMap) getStream(key string, fn func(*Stream)) {
	if thread := p.ctx.thread; thread != nil {
		if ctxID := p.ctx.id; thread.lockByContext(ctxID, 3) {
			defer thread.unlockByContext(ctxID)

			for _, v := range p.items {
				if v.key == key {
					p.stream.setReadPosUnsafe(v.pos)
					fn(p.stream)
					return
				}
			}
		}
	}
}

// Size ...
func (p LazyMap) Size() int {
	return len(p.items)
}

// GetBool ...
func (p LazyMap) GetBool(key string) (ret bool, ok bool) {
	p.getStream(key, func(stream *Stream) {
		ret, ok = p.stream.ReadBool()
	})
	return
}

// GetInt64 ...
func (p LazyMap) GetInt64(key string) (ret int64, ok bool) {
	p.getStream(key, func(stream *Stream) {
		ret, ok = p.stream.ReadInt64()
	})
	return
}

// GetUint64 ...
func (p LazyMap) GetUint64(key string) (ret uint64, ok bool) {
	p.getStream(key, func(stream *Stream) {
		ret, ok = p.stream.ReadUint64()
	})
	return
}

// GetFloat64 ...
func (p LazyMap) GetFloat64(key string) (ret float64, ok bool) {
	p.getStream(key, func(stream *Stream) {
		ret, ok = p.stream.ReadFloat64()
	})
	return
}

// GetString ...
func (p LazyMap) GetString(key string) (ret string, ok bool) {
	p.getStream(key, func(stream *Stream) {
		ret, ok = p.stream.ReadString()
	})
	return
}

// GetBytes ...
func (p LazyMap) GetBytes(key string) (ret []byte, ok bool) {
	p.getStream(key, func(stream *Stream) {
		ret, ok = p.stream.ReadBytes()
	})
	return
}

// Get ...
func (p LazyMap) Get(key string) (ret interface{}, ok bool) {
	p.getStream(key, func(stream *Stream) {
		ret, ok = p.stream.Read()
	})
	return
}
