package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"reflect"
	"testing"
	"time"
)

func TestGetFuncKind(t *testing.T) {
	assert := base.NewAssert(t)

	t.Run("fn not func", func(t *testing.T) {
		v := 3
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrActionHandler.AddDebug(
				"handler must be a function",
			))
	})

	t.Run("fn arguments length is zero", func(t *testing.T) {
		v := func() {}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrActionHandler.AddDebug(
				"handler 1st argument type must be rpc.Runtime",
			))
	})

	t.Run("fn 1st argument is not rpc.Runtime", func(t *testing.T) {
		v := func(_ chan bool) {}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrActionHandler.AddDebug(
				"handler 1st argument type must be rpc.Runtime",
			))
	})

	t.Run("fn without return", func(t *testing.T) {
		v := func(rt Runtime) {}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrActionHandler.AddDebug(
				"handler return type must be rpc.Return",
			))
	})

	t.Run("fn return multiply value", func(t *testing.T) {
		v := func(rt Runtime) (Return, bool) { return emptyReturn, true }
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrActionHandler.AddDebug(
				"handler return type must be rpc.Return",
			))
	})

	t.Run("fn return type not supported", func(t *testing.T) {
		v := func(rt Runtime, _ bool) bool { return true }
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrActionHandler.AddDebug(
				"handler return type must be rpc.Return",
			))
	})

	t.Run("2nd argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ int32, _ int64, _ uint64, _ float64, _ string, _ Bytes,
			_ Array, _ Map, _ RTValue, _ RTArray, _ RTMap,
		) Return {
			return rt.Reply(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrActionHandler.AddDebug(
				"handler 2nd argument type int32 is not supported",
			))
	})

	t.Run("3rd argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int32, _ uint64, _ float64, _ string, _ Bytes,
			_ Array, _ Map, _ RTValue, _ RTArray, _ RTMap,
		) Return {
			return rt.Reply(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrActionHandler.AddDebug(
				"handler 3rd argument type int32 is not supported",
			))
	})

	t.Run("4th argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ int32, _ float64, _ string, _ Bytes,
			_ Array, _ Map, _ RTValue, _ RTArray, _ RTMap,
		) Return {
			return rt.Reply(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrActionHandler.AddDebug(
				"handler 4th argument type int32 is not supported",
			))
	})

	t.Run("5th argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ uint64, _ int32, _ string, _ Bytes,
			_ Array, _ Map, _ RTValue, _ RTArray, _ RTMap,
		) Return {
			return rt.Reply(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrActionHandler.AddDebug(
				"handler 5th argument type int32 is not supported",
			))
	})

	t.Run("6th argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ uint64, _ float64, _ int32, _ Bytes,
			_ Array, _ Map, _ RTValue, _ RTArray, _ RTMap,
		) Return {
			return rt.Reply(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrActionHandler.AddDebug(
				"handler 6th argument type int32 is not supported",
			))
	})

	t.Run("7th argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ uint64, _ float64, _ string, _ int32,
			_ Array, _ Map, _ RTValue, _ RTArray, _ RTMap,
		) Return {
			return rt.Reply(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrActionHandler.AddDebug(
				"handler 7th argument type int32 is not supported",
			))
	})

	t.Run("8th argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ uint64, _ float64, _ string, _ Bytes,
			_ int32, _ Map, _ RTValue, _ RTArray, _ RTMap,
		) Return {
			return rt.Reply(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrActionHandler.AddDebug(
				"handler 8th argument type int32 is not supported",
			))
	})

	t.Run("9th argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ uint64, _ float64, _ string, _ Bytes,
			_ Array, _ int32, _ RTValue, _ RTArray, _ RTMap,
		) Return {
			return rt.Reply(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrActionHandler.AddDebug(
				"handler 9th argument type int32 is not supported",
			))
	})

	t.Run("10th argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ uint64, _ float64, _ string, _ Bytes,
			_ Array, _ Map, _ int32, _ RTArray, _ RTMap,
		) Return {
			return rt.Reply(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrActionHandler.AddDebug(
				"handler 10th argument type int32 is not supported",
			))
	})

	t.Run("11th argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ uint64, _ float64, _ string, _ Bytes,
			_ Array, _ Map, _ RTValue, _ int32, _ RTMap,
		) Return {
			return rt.Reply(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrActionHandler.AddDebug(
				"handler 11th argument type int32 is not supported",
			))
	})

	t.Run("12th argument unsupported", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ uint64, _ float64, _ string, _ Bytes,
			_ Array, _ Map, _ RTValue, _ RTArray, _ int32,
		) Return {
			return rt.Reply(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			errors.ErrActionHandler.AddDebug(
				"handler 12th argument type int32 is not supported",
			))
	})

	t.Run("test ok", func(t *testing.T) {
		v := func(rt Runtime,
			_ bool, _ int64, _ uint64, _ float64, _ string, _ Bytes,
			_ Array, _ Map, _ RTValue, _ RTArray, _ RTMap,
		) Return {
			return rt.Reply(true)
		}
		assert(getFuncKind(reflect.ValueOf(v))).Equal("BIUFSXAMVYZ", nil)
	})
}

func TestConvertTypeToString(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(convertTypeToString(nil)).Equal("<nil>")
		assert(convertTypeToString(runtimeType)).Equal("rpc.Runtime")
		assert(convertTypeToString(returnType)).Equal("rpc.Return")
		assert(convertTypeToString(boolType)).Equal("rpc.Bool")
		assert(convertTypeToString(int64Type)).Equal("rpc.Int64")
		assert(convertTypeToString(uint64Type)).Equal("rpc.Uint64")
		assert(convertTypeToString(float64Type)).Equal("rpc.Float64")
		assert(convertTypeToString(stringType)).Equal("rpc.String")
		assert(convertTypeToString(bytesType)).Equal("rpc.Bytes")
		assert(convertTypeToString(arrayType)).Equal("rpc.Array")
		assert(convertTypeToString(mapType)).Equal("rpc.Map")
		assert(convertTypeToString(rtValueType)).Equal("rpc.RTValue")
		assert(convertTypeToString(rtArrayType)).Equal("rpc.RTArray")
		assert(convertTypeToString(rtMapType)).Equal("rpc.RTMap")
		assert(convertTypeToString(reflect.ValueOf(make(chan bool)).Type())).
			Equal("chan bool")
	})
}

func TestMakeRequestStream(t *testing.T) {
	t.Run("write error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(MakeRequestStream("#", "", make(chan bool))).Equal(
			nil,
			errors.ErrUnsupportedValue.AddDebug(
				"2nd argument: value is not supported",
			),
		)
	})

	t.Run("arguments is empty", func(t *testing.T) {
		assert := base.NewAssert(t)
		v, err := MakeRequestStream("#", "from")
		assert(v).IsNotNil()
		assert(err).IsNil()
		assert(v.ReadString()).Equal("#", nil)
		assert(v.ReadString()).Equal("from", nil)
		assert(v.IsReadFinish()).IsTrue()
		v.Release()
	})

	t.Run("arguments is not empty", func(t *testing.T) {
		assert := base.NewAssert(t)
		v, err := MakeRequestStream("#", "from", false)
		assert(v).IsNotNil()
		assert(err).IsNil()
		assert(v.ReadString()).Equal("#", nil)
		assert(v.ReadString()).Equal("from", nil)
		assert(v.ReadBool()).Equal(false, nil)
		assert(v.IsReadFinish()).IsTrue()
		v.Release()
	})
}

func TestParseResponseStream(t *testing.T) {
	t.Run("errCode format error", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewStream()
		v.WriteInt64(3)
		assert(ParseResponseStream(v)).Equal(nil, errors.ErrStream)
	})

	t.Run("Read ret error", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewStream()
		v.WriteUint64(0)
		assert(ParseResponseStream(v)).Equal(false, errors.ErrStream)
	})

	t.Run("Read ret ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewStream()
		v.WriteUint64(0)
		v.WriteBool(true)
		assert(ParseResponseStream(v)).Equal(true, nil)
	})

	t.Run("error message Read error", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewStream()
		v.WriteUint64(uint64(base.ErrorTypeSecurity))
		v.WriteBool(true)
		assert(ParseResponseStream(v)).Equal(nil, errors.ErrStream)
	})

	t.Run("error stream is not finish", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewStream()
		v.WriteUint64(errors.ErrStream.GetCode())
		v.WriteString(errors.ErrStream.GetMessage())
		v.WriteBool(true)
		assert(ParseResponseStream(v)).Equal(nil, errors.ErrStream)
	})

	t.Run("error stream ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewStream()
		v.WriteUint64(errors.ErrStream.GetCode())
		v.WriteString(errors.ErrStream.GetMessage())
		assert(ParseResponseStream(v)).Equal(nil, errors.ErrStream)
	})
}

func getFakeProcessor(debug bool) *Processor {
	processor, _ := NewProcessor(
		debug,
		1,
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
		func(stream *Stream) {},
		func(thread *rpcThread) {},
	)
}

type testProcessorHelper struct {
	streamCH  chan *Stream
	processor *Processor
}

func newTestProcessorHelper(
	isDebug bool,
	numOfThreads int,
	maxNodeDepth int16,
	maxCallDepth int16,
	fnCache ActionCache,
	closeTimeout time.Duration,
	mountServices []*ServiceMeta,
) *testProcessorHelper {
	streamCH := make(chan *Stream, 102400)
	fnOnReturnStream := func(stream *Stream) {
		select {
		case streamCH <- stream:
			return
		case <-time.After(time.Second):
			// prevent capture
			go func() {
				panic("streamCH is full")
			}()
		}
	}

	processor, _ := NewProcessor(
		isDebug,
		numOfThreads,
		maxNodeDepth,
		maxCallDepth,
		fnCache,
		closeTimeout,
		mountServices,
		fnOnReturnStream,
	)

	return &testProcessorHelper{
		streamCH:  streamCH,
		processor: processor,
	}
}

func (p *testProcessorHelper) GetStream() *Stream {
	return <-p.streamCH
}

func (p *testProcessorHelper) GetProcessor() *Processor {
	return p.processor
}

func (p *testProcessorHelper) Close() {
	if p.processor != nil {
		p.processor.Close()
		p.processor = nil
	}
}

func testWithProcessorAndRuntime(
	isDebug bool,
	fn func(processor *Processor, rt Runtime) Return,
) *Stream {
	helper := (*testProcessorHelper)(nil)
	helper = newTestProcessorHelper(
		isDebug,
		1,
		16,
		16,
		nil,
		5*time.Second,
		[]*ServiceMeta{{
			name: "test",
			service: NewService().
				On("Eval", func(rt Runtime) Return {
					return fn(helper.processor, rt)
				}).
				On("SayHello", func(rt Runtime, name string) Return {
					return rt.Reply("hello " + name)
				}),
			fileLine: "",
		}},
	)
	defer helper.Close()

	stream, _ := MakeRequestStream("#.test:Eval", "")
	helper.GetProcessor().PutStream(stream)
	return <-helper.streamCH
}
