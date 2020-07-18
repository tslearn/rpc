package internal

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

type ContextObject struct {
	thread unsafe.Pointer
}

func (p *ContextObject) stop() {
	atomic.StorePointer(&p.thread, nil)
}

func (p *ContextObject) writeError(message string, debug string) *ReturnObject {
	if thread := (*rpcThread)(p.thread); thread != nil {
		execStream := thread.outStream
		execStream.SetWritePos(streamBodyPos)
		execStream.WriteBool(false)
		execStream.WriteString(message)
		execStream.WriteString(debug)
		thread.execSuccessful = false
	}
	return nilReturn
}

// OK get success ReturnObject  by value
func (p *ContextObject) OK(value interface{}) *ReturnObject {
	if thread := (*rpcThread)(p.thread); thread != nil {
		stream := thread.outStream
		stream.SetWritePos(streamBodyPos)
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

func (p *ContextObject) Error(err Error) *ReturnObject {
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

func (p *ContextObject) Errorf(format string, a ...interface{}) *ReturnObject {
	return p.Error(NewError(
		fmt.Sprintf(format, a...),
	).AddDebug(GetStackString(1)))
}
