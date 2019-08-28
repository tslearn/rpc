package core

import (
	"fmt"
)

//type rpcInnerContext struct {
//	stream *rpcStream
//	thread *rpcThread
//}

type rpcContext struct {
	thread *rpcThread
}

func (p *rpcContext) ok() bool {
	return p != nil && p.thread != nil
}

func (p *rpcContext) getCacheStream() *rpcStream {
	if p != nil && p.thread != nil {
		return p.thread.inStream
	}
	return nil
}

func (p *rpcContext) close() bool {
	if p.thread != nil {
		p.thread = nil
		return true
	}
	return false
}

func (p *rpcContext) writeError(message string, debug string) *rpcReturn {
	if p.thread != nil {
		logger := p.thread.processor.logger
		logger.Error(NewRPCErrorByDebug(message, debug))
		execStream := p.thread.outStream
		execStream.SetWritePos(17)
		execStream.WriteBool(false)
		execStream.WriteString(message)
		execStream.WriteString(debug)
		p.thread.execSuccessful = false
	}
	return nilReturn
}

// OK get success Return  by value
func (p *rpcContext) OK(value interface{}) *rpcReturn {
	if p.thread != nil {
		stream := p.thread.outStream
		stream.SetWritePos(17)
		stream.WriteBool(true)
		if stream.Write(value) == RPCStreamWriteOK {
			p.thread.execSuccessful = true
			return nilReturn
		} else {
			return p.writeError(
				"return type is error",
				GetStackString(1),
			)
		}
	}
	return nilReturn
}

//// Failed get failed Return by code and message
//func (p *rpcContext) Error(a ...interface{}) *rpcReturn {
//	if p.inner != nil {
//		serverThread := p.inner.serverThread
//		if serverThread != nil {
//			message := fmt.Sprint(a...)
//			err := NewRPCErrorByDebug(message, GetStackString(1))
//			if serverThread.execEchoNode != nil &&
//				serverThread.execEchoNode.debugString != "" {
//				err.AddDebug(serverThread.execEchoNode.debugString)
//			}
//			return p.writeError(err.GetMessage(), err.GetDebug())
//		}
//	}
//	return nilReturn
//}

func (p *rpcContext) Errorf(format string, a ...interface{}) *rpcReturn {
	if p.thread != nil {
		message := fmt.Sprintf(format, a...)
		err := NewRPCErrorByDebug(message, GetStackString(1))
		if p.thread.execEchoNode != nil &&
			p.thread.execEchoNode.debugString != "" {
			err.AddDebug(p.thread.execEchoNode.debugString)
		}
		return p.writeError(err.GetMessage(), err.GetDebug())
	}
	return nilReturn
}
