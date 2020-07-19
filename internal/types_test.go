package internal

var getFakeOnEvalFinish = func() func(*rpcThread, *Stream, bool) {
	return func(thread *rpcThread, stream *Stream, b bool) {}
}

var getFakeProcessor = func() *Processor {
	return NewProcessor(true, 8192, 32, 32, nil)
}

var getFakeThread = func() *rpcThread {
	return newThread(getFakeProcessor(), getFakeOnEvalFinish())
}
