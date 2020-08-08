package internal

import (
	"io/ioutil"
	"sync/atomic"
	"time"
	"unsafe"
)

type testFuncCache struct{}

func (p *testFuncCache) Get(fnString string) ReplyCacheFunc {
	switch fnString {
	case "S":
		return func(ctx Context, stream *Stream, fn interface{}) bool {
			if arg0, ok := stream.ReadString(); !ok {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				fn.(func(Context, String) Return)(ctx, arg0)
				return true
			}
		}
	case "I":
		return func(ctx Context, stream *Stream, fn interface{}) bool {
			if arg0, ok := stream.ReadInt64(); !ok {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				fn.(func(Context, Int64) Return)(ctx, arg0)
				return true
			}
		}
	case "M":
		return func(ctx Context, stream *Stream, fn interface{}) bool {
			if arg0, ok := stream.ReadMap(); !ok {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				fn.(func(Context, Map) Return)(ctx, arg0)
				return true
			}
		}
	case "BIUFSXAM":
		return func(ctx Context, stream *Stream, fn interface{}) bool {
			if arg0, ok := stream.ReadBool(); !ok {
				return false
			} else if arg1, ok := stream.ReadInt64(); !ok {
				return false
			} else if arg2, ok := stream.ReadUint64(); !ok {
				return false
			} else if arg3, ok := stream.ReadFloat64(); !ok {
				return false
			} else if arg4, ok := stream.ReadString(); !ok {
				return false
			} else if arg5, ok := stream.ReadBytes(); !ok {
				return false
			} else if arg6, ok := stream.ReadArray(); !ok {
				return false
			} else if arg7, ok := stream.ReadMap(); !ok {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				fn.(func(
					Context, Bool, Int64, Uint64,
					Float64, String, Bytes, Array, Map,
				) Return)(ctx, arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
				return true
			}
		}
	default:
		return nil
	}
}

func testReadFromFile(filePath string) (string, Error) {
	ret, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", NewKernelPanic(err.Error())
	}
	return string(ret), nil
}

func getFakeOnEvalBack() func(*Stream) {
	return func(stream *Stream) {}
}

func getFakeOnEvalFinish() func(*rpcThread) {
	return func(thread *rpcThread) {}
}

func getFakeProcessor(debug bool) *Processor {
	processor := NewProcessor(
		debug,
		1024,
		32,
		32,
		nil,
		5*time.Second,
		nil,
		func(stream *Stream) {},
	)
	processor.Close()
	return processor
}

func getFakeThread(debug bool) *rpcThread {
	return newThread(
		getFakeProcessor(debug),
		5*time.Second,
		getFakeOnEvalBack(),
		getFakeOnEvalFinish(),
	)
}

func getFakeContext(debug bool) Context {
	return &ContextObject{thread: unsafe.Pointer(getFakeThread(debug))}
}

func testRunWithSubscribePanic(fn func()) Error {
	ch := make(chan Error, 1)
	sub := subscribePanic(func(err Error) {
		ch <- err
	})
	defer sub.Close()

	fn()

	select {
	case err := <-ch:
		return err
	default:
		return nil
	}
}

func testRunWithCatchPanic(fn func()) (ret interface{}) {
	defer func() {
		ret = recover()
	}()

	fn()
	return
}

func testRunWithProcessor(
	isDebug bool,
	fnCache ReplyCache,
	handler interface{},
	getStream func(processor *Processor) *Stream,
) (ret interface{}, retError Error, retPanic Error) {
	helper := newTestProcessorReturnHelper()
	service := NewService().Reply("Eval", handler)

	if processor := NewProcessor(
		isDebug,
		1024,
		16,
		16,
		fnCache,
		5*time.Second,
		[]*ServiceMeta{{
			name:     "test",
			service:  service,
			fileLine: "",
		}},
		helper.GetFunction(),
	); processor == nil {
		panic("internal error")
	} else if inStream := getStream(processor); inStream == nil {
		panic("internal error")
	} else {
		processor.PutStream(inStream)

		helper.WaitForFirstStream()

		if !processor.Close() {
			panic("internal error")
		}

		retArray, errorArray, panicArray := helper.GetReturn()

		if len(retArray) > 1 || len(errorArray) > 1 || len(panicArray) > 1 {
			panic("internal error")
		}

		if len(retArray) == 1 {
			ret = retArray[0]
		}

		if len(errorArray) == 1 {
			retError = errorArray[0]
		}

		if len(panicArray) == 1 {
			retPanic = panicArray[0]
		}

		return
	}
}

func testRunOnContext(
	isDebug bool,
	fn func(processor *Processor, ctx Context) Return,
) (interface{}, Error, Error) {
	callProcessor := (*Processor)(nil)
	return testRunWithProcessor(
		isDebug,
		nil,
		func(ctx Context) Return {
			return fn(callProcessor, ctx)
		},
		func(processor *Processor) *Stream {
			callProcessor = processor
			stream := NewStream()
			stream.WriteString("#.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("")
			return stream
		},
	)
}

type testProcessorReturnHelper struct {
	streamCH       chan *Stream
	firstReceiveCH chan bool
	isFirst        int32
}

func newTestProcessorReturnHelper() *testProcessorReturnHelper {
	return &testProcessorReturnHelper{
		streamCH:       make(chan *Stream, 102400),
		firstReceiveCH: make(chan bool, 1),
		isFirst:        0,
	}
}

func (p *testProcessorReturnHelper) GetFunction() func(stream *Stream) {
	return func(stream *Stream) {
		if atomic.CompareAndSwapInt32(&p.isFirst, 0, 1) {
			p.firstReceiveCH <- true
		}

		stream.SetReadPosToBodyStart()
		if kind, ok := stream.ReadUint64(); ok {
			if kind == uint64(ErrorKindTransport) {
				panic("it makes onEvalFinish panic")
			}
		}

		select {
		case p.streamCH <- stream:
			return
		case <-time.After(time.Second):
			// prevent capture
			go func() {
				panic("streamCH is full")
			}()
		}
	}
}

func (p *testProcessorReturnHelper) WaitForFirstStream() {
	<-p.firstReceiveCH
}

func (p *testProcessorReturnHelper) GetReturn() ([]Any, []Error, []Error) {
	retArray := make([]Any, 0)
	errorArray := make([]Error, 0)
	panicArray := make([]Error, 0)
	reportPanic := func(message string) {
		go func() {
			panic("message")
		}()
	}
	close(p.streamCH)
	for stream := range p.streamCH {
		stream.SetReadPosToBodyStart()
		if kind, ok := stream.ReadUint64(); !ok {
			reportPanic("stream is bad")
		} else if ErrorKind(kind) == ErrorKindNone {
			if v, ok := stream.Read(); ok {
				retArray = append(retArray, v)
			} else {
				reportPanic("read value error")
			}
		} else {
			if message, ok := stream.ReadString(); !ok {
				reportPanic("read message error")
			} else if debug, ok := stream.ReadString(); !ok {
				reportPanic("read debug error")
			} else {
				err := NewError(ErrorKind(kind), message, debug)

				switch ErrorKind(kind) {
				case ErrorKindProtocol:
					fallthrough
				case ErrorKindTransport:
					fallthrough
				case ErrorKindReply:
					errorArray = append(errorArray, err)
				case ErrorKindReplyPanic:
					fallthrough
				case ErrorKindRuntimePanic:
					fallthrough
				case ErrorKindKernelPanic:
					panicArray = append(panicArray, err)
				default:
					reportPanic("kind error")
				}
			}
		}
		stream.Release()
	}
	return retArray, errorArray, panicArray
}
