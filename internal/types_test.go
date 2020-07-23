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

func getFakeContext(debug bool) *ContextObject {
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
) (interface{}, Error, Error, string) {
	done := make(chan bool)
	ret := interface{}(nil)
	retError := Error(nil)
	retPanic := Error(nil)
	codeSource := ""

	processor := getFakeProcessor(debug)

	if err := processor.AddService(
		"test",
		NewService().
			Reply("Eval", func(ctx Context) Return {
				defer func() {
					done <- true
				}()
				ret, fileLine := fn(ctx), GetFileLine(0)
				codeSource = fileLine
				return ret
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
	return ret, retError, retPanic, codeSource
}
