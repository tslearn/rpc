package lab

import "time"

type threadPool struct {
	freeThreads chan int
}

func newThreadPool(size int) *threadPool {
	ret := &threadPool{
		freeThreads: make(chan int, size),
	}

	for i := 0; i < size; i++ {
		ret.freeThreads <- i
	}

	return ret
}

func (p *threadPool) Test() {
	v := p.freeThreads
	time.Sleep(10 * time.Millisecond)
	p.freeThreads <- v
}
