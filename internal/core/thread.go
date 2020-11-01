package core

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"math/rand"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	rpcThreadFrameCache = &sync.Pool{
		New: func() interface{} {
			return &rpcThreadFrame{
				stream:     nil,
				depth:      0,
				actionNode: nil,
				from:       "",
				retStatus:  0,
				lockStatus: 0,
			}
		},
	}
	zeroCacheEntry = [8]cacheEntry{}
)

type rpcThreadFrame struct {
	stream             *Stream
	actionNode         unsafe.Pointer
	from               string
	depth              uint16
	cacheArrayItemsPos uint32
	cacheMapItemsPos   uint32
	cacheArrayEntryPos uint32
	cacheMapEntryPos   uint32
	retStatus          uint32
	lockStatus         uint64
	parentRTWritePos   int
	next               *rpcThreadFrame
}

func newRPCThreadFrame() *rpcThreadFrame {
	return rpcThreadFrameCache.Get().(*rpcThreadFrame)
}

func (p *rpcThreadFrame) Reset() {
	p.stream = nil
	atomic.StorePointer(&p.actionNode, nil)
	p.from = ""
	p.cacheArrayItemsPos = 0
	p.cacheMapItemsPos = 0
	p.cacheArrayEntryPos = 0
	p.cacheMapEntryPos = 0
	p.parentRTWritePos = streamPosBody
	p.next = nil
}

func (p *rpcThreadFrame) Release() {
	p.Reset()
	rpcThreadFrameCache.Put(p)
}

type cacheEntry struct {
	arrayItems []posRecord
	mapItems   []mapItem
	mapLength  uint32
}

type rpcThread struct {
	processor       *Processor
	inputCH         chan *Stream
	closeCH         unsafe.Pointer
	closeTimeout    time.Duration
	top             *rpcThreadFrame
	rootFrame       rpcThreadFrame
	sequence        uint64
	rtStream        *Stream
	cacheEntry      [8]cacheEntry
	cacheArrayItems []posRecord
	cacheMapItems   []mapItem
	sync.Mutex
}

func newRTArray(rt Runtime, size int) (ret RTArray) {
	if thread := rt.thread; thread != nil && size >= 0 {
		ret.rt = rt
		frame := thread.top

		if frame.cacheArrayEntryPos < 8 {
			ret.items = &thread.cacheEntry[frame.cacheArrayEntryPos].arrayItems
			frame.cacheArrayEntryPos++
		} else {
			ret.items = new([]posRecord)
		}

		end := int(frame.cacheArrayItemsPos) + size
		if end <= len(thread.cacheArrayItems) {
			*ret.items = thread.cacheArrayItems[frame.cacheArrayItemsPos:]
			itemsHeader := (*reflect.SliceHeader)(unsafe.Pointer(ret.items))
			itemsHeader.Len = 0
			itemsHeader.Cap = size
			frame.cacheArrayItemsPos = uint32(end)
		} else {
			*ret.items = make([]posRecord, 0, size)
		}
	}

	return
}

func newRTMap(rt Runtime, size int) (ret RTMap) {
	if thread := rt.thread; thread != nil && size >= 0 {
		ret.rt = rt
		frame := thread.top

		if frame.cacheMapEntryPos < 8 {
			ret.items = &thread.cacheEntry[frame.cacheMapEntryPos].mapItems
			ret.length = &thread.cacheEntry[frame.cacheMapEntryPos].mapLength
			*ret.length = 0
			frame.cacheMapEntryPos++
		} else {
			ret.items = new([]mapItem)
			ret.length = new(uint32)
		}

		end := int(frame.cacheMapItemsPos) + size
		if end <= len(thread.cacheMapItems) {
			*ret.items = thread.cacheMapItems[frame.cacheMapItemsPos:]
			itemsHeader := (*reflect.SliceHeader)(unsafe.Pointer(ret.items))
			itemsHeader.Len = 0
			itemsHeader.Cap = size
			frame.cacheMapItemsPos = uint32(end)
		} else {
			*ret.items = make([]mapItem, 0, size)
		}
	}

	return
}

func newThread(
	processor *Processor,
	closeTimeout time.Duration,
	bufferSize uint32,
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
		numOfArray := int(bufferSize) / (2 * sizeOfPosRecord)
		numOfMap := int(bufferSize) / (2 * sizeOfMapItem)

		thread := &rpcThread{
			processor:       processor,
			inputCH:         inputCH,
			closeCH:         unsafe.Pointer(&closeCH),
			closeTimeout:    timeout,
			sequence:        1<<48 + (rand.Uint64()%(1<<56)/2)*2,
			rtStream:        NewStream(),
			cacheArrayItems: make([]posRecord, numOfArray, numOfArray),
			cacheMapItems:   make([]mapItem, numOfMap, numOfMap),
		}

		thread.top = &thread.rootFrame
		retCH <- thread

		for stream := <-inputCH; stream != nil; stream = <-inputCH {
			thread.Eval(stream, onEvalBack)
			onEvalFinish(thread)
			thread.Reset()
		}

		closeCH <- true
	}()

	return <-retCH
}

func (p *rpcThread) Reset() {
	p.rtStream.Reset()
	p.rootFrame.Reset()
	p.cacheEntry = zeroCacheEntry
}

func (p *rpcThread) Close() bool {
	p.Lock()
	defer p.Unlock()

	if chPtr := (*chan bool)(atomic.LoadPointer(&p.closeCH)); chPtr != nil {
		atomic.StorePointer(&p.closeCH, nil)
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

func (p *rpcThread) lock(rtID uint64) *rpcThread {
	lockPtr := &p.top.lockStatus
	for !atomic.CompareAndSwapUint64(lockPtr, rtID, rtID+1) {
		currStatus := atomic.LoadUint64(lockPtr)
		for currStatus == rtID || currStatus == rtID+1 {
			runtime.Gosched()
			if atomic.CompareAndSwapUint64(lockPtr, rtID, rtID+1) {
				return p
			}
		}

		return nil
	}

	return p
}

func (p *rpcThread) unlock(rtID uint64) {
	atomic.CompareAndSwapUint64(&p.top.lockStatus, rtID+1, rtID)
}

func (p *rpcThread) pushFrame() {
	frame := newRPCThreadFrame()
	frame.cacheArrayItemsPos = p.top.cacheArrayItemsPos
	frame.cacheMapItemsPos = p.top.cacheMapItemsPos
	frame.cacheArrayEntryPos = p.top.cacheArrayEntryPos
	frame.cacheMapEntryPos = p.top.cacheMapEntryPos
	frame.parentRTWritePos = p.rtStream.GetWritePos()
	frame.next = p.top
	p.top = frame
}

func (p *rpcThread) popFrame() {
	if frame := p.top.next; frame != nil {
		p.rtStream.SetWritePos(p.top.parentRTWritePos)
		p.top.Release()
		p.top = frame
	}
}

func (p *rpcThread) GetActionNode() *rpcActionNode {
	return (*rpcActionNode)(atomic.LoadPointer(&p.top.actionNode))
}

func (p *rpcThread) GetExecActionNodePath() string {
	if node := p.GetActionNode(); node != nil {
		return node.path
	}
	return ""
}

func (p *rpcThread) GetExecActionDebug() string {
	if node := p.GetActionNode(); node != nil && node.meta != nil {
		return base.ConcatString(node.path, " ", node.meta.fileLine)
	}
	return ""
}

func (p *rpcThread) Write(value interface{}, skip uint, debug bool) Return {
	frame := p.top

	if frame.retStatus != 0 {
		value = errors.ErrRuntimeReplyHasBeenCalled
	}

	stream := frame.stream
	stream.SetWritePosToBodyStart()
	stream.WriteUint64(0)

	if reason := stream.Write(value); reason == StreamWriteOK {
		frame.retStatus = 1
		return emptyReturn
	} else if err, ok := value.(*base.Error); ok {
		if err == nil {
			err = errors.ErrUnsupportedValue.AddDebug("value is nil")
		}
		if debug {
			err = err.AddDebug(base.AddFileLine(p.GetExecActionNodePath(), skip+1))
		}
		frame.retStatus = 2
		stream.SetWritePosToBodyStart()
		stream.WriteUint64(err.GetCode())
		stream.WriteString(err.GetMessage())
		return emptyReturn
	} else if rtValue, ok := value.(RTValue); ok && rtValue.err != nil {
		if debug {
			rtValue.err = rtValue.err.AddDebug(
				base.AddFileLine(p.GetExecActionNodePath(), skip+1),
			)
		}
		frame.retStatus = 2
		stream.SetWritePosToBodyStart()
		stream.WriteUint64(rtValue.err.GetCode())
		stream.WriteString(rtValue.err.GetMessage())
		return emptyReturn
	} else {
		frame.retStatus = 2
		err = errors.ErrUnsupportedValue.AddDebug(reason)
		if debug {
			err = err.AddDebug(
				base.AddFileLine(p.GetExecActionNodePath(), skip+1),
			)
		}
		stream.SetWritePosToBodyStart()
		stream.WriteUint64(err.GetCode())
		stream.WriteString(err.GetMessage())
		return emptyReturn
	}
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
	execActionNode := (*rpcActionNode)(nil)
	ok := true

	defer func() {
		if v := recover(); v != nil {
			// write runtime error
			p.Write(
				errors.ErrActionPanic.
					AddDebug(fmt.Sprintf("runtime error: %v", v)).
					AddDebug(p.GetExecActionDebug()).AddDebug(string(debug.Stack())),
				0,
				false,
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
			p.Write(
				errors.ErrRuntimeExternalReturn.AddDebug(p.GetExecActionDebug()),
				0,
				false,
			)
		} else {
			// count
			if execActionNode != nil {
				execActionNode.indicator.Count(
					base.TimeNow().Sub(timeStart),
					frame.retStatus == 1,
				)
			}
		}

		// callback
		inStream.SetReadPosToBodyStart()
		onEvalBack(inStream)
	}()

	// set exec action node
	actionPath, _, err := inStream.readUnsafeString()
	if err != nil {
		return p.Write(err, 0, false)
	} else if execActionNode, ok = p.processor.actionsMap[actionPath]; !ok {
		return p.Write(
			errors.ErrTargetNotExist.
				AddDebug(base.ConcatString("target ", actionPath, " does not exist")),
			0,
			false,
		)
	} else {
		atomic.StorePointer(&frame.actionNode, unsafe.Pointer(execActionNode))
	}

	if frame.depth >= p.processor.maxCallDepth {
		return p.Write(
			errors.ErrCallOverflow.
				AddDebug(base.ConcatString(
					"call ",
					actionPath,
					" level(",
					strconv.FormatUint(uint64(frame.depth), 10),
					") overflows",
				)).
				AddDebug(p.GetExecActionDebug()),
			0,
			false,
		)
	} else if frame.from, _, err = inStream.readUnsafeString(); err != nil {
		return p.Write(err, 0, false)
	} else {
		// create context
		rt := Runtime{id: rtID, thread: p}
		// save argsPos
		argsStreamPos := inStream.GetReadPos()

		if fnCache := execActionNode.cacheFN; fnCache != nil {
			ok = fnCache(rt, inStream, execActionNode.meta.handler)
			if ok {
				return emptyReturn
			}
		} else {
			args := make([]reflect.Value, 0, 8)
			args = append(args, reflect.ValueOf(rt))
			for i := 1; i < len(execActionNode.argTypes); i++ {
				var rv reflect.Value

				switch execActionNode.argTypes[i] {
				case int64Type:
					if v, err := inStream.ReadInt64(); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case uint64Type:
					if v, err := inStream.ReadUint64(); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case float64Type:
					if v, err := inStream.ReadFloat64(); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case boolType:
					if v, err := inStream.ReadBool(); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case stringType:
					if v, err := inStream.ReadString(); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case bytesType:
					if v, err := inStream.ReadBytes(); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case arrayType:
					if v, err := inStream.ReadArray(); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case rtArrayType:
					if v, err := inStream.ReadRTArray(rt); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case mapType:
					if v, err := inStream.ReadMap(); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case rtMapType:
					if v, err := inStream.ReadRTMap(rt); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						ok = false
					}
				case rtValueType:
					if v, err := inStream.ReadRTValue(rt); err == nil {
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
				execActionNode.reflectFn.Call(args)
				return emptyReturn
			}
		}

		if _, err := inStream.Read(); err != nil {
			return p.Write(err, 0, false)
		} else if !p.processor.isDebug {
			return p.Write(
				errors.ErrArgumentsNotMatch.
					AddDebug(base.ConcatString(
						actionPath,
						" action arguments does not match",
					)),
				0,
				false,
			)
		} else {
			remoteArgsType := make([]string, 0)
			remoteArgsType = append(remoteArgsType, convertTypeToString(runtimeType))
			inStream.SetReadPos(argsStreamPos)
			for inStream.CanRead() {
				if val, err := inStream.Read(); err != nil {
					return p.Write(err, 0, false)
				} else if val != nil {
					remoteArgsType = append(
						remoteArgsType,
						convertTypeToString(reflect.ValueOf(val).Type()),
					)
				} else {
					if len(remoteArgsType) < len(execActionNode.argTypes) {
						argType := execActionNode.argTypes[len(remoteArgsType)]
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

			return p.Write(
				errors.ErrArgumentsNotMatch.AddDebug(base.ConcatString(
					actionPath,
					" action arguments does not match\nwant: ",
					execActionNode.callString,
					"\ngot: ",
					actionPath,
					"(", strings.Join(remoteArgsType, ", "), ") ",
					convertTypeToString(returnType),
				)).AddDebug(p.GetExecActionDebug()),
				0,
				false,
			)
		}
	}
}
