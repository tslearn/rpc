package internal

import (
	"errors"
	"github.com/rpccloud/rpc/internal/util"
	"testing"
)

func TestContextObject_OK(t *testing.T) {
	assert := util.NewAssert(t)

	// Test(1)
	source1 := ""
	assert(testRunWithSubscribePanic(func() {
		_, source1 = Runtime{}.OK(true), util.GetFileLine(0)
	})).Equals(
		NewReplyPanic("Runtime is illegal in current goroutine").AddDebug(source1),
	)

	// Test(2) return value type error
	source2 := ""
	assert(testRunOnContext(true, func(_ *Processor, rt Runtime) Return {
		// return value type error
		ret, source := rt.OK(make(chan bool)), util.GetFileLine(0)
		source2 = source
		return ret
	})).Equals(
		nil,
		nil,
		NewReplyPanic("value type is not supported").
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
		ret, source := rt.OK(make(chan bool)), util.GetFileLine(0)
		source4 = source
		return ret
	})).Equals(
		nil,
		nil,
		NewReplyPanic("Runtime.OK has been called before").
			AddDebug("#.test:Eval "+source4),
	)
}

func TestContextObject_Error(t *testing.T) {
	assert := util.NewAssert(t)

	// Test(1)
	source1 := ""
	assert(testRunWithSubscribePanic(func() {
		ret, source := Runtime{}.Error(errors.New("error")), util.GetFileLine(0)
		source1 = source
		assert(ret).Equals(emptyReturn)
	})).Equals(NewReplyPanic(
		"Runtime is illegal in current goroutine",
	).AddDebug(source1))

	// Test(2)
	source2 := ""
	assert(testRunWithSubscribePanic(func() {
		err := NewReplyError("error")
		ret, source := (&Runtime{thread: nil}).Error(err), util.GetFileLine(0)
		source2 = source
		assert(ret).Equals(emptyReturn)
	})).Equals(
		NewReplyPanic("Runtime is illegal in current goroutine").AddDebug(source2),
	)

	// Test(3)
	source3 := ""
	assert(testRunOnContext(false, func(_ *Processor, rt Runtime) Return {
		ret, source := rt.Error(NewReplyError("error")), util.GetFileLine(0)
		source3 = rt.thread.GetReplyNode().path + " " + source
		assert(ret).Equals(emptyReturn)
		return ret
	})).Equals(
		nil,
		NewReplyError("error").AddDebug(source3),
		nil,
	)

	// Test(4)
	source4 := ""
	assert(testRunOnContext(false, func(_ *Processor, rt Runtime) Return {
		ret, source := rt.Error(errors.New("error")), util.GetFileLine(0)
		source4 = rt.thread.GetReplyNode().path + " " + source
		assert(ret).Equals(emptyReturn)
		return ret
	})).Equals(
		nil,
		NewReplyError("error").AddDebug(source4),
		nil,
	)

	// Test(5)
	source5 := ""
	assert(testRunOnContext(false, func(_ *Processor, rt Runtime) Return {
		ret, source := rt.Error(nil), util.GetFileLine(0)
		source5 = source
		assert(ret).Equals(emptyReturn)
		return ret
	})).Equals(
		nil,
		NewReplyError("argument should not nil").
			AddDebug("#.test:Eval "+source5),
		nil,
	)

	// Test(6) rpcThreadReturnStatusAlreadyCalled
	source6 := ""
	assert(testRunOnContext(true, func(_ *Processor, rt Runtime) Return {
		// OK called twice
		rt.Error(errors.New("error"))
		ret, source := rt.Error(errors.New("error")), util.GetFileLine(0)
		source6 = source
		return ret
	})).Equals(
		nil,
		nil,
		NewReplyPanic("Runtime.Error has been called before").
			AddDebug("#.test:Eval "+source6),
	)
}
