package internal

import (
	"errors"
	"testing"
)

func TestContextObject_OK(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	source1 := ""
	assert(testRunWithSubscribePanic(func() {
		_, source1 = Context{}.OK(true), GetFileLine(0)
	})).Equals(
		NewReplyPanic("Context is illegal in current goroutine").AddDebug(source1),
	)

	// Test(2) return value type error
	source2 := ""
	assert(testRunOnContext(true, func(_ *Processor, ctx Context) Return {
		// return value type error
		ret, source := ctx.OK(make(chan bool)), GetFileLine(0)
		source2 = source
		return ret
	})).Equals(
		nil,
		nil,
		NewReplyPanic("value type (chan bool) is not supported").
			AddDebug("#.test:Eval "+source2),
	)

	// Test(3) OK
	assert(testRunOnContext(true, func(_ *Processor, ctx Context) Return {
		return ctx.OK(true)
	})).Equals(true, nil, nil)

	// Test(4) rpcThreadReturnStatusAlreadyCalled
	source4 := ""
	assert(testRunOnContext(true, func(_ *Processor, ctx Context) Return {
		// OK called twice
		ctx.OK(true)
		ret, source := ctx.OK(make(chan bool)), GetFileLine(0)
		source4 = source
		return ret
	})).Equals(
		nil,
		nil,
		NewReplyPanic("Context.OK or Context.Error has been called before").
			AddDebug("#.test:Eval "+source4),
	)

	// Test(5) rpcThreadReturnStatusContextError
	source5 := ""
	assert(testRunOnContext(true, func(_ *Processor, ctx Context) Return {
		ctx.OK(true)
		ctx.id = 0
		ret, source := ctx.OK(make(chan bool)), GetFileLine(0)
		source5 = source
		return ret
	})).Equals(
		nil,
		nil,
		NewReplyPanic("Context is illegal in current goroutine").
			AddDebug("#.test:Eval "+source5),
	)
}

func TestContextObject_Error(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	source1 := ""
	assert(testRunWithSubscribePanic(func() {
		ret, source := Context{}.Error(errors.New("error")), GetFileLine(0)
		source1 = source
		assert(ret).Equals(nilReturn)
	})).Equals(NewReplyPanic(
		"Context is illegal in current goroutine",
	).AddDebug(source1))

	// Test(2)
	source2 := ""
	assert(testRunWithSubscribePanic(func() {
		err := NewReplyError("error")
		ret, source := (&Context{thread: nil}).Error(err), GetFileLine(0)
		source2 = source
		assert(ret).Equals(nilReturn)
	})).Equals(
		NewReplyPanic("Context is illegal in current goroutine").AddDebug(source2),
	)

	// Test(3)
	source3 := ""
	assert(testRunOnContext(false, func(_ *Processor, ctx Context) Return {
		ret, source := ctx.Error(NewReplyError("error")), GetFileLine(0)
		source3 = ctx.thread.GetReplyNode().path + " " + source
		assert(ret).Equals(nilReturn)
		return ret
	})).Equals(
		nil,
		NewReplyError("error").AddDebug(source3),
		nil,
	)

	// Test(4)
	source4 := ""
	assert(testRunOnContext(false, func(_ *Processor, ctx Context) Return {
		ret, source := ctx.Error(errors.New("error")), GetFileLine(0)
		source4 = ctx.thread.GetReplyNode().path + " " + source
		assert(ret).Equals(nilReturn)
		return ret
	})).Equals(
		nil,
		NewReplyError("error").AddDebug(source4),
		nil,
	)

	// Test(5)
	source5 := ""
	assert(testRunOnContext(false, func(_ *Processor, ctx Context) Return {
		ret, source := ctx.Error(nil), GetFileLine(0)
		source5 = source
		assert(ret).Equals(nilReturn)
		return ret
	})).Equals(
		nil,
		NewReplyError("argument should not nil").
			AddDebug("#.test:Eval "+source5),
		nil,
	)
}
