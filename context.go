package common

import (
	"fmt"
)

type serverContext struct {
	thread *rpcThread
}

type Context = *serverContext

func (p *serverContext) writeError(message string, debug string) Return {
	stream := p.thread.execStream
	stream.SetWritePos(13)
	stream.WriteBool(false)
	stream.WriteString(message)
	stream.WriteString(debug)
	p.thread.execSuccessful = false
	return nilReturn
}

// OK get success Return  by value
func (p *serverContext) OK(value interface{}) Return {
	stream := p.thread.execStream
	stream.SetWritePos(13)
	stream.WriteBool(true)
	if stream.Write(value) == common.RPCStreamWriteOK {
		p.thread.execSuccessful = true
		return nilReturn
	} else {
		return p.writeError(
			"return type is error",
			common.GetStackString(1),
		)
	}
}

// Failed get failed Return by code and message
func (p *serverContext) Error(a ...interface{}) Return {
	message := fmt.Sprint(a...)
	err := common.NewRPCErrorWithDebug(message, common.GetStackString(1))
	if p.thread != nil &&
		p.thread.execEchoNode != nil &&
		p.thread.execEchoNode.debugString != "" {
		err.AddDebug(p.thread.execEchoNode.debugString)
	}

	return p.writeError(err.GetMessage(), err.GetDebug())
}

func (p *serverContext) Errorf(format string, a ...interface{}) Return {
	message := fmt.Sprintf(format, a...)
	err := common.NewRPCErrorWithDebug(message, common.GetStackString(1))
	if p.thread != nil &&
		p.thread.execEchoNode != nil &&
		p.thread.execEchoNode.debugString != "" {
		err.AddDebug(p.thread.execEchoNode.debugString)
	}
	return p.writeError(err.GetMessage(), err.GetDebug())
}
