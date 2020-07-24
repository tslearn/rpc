package internal

import "unsafe"

func getFakeOnEvalFinish() func(*rpcThread, *Stream) {
	return func(thread *rpcThread, stream *Stream) {}
}

func getFakeProcessor(debug bool) *Processor {
	return NewProcessor(debug, 1024, 32, 32, nil)
}

func getFakeThread(debug bool) *rpcThread {
	return newThread(getFakeProcessor(debug), getFakeOnEvalFinish())
}

func getFakeContext(debug bool) Context {
	return &ContextObject{thread: unsafe.Pointer(getFakeThread(debug))}
}

func testRunAndCatchPanic(fn func()) Error {
	ch := make(chan Error, 1)
	sub := SubscribePanic(func(err Error) {
		ch <- err
	})
	defer sub.Close()

	fn()

	select {
	case err := <-ch:
		return err
	default:
		return nil
	}
}

func testRunOnContext(
	debug bool,
	fn func(ctx Context) Return,
) (interface{}, Error, Error) {
	done := make(chan bool)
	ret := interface{}(nil)
	retError := Error(nil)
	retPanic := Error(nil)

	processor := getFakeProcessor(debug)

	if err := processor.AddService(
		"test",
		NewService().
			Reply("Eval", func(ctx Context) Return {
				defer func() {
					done <- true
				}()
				return fn(ctx)
			}),
		"",
	); err != nil {
		panic(err)
	}

	if err := processor.Start(
		func(stream *Stream) {
			stream.SetReadPosToBodyStart()
			if stream.GetStreamKind() == StreamKindResponseOK {
				if v, ok := stream.Read(); ok {
					ret = v
				} else {
					panic("internal error")
				}
			} else {
				if errKind, ok := stream.ReadUint64(); !ok {
					panic("internal error")
				} else if message, ok := stream.ReadString(); !ok {
					panic("internal error")
				} else if debug, ok := stream.ReadString(); !ok {
					panic("internal error")
				} else {
					err := NewError(ErrorKind(errKind), message, debug)
					if stream.GetStreamKind() == StreamKindResponseError {
						retError = err
					} else if stream.GetStreamKind() == StreamKindResponsePanic {
						retPanic = err
					}
				}
			}
			stream.Release()
		},
	); err != nil {
		panic(err)
	}

	// put the stream
	stream := NewStream()
	stream.SetStreamKind(StreamKindRequest)
	stream.WriteString("$.test:Eval")
	stream.WriteUint64(3)
	stream.WriteString("#")
	processor.PutStream(stream)

	// wait for finish
	<-done
	if err := processor.Stop(); err != nil {
		panic(err)
	}
	return ret, retError, retPanic
}

func runWithPanicCatch(fn func()) (ret interface{}) {
	defer func() {
		ret = recover()
	}()

	fn()
	return
}

func runWithProcessor(
	isDebug bool,
	handler interface{},
	fnCache ReplyCache,
	getStream func(processor *Processor) *Stream,
) (ret interface{}, retError Error, retPanic Error) {
	done := make(chan bool)
	fnDealStream := func(stream *Stream) {
		stream.SetReadPosToBodyStart()
		if stream.GetStreamKind() == StreamKindResponseOK {
			if v, ok := stream.Read(); ok {
				if ret != nil {
					panic("internal error")
				} else {
					ret = v
					done <- true
				}
			} else {
				panic("internal error")
			}
		} else {
			if errKind, ok := stream.ReadUint64(); !ok {
				panic("internal error")
			} else if message, ok := stream.ReadString(); !ok {
				panic("internal error")
			} else if debug, ok := stream.ReadString(); !ok {
				panic("internal error")
			} else {
				err := NewError(ErrorKind(errKind), message, debug)
				if stream.GetStreamKind() == StreamKindResponseError {
					if retError != nil {
						panic("internal error")
					} else {
						retError = err
						done <- true
					}
				} else if stream.GetStreamKind() == StreamKindResponsePanic {
					if retPanic != nil {
						panic("internal error")
					} else {
						retPanic = err
						done <- true
					}
				}
			}
		}
		stream.Release()
	}

	processor := NewProcessor(
		isDebug,
		8192,
		16,
		16,
		fnCache,
	)
	if err := processor.AddService(
		"test",
		NewService().Reply("Eval", handler),
		"",
	); err != nil {
		panic(err)
	}

	if err := processor.Start(func(stream *Stream) {
		fnDealStream(stream)
	}); err != nil {
		panic(err)
	}

	// wait for finish
	<-done
	if err := processor.Stop(); err != nil {
		panic(err)
	}
	return
}
