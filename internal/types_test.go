package internal

import (
	"time"
)

func getFakeOnEvalFinish() func(*rpcThread, *Stream, bool) {
	return func(thread *rpcThread, stream *Stream, b bool) {}
}

func getFakeProcessor(debug bool) *Processor {
	return NewProcessor(debug, 1024, 32, 32, nil)
}

func getFakeThread(debug bool) *rpcThread {
	return newThread(getFakeProcessor(debug), getFakeOnEvalFinish())
}

func testRunAndCatchPanic(fn func(), timeout time.Duration) Error {
	ch := make(chan Error, 1)
	sub := SubscribePanic(func(err Error) {
		ch <- err
	})
	defer sub.Close()

	fn()

	select {
	case err := <-ch:
		return err
	case <-time.After(timeout):
		return nil
	}
}

func testRunOnContext(debug bool, fn func(ctx Context)) {
	//processor := getFakeProcessor(debug)
	//_ = processor.Start(
	//  func(stream *Stream, ok bool) {
	//    if ok {
	//      atomic.AddUint64(&success, 1)
	//    } else {
	//      atomic.AddUint64(&failed, 1)
	//    }
	//    stream.Release()
	//  },
	//)
	//_ = processor.AddService(
	//  "user",
	//  NewService().
	//    Reply("sayHello", func(
	//      ctx *ContextObject,
	//      name String,
	//    ) *ReturnObject {
	//      return ctx.OK(name)
	//    }),
	//  "",
	//)
}
