package internal

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	rpcThreadReturnStatusOK            = 0
	rpcThreadReturnStatusAlreadyCalled = 1
	rpcThreadReturnStatusContextError  = 2
)

var rpcThreadFrameCache = &sync.Pool{
	New: func() interface{} {
		return &rpcThreadFrame{
			stream:    nil,
			depth:     0,
			replyNode: nil,
			args:      make([]reflect.Value, 0, 8),
			from:      "",
			ok:        false,
		}
	},
}

type rpcThreadFrame struct {
	stream    *Stream
	depth     uint64
	replyNode unsafe.Pointer
	args      []reflect.Value
	from      string
	ok        bool
	status    uint64
	next      *rpcThreadFrame
}

func newRPCThreadFrame() *rpcThreadFrame {
	return rpcThreadFrameCache.Get().(*rpcThreadFrame)
}

func (p *rpcThreadFrame) Reset() {
	p.stream = nil
	atomic.StorePointer(&p.replyNode, nil)
	p.from = ""
	p.args = p.args[:0]
}

func (p *rpcThreadFrame) Release() {
	p.Reset()
	p.next = nil
	rpcThreadFrameCache.Put(p)
}

type rpcThread struct {
	processor    *Processor
	inputCH      chan *Stream
	closeCH      unsafe.Pointer
	closeTimeout time.Duration
	top          *rpcThreadFrame
	sequence     uint64
	Lock
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
	if timeout < 3*time.Second {
		timeout = 3 * time.Second
	}

	inputCH := make(chan *Stream)
	closeCH := make(chan bool, 1)
	go func() {
		thread := &rpcThread{
			processor:    processor,
			inputCH:      inputCH,
			closeCH:      unsafe.Pointer(&closeCH),
			closeTimeout: timeout,
			top:          newRPCThreadFrame(),
			sequence:     rand.Uint64() % (1 << 56) / 2 * 2,
		}

		retCH <- thread

		for stream := <-inputCH; stream != nil; stream = <-inputCH {
			thread.Eval(stream, onEvalBack, onEvalFinish)
		}

		closeCH <- true
	}()

	return <-retCH
}

func (p *rpcThread) Close() bool {
	return p.CallWithLock(func() interface{} {
		if chPtr := (*chan bool)(atomic.LoadPointer(&p.closeCH)); chPtr != nil {
			atomic.StorePointer(&p.closeCH, nil)
			time.Sleep(500 * time.Millisecond)
			close(p.inputCH)
			select {
			case <-*chPtr:
				p.top.Release()
				p.top = nil
				return true
			case <-time.After(p.closeTimeout):
				return false
			}
		}

		return false
	}).(bool)
}

func (p *rpcThread) pushFrame(rtID uint64) bool {
	if status := atomic.LoadUint64(&p.top.status); (status/2)*2 == rtID {
		frame := newRPCThreadFrame()
		frame.next = p.top
		p.top = frame
		return true
	} else {
		return false
	}
}

func (p *rpcThread) popFrame() bool {
	if frame := p.top.next; frame != nil {
		p.top.Release()
		p.top = frame
		return true
	} else {
		return false
	}
}

func (p *rpcThread) setReturn(contextID uint64) int {
	frame := p.top
	if atomic.CompareAndSwapUint64(&frame.status, contextID, contextID+1) {
		return rpcThreadReturnStatusOK
	} else if delta := atomic.LoadUint64(&frame.status) - contextID; delta == 1 {
		return rpcThreadReturnStatusAlreadyCalled
	} else {
		return rpcThreadReturnStatusContextError
	}
}

func (p *rpcThread) GetReplyNode() *rpcReplyNode {
	return (*rpcReplyNode)(atomic.LoadPointer(&p.top.replyNode))
}

func (p *rpcThread) GetExecReplyNodePath() string {
	if node := p.GetReplyNode(); node != nil {
		return node.path
	}
	return ""
}

func (p *rpcThread) GetExecReplyDebug() string {
	if node := p.GetReplyNode(); node != nil && node.meta != nil {
		return ConcatString(node.path, " ", node.meta.fileLine)
	}
	return ""
}

func (p *rpcThread) returnError(rtID uint64, err Error) Return {
	atomic.StoreUint64(&p.top.status, rtID+1)
	return p.WriteError(err)
}

func (p *rpcThread) WriteError(err Error) Return {
	stream := p.top.stream
	stream.SetWritePosToBodyStart()
	stream.WriteUint64(uint64(err.GetKind()))
	stream.WriteString(err.GetMessage())
	stream.WriteString(err.GetDebug())
	p.top.ok = false
	return emptyReturn
}

func (p *rpcThread) PutStream(stream *Stream) (ret bool) {
	defer func() {
		if v := recover(); v != nil {
			ret = false
		}
	}()
	if atomic.LoadPointer(&p.closeCH) != nil {
		p.inputCH <- stream
		return true
	}
	return false
}

func (p *rpcThread) Eval(
	inStream *Stream,
	onEvalBack func(*Stream),
	onEvalFinish func(*rpcThread),
) Return {
	timeStart := TimeNow()
	frame := p.top
	frame.ok = true
	frame.stream = inStream
	p.sequence += 2
	rtID := p.sequence
	frame.status = rtID
	execReplyNode := (*rpcReplyNode)(nil)

	defer func() {
		if v := recover(); v != nil {
			// write runtime error
			p.returnError(
				rtID,
				NewReplyPanic(
					fmt.Sprintf("runtime error: %v", v),
				).AddDebug(p.GetExecReplyDebug()).AddDebug(string(debug.Stack())),
			)
		}

		func() {
			defer func() {
				if v := recover(); v != nil {
					// kernel error
					reportPanic(
						NewKernelPanic(fmt.Sprintf("kernel error: %v", v)).
							AddDebug(string(debug.Stack())),
					)
				}

				frame.Reset()
				onEvalFinish(p)
			}()

			if atomic.CompareAndSwapUint64(&frame.status, rtID+1, rtID+2) {
				// return ok
			} else if atomic.CompareAndSwapUint64(&frame.status, rtID, rtID+2) {
				// Runtime.OK or Runtime.Error not called
				p.WriteError(
					NewReplyPanic(
						"reply must return through Runtime.OK or Runtime.Error",
					).AddDebug(p.GetExecReplyDebug()),
				)
			} else {
				// code should not run here
				p.WriteError(
					NewKernelPanic("internal error").
						AddDebug(p.GetExecReplyDebug()).AddDebug(string(debug.Stack())),
				)
			}

			// callback
			inStream.SetReadPosToBodyStart()
			onEvalBack(inStream)

			// count
			if execReplyNode != nil {
				execReplyNode.indicator.Count(TimeNow().Sub(timeStart), frame.ok)
			}
		}()
	}()

	// set exec reply node
	replyPath, ok := inStream.ReadUnsafeString()
	if !ok {
		return p.returnError(rtID, NewProtocolError(ErrStringBadStream))
	} else if execReplyNode, ok = p.processor.repliesMap[replyPath]; !ok {
		return p.returnError(
			rtID,
			NewReplyError(ConcatString("target ", replyPath, " does not exist")),
		)
	} else {
		atomic.StorePointer(&frame.replyNode, unsafe.Pointer(execReplyNode))
	}

	if frame.depth, ok = inStream.ReadUint64(); !ok {
		return p.returnError(rtID, NewProtocolError(ErrStringBadStream))
	} else if frame.depth > p.processor.maxCallDepth {
		return p.returnError(
			rtID,
			NewReplyError(ConcatString(
				"call ",
				replyPath,
				" level(",
				strconv.FormatUint(frame.depth, 10),
				") overflows",
			)).AddDebug(p.GetExecReplyDebug()),
		)
	} else if frame.from, ok = inStream.ReadUnsafeString(); !ok {
		return p.returnError(rtID, NewProtocolError(ErrStringBadStream))
	} else {
		// create context
		rt := Runtime{id: rtID, thread: p}
		// save argsPos
		argsStreamPos := inStream.GetReadPos()

		if fnCache := execReplyNode.cacheFN; fnCache != nil {
			ok = fnCache(rt, inStream, execReplyNode.meta.handler)
			if ok {
				return emptyReturn
			}
		} else {
			frame.args = append(frame.args, reflect.ValueOf(rt))
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
					frame.args = append(frame.args, rv)
				}
			}

			if !inStream.IsReadFinish() {
				ok = false
			}

			if ok {
				inStream.SetWritePosToBodyStart()
				execReplyNode.reflectFn.Call(frame.args)
				return emptyReturn
			}
		}

		if _, ok := inStream.Read(); !ok {
			return p.returnError(rtID, NewProtocolError(ErrStringBadStream))
		} else if !p.processor.isDebug {
			return p.returnError(
				rtID,
				NewReplyError(ConcatString(
					replyPath,
					" reply arguments does not match",
				)),
			)
		} else {
			remoteArgsType := make([]string, 0)
			remoteArgsType = append(remoteArgsType, convertTypeToString(contextType))
			inStream.setReadPosUnsafe(argsStreamPos)
			for inStream.CanRead() {
				if val, ok := inStream.Read(); !ok {
					return p.returnError(rtID, NewProtocolError(ErrStringBadStream))
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

			return p.returnError(
				rtID,
				NewReplyError(ConcatString(
					replyPath,
					" reply arguments does not match\nwant: ",
					execReplyNode.callString,
					"\ngot: ",
					replyPath,
					"(", strings.Join(remoteArgsType, ", "), ") ",
					convertTypeToString(returnType),
				)).AddDebug(p.GetExecReplyDebug()),
			)
		}
	}
}
