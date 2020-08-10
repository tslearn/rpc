package internal

import (
	"errors"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

func TestNewThread(t *testing.T) {
	assert := NewAssert(t)
	fakeEvalBack := getFakeOnEvalBack()
	fakeEvalFinish := getFakeOnEvalFinish()

	// Test(1) processor is nil
	assert(newThread(nil, 5*time.Second, fakeEvalBack, nil)).IsNil()

	// Test(2) processor is nil
	assert(newThread(nil, 5*time.Second, nil, fakeEvalFinish)).IsNil()

	// Test(3) onEvalBack is nil
	assert(newThread(getFakeProcessor(true), 5*time.Second, nil, nil)).IsNil()

	// Test(4) debug thread
	thread4 := getFakeThread(true)
	defer thread4.Close()
	assert(thread4.processor).IsNotNil()
	assert(thread4.inputCH).IsNotNil()
	assert(thread4.closeCH).IsNotNil()
	assert(thread4.execStream).IsNotNil()
	assert(thread4.execDepth).IsNotNil()
	assert(thread4.execReplyNode).IsNil()
	assert(thread4.execArgs).IsNotNil()
	assert(thread4.execFrom).Equals("")

	// Test(5) release thread
	thread5 := getFakeThread(false)
	defer thread5.Close()
	assert(thread5.processor).IsNotNil()
	assert(thread5.inputCH).IsNotNil()
	assert(thread5.closeCH).IsNotNil()
	assert(thread5.execStream).IsNotNil()
	assert(thread5.execDepth).IsNotNil()
	assert(thread5.execReplyNode).IsNil()
	assert(thread5.execArgs).IsNotNil()
	assert(thread5.execFrom).Equals("")

	// Test(6) closeTimeout is less than 1 * time.Second
	thread6 := newThread(
		getFakeProcessor(true),
		-500*time.Millisecond,
		getFakeOnEvalBack(),
		getFakeOnEvalFinish(),
	)
	assert(thread6.closeTimeout).Equals(time.Second)
}

func TestRpcThread_Close(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	thread1 := getFakeThread(true)
	assert(thread1.Close()).IsTrue()

	// Test(2) cant close
	assert(testRunWithCatchPanic(func() {
		_, _, _ = testRunWithProcessor(true, nil,
			func(ctx Context, name string) Return {
				time.Sleep(8 * time.Second)
				return ctx.OK("hello " + name)
			},
			func(processor *Processor) *Stream {
				go func() {
					time.Sleep(time.Second)
					assert(processor.Close()).IsFalse()
				}()
				stream := NewStream()
				stream.WriteString("#.test:Eval")
				stream.WriteUint64(3)
				stream.WriteString("#")
				stream.WriteString("world")
				return stream
			},
		)
	})).IsNotNil()

	// Test(3) Close twice
	thread3 := getFakeThread(true)
	assert(thread3.Close()).IsTrue()
	assert(thread3.Close()).IsFalse()
}

func TestRpcThread_GetExecReplyNodePath(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	thread1 := getFakeThread(true)
	defer thread1.Close()
	assert(thread1.GetExecReplyNodePath()).Equals("")

	// Test(2)
	thread2 := getFakeThread(true)
	defer thread2.Close()
	thread2.execReplyNode = unsafe.Pointer(&rpcReplyNode{path: "#.test:Eval"})
	assert(thread2.GetExecReplyNodePath()).Equals("#.test:Eval")
}

func TestRpcThread_GetExecReplyFileLine(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	thread1 := getFakeThread(true)
	defer thread1.Close()
	assert(thread1.GetExecReplyDebug()).Equals("")

	// Test(2)
	thread2 := getFakeThread(true)
	defer thread2.Close()
	thread2.execReplyNode = unsafe.Pointer(&rpcReplyNode{
		path: "#.test:Eval",
		meta: &rpcReplyMeta{fileLine: "/test_file:234"},
	})
	assert(thread2.GetExecReplyDebug()).Equals("#.test:Eval /test_file:234")
}

func TestRpcThread_WriteError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1) ok
	source1 := ""
	assert(testRunWithProcessor(true, nil,
		func(ctx Context, name string) Return {
			ret, source := ctx.Error(errors.New("error")), GetFileLine(0)
			source1 = source
			return ret
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
	)).Equals(nil, NewReplyError("error").AddDebug("#.test:Eval "+source1), nil)
}

func TestRpcThread_WriteOK(t *testing.T) {
	assert := NewAssert(t)

	// Test(1) value is endless loop
	source1 := ""
	assert(testRunWithProcessor(true, nil,
		func(ctx Context, name string) Return {
			v := make(Map)
			v["v"] = v
			ret, source := ctx.OK(v), GetFileLine(0)
			source1 = source
			return ret
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
	)).Equals(
		nil,
		nil,
		NewReplyPanic("value[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"]"+
			"[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"]"+
			"[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"]"+
			"[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"]"+
			"[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"]"+
			"[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"]"+
			"[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"] is too complicated").
			AddDebug("#.test:Eval "+source1),
	)

	// Test(2) value is not support
	source2 := ""
	assert(testRunWithProcessor(true, nil,
		func(ctx Context, name string) Return {
			ret, source := ctx.OK(make(chan bool)), GetFileLine(0)
			source2 = source
			return ret
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
	)).Equals(
		nil,
		nil,
		NewReplyPanic("value type (chan bool) is not supported").
			AddDebug("#.test:Eval "+source2),
	)

	// Test(3) ok
	assert(testRunWithProcessor(true, nil,
		func(ctx Context, name string) Return {
			return ctx.OK("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
	)).Equals("hello world", nil, nil)
}

func TestRpcThread_PutStream(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	thread1 := getFakeThread(true)
	defer thread1.Close()
	assert(thread1.PutStream(NewStream())).IsTrue()

	// Test(2)
	thread2 := getFakeThread(true)
	thread2.Close()
	assert(thread2.PutStream(NewStream())).IsFalse()
}

func TestRpcThread_Eval1(t *testing.T) {
	assert := NewAssert(t)

	// Test(1) read reply path error
	assert(testRunWithProcessor(true, nil,
		func(ctx Context, name string) Return {
			return ctx.OK("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			// path format error
			stream.WriteBytes([]byte("#.test:Eval"))
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
	)).Equals(nil, NewProtocolError(ErrStringBadStream), nil)
}

func TestRpcThread_setReturn(t *testing.T) {
	assert := NewAssert(t)

	thread := getFakeThread(false)

	// Test(1)
	assert(thread.setReturn(atomic.LoadUint64(&thread.sequence))).
		Equals(rpcThreadReturnStatusOK)

	// Test(2)
	assert(thread.setReturn(atomic.LoadUint64(&thread.sequence) - 1)).
		Equals(rpcThreadReturnStatusAlreadyCalled)

	// Test(3)
	assert(thread.setReturn(atomic.LoadUint64(&thread.sequence) - 2)).
		Equals(rpcThreadReturnStatusContextError)

	// Test(4)
	assert(thread.setReturn(atomic.LoadUint64(&thread.sequence) + 1)).
		Equals(rpcThreadReturnStatusContextError)
}

func TestRpcThread_Eval(t *testing.T) {
	assert := NewAssert(t)

	// Test(0) basic
	assert(testRunWithProcessor(true, nil,
		func(ctx Context, name string) Return {
			return ctx.OK("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
	)).Equals("hello world", nil, nil)

	// Test(1) read reply path error
	assert(testRunWithProcessor(true, nil,
		func(ctx Context, name string) Return {
			return ctx.OK("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			// path format error
			stream.WriteBytes([]byte("#.test:Eval"))
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
	)).Equals(nil, NewProtocolError(ErrStringBadStream), nil)

	// Test(2) reply path is not mounted
	assert(testRunWithProcessor(true, nil,
		func(ctx Context, name string) Return {
			return ctx.OK("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			// reply path is not mounted
			stream.WriteString("#.system:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
	)).Equals(nil, NewReplyError("target #.system:Eval does not exist"), nil)

	// Test(3) depth data format error
	assert(testRunWithProcessor(true, nil,
		func(ctx Context, name string) Return {
			return ctx.OK("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
			// depth type error
			stream.WriteInt64(3)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
	)).Equals(nil, NewProtocolError(ErrStringBadStream), nil)

	// Test(4) depth is overflow
	ret4, error4, panic4 := testRunWithProcessor(true, nil,
		func(ctx Context, name string) Return {
			return ctx.OK("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
			//  depth is overflow
			stream.WriteUint64(17)
			stream.WriteString("#")
			stream.WriteString("world")
			return stream
		},
	)
	assert(ret4, panic4).IsNil()
	assert(error4.GetKind()).Equals(ErrorKindReply)
	assert(error4.GetMessage()).
		Equals("call #.test:Eval level(17) overflows")
	assert(strings.Contains(error4.GetDebug(), "#.test:Eval")).IsTrue()
	assert(strings.Contains(error4.GetDebug(), "type_test.go")).IsTrue()

	// Test(5) execFrom data format error
	assert(testRunWithProcessor(true, nil,
		func(ctx Context, name string) Return {
			return ctx.OK("hello " + name)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
			stream.WriteUint64(3)
			// execFrom data format error
			stream.WriteBool(true)
			stream.WriteString("world")
			return stream
		},
	)).Equals(nil, NewProtocolError(ErrStringBadStream), nil)

	// Test(6) ok call with all type value
	assert(testRunWithProcessor(true, nil,
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
		NewReplyError("#.test:Eval reply arguments does not match"),
		nil,
	)

	// Test(8) error with 1st param
	ret8, error8, panic8 := testRunWithProcessor(true, nil,
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
		"#.test:Eval reply arguments does not match\n" +
			"want: #.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
			"got: #.test:Eval(rpc.Context, rpc.Int64, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return",
	)

	assert(strings.Contains(error8.GetDebug(), "#.test:Eval")).IsTrue()
	assert(strings.Contains(error8.GetDebug(), "type_test.go:")).IsTrue()

	// Test(9) error with 2nd param
	assert(testRunWithProcessor(false, nil,
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
		NewReplyError("#.test:Eval reply arguments does not match"),
		nil,
	)

	// Test(10) error with 2nd param
	ret10, error10, panic10 := testRunWithProcessor(true, nil,
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
		"#.test:Eval reply arguments does not match\n" +
			"want: #.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
			"got: #.test:Eval(rpc.Context, rpc.Bool, rpc.Bool, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return",
	)
	assert(strings.Contains(error10.GetDebug(), "#.test:Eval")).IsTrue()
	assert(strings.Contains(error10.GetDebug(), "type_test.go:")).IsTrue()

	// Test(11) error with 3rd param
	assert(testRunWithProcessor(false, nil,
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
		NewReplyError("#.test:Eval reply arguments does not match"),
		nil,
	)

	// Test(12) error with 3rd param
	ret12, error12, panic12 := testRunWithProcessor(true, nil,
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
		"#.test:Eval reply arguments does not match\n" +
			"want: #.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
			"got: #.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Bool, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return",
	)
	assert(strings.Contains(error12.GetDebug(), "#.test:Eval")).IsTrue()
	assert(strings.Contains(error12.GetDebug(), "type_test.go:")).IsTrue()

	// Test(13) error with 4th param
	assert(testRunWithProcessor(false, nil,
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
		NewReplyError("#.test:Eval reply arguments does not match"),
		nil,
	)

	// Test(14) error with 4th param
	ret14, error14, panic14 := testRunWithProcessor(true, nil,
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
		"#.test:Eval reply arguments does not match\n" +
			"want: #.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
			"got: #.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Bool, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return",
	)
	assert(strings.Contains(error14.GetDebug(), "#.test:Eval")).IsTrue()
	assert(strings.Contains(error14.GetDebug(), "type_test.go:")).IsTrue()

	// Test(15) error with 5th param
	assert(testRunWithProcessor(false, nil,
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
		NewReplyError("#.test:Eval reply arguments does not match"),
		nil,
	)

	// Test(16) error with 5th param
	ret16, error16, panic16 := testRunWithProcessor(true, nil,
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
		"#.test:Eval reply arguments does not match\n" +
			"want: #.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
			"got: #.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.Bool, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return",
	)
	assert(strings.Contains(error16.GetDebug(), "#.test:Eval")).IsTrue()
	assert(strings.Contains(error16.GetDebug(), "type_test.go:")).IsTrue()

	// Test(17) error with 6th param
	assert(testRunWithProcessor(false, nil,
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
		NewReplyError("#.test:Eval reply arguments does not match"),
		nil,
	)

	// Test(18) error with 6th param
	ret18, error18, panic18 := testRunWithProcessor(true, nil,
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
		"#.test:Eval reply arguments does not match\n" +
			"want: #.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
			"got: #.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bool, rpc.Array, rpc.Map) rpc.Return",
	)
	assert(strings.Contains(error18.GetDebug(), "#.test:Eval")).IsTrue()
	assert(strings.Contains(error18.GetDebug(), "type_test.go:")).IsTrue()

	// Test(19) error with 7th param
	assert(testRunWithProcessor(false, nil,
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
		NewReplyError("#.test:Eval reply arguments does not match"),
		nil,
	)

	// Test(20) error with 6th param
	ret20, error20, panic20 := testRunWithProcessor(true, nil,
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
		"#.test:Eval reply arguments does not match\n" +
			"want: #.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
			"got: #.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Bool, rpc.Map) rpc.Return",
	)
	assert(strings.Contains(error20.GetDebug(), "#.test:Eval")).IsTrue()
	assert(strings.Contains(error20.GetDebug(), "type_test.go:")).IsTrue()

	// Test(21) error with 8th param
	assert(testRunWithProcessor(false, nil,
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
		NewReplyError("#.test:Eval reply arguments does not match"),
		nil,
	)

	// Test(22) error with 8th param
	ret22, error22, panic22 := testRunWithProcessor(true, nil,
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
		"#.test:Eval reply arguments does not match\n" +
			"want: #.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
			"got: #.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Bool) rpc.Return",
	)
	assert(strings.Contains(error22.GetDebug(), "#.test:Eval")).IsTrue()
	assert(strings.Contains(error22.GetDebug(), "type_test.go:")).IsTrue()

	// Test(23) nil rpcBytes
	assert(testRunWithProcessor(true, nil,
		func(ctx Context, a Bytes) Return {
			if a != nil {
				return ctx.Error(errors.New("param is not nil"))
			}
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(nil)
			return stream
		},
	)).Equals(true, nil, nil)

	// Test(24) nil rpcArray
	assert(testRunWithProcessor(true, nil,
		func(ctx Context, a Array) Return {
			if a != nil {
				return ctx.Error(errors.New("param is not nil"))
			}
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(nil)
			return stream
		},
	)).Equals(true, nil, nil)

	// Test(25) nil rpcMap
	assert(testRunWithProcessor(true, nil,
		func(ctx Context, a Map) Return {
			if a != nil {
				return ctx.Error(errors.New("param is not nil"))
			}
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(nil)
			return stream
		},
	)).Equals(true, nil, nil)

	// Test(26) unsupported type
	assert(testRunWithProcessor(false, nil,
		func(ctx Context, a bool) Return {
			return ctx.OK(a)
		},
		func(processor *Processor) *Stream {
			replyNode := processor.repliesMap["#.test:Eval"]
			replyNode.argTypes[1] = reflect.ValueOf(int16(0)).Type()
			stream := NewStream()
			stream.WriteString("#.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			return stream
		},
	)).Equals(
		nil,
		NewReplyError("#.test:Eval reply arguments does not match"),
		nil,
	)

	// Test(27) test
	ret27, error27, panic27 := testRunWithProcessor(true, nil,
		func(ctx Context, bVal bool, rpcMap Map) Return {
			return ctx.OK(bVal)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
		"#.test:Eval reply arguments does not match\n" +
			"want: #.test:Eval(rpc.Context, rpc.Bool, rpc.Map) rpc.Return\n" +
			"got: #.test:Eval(rpc.Context, <nil>, rpc.Map, <nil>) rpc.Return",
	)
	assert(strings.Contains(error27.GetDebug(), "#.test:Eval")).IsTrue()
	assert(strings.Contains(error27.GetDebug(), "type_test.go:")).IsTrue()

	// Test(28) badStream
	assert(testRunWithProcessor(true, nil,
		func(ctx Context, bVal bool, rpcMap Map) Return {
			return ctx.OK(bVal)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write("helloWorld")
			stream.SetWritePos(stream.GetWritePos() - 1)
			return stream
		},
	)).Equals(nil, NewProtocolError(ErrStringBadStream), nil)

	// Test(29) call function error
	ret29, error29, panic29 := testRunWithProcessor(false, nil,
		func(ctx Context, bVal bool) Return {
			if bVal {
				panic("this is a error")
			}
			return ctx.OK(bVal)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			return stream
		},
	)
	assert(ret29, error29).IsNil()
	assert(panic29).IsNotNil()
	if panic29 != nil {
		assert(panic29.GetKind()).Equals(ErrorKindReplyPanic)
		assert(panic29.GetMessage()).
			Equals("runtime error: this is a error")
		assert(strings.Contains(panic29.GetDebug(), "#.test:Eval")).IsTrue()
		assert(strings.Contains(panic29.GetDebug(), "thread_test.go")).IsTrue()
	}

	// Test(30) call function error
	ret30, error30, panic30 := testRunWithProcessor(true, nil,
		func(ctx Context, bVal bool) Return {
			if bVal {
				panic("this is a error")
			}
			return ctx.OK(bVal)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			return stream
		},
	)
	assert(ret30, error30).IsNil()
	assert(panic30).IsNotNil()
	if panic30 != nil {
		assert(panic30.GetKind()).Equals(ErrorKindReplyPanic)
		assert(panic30.GetMessage()).
			Equals("runtime error: this is a error")
		assert(strings.Contains(panic30.GetDebug(), "#.test:Eval")).IsTrue()
		assert(strings.Contains(panic30.GetDebug(), "thread_test.go")).IsTrue()
	}

	// Test(31) return TransportError to make onEvalFinish panic
	ret31, error31, panic31 := testRunWithProcessor(true, nil,
		func(ctx Context, bVal bool) Return {
			return ctx.Error(NewTransportError("it makes onEvalFinish panic"))
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			return stream
		},
	)
	assert(ret31, error31).Equals(nil, nil)
	assert(panic31).IsNotNil()
	if panic31 != nil {
		assert(panic31.GetKind()).Equals(ErrorKindKernelPanic)
		assert(panic31.GetMessage()).
			Equals("kernel error: it makes onEvalFinish panic")
		assert(strings.Contains(panic31.GetDebug(), "type_test.go")).IsTrue()
	}

	// Test(32) return without ctx
	ret32, error32, panic32 := testRunWithProcessor(true, nil,
		func(ctx Context, bVal bool) Return {
			return Return{}
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.Write(true)
			return stream
		},
	)
	assert(ret32, error32).Equals(nil, nil)
	assert(panic32).IsNotNil()
	if panic32 != nil {
		assert(panic32.GetKind()).Equals(ErrorKindReplyPanic)
		assert(panic32.GetMessage()).
			Equals("reply must return through Context.OK or Context.Error")
		assert(strings.Contains(panic32.GetDebug(), "type_test.go")).IsTrue()
	} else {
		assert().Fail("nil)")
	}

	// Test(33) ok call with  cache
	assert(testRunWithProcessor(true, &testFuncCache{},
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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

	// Test(34) stream is not finish
	ret34, error34, panic34 := testRunWithProcessor(true, nil,
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
			// error
			stream.Write(true)
			return stream
		},
	)
	assert(ret34, panic34).Equals(nil, nil)
	assert(error34).IsNotNil()
	assert(error34.GetMessage())

	assert(error34.GetKind()).Equals(ErrorKindReply)
	assert(error34.GetMessage()).Equals(
		"#.test:Eval reply arguments does not match\n" +
			"want: #.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map) rpc.Return\n" +
			"got: #.test:Eval(rpc.Context, rpc.Bool, rpc.Int64, rpc.Uint64, " +
			"rpc.Float64, rpc.String, rpc.Bytes, rpc.Array, rpc.Map, rpc.Bool) " +
			"rpc.Return",
	)
	assert(strings.Contains(error34.GetDebug(), "#.test:Eval")).IsTrue()
	assert(strings.Contains(error34.GetDebug(), "type_test.go:")).IsTrue()

	// Test(35) bad stream
	assert(testRunWithProcessor(true, &testFuncCache{},
		func(ctx Context,
			b bool, i int64, u uint64, f float64, s string,
			x Bytes, a Array, m Map,
		) Return {
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
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
			// error
			stream.Write(true)
			stream.Write(Map{"name": "world"})
			stream.SetWritePos(stream.GetWritePos() - 2)

			return stream
		},
	)).Equals(nil, NewProtocolError(ErrStringBadStream), nil)

	// Test(36) sequence error
	ret36, error36, panic36 := testRunWithProcessor(true, &testFuncCache{},
		func(ctx Context) Return {
			ctx.OK(true)
			// hack sequence error
			ctx.id = ctx.id + 1
			return ctx.OK(true)
		},
		func(_ *Processor) *Stream {
			stream := NewStream()
			stream.WriteString("#.test:Eval")
			stream.WriteUint64(3)
			stream.WriteString("#")
			return stream
		},
	)

	assert(ret36, error36).IsNil()
	assert(panic36).IsNotNil()

	assert(panic36.GetKind()).Equals(ErrorKindKernelPanic)
	assert(panic36.GetMessage()).Equals("internal error")
	assert(strings.Contains(panic36.GetDebug(), "goroutine"))
	assert(strings.Contains(panic36.GetDebug(), "[running]"))
	assert(strings.Contains(panic36.GetDebug(), "Eval"))
	assert(strings.Contains(panic36.GetDebug(), "thread.go"))
}
