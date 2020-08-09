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
		ret, source := Context{}.OK(true), GetFileLine(0)
		source1 = source
		assert(ret).Equals(nilReturn)
	})).Equals(
		NewReplyPanic("Context is illegal in current goroutine").AddDebug(source1),
	)

	// Test(2)
	source2 := ""
	assert(testRunWithSubscribePanic(func() {
		ret, source := (&Context{thread: nil}).OK(true), GetFileLine(0)
		source2 = source
		assert(ret).Equals(nilReturn)
	})).Equals(
		NewReplyPanic("Context is illegal in current goroutine").AddDebug(source2),
	)

	// Test(3)
	assert(testRunOnContext(true, func(_ *Processor, ctx Context) Return {
		ret := ctx.OK(true)
		assert(ret).Equals(nilReturn)
		return ret
	})).Equals(true, nil, nil)
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
