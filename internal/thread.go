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
			goroutineId:   0,
			processor:     processor,
			inputCH:       make(chan *Stream),
			closeCH:       make(chan bool, 1),
			execStream:    NewStream(),
			execDepth:     0,
			execReplyNode: nil,
			execArgs:      make([]reflect.Value, 0, 16),
			execStatus:    rpcThreadExecNone,
			execFrom:      "",
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
		if p.closeCH == nil {
			return false
		} else {
			close(p.inputCH)
			select {
			case <-time.After(20 * time.Second):
				p.closeCH = nil
				return false
			case <-p.closeCH:
				p.execStream.Release()
				p.execStream = nil
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
	if stream := p.execStream; stream != nil {
		stream := p.execStream
		stream.SetWritePosToBodyStart()
		stream.SetStreamKind(StreamKindResponseError)
		stream.WriteUint64(uint64(err.GetKind()))
		stream.WriteString(err.GetMessage())
		stream.WriteString(err.GetDebug())
		p.execStatus = rpcThreadExecFailed
		return nilReturn
	} else {
		ReportPanic(
			NewKernelError(ErrStringUnexpectedNil).AddDebug(string(debug.Stack())),
		)
		return nilReturn
	}
}

func (p *rpcThread) WriteOK(value interface{}, skip uint) Return {
	if stream := p.execStream; stream != nil {
		stream.SetWritePosToBodyStart()
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
	} else {
		ReportPanic(
			NewKernelError(ErrStringUnexpectedNil).AddDebug(string(debug.Stack())),
		)
		return nilReturn
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
) Return {
	timeStart := TimeNow()
	inStream.SetReadPosToBodyStart()
	p.execStatus = rpcThreadExecNone
	// copy head
	copy(p.execStream.GetHeader(), inStream.GetHeader())
	// create context
	ctx := &ContextObject{
		thread: unsafe.Pointer(p),
	}

	defer func() {
		if v := recover(); v != nil {
			ReportPanic(
				NewReplyPanic(ErrStringRuntime).AddDebug(string(debug.Stack())),
			)
			p.WriteError(
				NewReplyError(ErrStringRuntime),
			)
		}

		func() {
			defer func() {
				if v := recover(); v != nil {
					ReportPanic(
						NewKernelError(ErrStringRuntime).AddDebug(string(debug.Stack())),
					)
				}
			}()

			if p.execReplyNode != nil {
				if p.execStatus == rpcThreadExecNone {
					p.WriteError(
						NewReplyPanic(ConcatString(
							"rpc: ",
							p.execReplyNode.GetPath(),
							" must return through Context.OK or Context.Error",
						)),
					)
				}

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

	// read reply path
	if replyPath, ok := inStream.ReadUnsafeString(); !ok {
		return p.WriteError(NewProtocolError(ErrStringBadStream))
	} else if p.execReplyNode, ok = p.processor.repliesMap[replyPath]; !ok {
		return p.WriteError(
			NewReplyError(ConcatString("rpc: target ", replyPath, " does not exist")),
		)
	} else if p.execDepth, ok = inStream.ReadUint64(); !ok {
		return p.WriteError(NewProtocolError(ErrStringBadStream))
	} else if p.execDepth > p.processor.maxCallDepth {
		return p.WriteError(
			NewReplyError(ConcatString("rpc: call ", replyPath, " overflows")),
		)
	} else if p.execFrom, ok = inStream.ReadUnsafeString(); !ok {
		return p.WriteError(NewProtocolError(ErrStringBadStream))
	} else {
		if fnCache := p.execReplyNode.cacheFN; fnCache != nil {
			ok = fnCache(ctx, inStream, p.execReplyNode.replyMeta.handler)
		} else {
			p.execArgs = append(p.execArgs, reflect.ValueOf(ctx))
			for i := 1; i < len(p.execReplyNode.argTypes); i++ {
				var rv reflect.Value

				switch p.execReplyNode.argTypes[i] {
				case int64Type:
					if iVar, success := inStream.ReadInt64(); success {
						rv = reflect.ValueOf(iVar)
					} else {
						ok = false
					}
				case uint64Type:
					if uVar, success := inStream.ReadUint64(); success {
						rv = reflect.ValueOf(uVar)
					} else {
						ok = false
					}
				case float64Type:
					if fVar, success := inStream.ReadFloat64(); success {
						rv = reflect.ValueOf(fVar)
					} else {
						ok = false
					}
				case boolType:
					if bVar, success := inStream.ReadBool(); success {
						rv = reflect.ValueOf(bVar)
					} else {
						ok = false
					}
				case stringType:
					if sVar, success := inStream.ReadString(); success {
						rv = reflect.ValueOf(sVar)
					} else {
						ok = false
					}
				case bytesType:
					if xVar, success := inStream.ReadBytes(); success {
						rv = reflect.ValueOf(xVar)
					} else {
						ok = false
					}
				case arrayType:
					if aVar, success := inStream.ReadArray(); success {
						rv = reflect.ValueOf(aVar)
					} else {
						ok = false
					}
				case mapType:
					if mVar, success := inStream.ReadMap(); success {
						rv = reflect.ValueOf(mVar)
					} else {
						ok = false
					}
				default:
					ok = false
				}

				if !ok {
					break
				} else {
					p.execArgs = append(p.execArgs, rv)
				}
			}

			if ok {
				if !inStream.IsReadFinish() {
					ok = false
				} else {
					p.execReplyNode.reflectFn.Call(p.execArgs)
				}
			}
		}

		if ok {
			return nilReturn
		} else if !p.IsDebug() {
			return p.WriteError(
				NewReplyError(ConcatString(
					"rpc: reply ",
					replyPath,
					" arguments does not match",
				)),
			)
		} else {
			remoteArgsType := make([]string, 0, 0)
			remoteArgsType = append(remoteArgsType, convertTypeToString(contextType))
			for inStream.CanRead() {
				if val, ok := inStream.Read(); !ok {
					return p.WriteError(NewProtocolError(ErrStringBadStream))
				} else if val != nil {
					remoteArgsType = append(
						remoteArgsType,
						convertTypeToString(reflect.ValueOf(val).Type()),
					)
				} else {
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
				}
			}

			return p.WriteError(
				NewReplyError(ConcatString(
					"rpc: reply ",
					replyPath,
					" arguments does not match",
				)).AddDebug(fmt.Sprintf(
					"want: %s(%s) %s\ngot: %s",
					replyPath,
					strings.Join(remoteArgsType, ", "),
					convertTypeToString(returnType),
					p.execReplyNode.callString,
				)),
			)
		}
	}
}
