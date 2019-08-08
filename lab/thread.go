package lab

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tslearn/rpcc/core"
)

const (
	NumOfThreadPerBlock = 1
	NumOfBlockPerSlot   = 8192
	NumOfThreadGCSwipe  = 8192
	NumOfSlotPerCore    = 4
	NumOfMinSlot        = 4
	NumOfMaxSlot        = 128
)

var count = uint64(0)

var speedCounter = core.NewSpeedCounter()

func consume(pool *rpcThreadPool, stream *core.RPCStream) {
	speedCounter.Add(1)
}

type rpcThread struct {
	ch     chan *core.RPCStream
	status int32
}

func newThread() *rpcThread {
	return &rpcThread{
		ch:     make(chan *core.RPCStream),
		status: 0,
	}
}

func (p *rpcThread) toRun() bool {
	return atomic.CompareAndSwapInt32(&p.status, 0, 2)
}

func (p *rpcThread) toSweep() bool {
	return atomic.CompareAndSwapInt32(&p.status, 2, 1)
}

func (p *rpcThread) toFree() bool {
	return atomic.CompareAndSwapInt32(&p.status, 1, 0)
}

func (p *rpcThread) start(pool *rpcThreadPool) {
	go func() {
		for node := <-p.ch; node != nil; node = <-p.ch {
			consume(pool, node)
			p.toSweep()
		}
	}()
}

func (p *rpcThread) put(stream *core.RPCStream) {
	p.ch <- stream
}

func (p *rpcThread) stop() {
	close(p.ch)
}

type rpcThreadArray [NumOfThreadPerBlock]*rpcThread

type rpcThreadSlot struct {
	size       uint32
	threads    []*rpcThread
	cacheArray chan *rpcThreadArray
	emptyArray chan *rpcThreadArray
	sync.Mutex
}

func newThreadSlot() *rpcThreadSlot {
	size := uint32(NumOfBlockPerSlot * NumOfThreadPerBlock)
	return &rpcThreadSlot{
		size:    size,
		threads: make([]*rpcThread, size, size),
	}
}

func (p *rpcThreadSlot) start(pool *rpcThreadPool) {
	p.Lock()
	defer p.Unlock()

	p.cacheArray = make(chan *rpcThreadArray, NumOfBlockPerSlot)
	p.emptyArray = make(chan *rpcThreadArray, NumOfBlockPerSlot)

	for i := 0; i < NumOfBlockPerSlot; i++ {
		threadArray := &rpcThreadArray{}
		for n := 0; n < NumOfThreadPerBlock; n++ {
			threadPos := i*NumOfThreadPerBlock + n
			thread := newThread()
			p.threads[threadPos] = thread
			threadArray[n] = thread
			thread.start(pool)
		}
		p.cacheArray <- threadArray
	}

	go p.gc()
}

func (p *rpcThreadSlot) stop() {
	p.Lock()
	defer p.Unlock()

	for i := uint32(0); i < p.size; i++ {
		p.threads[i].stop()
		p.threads[i] = nil
	}

	close(p.cacheArray)
	close(p.emptyArray)
}

func (p *rpcThreadSlot) gc() {
	gIndex := 0
	arr := <-p.emptyArray
	arrIndex := 0
	for arr != nil {
		for i := 0; i < NumOfThreadGCSwipe; i++ {
			gIndex = (gIndex + 1) % len(p.threads)
			if p.threads[gIndex].status == 1 {
				if p.threads[gIndex].toFree() {
					arr[arrIndex] = p.threads[gIndex]
					arrIndex++
					if arrIndex == NumOfThreadPerBlock {
						p.cacheArray <- arr
						arr = <-p.emptyArray
						arrIndex = 0
					}
				}
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func (p *rpcThreadSlot) put(stream *core.RPCStream) {

	arrayOfThreads := <-p.cacheArray

	for i := 0; i < NumOfThreadPerBlock; i++ {
		thread := arrayOfThreads[i]
		thread.toRun()
		thread.put(stream)
		arrayOfThreads[i] = nil
	}

	p.emptyArray <- arrayOfThreads
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
type rpcThreadPool struct {
	size  uint32
	slots []*rpcThreadSlot
	sync.Mutex
}

func newThreadPool() *rpcThreadPool {
	size := uint32(runtime.NumCPU() * NumOfSlotPerCore)
	if size < NumOfMinSlot {
		size = NumOfMinSlot
	}
	if size > NumOfMaxSlot {
		size = NumOfMaxSlot
	}
	return &rpcThreadPool{
		size:  size,
		slots: make([]*rpcThreadSlot, size, size),
	}
}

func (p *rpcThreadPool) start() {
	p.Lock()
	defer p.Unlock()

	for i := uint32(0); i < p.size; i++ {
		p.slots[i] = newThreadSlot()
		p.slots[i].start(p)
	}
}

func (p *rpcThreadPool) stop() {
	p.Lock()
	defer p.Unlock()

	for i := uint32(0); i < p.size; i++ {
		p.slots[i].stop()
		p.slots[i] = nil
	}
}

func (p *rpcThreadPool) put(
	stream *core.RPCStream,
	goroutineFixedRand *rand.Rand,
) {
	// get a random uint32
	randUint32 := uint32(0)
	if goroutineFixedRand == nil {
		randUint32 = core.GetRandUint32()
	} else {
		randUint32 = goroutineFixedRand.Uint32()
	}
	// put stream in a random slot
	p.slots[randUint32%p.size].put(stream)
}

func ThreadPoolProfile() {
	stream := core.NewRPCStream()
	pools := newThreadPool()
	pools.start()
	time.Sleep(3 * time.Second)

	startTime := time.Now()

	n := 4
	finish := make(chan bool, n)

	speedCounter.Calculate()

	for i := 0; i < n; i++ {
		go func() {
			r := rand.New(rand.NewSource(rand.Int63()))
			for j := 0; j < 2000000; j++ {
				pools.put(stream, r)
			}
			finish <- true
		}()
	}

	for i := 0; i < n; i++ {
		<-finish
	}

	fmt.Println("SpeedCounter : ", speedCounter.Calculate())
	fmt.Println("time used: ", time.Now().Sub(startTime))
}
