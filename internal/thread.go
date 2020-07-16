package internal

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"time"
	"unsafe"
)

type rpcThread struct {
	processor      *Processor
	isRunning      bool
	ch             chan *RPCStream
	inStream       *RPCStream
	outStream      *RPCStream
	execDepth      uint64
	execReplyNode  *replyNode
	execArgs       []reflect.Value
	execSuccessful bool
	from           string
	closeCH        chan bool
	Lock
}

func newThread(
	processor *Processor,
	onEvalFinish func(*rpcThread, *RPCStream, bool),
	onPanic func(v interface{}, debug string),
) *rpcThread {
	if processor == nil || onEvalFinish == nil || onPanic == nil {
		return nil
	}

	ret := &rpcThread{
		processor:      processor,
		isRunning:      true,
		ch:             make(chan *RPCStream, 1),
		inStream:       nil,
		outStream:      NewRPCStream(),
		execDepth:      0,
		execReplyNode:  nil,
		execArgs:       make([]reflect.Value, 0, 16),
		execSuccessful: false,
		from:           "",
		closeCH:        make(chan bool),
	}

	go func() {
		for stream := <-ret.ch; stream != nil; stream = <-ret.ch {
			ret.eval(stream, onEvalFinish, onPanic)
		}
		ret.closeCH <- true
	}()

	return ret
}

func (p *rpcThread) Stop() bool {
	return p.CallWithLock(func() interface{} {
		if !p.isRunning {
			return false
		} else {
			p.isRunning = false
			close(p.ch)
			select {
			case <-time.After(10 * time.Second):
				return false
			case <-p.closeCH:
				return true
			}
		}
	}).(bool)
}

func (p *rpcThread) PutStream(stream *RPCStream) bool {
	p.ch <- stream
	return true
}

func (p *rpcThread) eval(
	inStream *RPCStream,
	onEvalFinish func(*rpcThread, *RPCStream, bool),
	onPanic func(v interface{}, debug string),
) *RPCReturn {
	timeStart := TimeNowNS()
	// create context
	p.inStream = inStream
	p.execSuccessful = false
	ctx := &RPCContext{thread: unsafe.Pointer(p)}

	defer func() {
		if v := recover(); v != nil {
			onPanic(v, string(debug.Stack()))
			if p.execReplyNode != nil {
				ctx.writeError(
					fmt.Sprintf(
						"Reply: %s: runtime error: %s",
						p.execReplyNode.callString,
						v,
					),
					GetStackString(1),
				)
			}
		}

		if p.execReplyNode != nil {
			p.execReplyNode.indicator.Count(
				time.Duration(TimeNowNS()-timeStart),
				p.from,
				p.execSuccessful,
			)
		}
		ctx.stop()
		inStream.Reset()
		retStream := p.outStream
		p.outStream = inStream
		p.from = ""
		p.execDepth = 0
		p.execReplyNode = nil
		p.execArgs = p.execArgs[:0]
		onEvalFinish(p, retStream, p.execSuccessful)
	}()

	// copy head
	copy(p.outStream.GetHeader(), inStream.GetHeader())

	// read reply path
	replyPath, ok := inStream.ReadUnsafeString()
	if !ok {
		return ctx.writeError("rpc data format error", "")
	}
	if p.execReplyNode, ok = p.processor.repliesMap[replyPath]; !ok {
		return ctx.writeError(
			fmt.Sprintf("rpc-server: reply path %s is not mounted", replyPath),
			"",
		)
	}

	// read depth
	if p.execDepth, ok = inStream.ReadUint64(); !ok {
		return ctx.writeError("rpc data format error", "")
	}
	if p.execDepth > p.processor.maxCallDepth {
		return ctx.Errorf(
			"rpc current call depth(%d) is overflow. limited(%d)",
			p.execDepth,
			p.processor.maxCallDepth,
		)
	}

	// read from
	if p.from, ok = inStream.ReadUnsafeString(); !ok {
		return ctx.writeError("rpc data format error", "")
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
				return ctx.writeError("rpc data format error", "")
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

		return ctx.writeError(
			fmt.Sprintf(
				"rpc reply arguments not match\nCalled: %s(%s) %s\nRequired: %s",
				replyPath,
				strings.Join(remoteArgsType, ", "),
				convertTypeToString(returnType),
				p.execReplyNode.callString,
			),
			p.execReplyNode.debugString,
		)
	}

	return nilReturn
}
