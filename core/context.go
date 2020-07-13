package core

import (
	"fmt"
	"github.com/tslearn/rpcc/util"
	"sync/atomic"
	"unsafe"
)

type rpcContext struct {
	thread unsafe.Pointer
}

func (p *rpcContext) getThread() *rpcThread {
	return (*rpcThread)(p.thread)
}

func (p *rpcContext) stop() {
	atomic.StorePointer(&p.thread, nil)
}

func (p *rpcContext) writeError(message string, debug string) *rpcReturn {
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
func (p *rpcContext) OK(value interface{}) *rpcReturn {
	if thread := p.getThread(); thread != nil {
		stream := thread.outStream
		stream.SetWritePos(StreamBodyPos)
		stream.WriteBool(true)

		if stream.Write(value) != RPCStreamWriteOK {
			return p.writeError(
				"return type is error",
				util.GetStackString(1),
			)
		}

		thread.execSuccessful = true
	}
	return nilReturn
}

func (p *rpcContext) Error(err RPCError) *rpcReturn {
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

func (p *rpcContext) Errorf(format string, a ...interface{}) *rpcReturn {
	return p.Error(
		NewRPCErrorByDebug(
			fmt.Sprintf(format, a...),
			util.GetStackString(1),
		))
}
