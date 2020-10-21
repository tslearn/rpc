package core

//
//import (
//	"errors"
//	"fmt"
//	"github.com/rpccloud/rpc/internal/base"
//	"io/ioutil"
//	"reflect"
//	"strings"
//	"sync/atomic"
//	"testing"
//	"time"
//)
//
//func TestRTPosRecord(t *testing.T) {
//	assert := base.NewAssert(t)
//
//	posArray := []int64{
//		0,
//		1,
//		1000,
//		0x00FFFFFFFFFFFFFE,
//		0x00FFFFFFFFFFFFFF,
//		0x7FFFFFFFFFFFFFFE,
//		0x7FFFFFFFFFFFFFFF,
//	}
//
//	flagArray := []bool{
//		false,
//		true,
//	}
//
//	for i := 0; i < len(posArray); i++ {
//		for j := 0; j < len(flagArray); j++ {
//			pos := posArray[i]
//			flag := flagArray[j]
//
//			record := makePosRecord(pos, flag)
//
//			fmt.Println("record", record)
//			assert(record.getPos()).Equal(pos)
//			assert(record.isString()).Equal(flag)
//
//			fmt.Println(record)
//		}
//	}
//}
//
//func BenchmarkRTPosRecord(b *testing.B) {
//	pos := int64(128)
//	flag := true
//	record := posRecord(0)
//
//	b.ReportAllocs()
//
//	for i := 0; i < b.N; i++ {
//		record = makePosRecord(pos, flag)
//		pos = record.getPos()
//		flag = record.isString()
//	}
//
//	fmt.Println(record)
//}
//
//func BenchmarkAddPrefixPerLine(b *testing.B) {
//	b.ReportAllocs()
//	for n := 0; n < b.N; n++ {
//		base.AddPrefixPerLine("a\nb\nc", "test")
//	}
//}
//
//func BenchmarkConcatString(b *testing.B) {
//	b.ReportAllocs()
//	for n := 0; n < b.N; n++ {
//		base.ConcatString("a", "b")
//	}
//}
//
//func BenchmarkGetCodePosition(b *testing.B) {
//	b.ReportAllocs()
//	b.RunParallel(func(pb *testing.PB) {
//		for pb.Next() {
//			base.AddFileLine("test", 0)
//		}
//	})
//}
//
//func BenchmarkGetRandString(b *testing.B) {
//	b.ReportAllocs()
//	for n := 0; n < b.N; n++ {
//		base.GetRandString(128)
//	}
//}
//
//func BenchmarkTimeNow(b *testing.B) {
//	b.ReportAllocs()
//	for n := 0; n < b.N; n++ {
//		base.TimeNow()
//	}
//}
//
//func BenchmarkTimeNowISOString(b *testing.B) {
//	b.ReportAllocs()
//	for n := 0; n < b.N; n++ {
//		base.TimeNowISOString()
//	}
//}

type testFuncCache struct{}

func (p *testFuncCache) Get(fnString string) ActionCacheFunc {
	switch fnString {
	case "":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime) Return)(rt)
				return true
			}
		}
	case "S":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, err := stream.ReadString(); err != nil {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, String) Return)(rt, arg0)
				return true
			}
		}
	case "I":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, err := stream.ReadInt64(); err != nil {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, Int64) Return)(rt, arg0)
				return true
			}
		}
	case "M":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, err := stream.ReadMap(); err != nil {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, Map) Return)(rt, arg0)
				return true
			}
		}
	case "V":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, err := stream.ReadRTValue(rt); err != nil {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, RTValue) Return)(rt, arg0)
				return true
			}
		}
	case "Y":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, err := stream.ReadRTArray(rt); err != nil {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, RTArray) Return)(rt, arg0)
				return true
			}
		}
	case "Z":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, err := stream.ReadRTMap(rt); err != nil {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, RTMap) Return)(rt, arg0)
				return true
			}
		}
	case "BIUFSXAM":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, err := stream.ReadBool(); err != nil {
				return false
			} else if arg1, err := stream.ReadInt64(); err != nil {
				return false
			} else if arg2, err := stream.ReadUint64(); err != nil {
				return false
			} else if arg3, err := stream.ReadFloat64(); err != nil {
				return false
			} else if arg4, err := stream.ReadString(); err != nil {
				return false
			} else if arg5, err := stream.ReadBytes(); err != nil {
				return false
			} else if arg6, err := stream.ReadArray(); err != nil {
				return false
			} else if arg7, err := stream.ReadMap(); err != nil {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(
					Runtime, Bool, Int64, Uint64,
					Float64, String, Bytes, Array, Map,
				) Return)(rt, arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
				return true
			}
		}
	default:
		return nil
	}
}

//func testReadFromFile(filePath string) (string, base.Error) {
//	ret, err := ioutil.ReadFile(filePath)
//	if err != nil {
//		return "", base.NewKernelPanic(err.Error())
//	}
//
//	// for windows, remove \r
//	return strings.Replace(string(ret), "\r", "", -1), nil
//}
//
//func getFakeOnEvalBack() func(*Stream) {
//	return func(stream *Stream) {}
//}
//
//func getFakeOnEvalFinish() func(*rpcThread) {
//	return func(thread *rpcThread) {}
//}
//
//func getFakeProcessor(debug bool) *Processor {
//	processor := NewProcessor(
//		debug,
//		1024,
//		32,
//		32,
//		nil,
//		5*time.Second,
//		nil,
//		func(stream *Stream) {},
//	)
//	processor.Close()
//	return processor
//}
//
//func getFakeThread(debug bool) *rpcThread {
//	return newThread(
//		getFakeProcessor(debug),
//		5*time.Second,
//		getFakeOnEvalBack(),
//		getFakeOnEvalFinish(),
//	)
//}
//
//func testRunWithProcessor(
//	isDebug bool,
//	fnCache ActionCache,
//	handler interface{},
//	getStream func(processor *Processor) *Stream,
//	onTest func(processor *Processor),
//) (ret interface{}, retError base.Error, retPanic base.Error) {
//	helper := newTestProcessorReturnHelper()
//	service := NewService().On("Eval", handler)
//
//	if processor := NewProcessor(
//		isDebug,
//		1024,
//		16,
//		16,
//		fnCache,
//		5*time.Second,
//		[]*ServiceMeta{{
//			name:     "test",
//			service:  service,
//			fileLine: "",
//		}},
//		helper.GetFunction(),
//	); processor == nil {
//		panic("internal error")
//	} else if inStream := getStream(processor); inStream == nil {
//		panic("internal error")
//	} else {
//		processor.PutStream(inStream)
//		if onTest != nil {
//			onTest(processor)
//		}
//
//		helper.WaitForFirstStream()
//
//		if !processor.Close() {
//			panic("internal error")
//		}
//
//		retArray, errorArray, panicArray := helper.GetReturn()
//
//		if len(retArray) > 1 || len(errorArray) > 1 || len(panicArray) > 1 {
//			panic("internal error")
//		}
//
//		if len(retArray) == 1 {
//			ret = retArray[0]
//		}
//
//		if len(errorArray) == 1 {
//			retError = errorArray[0]
//		}
//
//		if len(panicArray) == 1 {
//			retPanic = panicArray[0]
//		}
//
//		return
//	}
//}
//
//func testRunOnContext(
//	isDebug bool,
//	fn func(processor *Processor, rt Runtime) Return,
//) (interface{}, base.Error, base.Error) {
//	processorCH := make(chan *Processor)
//	return testRunWithProcessor(
//		isDebug,
//		nil,
//		func(rt Runtime) Return {
//			return fn(<-processorCH, rt)
//		},
//		func(processor *Processor) *Stream {
//			stream := NewStream()
//			stream.SetDepth(3)
//			stream.WriteString("#.test:Eval")
//			stream.WriteString("")
//			return stream
//		},
//		func(processor *Processor) {
//			processorCH <- processor
//		},
//	)
//}
//
//type testProcessorReturnHelper struct {
//	streamCH       chan *Stream
//	firstReceiveCH chan bool
//	isFirst        int32
//}
//
//func newTestProcessorReturnHelper() *testProcessorReturnHelper {
//	return &testProcessorReturnHelper{
//		streamCH:       make(chan *Stream, 102400),
//		firstReceiveCH: make(chan bool, 1),
//		isFirst:        0,
//	}
//}
//
//func (p *testProcessorReturnHelper) GetFunction() func(stream *Stream) {
//	return func(stream *Stream) {
//		if atomic.CompareAndSwapInt32(&p.isFirst, 0, 1) {
//			p.firstReceiveCH <- true
//		}
//
//		stream.SetReadPosToBodyStart()
//		if kind, ok := stream.ReadUint64(); ok {
//			if kind == uint64(base.ErrorKindTransport) {
//				panic("it makes onEvalFinish panic")
//			}
//		}
//
//		select {
//		case p.streamCH <- stream:
//			return
//		case <-time.After(time.Second):
//			// prevent capture
//			go func() {
//				panic("streamCH is full")
//			}()
//		}
//	}
//}
//
//func (p *testProcessorReturnHelper) WaitForFirstStream() {
//	<-p.firstReceiveCH
//}
//
//func (p *testProcessorReturnHelper) GetReturn() ([]Any, []base.Error, []base.Error) {
//	retArray := make([]Any, 0)
//	errorArray := make([]base.Error, 0)
//	panicArray := make([]base.Error, 0)
//	reportPanic := func(message string) {
//		go func() {
//			panic("message")
//		}()
//	}
//	close(p.streamCH)
//	for stream := range p.streamCH {
//		stream.SetReadPosToBodyStart()
//		if kind, ok := stream.ReadUint64(); !ok {
//			reportPanic("stream is bad")
//		} else if base.ErrorKind(kind) == base.ErrorKindNone {
//			if v, ok := stream.Read(); ok {
//				retArray = append(retArray, v)
//			} else {
//				reportPanic("read value error")
//			}
//		} else {
//			if message, ok := stream.ReadString(); !ok {
//				reportPanic("read message error")
//			} else if debug, ok := stream.ReadString(); !ok {
//				reportPanic("read debug error")
//			} else {
//				err := base.NewError(base.ErrorKind(kind), message, debug)
//
//				switch base.ErrorKind(kind) {
//				case base.ErrorKindProtocol:
//					fallthrough
//				case base.ErrorKindTransport:
//					fallthrough
//				case base.ErrorKindAction:
//					errorArray = append(errorArray, err)
//				case base.ErrorKindActionPanic:
//					fallthrough
//				case base.ErrorKindRuntimePanic:
//					fallthrough
//				case base.ErrorKindKernelPanic:
//					panicArray = append(panicArray, err)
//				default:
//					reportPanic("kind error")
//				}
//			}
//		}
//		stream.Release()
//	}
//	return retArray, errorArray, panicArray
//}
