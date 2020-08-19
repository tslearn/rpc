package internal

import (
	"net"
	"time"
)

// IStreamConn ...
type IStreamConn interface {
	ReadStream(timeout time.Duration, readLimit int64) (*Stream, Error)
	WriteStream(stream *Stream, timeout time.Duration) Error
	Close() Error
}

// IServerAdapter ...
type IServerAdapter interface {
	Open(onConnRun func(IStreamConn, net.Addr), onError func(uint64, Error))
	Close(onError func(uint64, Error))
}

// IClientAdapter ...
type IClientAdapter interface {
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

// Return ...
type Return struct{}

var emptyReturn = Return{}

// Context ...
type Context struct {
	id     uint64
	thread *rpcThread
}

// OK ...
func (p Context) OK(value interface{}) Return {
	if thread := p.thread; thread == nil {
		reportPanic(
			NewReplyPanic(
				"Context is illegal in current goroutine",
			).AddDebug(GetFileLine(1)),
		)
		return emptyReturn
	} else if code := thread.setReturn(p.id); code == rpcThreadReturnStatusOK {
		stream := thread.execStream
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
				"Context.OK or Context.Error has been called before",
			).AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
		return emptyReturn
	} else {
		thread.WriteError(
			NewReplyPanic(
				"Context is illegal in current goroutine",
			).AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
		return emptyReturn
	}
}

// Error ...
func (p Context) Error(value error) Return {
	if thread := p.thread; thread == nil {
		reportPanic(
			NewReplyPanic(
				"Context is illegal in current goroutine",
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
				"Context.OK or Context.Error has been called before",
			).AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
		return emptyReturn
	} else {
		thread.WriteError(
			NewReplyPanic(
				"Context is illegal in current goroutine",
			).AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
		return emptyReturn
	}
}

func (p Context) GetServiceData() interface{} {
	if thread := p.thread; thread == nil {
		reportPanic(
			NewReplyPanic(
				"Context is illegal in current goroutine",
			).AddDebug(GetFileLine(1)),
		)
		return nil
	} else if node := thread.GetReplyNode(); node == nil {
		reportPanic(
			NewReplyPanic(
				"Context is illegal in current goroutine",
			).AddDebug(GetFileLine(1)),
		)
		return nil
	} else {
		return node.service.addMeta.data
	}
}
