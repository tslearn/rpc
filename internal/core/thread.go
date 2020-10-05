package core

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
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

const rpcThreadCacheSize = 512

var rpcThreadFrameCache = &sync.Pool{
	New: func() interface{} {
		return &rpcThreadFrame{
			stream:     nil,
			depth:      0,
			replyNode:  nil,
			from:       "",
			retStatus:  0,
			lockStatus: 0,
		}
	},
}

type rpcThreadFrame struct {
	stream     *Stream
	replyNode  unsafe.Pointer
	from       string
	depth      uint16
	cachePos   uint16
	retStatus  uint32
	lockStatus uint64
	next       *rpcThreadFrame
}

func newRPCThreadFrame() *rpcThreadFrame {
	return rpcThreadFrameCache.Get().(*rpcThreadFrame)
}

func (p *rpcThreadFrame) Reset() {
	p.stream = nil
	p.cachePos = 0
	atomic.StorePointer(&p.replyNode, nil)
	p.from = ""
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
	rootFrame    rpcThreadFrame
	sequence     uint64
	rtStream     *Stream
	cache        [rpcThreadCacheSize]byte
	sync.Mutex
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
			sequence:     rand.Uint64() % (1 << 56) / 2 * 2,
			rtStream:     NewStream(),
		}

		thread.top = &thread.rootFrame

		retCH <- thread

		for stream := <-inputCH; stream != nil; stream = <-inputCH {
			thread.Eval(stream, onEvalBack)
			thread.rtStream.Reset()
			thread.rootFrame.Reset()
			onEvalFinish(thread)
		}

		closeCH <- true
	}()

	return <-retCH
}

func (p *rpcThread) Close() bool {
	p.Lock()
	defer p.Unlock()

	if chPtr := (*chan bool)(atomic.LoadPointer(&p.closeCH)); chPtr != nil {
		atomic.StorePointer(&p.closeCH, nil)
		time.Sleep(500 * time.Millisecond)
		close(p.inputCH)
		select {
		case <-*chPtr:
			p.rtStream.Release()
			p.rtStream = nil
			return true
		case <-time.After(p.closeTimeout):
			return false
		}
	}

	return false
}

func (p *rpcThread) malloc(numOfBytes int) unsafe.Pointer {
	if numOfBytes > 0 && rpcThreadCacheSize-int(p.top.cachePos) > numOfBytes {
		ret := unsafe.Pointer(&p.cache[p.top.cachePos])
		p.top.cachePos += uint16(numOfBytes)
		return ret
	} else {
		return nil
	}
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
	frame.cachePos = p.top.cachePos
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
		return base.ConcatString(node.path, " ", node.meta.fileLine)
	}
	return ""
}

func (p *rpcThread) WriteOK(value interface{}, skip uint) Return {
	if p.top.retStatus == 0 {
		stream := p.top.stream
		stream.SetWritePosToBodyStart()
		stream.WriteUint64(0)
		if reason := stream.Write(value); reason != StreamWriteOK {
			return p.WriteError(
				errors.ErrRuntimeArgumentNotSupported.
					AddDebug(base.ConcatString("value", reason)).
					AddDebug(base.AddFileLine(p.GetExecReplyNodePath(), skip+1)),
				skip+1,
			)
		}
		p.top.retStatus = 1
		return emptyReturn
	} else {
		return p.WriteError(nil, skip+1)
	}
}

func (p *rpcThread) WriteError(err *base.Error, skip uint) Return {
	stream := p.top.stream
	stream.SetWritePosToBodyStart()

	writeError := (*base.Error)(nil)
	if p.top.retStatus == 0 {
		writeError = err
	} else if p.top.retStatus == 1 {
		writeError = errors.ErrRuntimeOKHasBeenCalled.
			AddDebug(base.AddFileLine(p.GetExecReplyNodePath(), skip+1))
	} else {
		writeError = errors.ErrRuntimeErrorHasBeenCalled.
			AddDebug(base.AddFileLine(p.GetExecReplyNodePath(), skip+1))
	}

	stream.WriteUint64(writeError.GetCode())
	stream.WriteString(writeError.GetMessage())

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
) Return {
	timeStart := base.TimeNow()
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
				errors.ErrThreadRunReplyFatal.
					AddDebug(fmt.Sprintf("runtime error: %v", v)).
					AddDebug(p.GetExecReplyDebug()).AddDebug(string(debug.Stack())),
				0,
			)
		}

		defer func() {
			if v := recover(); v != nil {
				// kernel error
				base.PublishPanic(
					errors.ErrThreadEvalFatal.AddDebug(fmt.Sprintf("%v", v)).
						AddDebug(string(debug.Stack())),
				)
			}
		}()

		for !atomic.CompareAndSwapUint64(&frame.lockStatus, rtID, rtID+2) {
			time.Sleep(10 * time.Millisecond)
		}

		if frame.retStatus == 0 {
			p.WriteError(errors.ErrRuntimeExternalReturn.AddDebug(p.GetExecReplyDebug()), 0)
		} else {
			// count
			if execReplyNode != nil {
				execReplyNode.indicator.Count(
					base.TimeNow().Sub(timeStart),
					frame.retStatus == 1,
				)
			}
		}

		// callback
		inStream.SetReadPosToBodyStart()
		onEvalBack(inStream)
	}()

	// set exec reply node
	replyPath, _, ok := inStream.readUnsafeString()
	if !ok {
		return p.WriteError(errors.ErrBadStream, 0)
	} else if execReplyNode, ok = p.processor.repliesMap[replyPath]; !ok {
		return p.WriteError(
			errors.ErrThreadTargetNotExist.
				AddDebug(base.ConcatString("target ", replyPath, " does not exist")),
			0,
		)
	} else {
		atomic.StorePointer(&frame.replyNode, unsafe.Pointer(execReplyNode))
	}

	if frame.depth > p.processor.maxCallDepth {
		return p.WriteError(
			errors.ErrThreadCallDepthOverflow.
				AddDebug(base.ConcatString(
					"call ",
					replyPath,
					" level(",
					strconv.FormatUint(uint64(frame.depth), 10),
					") overflows",
				)).
				AddDebug(p.GetExecReplyDebug()),
			0,
		)
	} else if frame.from, _, ok = inStream.readUnsafeString(); !ok {
		return p.WriteError(errors.ErrBadStream, 0)
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
			args := make([]reflect.Value, 0, 8)
			args = append(args, reflect.ValueOf(rt))
			for i := 1; i < len(execReplyNode.argTypes); i++ {
				var rv reflect.Value

				switch execReplyNode.argTypes[i] {
				case int64Type:
					if v, success := inStream.ReadInt64(); success {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case uint64Type:
					if v, success := inStream.ReadUint64(); success {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case float64Type:
					if v, success := inStream.ReadFloat64(); success {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case boolType:
					if v, success := inStream.ReadBool(); success {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case stringType:
					if v, success := inStream.ReadString(); success {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case bytesType:
					if v, success := inStream.ReadBytes(); success {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case arrayType:
					if v, success := inStream.ReadArray(); success {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case rtArrayType:
					if v, success := inStream.ReadRTArray(rt); success {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case mapType:
					if v, success := inStream.ReadMap(); success {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case rtMapType:
					if v, success := inStream.ReadRTMap(rt); success {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case rtValueType:
					if v, success := inStream.ReadRTValue(rt); success {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				default:
					ok = false
				}

				if !ok {
					break
				} else {
					args = append(args, rv)
				}
			}

			if !inStream.IsReadFinish() {
				ok = false
			}

			if ok {
				inStream.SetWritePosToBodyStart()
				execReplyNode.reflectFn.Call(args)
				return emptyReturn
			}
		}

		if _, ok := inStream.Read(); !ok {
			return p.WriteError(errors.ErrBadStream, 0)
		} else if !p.processor.isDebug {
			return p.WriteError(
				errors.ErrThreadArgumentsNotMatch.
					AddDebug(base.ConcatString(
						replyPath,
						" reply arguments does not match",
					)),
				0,
			)
		} else {
			remoteArgsType := make([]string, 0)
			remoteArgsType = append(remoteArgsType, convertTypeToString(runtimeType))
			inStream.SetReadPos(argsStreamPos)
			for inStream.CanRead() {
				if val, ok := inStream.Read(); !ok {
					return p.WriteError(errors.ErrBadStream, 0)
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
				errors.ErrThreadArgumentsNotMatch.AddDebug(base.ConcatString(
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
