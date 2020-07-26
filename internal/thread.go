package internal

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"sync/atomic"
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
	execReplyNode unsafe.Pointer
	execArgs      []reflect.Value
	execStatus    int
	execFrom      string
	Lock
}

func (p *rpcThread) GetReplyNode() *rpcReplyNode {
	return (*rpcReplyNode)(atomic.LoadPointer(&p.execReplyNode))
}

func newThread(
	processor *Processor,
	onEvalBack func(*Stream),
	onEvalFinish func(*rpcThread),
) *rpcThread {
	if processor == nil || onEvalBack == nil || onEvalFinish == nil {
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
			thread.Eval(stream, onEvalBack, onEvalFinish)
		}

		thread.closeCH <- true
	}()

	return <-retCH
}

func (p *rpcThread) Close() bool {
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
	if node := p.GetReplyNode(); node != nil {
		return node.GetPath()
	}
	return ""
}

func (p *rpcThread) GetExecReplyNodeDebug() string {
	if node := p.GetReplyNode(); node != nil {
		return node.GetDebug()
	}
	return ""
}

func (p *rpcThread) WriteError(err Error) Return {
	if stream := p.execStream; stream != nil {
		stream.SetWritePosToBodyStart()
		stream.SetStreamKind(StreamKindResponseError)
		stream.WriteUint64(uint64(err.GetKind()))
		stream.WriteString(err.GetMessage())
		stream.WriteString(err.GetDebug())
		p.execStatus = rpcThreadExecFailed
		return nilReturn
	} else {
		ReportPanic(
			NewKernelFatal(ErrStringUnexpectedNil).AddDebug(string(debug.Stack())),
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
		} else {
			ReportPanic(
				NewReplyFatal(ConcatString("rpc: ", checkValue(value, "value", 64))).
					AddDebug(AddFileLine(p.GetExecReplyNodePath(), skip)),
			)
			return p.WriteError(
				NewReplyError("rpc: reply return value error").
					AddDebug(AddFileLine(p.GetExecReplyNodePath(), skip)),
			)
		}
	} else {
		ReportPanic(
			NewKernelFatal(ErrStringUnexpectedNil).AddDebug(string(debug.Stack())),
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
	onEvalBack func(*Stream),
	onEvalFinish func(*rpcThread),
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
	hasFuncReturn := false
	execReplyNode := (*rpcReplyNode)(nil)

	defer func() {
		if v := recover(); v != nil {
			// report panic
			ReportPanic(
				NewReplyFatal(
					fmt.Sprintf("rpc: %s runtime error: %v", p.GetExecReplyNodePath(), v),
				).AddDebug(string(debug.Stack())),
			)

			// write runtime error
			p.WriteError(
				NewReplyError(
					ConcatString("rpc: ", p.GetExecReplyNodePath(), " runtime error"),
				).AddDebug(p.GetExecReplyNodeDebug()),
			)
		}

		func() {
			defer func() {
				if v := recover(); v != nil {
					ReportPanic(
						NewKernelFatal(fmt.Sprintf("rpc: kernel error: %v", v)).
							AddDebug(string(debug.Stack())),
					)
				}
			}()

			if execReplyNode != nil {
				if hasFuncReturn && p.execStatus == rpcThreadExecNone {
					p.WriteError(
						NewReplyFatal(ConcatString(
							"rpc: ",
							execReplyNode.GetPath(),
							" must return through Context.OK or Context.Error",
						)).AddDebug(p.GetExecReplyNodeDebug()),
					)
				}

				execReplyNode.indicator.Count(
					TimeNow().Sub(timeStart),
					p.execStatus == rpcThreadExecSuccess,
				)
			}

			ctx.stop()
			inStream.Reset()
			retStream := p.execStream
			p.execStream = inStream
			onEvalBack(retStream)
			p.execFrom = ""
			p.execDepth = 0
			p.execArgs = p.execArgs[:0]
			atomic.StorePointer(&p.execReplyNode, nil)
			onEvalFinish(p)
		}()
	}()

	// read reply path
	if replyPath, ok := inStream.ReadUnsafeString(); !ok {
		return p.WriteError(NewProtocolError(ErrStringBadStream))
	} else if execReplyNode, ok = p.processor.repliesMap[replyPath]; !ok {
		return p.WriteError(
			NewReplyError(ConcatString("rpc: target ", replyPath, " does not exist")),
		)
	} else if p.execDepth, ok = inStream.ReadUint64(); !ok {
		return p.WriteError(NewProtocolError(ErrStringBadStream))
	} else if p.execDepth > p.processor.maxCallDepth {
		return p.WriteError(
			NewReplyError(ConcatString(
				"rpc: call ",
				replyPath,
				" level(",
				strconv.FormatUint(p.execDepth, 10),
				") overflows",
			)).AddDebug(execReplyNode.GetDebug()),
		)
	} else if p.execFrom, ok = inStream.ReadUnsafeString(); !ok {
		return p.WriteError(NewProtocolError(ErrStringBadStream))
	} else {
		p.execReplyNode = unsafe.Pointer(execReplyNode)
		argsStreamPos := inStream.GetReadPos()

		if fnCache := execReplyNode.cacheFN; fnCache != nil {
			ok = fnCache(ctx, inStream, execReplyNode.replyMeta.handler)
			hasFuncReturn = true
			if ok && inStream.IsReadFinish() {
				return nilReturn
			}
		} else {
			p.execArgs = append(p.execArgs, reflect.ValueOf(ctx))
			for i := 1; i < len(execReplyNode.argTypes); i++ {
				var rv reflect.Value

				switch execReplyNode.argTypes[i] {
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

			if ok && inStream.IsReadFinish() {
				execReplyNode.reflectFn.Call(p.execArgs)
				hasFuncReturn = true
				return nilReturn
			}
		}

		if ok && !inStream.IsReadFinish() {
			return p.WriteError(NewProtocolError(ErrStringBadStream))
		} else if !p.IsDebug() {
			return p.WriteError(
				NewReplyError(ConcatString(
					"rpc: ",
					replyPath,
					" reply arguments does not match",
				)),
			)
		} else {
			remoteArgsType := make([]string, 0, 0)
			remoteArgsType = append(remoteArgsType, convertTypeToString(contextType))
			inStream.setReadPosUnsafe(argsStreamPos)
			for inStream.CanRead() {
				if val, ok := inStream.Read(); !ok {
					return p.WriteError(NewProtocolError(ErrStringBadStream))
				} else if val != nil {
					remoteArgsType = append(
						remoteArgsType,
						convertTypeToString(reflect.ValueOf(val).Type()),
					)
				} else {
					if len(remoteArgsType) < len(execReplyNode.argTypes) {
						argType := execReplyNode.argTypes[len(remoteArgsType)]
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
					"rpc: ",
					replyPath,
					" reply arguments does not match\nwant: ",
					execReplyNode.callString,
					"\ngot: ",
					replyPath,
					"(", strings.Join(remoteArgsType, ", "), ") ",
					convertTypeToString(returnType),
				)).AddDebug(execReplyNode.GetDebug()),
			)
		}
	}
}
