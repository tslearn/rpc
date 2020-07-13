package internal

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

type RPCContext struct {
	thread unsafe.Pointer
}

func (p *RPCContext) getThread() *rpcThread {
	return (*rpcThread)(p.thread)
}

func (p *RPCContext) stop() {
	atomic.StorePointer(&p.thread, nil)
}

func (p *RPCContext) writeError(message string, debug string) *RPCReturn {
	if thread := p.getThread(); thread != nil {
		execStream := thread.outStream
		execStream.SetWritePos(StreamBodyPos)
		execStream.WriteBool(false)
		execStream.WriteString(message)
		execStream.WriteString(debug)
		thread.execSuccessful = false
	}
	return nilReturn
}

// OK get success RPCReturn  by value
func (p *RPCContext) OK(value interface{}) *RPCReturn {
	if thread := p.getThread(); thread != nil {
		stream := thread.outStream
		stream.SetWritePos(StreamBodyPos)
		stream.WriteBool(true)

		if stream.Write(value) != RPCStreamWriteOK {
			return p.writeError(
				"return type is error",
				GetStackString(1),
			)
		}

		thread.execSuccessful = true
	}
	return nilReturn
}

func (p *RPCContext) Error(err RPCError) *RPCReturn {
	if err == nil {
		return nilReturn
	}

	if thread := p.getThread(); thread != nil &&
		thread.execEchoNode != nil &&
		thread.execEchoNode.debugString != "" {
		err.AddDebug(thread.execEchoNode.debugString)
	}

	return p.writeError(err.GetMessage(), err.GetDebug())
}

func (p *RPCContext) Errorf(format string, a ...interface{}) *RPCReturn {
	return p.Error(
		NewRPCErrorByDebug(
			fmt.Sprintf(format, a...),
			GetStackString(1),
		))
}
