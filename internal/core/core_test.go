package core

import (
	"reflect"
	"testing"
	"time"

	"github.com/rpccloud/rpc/internal/base"
)

func TestGetFuncKind(t *testing.T) {
	assert := base.NewAssert(t)

	t.Run("fn not func", func(t *testing.T) {
		v := 3
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			base.ErrActionHandler.AddDebug(
				"handler must be a function",
			))
	})

	t.Run("fn arguments length is zero", func(t *testing.T) {
		v := func() {}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			base.ErrActionHandler.AddDebug(
				"handler 1st argument type must be rpc.Runtime",
			))
	})

	t.Run("fn 1st argument is not rpc.Runtime", func(t *testing.T) {
		v := func(_ chan bool) {}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			base.ErrActionHandler.AddDebug(
				"handler 1st argument type must be rpc.Runtime",
			))
	})

	t.Run("fn without return", func(t *testing.T) {
		v := func(rt Runtime) {}
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			base.ErrActionHandler.AddDebug(
				"handler return type must be rpc.Return",
			))
	})

	t.Run("fn return multiply value", func(t *testing.T) {
		v := func(rt Runtime) (Return, bool) { return emptyReturn, true }
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			base.ErrActionHandler.AddDebug(
				"handler return type must be rpc.Return",
			))
	})

	t.Run("fn return type not supported", func(t *testing.T) {
		v := func(rt Runtime, _ bool) bool { return true }
		assert(getFuncKind(reflect.ValueOf(v))).Equal(
			"",
			base.ErrActionHandler.AddDebug(
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
			base.ErrActionHandler.AddDebug(
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
			base.ErrActionHandler.AddDebug(
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
			base.ErrActionHandler.AddDebug(
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
			base.ErrActionHandler.AddDebug(
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
			base.ErrActionHandler.AddDebug(
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
			base.ErrActionHandler.AddDebug(
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
			base.ErrActionHandler.AddDebug(
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
			base.ErrActionHandler.AddDebug(
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
			base.ErrActionHandler.AddDebug(
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
			base.ErrActionHandler.AddDebug(
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
			base.ErrActionHandler.AddDebug(
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

func TestGetFastKey(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(getFastKey("")).Equal(uint32(0))
		assert(getFastKey("0")).Equal(uint32('0'<<16 | '0'<<8 | '0'))
		assert(getFastKey("01")).Equal(uint32('0'<<16 | '1'<<8 | '1'))
		assert(getFastKey("012")).Equal(uint32('0'<<16 | '1'<<8 | '2'))
		assert(getFastKey("0123")).Equal(uint32('0'<<16 | '2'<<8 | '3'))
		assert(getFastKey("01234")).Equal(uint32('0'<<16 | '2'<<8 | '4'))
	})
}

func TestMakeRequestStream(t *testing.T) {
	t.Run("write error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(MakeRequestStream(false, 0, "#", "", make(chan bool))).Equal(
			nil,
			base.ErrUnsupportedValue.AddDebug(
				"2nd argument: value type(chan bool) is not supported",
			),
		)
	})

	t.Run("arguments is empty", func(t *testing.T) {
		assert := base.NewAssert(t)
		v, err := MakeRequestStream(false, 2, "#", "from")
		assert(v).IsNotNil()
		assert(err).IsNil()
		assert(v.HasStatusBitDebug()).IsFalse()
		assert(v.GetDepth()).Equal(uint16(2))
		assert(v.ReadString()).Equal("#", nil)
		assert(v.ReadString()).Equal("from", nil)
		assert(v.IsReadFinish()).IsTrue()
		v.Release()
	})

	t.Run("arguments is not empty", func(t *testing.T) {
		assert := base.NewAssert(t)
		v, err := MakeRequestStream(true, 5, "#", "from", false)
		assert(v).IsNotNil()
		assert(err).IsNil()
		assert(v.HasStatusBitDebug()).IsTrue()
		assert(v.GetDepth()).Equal(uint16(5))
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
		assert(ParseResponseStream(v)).Equal(nil, base.ErrStream)
	})

	t.Run("Read ret error", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewStream()
		v.WriteUint64(0)
		assert(ParseResponseStream(v)).Equal(nil, base.ErrStream)
	})

	t.Run("Read ret ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewStream()
		v.SetKind(DataStreamResponseOK)
		v.WriteBool(true)
		assert(ParseResponseStream(v)).Equal(true, nil)
	})

	t.Run("error code overflows", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewStream()
		v.WriteUint64(1 << 32)
		v.WriteString(base.ErrStream.GetMessage())
		assert(ParseResponseStream(v)).Equal(nil, base.ErrStream)
	})

	t.Run("error message Read error", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewStream()
		v.WriteUint64(uint64(base.ErrorTypeSecurity))
		v.WriteBool(true)
		assert(ParseResponseStream(v)).Equal(nil, base.ErrStream)
	})

	t.Run("error stream is not finish", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewStream()
		v.WriteUint64(uint64(base.ErrStream.GetCode()))
		v.WriteString(base.ErrStream.GetMessage())
		v.WriteBool(true)
		assert(ParseResponseStream(v)).Equal(nil, base.ErrStream)
	})

	t.Run("error stream ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewStream()
		v.WriteUint64(uint64(base.ErrStream.GetCode()))
		v.WriteString(base.ErrStream.GetMessage())
		assert(ParseResponseStream(v)).Equal(nil, base.ErrStream)
	})
}

type testProcessorHelper struct {
	streamCH  chan *Stream
	processor *Processor
}

func newTestProcessorHelper(
	numOfThreads int,
	maxNodeDepth int16,
	maxCallDepth int16,
	threadBufferSize uint32,
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
		numOfThreads,
		maxNodeDepth,
		maxCallDepth,
		threadBufferSize,
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
	fn func(processor *Processor, rt Runtime) Return,
	data Map,
) *Stream {
	helper := (*testProcessorHelper)(nil)
	helper = newTestProcessorHelper(
		1,
		16,
		16,
		2048,
		nil,
		3*time.Second,
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
			data:     data,
		}},
	)
	defer helper.Close()

	stream, _ := MakeRequestStream(true, 0, "#.test:Eval", "")
	helper.GetProcessor().PutStream(stream)
	return <-helper.streamCH
}

func testReplyWithSource(
	debug bool,
	fnCache ActionCache,
	data Map,
	handler interface{},
	args ...interface{},
) (*Stream, string) {
	service, source := NewService().On("Eval", handler), base.GetFileLine(0)
	helper := newTestProcessorHelper(
		1,
		16,
		16,
		1024,
		fnCache,
		3*time.Second,
		[]*ServiceMeta{{
			name:     "test",
			service:  service,
			fileLine: "",
			data:     data,
		}},
	)
	defer helper.Close()

	if len(args) == 1 {
		if s, ok := args[0].(*Stream); ok {
			helper.GetProcessor().PutStream(s)
			return <-helper.streamCH, source
		}
	}
	stream, _ := MakeRequestStream(debug, 0, "#.test:Eval", "", args...)
	helper.GetProcessor().PutStream(stream)
	return <-helper.streamCH, source
}

func testReply(
	debug bool,
	fnCache ActionCache,
	data Map,
	handler interface{},
	args ...interface{},
) (Any, *base.Error) {
	retStream, _ := testReplyWithSource(debug, fnCache, data, handler, args...)
	return ParseResponseStream(retStream)
}
