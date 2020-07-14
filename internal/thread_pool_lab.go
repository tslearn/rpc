package internal

import (
	"fmt"
	"strings"
	"sync/atomic"
)

const numOfThreadGroup = 64

// rpcThreadPool
type rpcThreadPoolLab struct {
	processor          *RPCProcessor
	threads            []*rpcThread
	freeThreadsCHGroup []chan *rpcThread
	readPos            uint64
	writePos           uint64
	RPCLock
}

func newThreadPoolLab(processor *RPCProcessor, size uint) *rpcThreadPoolLab {
	if size == 0 {
		size = 1
	}
	size = ((size + numOfThreadGroup - 1) / numOfThreadGroup) * numOfThreadGroup
	return &rpcThreadPoolLab{
		processor:          processor,
		threads:            make([]*rpcThread, size, size),
		freeThreadsCHGroup: nil,
	}
}

func (p *rpcThreadPoolLab) Start() RPCError {
	return ConvertToRPCError(p.CallWithLock(func() interface{} {
		size := len(p.threads)

		if p.freeThreadsCHGroup != nil {
			return NewRPCError("rpcThreadPool: Start: it has already benn started")
		} else {
			freeThreadsCHGroup := make([]chan *rpcThread, numOfThreadGroup, numOfThreadGroup)
			p.freeThreadsCHGroup = freeThreadsCHGroup
			for i := 0; i < numOfThreadGroup; i++ {
				p.freeThreadsCHGroup[i] = make(chan *rpcThread, len(p.threads)/numOfThreadGroup)
			}

			for i := 0; i < size; i++ {
				thread := newThread(p.processor, func(thread *rpcThread, stream *RPCStream, success bool) {
					p.processor.callback(stream, success)
					freeThreadsCHGroup[atomic.AddUint64(&p.writePos, 1)%numOfThreadGroup] <- thread
				})
				p.threads[i] = thread
				p.freeThreadsCHGroup[i%numOfThreadGroup] <- thread
			}
			return nil
		}
	}))
}

func (p *rpcThreadPoolLab) Stop() RPCError {
	return ConvertToRPCError(p.CallWithLock(func() interface{} {
		if p.freeThreadsCHGroup == nil {
			return NewRPCError("rpcThreadPool: Start: it has already benn stopped")
		} else {
			p.freeThreadsCHGroup = nil
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

func (p *rpcThreadPoolLab) PutStream(stream *RPCStream) bool {
	if freeThreadsCHGroup := p.freeThreadsCHGroup; freeThreadsCHGroup != nil {
		thread := <-freeThreadsCHGroup[atomic.AddUint64(&p.readPos, 1)%numOfThreadGroup]
		return thread.PutStream(stream)
	} else {
		return false
	}
}
