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

var rpcThreadFrameCache = &sync.Pool{
	New: func() interface{} {
		return &rpcThreadFrame{
			stream:     nil,
			depth:      0,
			replyNode:  nil,
			args:       make([]reflect.Value, 0, 8),
			from:       "",
			retStatus:  0,
			lockStatus: 0,
		}
	},
}

type rpcThreadFrame struct {
	stream     *Stream
	depth      uint16
	replyNode  unsafe.Pointer
	args       []reflect.Value
	from       string
	retStatus  uint32
	lockStatus uint64
	next       *rpcThreadFrame
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

func (p *rpcThread) lock(rtID uint64) *rpcThread {
	for !atomic.CompareAndSwapUint64(&p.top.lockStatus, rtID, rtID+1) {
		currStatus := atomic.LoadUint64(&p.top.lockStatus)
		if currStatus == rtID || rtID == rtID+1 {
			time.Sleep(10 * time.Millisecond)
		} else {
			return nil
		}
	}

	return p
}

func (p *rpcThread) unlock(rtID uint64) {
	atomic.CompareAndSwapUint64(&p.top.lockStatus, rtID+1, rtID)
}

func (p *rpcThread) pushFrame() {
	frame := newRPCThreadFrame()
	frame.next = p.top
	p.top = frame
}

func (p *rpcThread) popFrame() {
	if frame := p.top.next; frame != nil {
		p.top.Release()
		p.top = frame
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

func (p *rpcThread) WriteOK(value interface{}, skip uint) Return {
	if p.top.retStatus == 0 {
		stream := p.top.stream
		stream.SetWritePosToBodyStart()
		stream.WriteUint64(uint64(ErrorKindNone))
		if reason := stream.Write(value); reason != StreamWriteOK {
			return p.WriteError(
				NewReplyPanic(ConcatString("value", reason)).
					AddDebug(AddFileLine(p.GetExecReplyNodePath(), skip+1)),
				skip+1,
			)
		}
		p.top.retStatus = 1
		return emptyReturn
	} else {
		return p.WriteError(nil, skip+1)
	}
}

func (p *rpcThread) WriteError(err Error, skip uint) Return {
	stream := p.top.stream
	stream.SetWritePosToBodyStart()

	if p.top.retStatus == 0 {
		stream.WriteUint64(uint64(err.GetKind()))
		stream.WriteString(err.GetMessage())
		stream.WriteString(err.GetDebug())
	} else if p.top.retStatus == 1 {
		stream.WriteUint64(uint64(ErrorKindReplyPanic))
		stream.WriteString("Runtime.OK has been called before")
		stream.WriteString(AddFileLine(p.GetExecReplyNodePath(), skip+1))
	} else {
		stream.WriteUint64(uint64(ErrorKindReplyPanic))
		stream.WriteString("Runtime.Error has been called before")
		stream.WriteString(AddFileLine(p.GetExecReplyNodePath(), skip+1))
	}

	p.top.retStatus = 2
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
	frame.stream = inStream
	p.sequence += 2
	rtID := p.sequence
	frame.lockStatus = rtID
	frame.retStatus = 0
	frame.depth = inStream.GetDepth()
	execReplyNode := (*rpcReplyNode)(nil)

	defer func() {
		if v := recover(); v != nil {
			// write runtime error
			p.WriteError(
				NewReplyPanic(
					fmt.Sprintf("runtime error: %v", v),
				).AddDebug(p.GetExecReplyDebug()).AddDebug(string(debug.Stack())),
				0,
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

			for !atomic.CompareAndSwapUint64(&frame.lockStatus, rtID, rtID+2) {
				time.Sleep(10 * time.Millisecond)
			}

			if frame.retStatus == 0 {
				p.WriteError(
					NewReplyPanic(
						"reply must return through Runtime.OK or Runtime.Error",
					).AddDebug(p.GetExecReplyDebug()),
					0,
				)
			} else {
				// count
				if execReplyNode != nil {
					execReplyNode.indicator.Count(
						TimeNow().Sub(timeStart),
						frame.retStatus == 1,
					)
				}
			}

			// callback
			inStream.SetReadPosToBodyStart()
			onEvalBack(inStream)
		}()
	}()

	// set exec reply node
	replyPath, ok := inStream.ReadUnsafeString()
	if !ok {
		return p.WriteError(NewProtocolError(ErrStringBadStream), 0)
	} else if execReplyNode, ok = p.processor.repliesMap[replyPath]; !ok {
		return p.WriteError(
			NewReplyError(ConcatString("target ", replyPath, " does not exist")),
			0,
		)
	} else {
		atomic.StorePointer(&frame.replyNode, unsafe.Pointer(execReplyNode))
	}

	if frame.depth > p.processor.maxCallDepth {
		return p.WriteError(
			NewReplyError(ConcatString(
				"call ",
				replyPath,
				" level(",
				strconv.FormatUint(uint64(frame.depth), 10),
				") overflows",
			)).AddDebug(p.GetExecReplyDebug()),
			0,
		)
	} else if frame.from, ok = inStream.ReadUnsafeString(); !ok {
		return p.WriteError(NewProtocolError(ErrStringBadStream), 0)
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
			return p.WriteError(NewProtocolError(ErrStringBadStream), 0)
		} else if !p.processor.isDebug {
			return p.WriteError(
				NewReplyError(ConcatString(
					replyPath,
					" reply arguments does not match",
				)),
				0,
			)
		} else {
			remoteArgsType := make([]string, 0)
			remoteArgsType = append(remoteArgsType, convertTypeToString(contextType))
			inStream.setReadPosUnsafe(argsStreamPos)
			for inStream.CanRead() {
				if val, ok := inStream.Read(); !ok {
					return p.WriteError(NewProtocolError(ErrStringBadStream), 0)
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
				)).AddDebug(p.GetExecReplyDebug()),
				0,
			)
		}
	}
}
