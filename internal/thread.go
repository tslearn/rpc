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
	goroutineID   int64
	processor     *Processor
	inputCH       chan *Stream
	closeCH       chan bool
	closeTimeout  time.Duration
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
	closeTimeout time.Duration,
	onEvalBack func(*Stream),
	onEvalFinish func(*rpcThread),
) *rpcThread {
	if processor == nil || onEvalBack == nil || onEvalFinish == nil {
		return nil
	}

	retCH := make(chan *rpcThread)

	timeout := closeTimeout
	if timeout < time.Second {
		timeout = time.Second
	}

	go func() {
		thread := &rpcThread{
			goroutineID:   0,
			processor:     processor,
			inputCH:       make(chan *Stream),
			closeCH:       make(chan bool, 1),
			closeTimeout:  timeout,
			execStream:    NewStream(),
			execDepth:     0,
			execReplyNode: nil,
			execArgs:      make([]reflect.Value, 0, 16),
			execStatus:    rpcThreadExecNone,
			execFrom:      "",
		}

		if processor.isDebug {
			thread.goroutineID = CurrentGoroutineID()
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
		if p.closeCH != nil {
			defer func() {
				p.closeCH = nil
			}()

			close(p.inputCH)
			select {
			case <-p.closeCH:
				p.execStream.Release()
				p.execStream = nil
				return true
			case <-time.After(p.closeTimeout):
				return false
			}
		}

		return false
	}).(bool)
}

func (p *rpcThread) GetGoroutineID() int64 {
	return p.goroutineID
}

func (p *rpcThread) GetExecReplyNodePath() string {
	if node := p.GetReplyNode(); node != nil {
		return node.path
	}
	return ""
}

func (p *rpcThread) GetExecReplyFileLine() string {
	if node := p.GetReplyNode(); node != nil && node.meta != nil {
		return ConcatString(node.path, " ", node.meta.fileLine)
	}
	return ""
}

func (p *rpcThread) WriteError(err Error) Return {
	stream := p.execStream
	stream.SetWritePosToBodyStart()
	stream.WriteUint64(uint64(err.GetKind()))
	stream.WriteString(err.GetMessage())
	stream.WriteString(err.GetDebug())
	p.execStatus = rpcThreadExecFailed
	return nilReturn
}

func (p *rpcThread) WriteOK(value interface{}, skip uint) Return {
	stream := p.execStream
	stream.SetWritePosToBodyStart()
	stream.WriteUint64(uint64(ErrorKindNone))
	if stream.Write(value) != StreamWriteOK {
		p.processor.Panic(
			NewReplyPanic(checkValue(value, "value", 64)).
				AddDebug(AddFileLine(p.GetExecReplyNodePath(), skip)),
		)
		return p.WriteError(
			NewReplyError("reply return value error").
				AddDebug(AddFileLine(p.GetExecReplyNodePath(), skip)),
		)
	}
	p.execStatus = rpcThreadExecSuccess
	return nilReturn
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
			p.processor.Panic(
				NewReplyPanic(
					fmt.Sprintf("runtime error: %v", v),
				).AddDebug(p.GetExecReplyFileLine()).AddDebug(string(debug.Stack())),
			)

			// write runtime error
			p.WriteError(
				NewReplyError("runtime error").AddDebug(p.GetExecReplyFileLine()),
			)
		}

		func() {
			defer func() {
				if v := recover(); v != nil {
					p.processor.Panic(
						NewKernelPanic(fmt.Sprintf("kernel error: %v", v)).
							AddDebug(string(debug.Stack())),
					)
				}
			}()

			if execReplyNode != nil {
				if hasFuncReturn && p.execStatus == rpcThreadExecNone {
					p.WriteError(
						NewReplyPanic(
							"reply must return through Context.OK or Context.Error",
						).AddDebug(p.GetExecReplyFileLine()),
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
			if p.execStatus != rpcThreadExecNone {
				onEvalBack(retStream)
			}
			p.execFrom = ""
			p.execDepth = 0
			p.execArgs = p.execArgs[:0]
			atomic.StorePointer(&p.execReplyNode, nil)
			onEvalFinish(p)
		}()
	}()

	// set exec reply node
	replyPath, ok := inStream.ReadUnsafeString()
	if !ok {
		return p.WriteError(NewProtocolError(ErrStringBadStream))
	} else if execReplyNode, ok = p.processor.repliesMap[replyPath]; !ok {
		return p.WriteError(
			NewReplyError(ConcatString("target ", replyPath, " does not exist")),
		)
	} else {
		p.execReplyNode = unsafe.Pointer(execReplyNode)
	}

	if p.execDepth, ok = inStream.ReadUint64(); !ok {
		return p.WriteError(NewProtocolError(ErrStringBadStream))
	} else if p.execDepth > p.processor.maxCallDepth {
		return p.WriteError(
			NewReplyError(ConcatString(
				"call ",
				replyPath,
				" level(",
				strconv.FormatUint(p.execDepth, 10),
				") overflows",
			)).AddDebug(p.GetExecReplyFileLine()),
		)
	} else if p.execFrom, ok = inStream.ReadUnsafeString(); !ok {
		return p.WriteError(NewProtocolError(ErrStringBadStream))
	} else {
		argsStreamPos := inStream.GetReadPos()

		if fnCache := execReplyNode.cacheFN; fnCache != nil {
			ok = fnCache(ctx, inStream, execReplyNode.meta.handler)
			hasFuncReturn = true
			if ok {
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

			if !inStream.IsReadFinish() {
				ok = false
			}

			if ok {
				execReplyNode.reflectFn.Call(p.execArgs)
				hasFuncReturn = true
				return nilReturn
			}
		}

		if ok && !inStream.IsReadFinish() {
			return p.WriteError(NewProtocolError(ErrStringBadStream))
		} else if !p.processor.isDebug {
			return p.WriteError(
				NewReplyError(ConcatString(
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
					replyPath,
					" reply arguments does not match\nwant: ",
					execReplyNode.callString,
					"\ngot: ",
					replyPath,
					"(", strings.Join(remoteArgsType, ", "), ") ",
					convertTypeToString(returnType),
				)).AddDebug(p.GetExecReplyFileLine()),
			)
		}
	}
}
