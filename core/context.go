package core

import (
	"fmt"
)

type rpcInnerContext struct {
	stream       *rpcStream
	serverThread *rpcThread
	clientThread *uint64
}

type rpcContext struct {
	inner *rpcInnerContext
}

func (p *rpcContext) getCacheStream() *rpcStream {
	if p != nil && p.inner != nil {
		return p.inner.stream
	}
	return nil
}

func (p *rpcContext) close() bool {
	if p.inner != nil {
		p.inner = nil
		return true
	}
	return false
}

func (p *rpcContext) writeError(message string, debug string) *rpcReturn {
	if p.inner != nil {
		serverThread := p.inner.serverThread
		if serverThread != nil {
			logger := serverThread.processor.logger
			logger.Error(NewRPCErrorWithDebug(message, debug))
			execStream := serverThread.execStream
			execStream.SetWritePos(17)
			execStream.WriteBool(false)
			execStream.WriteString(message)
			execStream.WriteString(debug)
			serverThread.execSuccessful = false
		}
	}
	return nilReturn
}

// OK get success Return  by value
func (p *rpcContext) OK(value interface{}) *rpcReturn {
	if p.inner != nil {
		serverThread := p.inner.serverThread
		if serverThread != nil {
			stream := serverThread.execStream
			stream.SetWritePos(17)
			stream.WriteBool(true)
			if stream.Write(value) == RPCStreamWriteOK {
				serverThread.execSuccessful = true
				return nilReturn
			} else {
				return p.writeError(
					"return type is error",
					GetStackString(1),
				)
			}
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
//			err := NewRPCErrorWithDebug(message, GetStackString(1))
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
	if p.inner != nil {
		serverThread := p.inner.serverThread
		if serverThread != nil {
			message := fmt.Sprintf(format, a...)
			err := NewRPCErrorWithDebug(message, GetStackString(1))
			if serverThread.execEchoNode != nil &&
				serverThread.execEchoNode.debugString != "" {
				err.AddDebug(serverThread.execEchoNode.debugString)
			}
			return p.writeError(err.GetMessage(), err.GetDebug())
		}
	}
	return nilReturn
}
