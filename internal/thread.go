package internal

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"time"
	"unsafe"
)

const (
	rpcThreadExecFailed  = -1
	rpcThreadExecNone    = 0
	rpcThreadExecSuccess = 1
)

type rpcThread struct {
	goroutineId   int64
	processor     *Processor
	inputCH       chan *Stream
	closeCH       chan bool
	execStream    *Stream
	execDepth     uint64
	execReplyNode *rpcReplyNode
	execArgs      []reflect.Value
	execStatus    int
	execFrom      string
	Lock
}

func newThread(
	processor *Processor,
	onEvalFinish func(*rpcThread, *Stream),
) *rpcThread {
	if processor == nil || onEvalFinish == nil {
		return nil
	}

	retCH := make(chan *rpcThread)

	go func() {
		thread := &rpcThread{
			processor:     processor,
			inputCH:       make(chan *Stream),
			execStream:    NewStream(),
			execDepth:     0,
			execReplyNode: nil,
			execArgs:      make([]reflect.Value, 0, 16),
			execStatus:    rpcThreadExecNone,
			execFrom:      "",
			closeCH:       make(chan bool, 1),
			goroutineId:   0,
		}

		if processor.IsDebug() {
			thread.goroutineId = CurrentGoroutineID()
		}

		retCH <- thread

		for stream := <-thread.inputCH; stream != nil; stream = <-thread.inputCH {
			thread.Eval(stream, onEvalFinish)
		}

		thread.closeCH <- true
	}()

	return <-retCH
}

func (p *rpcThread) Stop() bool {
	return p.CallWithLock(func() interface{} {
		if p.closeCH != nil {
			return false
		} else {
			close(p.inputCH)
			select {
			case <-time.After(10 * time.Second):
				p.closeCH = nil
				return false
			case <-p.closeCH:
				if p.execStream != nil {
					p.execStream.Release()
					p.execStream = nil
				}
				p.closeCH = nil
				return true
			}
		}
	}).(bool)
}

func (p *rpcThread) GetGoroutineId() int64 {
	return p.goroutineId
}

func (p *rpcThread) IsDebug() bool {
	return p.processor.IsDebug()
}

func (p *rpcThread) GetExecReplyNodePath() string {
	if node := p.execReplyNode; node != nil {
		return node.GetPath()
	}
	return ""
}

func (p *rpcThread) GetExecReplyNodeDebug() string {
	if node := p.execReplyNode; node != nil {
		return node.GetDebug()
	}
	return ""
}

func (p *rpcThread) WriteError(err Error) Return {
	stream := p.execStream
	stream.SetWritePos(streamBodyPos)
	stream.SetStreamKind(StreamKindResponseError)
	stream.WriteUint64(uint64(err.GetKind()))
	stream.WriteString(err.GetMessage())
	stream.WriteString(err.GetDebug())
	p.execStatus = rpcThreadExecFailed
	return nilReturn
}

func (p *rpcThread) WriteOK(value interface{}, skip uint) Return {
	stream := p.execStream
	stream.SetWritePos(streamBodyPos)
	stream.SetStreamKind(StreamKindResponseOK)

	if stream.Write(value) == StreamWriteOK {
		p.execStatus = rpcThreadExecSuccess
		return nilReturn
	} else if reason := CheckValue(value, 64); reason != "" {
		return p.WriteError(
			NewReplyError(ConcatString("rpc: ", reason)).
				AddDebug(AddFileLine(p.GetExecReplyNodePath(), skip)),
		)
	} else {
		return p.WriteError(
			NewReplyError("rpc: value is not supported").
				AddDebug(AddFileLine(p.GetExecReplyNodePath(), skip)),
		)
	}
}

func (p *rpcThread) PutStream(stream *Stream) (ret bool) {
	defer func() {
		if v := recover(); v != nil {
			ret = false
		} else {
			ret = true
		}
	}()
	p.inputCH <- stream
	return
}

func (p *rpcThread) Eval(
	inStream *Stream,
	onEvalFinish func(*rpcThread, *Stream),
) *ReturnObject {
	timeStart := TimeNow()
	// create context
	p.execStatus = rpcThreadExecNone
	ctx := &ContextObject{
		thread: unsafe.Pointer(p),
	}

	defer func() {
		if v := recover(); v != nil {
			if p.execReplyNode != nil {
				ReportPanic(
					NewReplyPanic(fmt.Sprintf(
						"rpc: %s: runtime error: %s",
						p.execReplyNode.callString,
						v,
					)).AddDebug(string(debug.Stack())),
				)
			}
		}

		func() {
			defer func() {
				if v := recover(); v != nil {
					ReportPanic(
						NewKernelError("rpc: kernel error").AddDebug(string(debug.Stack())),
					)
				}
			}()

			if p.execReplyNode != nil && p.execReplyNode.indicator != nil {
				p.execReplyNode.indicator.Count(
					TimeNow().Sub(timeStart),
					p.execStatus == rpcThreadExecSuccess,
				)
			}

			ctx.stop()
			inStream.Reset()
			retStream := p.execStream
			p.execStream = inStream
			p.execFrom = ""
			p.execDepth = 0
			p.execReplyNode = nil
			p.execArgs = p.execArgs[:0]
			onEvalFinish(p, retStream)
		}()
	}()

	// copy head
	copy(p.execStream.GetHeader(), inStream.GetHeader())

	// read reply path
	replyPath, ok := inStream.ReadUnsafeString()
	if !ok {
		return p.WriteError(NewProtocolError("rpc data format error"))
	}
	if p.execReplyNode, ok = p.processor.repliesMap[replyPath]; !ok {
		return p.WriteError(NewProtocolError(fmt.Sprintf(
			"rpc-server: reply path %s is not mounted",
			replyPath,
		)))
	}

	// read depth
	if p.execDepth, ok = inStream.ReadUint64(); !ok {
		return p.WriteError(NewProtocolError("rpc data format error"))
	}
	if p.execDepth > p.processor.maxCallDepth {
		return p.WriteError(NewProtocolError(fmt.Sprintf(
			"rpc current call depth(%d) is overflow. limited(%d)",
			p.execDepth,
			p.processor.maxCallDepth,
		)))
	}

	// read execFrom
	if p.execFrom, ok = inStream.ReadUnsafeString(); !ok {
		return p.WriteError(NewProtocolError("rpc data format error"))
	}

	// build callArgs
	argStartPos := inStream.GetReadPos()

	if fnCache := p.execReplyNode.cacheFN; fnCache != nil {
		ok = fnCache(ctx, inStream, p.execReplyNode.replyMeta.handler)
	} else {
		p.execArgs = append(p.execArgs, reflect.ValueOf(ctx))
		for i := 1; i < len(p.execReplyNode.argTypes); i++ {
			var rv reflect.Value

			switch p.execReplyNode.argTypes[i].Kind() {
			case reflect.Int64:
				if iVar, success := inStream.ReadInt64(); success {
					rv = reflect.ValueOf(iVar)
				} else {
					ok = false
				}
				break
			case reflect.Uint64:
				if uVar, success := inStream.ReadUint64(); success {
					rv = reflect.ValueOf(uVar)
				} else {
					ok = false
				}
				break
			case reflect.Float64:
				if fVar, success := inStream.ReadFloat64(); success {
					rv = reflect.ValueOf(fVar)
				} else {
					ok = false
				}
				break
			case reflect.Bool:
				if bVar, success := inStream.ReadBool(); success {
					rv = reflect.ValueOf(bVar)
				} else {
					ok = false
				}
				break
			case reflect.String:
				sVar, success := inStream.ReadString()
				if !success {
					ok = false
				} else {
					rv = reflect.ValueOf(sVar)
				}
				break
			default:
				switch p.execReplyNode.argTypes[i] {
				case bytesType:
					if xVar, success := inStream.ReadBytes(); success {
						rv = reflect.ValueOf(xVar)
					} else {
						ok = false
					}
					break
				case arrayType:
					if aVar, success := inStream.ReadArray(); success {
						rv = reflect.ValueOf(aVar)
					} else {
						ok = false
					}
					break
				case mapType:
					if mVar, success := inStream.ReadMap(); success {
						rv = reflect.ValueOf(mVar)
					} else {
						ok = false
					}
					break
				default:
					ok = false
				}
			}

			if !ok {
				break
			}

			p.execArgs = append(p.execArgs, rv)
		}

		if ok && !inStream.CanRead() {
			p.execReplyNode.reflectFn.Call(p.execArgs)
		}
	}

	// if stream data format error
	if !ok || inStream.CanRead() {
		inStream.SetReadPos(argStartPos)
		remoteArgsType := make([]string, 0, 0)
		remoteArgsType = append(remoteArgsType, convertTypeToString(contextType))
		for inStream.CanRead() {
			val, ok := inStream.Read()

			if !ok {
				return p.WriteError(NewProtocolError("rpc data format error"))
			}

			if val == nil {
				if len(remoteArgsType) < len(p.execReplyNode.argTypes) {
					argType := p.execReplyNode.argTypes[len(remoteArgsType)]

					if argType == bytesType ||
						argType == arrayType ||
						argType == mapType {
						remoteArgsType = append(
							remoteArgsType,
							convertTypeToString(argType),
						)
					} else {
						remoteArgsType = append(remoteArgsType, "<nil>")
					}
				} else {
					remoteArgsType = append(remoteArgsType, "<nil>")
				}
			} else {
				remoteArgsType = append(
					remoteArgsType,
					convertTypeToString(reflect.ValueOf(val).Type()),
				)
			}
		}

		return p.WriteError(NewProtocolError(fmt.Sprintf(
			"rpc reply arguments not match\nCalled: %s(%s) %s\nRequired: %s",
			replyPath,
			strings.Join(remoteArgsType, ", "),
			convertTypeToString(returnType),
			p.execReplyNode.callString,
		)).AddDebug(p.GetExecReplyNodeDebug()))
	}

	return nilReturn
}
