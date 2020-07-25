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
	//assert(len(rpcThread.inputCH)).Equals(0)
	//assert(cap(rpcThread.inputCH)).Equals(0)
	//assert(rpcThread.inStream).IsNil()
	//assert(rpcThread.execStream).IsNotNil()
	//assert(rpcThread.execDepth).Equals(uint64(0))
	//assert(rpcThread.execReplyNode).IsNil()
	//assert(len(rpcThread.execArgs)).Equals(0)
	//assert(cap(rpcThread.execArgs)).Equals(16)
	//assert(rpcThread.execSuccessful).IsFalse()
	//assert(rpcThread.execFrom).Equals("")
	//assert(len(rpcThread.closeCH)).Equals(0)
	//assert(cap(rpcThread.closeCH)).Equals(0)
}

func TestRpcThread_Eval(t *testing.T) {
	assert := NewAssert(t)

	// Test(0) basic
	assert(testRunWithProcessor(true, nil,
		func(ctx *ContextObject, name string) *ReturnObject {
			return ctx.OK("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
	)).Equals("hello world", nil, nil)

	// Test(1) read reply path error
	assert(testRunWithProcessor(true, nil,
		func(ctx *ContextObject, name string) *ReturnObject {
			return ctx.OK("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			// path format error
			stream.WriteBytes([]byte("$.test:Eval"))
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
	)).Equals(nil, NewProtocolError(ErrStringBadStream), nil)

	// Test(2) reply path is not mounted
	assert(testRunWithProcessor(true, nil,
		func(ctx *ContextObject, name string) *ReturnObject {
			return ctx.OK("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			// reply path is not mounted
			stream.WriteString("$.system:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
	)).Equals(nil, NewReplyError("rpc: target $.system:Eval does not exist"), nil)

	// Test(3) depth data format error
	assert(testRunWithProcessor(true, nil,
		func(ctx *ContextObject, name string) *ReturnObject {
			return ctx.OK("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			// depth type error
			stream.WriteInt64(3)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
	)).Equals(nil, NewProtocolError(ErrStringBadStream), nil)

	// Test(4) depth is overflow
	assert(testRunWithProcessor(true, nil,
		func(ctx *ContextObject, name string) *ReturnObject {
			return ctx.OK("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			//  depth is overflow
			stream.WriteUint64(17)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
	)).Equals(
		nil,
		NewReplyError("rpc: call $.test:Eval level(17) overflows"),
		nil,
	)

	// Test(5) execFrom data format error
	assert(testRunWithProcessor(true, nil,
		func(ctx *ContextObject, name string) *ReturnObject {
			return ctx.OK("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			// execFrom data format error
			stream.WriteBool(true)
			stream.WriteString("world")
			return stream
		},
	)).Equals(nil, NewProtocolError(ErrStringBadStream), nil)

	// Test(6) ok call with all type value
	assert(testRunWithProcessor(true, nil,
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
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
	)).Equals(true, nil, nil)

	// Test(7) error with 1st param
	assert(testRunWithProcessor(false, nil,
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			// error rpc.Bool
			stream.Write(2)
			stream.Write(int64(3))
			stream.Write(uint64(3))
			stream.Write(float64(3))
			stream.Write("hello")
			stream.Write(([]byte)("world"))
			stream.Write(Array{1})
			stream.Write(Map{"name": "world"})
			return stream
		},
	)).Equals(
		nil,
		NewReplyError("rpc: $.test:Eval reply arguments does not match"),
		nil,
	)

	// Test(8) error with 1st param
	ret8, error8, panic8 := testRunWithProcessor(true, nil,
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			// error rpc.Bool
			stream.Write(2)
			stream.Write(int64(3))
			stream.Write(uint64(3))
			stream.Write(float64(3))
			stream.Write("hello")
			stream.Write(([]byte)("world"))
			stream.Write(Array{1})
			stream.Write(Map{"name": "world"})
			return stream
		},
	)
	assert(ret8, panic8).IsNil()
	assert(error8.GetKind()).Equals(ErrorKindReply)
	assert(error8.GetMessage()).Equals(
		"rpc: $.test:Eval reply arguments does not match\n" +
			"want: $.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
			"got: $.test:Eval(rpc.Context, rpc.Int64, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return",
	)
	assert(strings.Contains(error8.GetDebug(), "$.test:Eval")).IsTrue()
	assert(strings.Contains(error8.GetDebug(), "types_test.go:")).IsTrue()

	// Test(9) error with 2nd param
	assert(testRunWithProcessor(false, nil,
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			// error rpc.Int64
			stream.Write(true)
			stream.Write(uint64(3))
			stream.Write(float64(3))
			stream.Write("hello")
			stream.Write(([]byte)("world"))
			stream.Write(Array{1})
			stream.Write(Map{"name": "world"})
			return stream
		},
	)).Equals(
		nil,
		NewReplyError("rpc: $.test:Eval reply arguments does not match"),
		nil,
	)

	// Test(10) error with 2nd param
	ret10, error10, panic10 := testRunWithProcessor(true, nil,
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			// error rpc.Int64
			stream.Write(true)
			stream.Write(uint64(3))
			stream.Write(float64(3))
			stream.Write("hello")
			stream.Write(([]byte)("world"))
			stream.Write(Array{1})
			stream.Write(Map{"name": "world"})
			return stream
		},
	)
	assert(ret10, panic10).IsNil()
	assert(error10.GetKind()).Equals(ErrorKindReply)
	assert(error10.GetMessage()).Equals(
		"rpc: $.test:Eval reply arguments does not match\n" +
			"want: $.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
			"got: $.test:Eval(rpc.Context, rpc.Bool, rpc.Bool, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return",
	)
	assert(strings.Contains(error10.GetDebug(), "$.test:Eval")).IsTrue()
	assert(strings.Contains(error10.GetDebug(), "types_test.go:")).IsTrue()

	// Test(11) error with 3rd param
	assert(testRunWithProcessor(false, nil,
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(int64(3))
			// error rpc.Uint64
			stream.Write(true)
			stream.Write(float64(3))
			stream.Write("hello")
			stream.Write(([]byte)("world"))
			stream.Write(Array{1})
			stream.Write(Map{"name": "world"})
			return stream
		},
	)).Equals(
		nil,
		NewReplyError("rpc: $.test:Eval reply arguments does not match"),
		nil,
	)

	// Test(12) error with 3rd param
	ret12, error12, panic12 := testRunWithProcessor(true, nil,
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(int64(3))
			// error rpc.Uint64
			stream.Write(true)
			stream.Write(float64(3))
			stream.Write("hello")
			stream.Write(([]byte)("world"))
			stream.Write(Array{1})
			stream.Write(Map{"name": "world"})
			return stream
		},
	)
	assert(ret12, panic12).IsNil()
	assert(error12.GetKind()).Equals(ErrorKindReply)
	assert(error12.GetMessage()).Equals(
		"rpc: $.test:Eval reply arguments does not match\n" +
			"want: $.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
			"got: $.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Bool, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return",
	)
	assert(strings.Contains(error12.GetDebug(), "$.test:Eval")).IsTrue()
	assert(strings.Contains(error12.GetDebug(), "types_test.go:")).IsTrue()

	// Test(13) error with 4th param
	assert(testRunWithProcessor(false, nil,
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(int64(3))
			stream.Write(uint(3))
			// error rpc.Float64
			stream.Write(true)
			stream.Write("hello")
			stream.Write(([]byte)("world"))
			stream.Write(Array{1})
			stream.Write(Map{"name": "world"})
			return stream
		},
	)).Equals(
		nil,
		NewReplyError("rpc: $.test:Eval reply arguments does not match"),
		nil,
	)

	// Test(14) error with 4th param
	ret14, error14, panic14 := testRunWithProcessor(true, nil,
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(int64(3))
			stream.Write(uint(3))
			// error rpc.Float64
			stream.Write(true)
			stream.Write("hello")
			stream.Write(([]byte)("world"))
			stream.Write(Array{1})
			stream.Write(Map{"name": "world"})
			return stream
		},
	)
	assert(ret14, panic14).IsNil()
	assert(error14.GetKind()).Equals(ErrorKindReply)
	assert(error14.GetMessage()).Equals(
		"rpc: $.test:Eval reply arguments does not match\n" +
			"want: $.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
			"got: $.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Bool, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return",
	)
	assert(strings.Contains(error14.GetDebug(), "$.test:Eval")).IsTrue()
	assert(strings.Contains(error14.GetDebug(), "types_test.go:")).IsTrue()

	// Test(15) error with 5th param
	assert(testRunWithProcessor(false, nil,
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(int64(3))
			stream.Write(uint(3))
			stream.Write(float64(3))
			// error rpc.String
			stream.Write(true)
			stream.Write(([]byte)("world"))
			stream.Write(Array{1})
			stream.Write(Map{"name": "world"})
			return stream
		},
	)).Equals(
		nil,
		NewReplyError("rpc: $.test:Eval reply arguments does not match"),
		nil,
	)

	// Test(16) error with 5th param
	ret16, error16, panic16 := testRunWithProcessor(true, nil,
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(int64(3))
			stream.Write(uint(3))
			stream.Write(float64(3))
			// error rpc.String
			stream.Write(true)
			stream.Write(([]byte)("world"))
			stream.Write(Array{1})
			stream.Write(Map{"name": "world"})
			return stream
		},
	)
	assert(ret16, panic16).IsNil()
	assert(error16.GetKind()).Equals(ErrorKindReply)
	assert(error16.GetMessage()).Equals(
		"rpc: $.test:Eval reply arguments does not match\n" +
			"want: $.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
			"got: $.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.Bool, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return",
	)
	assert(strings.Contains(error16.GetDebug(), "$.test:Eval")).IsTrue()
	assert(strings.Contains(error16.GetDebug(), "types_test.go:")).IsTrue()

	// Test(17) error with 6th param
	assert(testRunWithProcessor(false, nil,
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(int64(3))
			stream.Write(uint(3))
			stream.Write(float64(3))
			stream.Write("hello")
			// error rpc.Bytes
			stream.Write(true)
			stream.Write(Array{1})
			stream.Write(Map{"name": "world"})
			return stream
		},
	)).Equals(
		nil,
		NewReplyError("rpc: $.test:Eval reply arguments does not match"),
		nil,
	)

	// Test(18) error with 6th param
	ret18, error18, panic18 := testRunWithProcessor(true, nil,
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(int64(3))
			stream.Write(uint(3))
			stream.Write(float64(3))
			stream.Write("hello")
			// error rpc.Bytes
			stream.Write(true)
			stream.Write(Array{1})
			stream.Write(Map{"name": "world"})
			return stream
		},
	)
	assert(ret18, panic18).IsNil()
	assert(error18.GetKind()).Equals(ErrorKindReply)
	assert(error18.GetMessage()).Equals(
		"rpc: $.test:Eval reply arguments does not match\n" +
			"want: $.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
			"got: $.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bool, rpc.Array, rpc.Map) rpc.Return",
	)
	assert(strings.Contains(error18.GetDebug(), "$.test:Eval")).IsTrue()
	assert(strings.Contains(error18.GetDebug(), "types_test.go:")).IsTrue()

	// Test(19) error with 7th param
	assert(testRunWithProcessor(false, nil,
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(int64(3))
			stream.Write(uint(3))
			stream.Write(float64(3))
			stream.Write("hello")
			stream.Write(([]byte)("world"))
			// error rpc.Array
			stream.Write(true)
			stream.Write(Map{"name": "world"})
			return stream
		},
	)).Equals(
		nil,
		NewReplyError("rpc: $.test:Eval reply arguments does not match"),
		nil,
	)

	// Test(20) error with 6th param
	ret20, error20, panic20 := testRunWithProcessor(true, nil,
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(int64(3))
			stream.Write(uint(3))
			stream.Write(float64(3))
			stream.Write("hello")
			stream.Write(([]byte)("world"))
			// error rpc.Array
			stream.Write(true)
			stream.Write(Map{"name": "world"})
			return stream
		},
	)
	assert(ret20, panic20).IsNil()
	assert(error20.GetKind()).Equals(ErrorKindReply)
	assert(error20.GetMessage()).Equals(
		"rpc: $.test:Eval reply arguments does not match\n" +
			"want: $.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
			"got: $.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Bool, rpc.Map) rpc.Return",
	)
	assert(strings.Contains(error20.GetDebug(), "$.test:Eval")).IsTrue()
	assert(strings.Contains(error20.GetDebug(), "types_test.go:")).IsTrue()

	// Test(21) error with 8th param
	assert(testRunWithProcessor(false, nil,
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(int64(3))
			stream.Write(uint(3))
			stream.Write(float64(3))
			stream.Write("hello")
			stream.Write(([]byte)("world"))
			stream.Write(Array{1})
			// error rpc.Map
			stream.Write(true)
			return stream
		},
	)).Equals(
		nil,
		NewReplyError("rpc: $.test:Eval reply arguments does not match"),
		nil,
	)

	// Test(22) error with 8th param
	ret22, error22, panic22 := testRunWithProcessor(true, nil,
		func(ctx *ContextObject,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) *ReturnObject {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			stream.Write(int64(3))
			stream.Write(uint(3))
			stream.Write(float64(3))
			stream.Write("hello")
			stream.Write(([]byte)("world"))
			stream.Write(Array{1})
			// error rpc.Map
			stream.Write(true)
			return stream
		},
	)
	assert(ret22, panic22).IsNil()
	assert(error22.GetKind()).Equals(ErrorKindReply)
	assert(error22.GetMessage()).Equals(
		"rpc: $.test:Eval reply arguments does not match\n" +
			"want: $.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
			"got: $.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Bool) rpc.Return",
	)
	assert(strings.Contains(error22.GetDebug(), "$.test:Eval")).IsTrue()
	assert(strings.Contains(error22.GetDebug(), "types_test.go:")).IsTrue()

	// Test(23) nil rpcBytes
	assert(testRunWithProcessor(true, nil,
		func(ctx *ContextObject, a Bytes) *ReturnObject {
			if a != nil {
				return ctx.Error(errors.New("param is not nil"))
			}
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(nil)
			return stream
		},
	)).Equals(true, nil, nil)

	// Test(24) nil rpcArray
	assert(testRunWithProcessor(true, nil,
		func(ctx *ContextObject, a Array) *ReturnObject {
			if a != nil {
				return ctx.Error(errors.New("param is not nil"))
			}
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(nil)
			return stream
		},
	)).Equals(true, nil, nil)

	// Test(25) nil rpcMap
	assert(testRunWithProcessor(true, nil,
		func(ctx *ContextObject, a Map) *ReturnObject {
			if a != nil {
				return ctx.Error(errors.New("param is not nil"))
			}
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(nil)
			return stream
		},
	)).Equals(true, nil, nil)

	// Test(26) unsupported type
	assert(testRunWithProcessor(false, nil,
		func(ctx *ContextObject, a bool) *ReturnObject {
			return ctx.OK(a)
		},
		func(processor *Processor) *Stream {
			replyNode := processor.repliesMap["$.test:Eval"]
			replyNode.argTypes[1] = reflect.ValueOf(int16(0)).Type()
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			return stream
		},
	)).Equals(
		nil,
		NewReplyError("rpc: $.test:Eval reply arguments does not match"),
		nil,
	)

	// Test(27) test
	ret27, error27, panic27 := testRunWithProcessor(true, nil,
		func(ctx *ContextObject, bVal bool, rpcMap Map) *ReturnObject {
			return ctx.OK(bVal)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(nil)
			stream.Write(nil)
			stream.Write(nil)
			return stream
		},
	)
	assert(ret27, panic27).IsNil()
	assert(error27.GetKind()).Equals(ErrorKindReply)
	assert(error27.GetMessage()).Equals(
		"rpc: $.test:Eval reply arguments does not match\n" +
			"want: $.test:Eval(rpc.Context, rpc.Bool, rpc.Map) rpc.Return\n" +
			"got: $.test:Eval(rpc.Context, <nil>, rpc.Map, <nil>) rpc.Return",
	)
	assert(strings.Contains(error27.GetDebug(), "$.test:Eval")).IsTrue()
	assert(strings.Contains(error27.GetDebug(), "types_test.go:")).IsTrue()

	// Test(28) badStream
	assert(testRunWithProcessor(true, nil,
		func(ctx *ContextObject, bVal bool, rpcMap Map) *ReturnObject {
			return ctx.OK(bVal)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write("helloWorld")
			stream.SetWritePos(stream.GetWritePos() - 1)
			return stream
		},
	)).Equals(nil, NewProtocolError(ErrStringBadStream), nil)

	// Test(29) call function error
	ret29, error29, panic29 := testRunWithProcessor(false, nil,
		func(ctx *ContextObject, bVal bool) *ReturnObject {
			if bVal {
				panic("this is a error")
			}
			return ctx.OK(bVal)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			return stream
		},
	)
	assert(ret29).IsNil()
	assert(error29, panic29).IsNotNil()
	if error29 != nil {
		assert(error29.GetKind()).Equals(ErrorKindReply)
		assert(error29.GetMessage()).Equals("rpc: $.test:Eval runtime error")
		assert(strings.Contains(error29.GetDebug(), "$.test:Eval")).IsTrue()
		assert(strings.Contains(error29.GetDebug(), "types_test.go")).IsTrue()
	}
	if panic29 != nil {
		assert(panic29.GetKind()).Equals(ErrorKindReplyPanic)
		assert(panic29.GetMessage()).
			Equals("rpc: $.test:Eval runtime error: this is a error")
		assert(strings.Contains(panic29.GetDebug(), "thread_test.go")).IsTrue()
	}

	// Test(30) call function error
	ret30, error30, panic30 := testRunWithProcessor(true, nil,
		func(ctx *ContextObject, bVal bool) *ReturnObject {
			if bVal {
				panic("this is a error")
			}
			return ctx.OK(bVal)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			return stream
		},
	)
	assert(ret30).IsNil()
	assert(error30, panic30).IsNotNil()
	if error30 != nil {
		assert(error30.GetKind()).Equals(ErrorKindReply)
		assert(error30.GetMessage()).Equals("rpc: $.test:Eval runtime error")
		assert(strings.Contains(error30.GetDebug(), "$.test:Eval")).IsTrue()
		assert(strings.Contains(error30.GetDebug(), "types_test.go")).IsTrue()
	}
	if panic30 != nil {
		assert(panic30.GetKind()).Equals(ErrorKindReplyPanic)
		assert(panic30.GetMessage()).
			Equals("rpc: $.test:Eval runtime error: this is a error")
		assert(strings.Contains(panic30.GetDebug(), "thread_test.go")).IsTrue()
	}

	// Test(31) return TransportError to make onEvalFinish panic
	ret31, error31, panic31 := testRunWithProcessor(true, nil,
		func(ctx *ContextObject, bVal bool) *ReturnObject {
			return ctx.Error(NewTransportError("it makes onEvalFinish panic"))
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			return stream
		},
	)
	assert(ret31, error31).Equals(nil, nil)
	assert(panic31).IsNotNil()
	if panic31 != nil {
		assert(panic31.GetKind()).Equals(ErrorKindKernel)
		assert(panic31.GetMessage()).
			Equals("rpc: kernel error: test panic")
		assert(strings.Contains(panic31.GetDebug(), "types_test.go")).IsTrue()
	}

	// Test(32) return without ctx
	ret32, error32, panic32 := testRunWithProcessor(true, nil,
		func(ctx *ContextObject, bVal bool) *ReturnObject {
			return Return(nil)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			return stream
		},
	)
	assert(ret32, panic32).Equals(nil, nil)
	assert(error32).IsNotNil()
	if error32 != nil {
		assert(error32.GetKind()).Equals(ErrorKindReplyPanic)
		assert(error32.GetMessage()).
			Equals("rpc: $.test:Eval must return through Context.OK or Context.Error")
		assert(strings.Contains(error32.GetDebug(), "types_test.go")).IsTrue()
	} else {
		assert().Fail("nil)")
	}
}

func TestRpcThread_Eval2(t *testing.T) {
	assert := NewAssert(t)

	// Test(32) return without ctx
	ret32, error32, panic32 := testRunWithProcessor(true, nil,
		func(ctx *ContextObject, bVal bool) *ReturnObject {
			return Return(nil)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("$.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			return stream
		},
	)
	assert(ret32, panic32).Equals(nil, nil)
	assert(error32).IsNotNil()
	if error32 != nil {
		assert(error32.GetKind()).Equals(ErrorKindReplyPanic)
		assert(error32.GetMessage()).
			Equals("rpc: $.test:Eval must return through Context.OK or Context.Error")
		assert(strings.Contains(error32.GetDebug(), "types_test.go")).IsTrue()
	} else {
		assert().Fail("nil)")
	}
}
