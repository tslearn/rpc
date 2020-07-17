package internal

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

type Context struct {
	thread unsafe.Pointer
}

func (p *Context) stop() {
	atomic.StorePointer(&p.thread, nil)
}

func (p *Context) writeError(message string, debug string) *Return {
	if thread := (*rpcThread)(p.thread); thread != nil {
		execStream := thread.outStream
		execStream.SetWritePos(StreamBodyPos)
		execStream.WriteBool(false)
		execStream.WriteString(message)
		execStream.WriteString(debug)
		thread.execSuccessful = false
	}
	return nilReturn
}

// OK get success Return  by value
func (p *Context) OK(value interface{}) *Return {
	if thread := (*rpcThread)(p.thread); thread != nil {
		stream := thread.outStream
		stream.SetWritePos(StreamBodyPos)
		stream.WriteBool(true)

		if stream.Write(value) != StreamWriteOK {
			return p.writeError(
				"return type is error",
				GetStackString(1),
			)
		}

		thread.execSuccessful = true
	}
	return nilReturn
}

func (p *Context) Error(err Error) *Return {
	if err == nil {
		return nilReturn
	}

	if thread := (*rpcThread)(p.thread); thread != nil &&
		thread.execReplyNode != nil &&
		thread.execReplyNode.debugString != "" {
		err.AddDebug(thread.execReplyNode.debugString)
	}

	return p.writeError(err.GetMessage(), err.GetDebug())
}

func (p *Context) Errorf(format string, a ...interface{}) *Return {
	return p.Error(NewError(
		fmt.Sprintf(format, a...),
	).AddDebug(GetStackString(1)))
}
