package internal

import (
	"fmt"
	"strings"
)

// rpcThreadPool
type rpcThreadPool struct {
	processor     *RPCProcessor
	threads       []*rpcThread
	freeThreadsCH chan *rpcThread
	RPCLock
}

func newThreadPool(processor *RPCProcessor, size uint) *rpcThreadPool {
	if size == 0 {
		size = 1
	}
	return &rpcThreadPool{
		processor:     processor,
		threads:       make([]*rpcThread, size, size),
		freeThreadsCH: nil,
	}
}

func (p *rpcThreadPool) Start() RPCError {
	return ConvertToRPCError(p.CallWithLock(func() interface{} {
		size := len(p.threads)

		if p.freeThreadsCH != nil {
			return NewRPCError("rpcThreadPool: Start: it has already benn started")
		} else {
			p.freeThreadsCH = make(chan *rpcThread, size)
			for i := 0; i < size; i++ {
				thread := newThread(p.processor, p.freeThreadsCH)
				p.threads[i] = thread
				p.freeThreadsCH <- thread
			}
			return nil
		}
	}))
}

func (p *rpcThreadPool) Stop() RPCError {
	return ConvertToRPCError(p.CallWithLock(func() interface{} {
		if p.freeThreadsCH == nil {
			return NewRPCError("rpcThreadPool: Start: it has already benn stopped")
		} else {
			numOfThreads := len(p.threads)
			closeCH := make(chan string, numOfThreads)

			for i := 0; i < numOfThreads; i++ {
				go func(idx int) {
					if !p.threads[idx].Stop() && p.threads[idx].execReplyNode != nil {
						closeCH <- p.threads[idx].execReplyNode.debugString
					} else {
						closeCH <- ""
					}
					p.threads[idx] = nil
				}(i)
			}

			// wait all thread stop
			errMap := make(map[string]int)
			for i := 0; i < numOfThreads; i++ {
				if errString := <-closeCH; errString != "" {
					if v, ok := errMap[errString]; ok {
						errMap[errString] = v + 1
					} else {
						errMap[errString] = 1
					}
				}
			}

			errList := make([]string, 0)

			for k, v := range errMap {
				if v > 1 {
					errList = append(errList, fmt.Sprintf(
						"%s (%d routines)",
						k,
						v,
					))
				} else {
					errList = append(errList, fmt.Sprintf(
						"%s (%d routine)",
						k,
						v,
					))
				}
			}

			if len(errList) > 0 {
				return NewRPCError(ConcatString(
					"rpcThreadPool: Stop: The following routine still running: \n\t",
					strings.Join(errList, "\n\t"),
				))
			} else {
				return nil
			}
		}
	}))
}

func (p *rpcThreadPool) PutStream(stream *RPCStream) bool {
	thread := <-p.freeThreadsCH
	return thread.PutStream(stream)
}

//// PutStream ...
//func (p *RPCProcessor) PutStream(stream *RPCStream) bool {
//  // PutStream stream in a random thread pool
//  threadPool := p.threadPools[int(rand.Int31())%len(p.threadPools)]
//  if threadPool != nil {
//    if thread := threadPool.allocThread(); thread != nil {
//      thread.put(stream)
//      return true
//    }
//    return false
//  }
//
//  return false
//}
