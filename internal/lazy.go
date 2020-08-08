package internal

import "runtime/debug"

type lazyArrayItem = int
type lazyMapItem struct {
	key string
	pos int
}

type lazyMemory struct {
	arrayBuffer []lazyArrayItem
	mapBuffer   []lazyMapItem
	arrayPos    int
	mapPos      int
}

func newLazyMemory() *lazyMemory {
	return &lazyMemory{
		arrayBuffer: make([]lazyArrayItem, 128, 128),
		arrayPos:    0,
		mapBuffer:   make([]lazyMapItem, 32, 32),
		mapPos:      0,
	}
}

func (p *lazyMemory) Reset() {
	p.arrayPos = 0
	p.mapPos = 0
}

func (p *lazyMemory) AllowArrayItem(size int) []lazyArrayItem {
	if end := p.arrayPos + size; end > 128 {
		return make([]lazyArrayItem, size, size)
	} else {
		p.arrayPos = end
		return p.arrayBuffer[end-size : end]
	}
}

func (p *lazyMemory) AllowMapItem(size int) []lazyMapItem {
	if end := p.mapPos + size; end > 32 {
		return make([]lazyMapItem, size, size)
	} else {
		p.mapPos = end
		return p.mapBuffer[end-size : end]
	}
}

type LazyArray struct {
	ctx    Context
	stream *Stream
	items  []lazyArrayItem
}

var emptyLazyArray = LazyArray{}

func newLazyArray(ctx Context, stream *Stream, size int) LazyArray {
	if thread := ctx.thread; thread == nil {
		reportPanic(
			NewKernelPanic("thread is nil").AddDebug(string(debug.Stack())),
		)
		return emptyLazyArray
	} else if stream == nil {
		reportPanic(
			NewKernelPanic("stream is nil").AddDebug(string(debug.Stack())),
		)
		return emptyLazyArray
	} else if ctxID := ctx.id; !thread.lockByContext(ctxID) {
		reportPanic(
			NewKernelPanic("internal error").AddDebug(string(debug.Stack())),
		)
		return emptyLazyArray
	} else {
		defer thread.unlockByContext(ctxID)
		return LazyArray{
			ctx:    ctx,
			stream: stream,
			items:  thread.lazyMemory.AllowArrayItem(size),
		}
	}
}

// Size ...
func (p LazyArray) Size() int {
	return len(p.items)
}

// GetNil ...
func (p LazyArray) GetNil(index int) bool {
	if index < 0 || index > len(p.items) {
		return false
	} else if !p.ctx.thread.lockByContext(p.ctx.id, 2) {
		return false
	} else {
		defer p.ctx.thread.unlockByContext(p.ctx.id)
		p.stream.setReadPosUnsafe(p.items[index])
		return p.stream.ReadNil()
	}
}

// GetBool ...
func (p LazyArray) GetBool(index int) (bool, bool) {
	if index < 0 || index > len(p.items) {
		return false, false
	} else if !p.ctx.thread.lockByContext(p.ctx.id, 2) {
		return false, false
	} else {
		defer p.ctx.thread.unlockByContext(p.ctx.id)
		p.stream.setReadPosUnsafe(p.items[index])
		return p.stream.ReadBool()
	}
}

// GetInt64 ...
func (p LazyArray) GetInt64(index int) (int64, bool) {
	if index < 0 || index > len(p.items) {
		return 0, false
	} else if !p.ctx.thread.lockByContext(p.ctx.id, 2) {
		return 0, false
	} else {
		defer p.ctx.thread.unlockByContext(p.ctx.id)
		p.stream.setReadPosUnsafe(p.items[index])
		return p.stream.ReadInt64()
	}
}

// GetUint64 ...
func (p LazyArray) GetUint64(index int) (uint64, bool) {
	if index < 0 || index > len(p.items) {
		return 0, false
	} else if !p.ctx.thread.lockByContext(p.ctx.id, 2) {
		return 0, false
	} else {
		defer p.ctx.thread.unlockByContext(p.ctx.id)
		p.stream.setReadPosUnsafe(p.items[index])
		return p.stream.ReadUint64()
	}
}

// GetFloat64 ...
func (p LazyArray) GetFloat64(index int) (float64, bool) {
	if index < 0 || index > len(p.items) {
		return 0, false
	} else if !p.ctx.thread.lockByContext(p.ctx.id, 2) {
		return 0, false
	} else {
		defer p.ctx.thread.unlockByContext(p.ctx.id)
		p.stream.setReadPosUnsafe(p.items[index])
		return p.stream.ReadFloat64()
	}
}

// GetString ...
func (p LazyArray) GetString(index int) (string, bool) {
	if index < 0 || index > len(p.items) {
		return "", false
	} else if !p.ctx.thread.lockByContext(p.ctx.id, 2) {
		return "", false
	} else {
		defer p.ctx.thread.unlockByContext(p.ctx.id)
		p.stream.setReadPosUnsafe(p.items[index])
		return p.stream.ReadString()
	}
}

// GetBytes ...
func (p LazyArray) GetBytes(index int) ([]byte, bool) {
	if index < 0 || index > len(p.items) {
		return nil, false
	} else if !p.ctx.thread.lockByContext(p.ctx.id, 2) {
		return nil, false
	} else {
		defer p.ctx.thread.unlockByContext(p.ctx.id)
		p.stream.setReadPosUnsafe(p.items[index])
		return p.stream.ReadBytes()
	}
}

// GetArray ...
func (p LazyArray) GetArray(index int) (LazyArray, bool) {
	if index < 0 || index > len(p.items) {
		return LazyArray{}, false
	} else if !p.ctx.thread.lockByContext(p.ctx.id, 2) {
		return LazyArray{}, false
	} else {
		defer p.ctx.thread.unlockByContext(p.ctx.id)
		p.stream.setReadPosUnsafe(p.items[index])
		return p.stream.ReadLazyArray()
	}
}

// GetMap ...
func (p LazyArray) GetRPCMap(index int) (LazyMap, bool) {
	if index < 0 || index > len(p.items) {
		return LazyMap{}, false
	} else if !p.ctx.thread.lockByContext(p.ctx.id, 2) {
		return LazyMap{}, false
	} else {
		defer p.ctx.thread.unlockByContext(p.ctx.id)
		p.stream.setReadPosUnsafe(p.items[index])
		return p.stream.ReadLazyMap()
	}
}

// Get ...
func (p LazyArray) Get(index int) (interface{}, bool) {
	if index < 0 || index > len(p.items) {
		return LazyMap{}, false
	} else if !p.ctx.thread.lockByContext(p.ctx.id, 2) {
		return LazyMap{}, false
	} else {
		defer p.ctx.thread.unlockByContext(p.ctx.id)
		p.stream.setReadPosUnsafe(p.items[index])
		return p.stream.ReadLazyMap()
	}
}
