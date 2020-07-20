package internal

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestNewThread(t *testing.T) {
	//assert := NewAssert(t)
	//
	//processor := NewProcessor(true, 8192, 16, 16, nil, nil)
	//threadPool := newThreadPool(processor, 8192)
	//
	//rpcThread := newThread(processor, processor,fre)
	//assert(rpcThread).IsNotNil()
	//assert(rpcThread.threadPool).Equals(threadPool)
	//assert(rpcThread.isRunning).Equals(true)
	//assert(rpcThread.threadPool.processor).Equals(processor)
	//assert(len(rpcThread.ch)).Equals(0)
	//assert(cap(rpcThread.ch)).Equals(0)
	//assert(rpcThread.inStream).IsNil()
	//assert(rpcThread.outStream).IsNotNil()
	//assert(rpcThread.execDepth).Equals(uint64(0))
	//assert(rpcThread.execReplyNode).IsNil()
	//assert(len(rpcThread.execArgs)).Equals(0)
	//assert(cap(rpcThread.execArgs)).Equals(16)
	//assert(rpcThread.execSuccessful).IsFalse()
	//assert(rpcThread.from).Equals("")
	//assert(len(rpcThread.closeCH)).Equals(0)
	//assert(cap(rpcThread.closeCH)).Equals(0)
}

func runWithProcessor(
	handler interface{},
	getStream func(processor *Processor) *Stream,
	onTest func(in *Stream, out *Stream, success bool),
) {
	//retStreamCH := make(chan *Stream)
	//retSuccessCH := make(chan bool)
	//processor := NewProcessor(
	//	16,
	//	16,
	//	func(stream *Stream, success bool) {
	//		retStreamCH <- stream
	//		retSuccessCH <- success
	//	},
	//	&testFuncCache{},
	//)
	//_ = processor.AddChild(
	//	"user",
	//	NewService().Reply("sayHello", true, handler),
	//	"",
	//)
	//
	//inStream := getStream(processor)
	//processor.Start()
	//processor.PutStream(inStream)
	//retStream := <-retStreamCH
	//retSuccess := <-retSuccessCH
	//onTest(inStream, retStream, retSuccess)
	//processor.Stop()
}

func TestRpcThread_eval(t *testing.T) {
	assert := NewAssert(t)

	// test basic
	runWithProcessor(
		func(ctx *ContextObject, name string) *ReturnObject {
			return ctx.Return("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.SetCallbackID(345343535345343535)
			stream.SetMachineID(345343535)
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).Equals(true)
			// inStream is reset
			assert(in.GetHeader()).Equals(out.GetHeader())
			assert(in.GetReadPos()).Equals(streamBodyPos)
			assert(in.GetWritePos()).Equals(streamBodyPos)
			assert(out.ReadBool()).Equals(true, true)
			assert(out.Read()).Equals("hello world", true)
			assert(out.CanRead()).IsFalse()
		},
	)

	// test read reply path error
	runWithProcessor(
		func(ctx *ContextObject, name string) *ReturnObject {
			return ctx.Return("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			// path format error
			stream.WriteBytes([]byte("$.user:sayHello"))
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).Equals(false)
			assert(out.ReadBool()).Equals(false, true)
			assert(out.Read()).Equals("rpc data format error", true)
			assert(out.Read()).Equals("", true)
			assert(out.CanRead()).IsFalse()
		},
	)

	// reply path is not mounted
	runWithProcessor(
		func(ctx *ContextObject, name string) *ReturnObject {
			return ctx.Return("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.system:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).Equals(false)
			assert(out.ReadBool()).Equals(false, true)
			assert(out.Read()).
				Equals("rpc-server: reply path $.system:sayHello is not mounted", true)
			assert(out.Read()).Equals("", true)
			assert(out.CanRead()).IsFalse()
		},
	)

	// depth data format error
	runWithProcessor(
		func(ctx *ContextObject, name string) *ReturnObject {
			return ctx.Return("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			// depth type error
			stream.WriteInt64(3)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).Equals(false)
			assert(out.ReadBool()).Equals(false, true)
			assert(out.Read()).Equals("rpc data format error", true)
			assert(out.Read()).Equals("", true)
			assert(out.CanRead()).IsFalse()
		},
	)

	// depth is overflow
	runWithProcessor(
		func(ctx *ContextObject, name string) *ReturnObject {
			return ctx.Return("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(17)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).Equals(false)
			assert(out.ReadBool()).Equals(false, true)
			assert(out.Read()).Equals(
				"rpc current call depth(17) is overflow. limited(16)",
				true,
			)
			dbgMessage, ok := out.Read()
			assert(ok).IsTrue()
			assert(strings.Contains(dbgMessage.(string), "$.user:sayHello")).IsTrue()
		},
	)

	// from data format error
	runWithProcessor(
		func(ctx *ContextObject, name string) *ReturnObject {
			return ctx.Return("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteBool(true)
			stream.WriteString("world")
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).Equals(false)
			assert(out.ReadBool()).Equals(false, true)
			assert(out.Read()).Equals("rpc data format error", true)
			assert(out.Read()).Equals("", true)
		},
	)

	// OK, call with all type value
	runWithProcessor(
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.Return(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(int64(3))
			stream.Write(uint64(3))
			stream.Write(float64(3))
			stream.Write("hello")
			stream.Write(([]byte)("world"))
			stream.Write(Array{1})
			stream.Write(Map{"name": "world"})
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).Equals(true)
			assert(out.ReadBool()).Equals(true, true)
			assert(out.Read()).Equals(true, true)
			assert(out.CanRead()).IsFalse()
		},
	)

	// error with 1st param
	runWithProcessor(
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.Return(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(3)
			stream.Write(int64(3))
			stream.Write(uint64(3))
			stream.Write(float64(3))
			stream.Write("hello")
			stream.Write(([]byte)("world"))
			stream.Write(Array{1})
			stream.Write(Map{"name": "world"})
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).IsFalse()
			assert(out.ReadBool()).Equals(false, true)
			assert(out.Read()).Equals("rpc reply arguments not match\n"+
				"Called: $.user:sayHello(rpc.ContextObject, rpc.Int64, rpc.Int64, rpc.Uint64, "+
				"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.ReturnObject\n"+
				"Required: $.user:sayHello(rpc.ContextObject, rpc.Bool, rpc.Int64, rpc.Uint64, "+
				"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.ReturnObject",
				true,
			)
			dbgMessage, ok := out.Read()
			assert(ok).IsTrue()
			assert(strings.Contains(dbgMessage.(string), "$.user:sayHello")).IsTrue()
			assert(out.CanRead()).IsFalse()
		},
	)

	// error with 2nd param
	runWithProcessor(
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.Return(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(true)
			stream.Write(uint64(3))
			stream.Write(float64(3))
			stream.Write("hello")
			stream.Write(([]byte)("world"))
			stream.Write(Array{1})
			stream.Write(Map{"name": "world"})
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).IsFalse()
			assert(out.ReadBool()).Equals(false, true)
			assert(out.Read()).Equals("rpc reply arguments not match\n"+
				"Called: $.user:sayHello(rpc.ContextObject, rpc.Bool, rpc.Bool, rpc.Uint64, "+
				"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.ReturnObject\n"+
				"Required: $.user:sayHello(rpc.ContextObject, rpc.Bool, rpc.Int64, rpc.Uint64, "+
				"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.ReturnObject",
				true,
			)
			dbgMessage, ok := out.Read()
			assert(ok).IsTrue()
			assert(strings.Contains(dbgMessage.(string), "$.user:sayHello")).IsTrue()
			assert(out.CanRead()).IsFalse()
		},
	)

	// error with 3rd param
	runWithProcessor(
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.Return(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(int64(3))
			stream.Write(true)
			stream.Write(float64(3))
			stream.Write("hello")
			stream.Write(([]byte)("world"))
			stream.Write(Array{1})
			stream.Write(Map{"name": "world"})
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).IsFalse()
			assert(out.ReadBool()).Equals(false, true)
			assert(out.Read()).Equals("rpc reply arguments not match\n"+
				"Called: $.user:sayHello(rpc.ContextObject, rpc.Bool, rpc.Int64, rpc.Bool, "+
				"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.ReturnObject\n"+
				"Required: $.user:sayHello(rpc.ContextObject, rpc.Bool, rpc.Int64, rpc.Uint64, "+
				"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.ReturnObject",
				true,
			)
			dbgMessage, ok := out.Read()
			assert(ok).IsTrue()
			assert(strings.Contains(dbgMessage.(string), "$.user:sayHello")).IsTrue()
			assert(out.CanRead()).IsFalse()
		},
	)

	// error with 4th param
	runWithProcessor(
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.Return(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(int64(3))
			stream.Write(uint(3))
			stream.Write(true)
			stream.Write("hello")
			stream.Write(([]byte)("world"))
			stream.Write(Array{1})
			stream.Write(Map{"name": "world"})
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).IsFalse()
			assert(out.ReadBool()).Equals(false, true)
			assert(out.Read()).Equals("rpc reply arguments not match\n"+
				"Called: $.user:sayHello(rpc.ContextObject, rpc.Bool, rpc.Int64, rpc.Uint64, "+
				"rpc.Bool, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.ReturnObject\n"+
				"Required: $.user:sayHello(rpc.ContextObject, rpc.Bool, rpc.Int64, rpc.Uint64, "+
				"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.ReturnObject",
				true,
			)
			dbgMessage, ok := out.Read()
			assert(ok).IsTrue()
			assert(strings.Contains(dbgMessage.(string), "$.user:sayHello")).IsTrue()
			assert(out.CanRead()).IsFalse()
		},
	)

	// error with 5th param
	runWithProcessor(
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.Return(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(int64(3))
			stream.Write(uint(3))
			stream.Write(float64(3))
			stream.Write(true)
			stream.Write(([]byte)("world"))
			stream.Write(Array{1})
			stream.Write(Map{"name": "world"})
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).IsFalse()
			assert(out.ReadBool()).Equals(false, true)
			assert(out.Read()).Equals("rpc reply arguments not match\n"+
				"Called: $.user:sayHello(rpc.ContextObject, rpc.Bool, rpc.Int64, rpc.Uint64, "+
				"rpc.Float64, rpc.Bool, rpc.Bytes, rpc.Array, rpc.Map) rpc.ReturnObject\n"+
				"Required: $.user:sayHello(rpc.ContextObject, rpc.Bool, rpc.Int64, rpc.Uint64, "+
				"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.ReturnObject",
				true,
			)
			dbgMessage, ok := out.Read()
			assert(ok).IsTrue()
			assert(strings.Contains(dbgMessage.(string), "$.user:sayHello")).IsTrue()
			assert(out.CanRead()).IsFalse()
		},
	)

	// error with 6th param
	runWithProcessor(
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.Return(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(int64(3))
			stream.Write(uint(3))
			stream.Write(float64(3))
			stream.Write("hello")
			stream.Write(true)
			stream.Write(Array{1})
			stream.Write(Map{"name": "world"})
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).IsFalse()
			assert(out.ReadBool()).Equals(false, true)
			assert(out.Read()).Equals("rpc reply arguments not match\n"+
				"Called: $.user:sayHello(rpc.ContextObject, rpc.Bool, rpc.Int64, rpc.Uint64, "+
				"rpc.Float64, rpc.String, rpc.Bool, rpc.Array, rpc.Map) rpc.ReturnObject\n"+
				"Required: $.user:sayHello(rpc.ContextObject, rpc.Bool, rpc.Int64, rpc.Uint64, "+
				"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.ReturnObject",
				true,
			)
			dbgMessage, ok := out.Read()
			assert(ok).IsTrue()
			assert(strings.Contains(dbgMessage.(string), "$.user:sayHello")).IsTrue()
			assert(out.CanRead()).IsFalse()
		},
	)

	// error with 7th param
	runWithProcessor(
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.Return(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(int64(3))
			stream.Write(uint(3))
			stream.Write(float64(3))
			stream.Write("hello")
			stream.Write(([]byte)("world"))
			stream.Write(true)
			stream.Write(Map{"name": "world"})
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).IsFalse()
			assert(out.ReadBool()).Equals(false, true)
			assert(out.Read()).Equals("rpc reply arguments not match\n"+
				"Called: $.user:sayHello(rpc.ContextObject, rpc.Bool, rpc.Int64, rpc.Uint64, "+
				"rpc.Float64, rpc.String, rpc.Bytes, rpc.Bool, rpc.Map) rpc.ReturnObject\n"+
				"Required: $.user:sayHello(rpc.ContextObject, rpc.Bool, rpc.Int64, rpc.Uint64, "+
				"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.ReturnObject",
				true,
			)
			dbgMessage, ok := out.Read()
			assert(ok).IsTrue()
			assert(strings.Contains(dbgMessage.(string), "$.user:sayHello")).IsTrue()
			assert(out.CanRead()).IsFalse()
		},
	)

	// error with 8th param
	runWithProcessor(
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.Return(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(int64(3))
			stream.Write(uint(3))
			stream.Write(float64(3))
			stream.Write("hello")
			stream.Write(([]byte)("world"))
			stream.Write(Array{1})
			stream.Write(true)
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).IsFalse()
			assert(out.ReadBool()).Equals(false, true)
			assert(out.Read()).Equals("rpc reply arguments not match\n"+
				"Called: $.user:sayHello(rpc.ContextObject, rpc.Bool, rpc.Int64, rpc.Uint64, "+
				"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Bool) rpc.ReturnObject\n"+
				"Required: $.user:sayHello(rpc.ContextObject, rpc.Bool, rpc.Int64, rpc.Uint64, "+
				"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.ReturnObject",
				true,
			)
			dbgMessage, ok := out.Read()
			assert(ok).IsTrue()
			assert(strings.Contains(dbgMessage.(string), "$.user:sayHello")).IsTrue()
			assert(out.CanRead()).IsFalse()
		},
	)

	// test nil rpcBytes
	runWithProcessor(
		func(ctx *ContextObject, a Bytes) *ReturnObject {
			if a != nil {
				return ctx.Return(errors.New("param is not nil"))
			}

			return ctx.Return(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(nil)
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).IsTrue()
			assert(out.ReadBool()).Equals(true, true)
			assert(out.Read()).Equals(true, true)
			assert(out.CanRead()).IsFalse()
		},
	)

	// test nil rpcArray
	runWithProcessor(
		func(ctx *ContextObject, a Array) *ReturnObject {
			if a != nil {
				return ctx.Return(errors.New("param is not nil"))
			}
			return ctx.Return(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(nil)
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).IsTrue()
			assert(out.ReadBool()).Equals(true, true)
			assert(out.Read()).Equals(true, true)
			assert(out.CanRead()).IsFalse()
		},
	)

	// test nil rpcMap
	runWithProcessor(
		func(ctx *ContextObject, a Map) *ReturnObject {
			if a != nil {
				return ctx.Return(errors.New("param is not nil"))
			}
			return ctx.Return(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(nil)
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).IsTrue()
			assert(out.ReadBool()).Equals(true, true)
			assert(out.Read()).Equals(true, true)
			assert(out.CanRead()).IsFalse()
		},
	)

	// test unsupported type
	runWithProcessor(
		func(ctx *ContextObject, a bool) *ReturnObject {
			return ctx.Return(a)
		},
		func(processor *Processor) *Stream {
			replyNode := processor.repliesMap["$.user:sayHello"]
			replyNode.argTypes[1] = reflect.ValueOf(int16(0)).Type()
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).IsFalse()
			assert(out.ReadBool()).Equals(false, true)
			assert(out.Read()).Equals("rpc reply arguments not match\n"+
				"Called: $.user:sayHello(rpc.ContextObject, rpc.Bool) rpc.ReturnObject\n"+
				"Required: $.user:sayHello(rpc.ContextObject, rpc.Bool) rpc.ReturnObject",
				true,
			)
			dbgMessage, ok := out.Read()
			assert(ok).IsTrue()
			assert(strings.Contains(dbgMessage.(string), "$.user:sayHello")).IsTrue()
			assert(out.CanRead()).IsFalse()
		},
	)

	// nil text
	runWithProcessor(
		func(ctx *ContextObject, bVal bool, rpcMap Map) *ReturnObject {
			return ctx.Return(bVal)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(nil)
			stream.Write(nil)
			stream.Write(nil)
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).IsFalse()
			assert(out.ReadBool()).Equals(false, true)
			assert(out.Read()).Equals("rpc reply arguments not match\n"+
				"Called: $.user:sayHello(rpc.ContextObject, <nil>, rpc.Map, <nil>) "+
				"rpc.ReturnObject\n"+
				"Required: $.user:sayHello(rpc.ContextObject, rpc.Bool, rpc.Map) rpc.ReturnObject",
				true,
			)
			dbgMessage, ok := out.Read()
			assert(ok).IsTrue()
			assert(strings.Contains(dbgMessage.(string), "$.user:sayHello")).IsTrue()
			assert(out.CanRead()).IsFalse()
		},
	)

	// stream is broken
	runWithProcessor(
		func(ctx *ContextObject, bVal bool, rpcMap Map) *ReturnObject {
			return ctx.Return(bVal)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write("helloWorld")
			stream.SetWritePos(stream.GetWritePos() - 1)
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).IsFalse()
			assert(out.ReadBool()).Equals(false, true)
			assert(out.Read()).Equals("rpc data format error", true)
			assert(out.Read()).Equals("", true)
			assert(out.CanRead()).IsFalse()
		},
	)

	// call function error
	runWithProcessor(
		func(ctx *ContextObject, bVal bool) *ReturnObject {
			if bVal {
				panic("this is a error")
			}
			return ctx.Return(bVal)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			return stream
		},
		func(in *Stream, out *Stream, success bool) {
			assert(success).IsFalse()
			assert(out.ReadBool()).Equals(false, true)
			assert(out.Read()).Equals(
				"rpc-server: $.user:sayHello(rpc.ContextObject, rpc.Bool) rpc.ReturnObject: "+
					"runtime error: this is a error",
				true,
			)
			dbgMessage, ok := out.Read()
			assert(
				strings.Contains(dbgMessage.(string), "TestRpcThread_eval"),
			).IsTrue()
			assert(ok).IsTrue()
			assert(out.CanRead()).IsFalse()
		},
	)
}
