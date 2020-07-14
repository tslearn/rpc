package internal

// rpcThreadPool
type rpcThreadPool struct {
	isRunning   bool
	processor   *RPCProcessor
	threads     []*rpcThread
	freeThreads chan *rpcThread
	RPCLock
}

func newThreadPool(processor *RPCProcessor) *rpcThreadPool {
	ret := &rpcThreadPool{
		isRunning: true,
		processor: processor,
		threads: make(
			[]*rpcThread,
			numOfThreadPerThreadPool,
			numOfThreadPerThreadPool,
		),
		freeThreads: make(chan *rpcThread, numOfThreadPerThreadPool),
	}

	for i := 0; i < numOfThreadPerThreadPool; i++ {
		thread := newThread(ret)
		ret.threads[i] = thread
		ret.freeThreads <- thread
	}

	return ret
}

func (p *rpcThreadPool) stop() (bool, []string) {
	errList := make([]string, 0)
	return p.CallWithLock(func() interface{} {
		if !p.isRunning {
			return false
		} else {
			p.isRunning = false
			numOfThreads := len(p.threads)
			closeCH := make(chan string, numOfThreads)
			// stop threads
			for i := 0; i < numOfThreads; i++ {
				go func(idx int) {
					if !p.threads[idx].stop() && p.threads[idx].execReplyNode != nil {
						closeCH <- p.threads[idx].execReplyNode.debugString
					} else {
						closeCH <- ""
					}
				}(i)
			}

			// wait all thread stop
			for i := 0; i < numOfThreads; i++ {
				if str := <-closeCH; str != "" {
					errList = append(errList, str)
				}
			}

			return len(errList) == 0
		}
	}).(bool), errList
}

func (p *rpcThreadPool) allocThread() *rpcThread {
	return <-p.freeThreads
}

func (p *rpcThreadPool) freeThread(thread *rpcThread) {
	if thread != nil {
		p.freeThreads <- thread
	}
}
