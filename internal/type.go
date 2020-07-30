package internal

import (
	"runtime/debug"
	"sync/atomic"
	"time"
	"unsafe"
)

// Bool ...
type Bool = bool

// Int64 ...
type Int64 = int64

// Uint64 ...
type Uint64 = uint64

// Float64 ...
type Float64 = float64

// String ...
type String = string

// Bytes ...
type Bytes = []byte

// Array ...
type Array = []interface{}

// Map common Map type
type Map = map[string]interface{}

// Any ...
type Any = interface{}

type IStreamConn interface {
	ReadStream(timeout time.Duration, readLimit int64) (*Stream, Error)
	WriteStream(stream *Stream, timeout time.Duration) Error
	Close() Error
}

type IAdapter interface {
	Open(onConnRun func(IStreamConn), onError func(Error))
	Close(onError func(Error))
}

// ReplyCache ...
type ReplyCache interface {
	Get(fnString string) ReplyCacheFunc
}

// ReplyCacheFunc ...
type ReplyCacheFunc = func(
	ctx Context,
	stream *Stream,
	fn interface{},
) bool

// Return ...
type Return = *ReturnObject

// ReturnObject ...
type ReturnObject struct{}

var nilReturn = (Return)(nil)

// Context ...
type Context = *ContextObject

// ContextObject ...
type ContextObject struct {
	thread unsafe.Pointer
}

var nilContext = (Context)(nil)

func (p *ContextObject) getThread() *rpcThread {
	if thread := (*rpcThread)(atomic.LoadPointer(&p.thread)); thread == nil {
		reportPanic(
			NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(GetFileLine(2)),
		)
		return nil
	} else if node := thread.execReplyNode; node == nil {
		thread.processor.Panic(
			NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(GetFileLine(2)),
		)
		return nil
	} else if !thread.processor.isDebug {
		return thread
	} else if thread.GetGoroutineID() != CurrentGoroutineID() {
		thread.processor.Panic(
			NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(GetFileLine(2)),
		)
		return nil
	} else {
		return thread
	}
}

func (p *ContextObject) stop() bool {
	if p == nil {
		reportPanic(
			NewKernelPanic("rpc: object is nil").AddDebug(string(debug.Stack())),
		)
		return false
	}

	atomic.StorePointer(&p.thread, nil)
	return true
}

// OK ...
func (p *ContextObject) OK(value interface{}) Return {
	if p == nil {
		reportPanic(
			NewReplyPanic("rpc: context is nil").AddDebug(GetFileLine(1)),
		)
		return nilReturn
	} else if thread := p.getThread(); thread == nil {
		// reportPanic has already run
		return nilReturn
	} else {
		return thread.WriteOK(value, 2)
	}
}

// Error ...
func (p *ContextObject) Error(value error) Return {
	if p == nil {
		reportPanic(
			NewReplyPanic("rpc: context is nil").AddDebug(GetFileLine(1)),
		)
		return nilReturn
	} else if thread := p.getThread(); thread == nil {
		// reportPanic has already run
		return nilReturn
	} else if err, ok := value.(Error); ok && err != nil {
		return thread.WriteError(
			err.AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
	} else if value != nil {
		return thread.WriteError(
			NewReplyError(value.Error()).
				AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
	} else {
		return thread.WriteError(
			NewReplyError("rpc: Context.Error() argument should not nil").
				AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
	}
}
