package core

import (
	"errors"
	"github.com/rpccloud/rpc/internal/base"
	"testing"
)

func TestContextObject_OK(t *testing.T) {
	assert := base.NewAssert(t)

	// Test(1)
	source1 := ""
	assert(base.TestRunWithSubscribePanic(func() {
		_, source1 = Runtime{}.OK(true), base.GetFileLine(0)
	})).Equals(
		base.NewReplyPanic("Runtime is illegal in current goroutine").AddDebug(source1),
	)

	// Test(2) return value type error
	source2 := ""
	assert(testRunOnContext(true, func(_ *Processor, rt Runtime) Return {
		// return value type error
		ret, source := rt.OK(make(chan bool)), base.GetFileLine(0)
		source2 = source
		return ret
	})).Equals(
		nil,
		nil,
		base.NewReplyPanic("value type is not supported").
			AddDebug("#.test:Eval "+source2),
	)

	// Test(3) OK
	assert(testRunOnContext(true, func(_ *Processor, rt Runtime) Return {
		return rt.OK(true)
	})).Equals(true, nil, nil)

	// Test(4) rpcThreadReturnStatusAlreadyCalled
	source4 := ""
	assert(testRunOnContext(true, func(_ *Processor, rt Runtime) Return {
		// OK called twice
		rt.OK(true)
		ret, source := rt.OK(make(chan bool)), base.GetFileLine(0)
		source4 = source
		return ret
	})).Equals(
		nil,
		nil,
		base.NewReplyPanic("Runtime.OK has been called before").
			AddDebug("#.test:Eval "+source4),
	)
}

func TestContextObject_Error(t *testing.T) {
	assert := base.NewAssert(t)

	// Test(1)
	source1 := ""
	assert(base.TestRunWithSubscribePanic(func() {
		ret, source := Runtime{}.Error(errors.New("error")), base.GetFileLine(0)
		source1 = source
		assert(ret).Equals(emptyReturn)
	})).Equals(base.NewReplyPanic(
		"Runtime is illegal in current goroutine",
	).AddDebug(source1))

	// Test(2)
	source2 := ""
	assert(base.TestRunWithSubscribePanic(func() {
		err := base.NewReplyError("error")
		ret, source := (&Runtime{thread: nil}).Error(err), base.GetFileLine(0)
		source2 = source
		assert(ret).Equals(emptyReturn)
	})).Equals(
		base.NewReplyPanic("Runtime is illegal in current goroutine").AddDebug(source2),
	)

	// Test(3)
	source3 := ""
	assert(testRunOnContext(false, func(_ *Processor, rt Runtime) Return {
		ret, source := rt.Error(base.NewReplyError("error")), base.GetFileLine(0)
		source3 = rt.thread.GetReplyNode().path + " " + source
		assert(ret).Equals(emptyReturn)
		return ret
	})).Equals(
		nil,
		base.NewReplyError("error").AddDebug(source3),
		nil,
	)

	// Test(4)
	source4 := ""
	assert(testRunOnContext(false, func(_ *Processor, rt Runtime) Return {
		ret, source := rt.Error(errors.New("error")), base.GetFileLine(0)
		source4 = rt.thread.GetReplyNode().path + " " + source
		assert(ret).Equals(emptyReturn)
		return ret
	})).Equals(
		nil,
		base.NewReplyError("error").AddDebug(source4),
		nil,
	)

	// Test(5)
	source5 := ""
	assert(testRunOnContext(false, func(_ *Processor, rt Runtime) Return {
		ret, source := rt.Error(nil), base.GetFileLine(0)
		source5 = source
		assert(ret).Equals(emptyReturn)
		return ret
	})).Equals(
		nil,
		base.NewReplyError("argument should not nil").
			AddDebug("#.test:Eval "+source5),
		nil,
	)

	// Test(6) rpcThreadReturnStatusAlreadyCalled
	source6 := ""
	assert(testRunOnContext(true, func(_ *Processor, rt Runtime) Return {
		// OK called twice
		rt.Error(errors.New("error"))
		ret, source := rt.Error(errors.New("error")), base.GetFileLine(0)
		source6 = source
		return ret
	})).Equals(
		nil,
		nil,
		base.NewReplyPanic("Runtime.Error has been called before").
			AddDebug("#.test:Eval "+source6),
	)
}
