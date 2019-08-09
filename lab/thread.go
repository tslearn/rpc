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
	NumOfThreadPerBlock = 16
	NumOfBlockPerSlot   = 512
	NumOfThreadGCSwipe  = 8192
	NumOfSlotPerCore    = 4
	NumOfMinSlot        = 4
	NumOfMaxSlot        = 128
)

////////////////////////////////////////////////////////////////////////////////
///////////////  rpcThread                                       ///////////////
////////////////////////////////////////////////////////////////////////////////
var speedCounter = core.NewSpeedCounter()

func consume(processor *rpcProcessor, stream *core.RPCStream) {
	speedCounter.Add(1)
}

type rpcThread struct {
	processor *rpcProcessor
	ch        chan *core.RPCStream
	execNS    int64

	execDepth      uint64
	execSuccessful bool
}

func newThread(processor *rpcProcessor) *rpcThread {
	ret := rpcThread{
		processor: processor,
		ch:        make(chan *core.RPCStream),
		execNS:    0,
	}

	return &ret
}

func (p *rpcThread) getRunDuration() time.Duration {
	timeNS := atomic.LoadInt64(&p.execNS)
	if timeNS == 0 {
		return 0
	}
	return core.TimeSpanFrom(timeNS)
}

func (p *rpcThread) toRun() bool {
	return atomic.CompareAndSwapInt64(&p.execNS, 0, core.TimeNowNS())
}

func (p *rpcThread) toSweep() {
	//p.execDepth = 0
	//p.execArgs = p.execArgs[:0]
	atomic.StoreInt64(&p.execNS, -1)
}

func (p *rpcThread) toFree() bool {
	return atomic.CompareAndSwapInt64(&p.execNS, -1, 0)
}

func (p *rpcThread) start() {
	go func() {
		for node := <-p.ch; node != nil; node = <-p.ch {
			consume(p.processor, node)
			p.toSweep()
		}
	}()
}

func (p *rpcThread) stop() {
	close(p.ch)
}

func (p *rpcThread) put(stream *core.RPCStream) {
	p.ch <- stream
}

////////////////////////////////////////////////////////////////////////////////
///////////////  rpcThreadSlot                                   ///////////////
////////////////////////////////////////////////////////////////////////////////
type rpcThreadArray [NumOfThreadPerBlock]*rpcThread
type rpcThreadSlot struct {
	isRunning  bool
	gcFinish   chan bool
	threads    []*rpcThread
	cacheArray chan *rpcThreadArray
	emptyArray chan *rpcThreadArray
	sync.Mutex
}

func newThreadSlot() *rpcThreadSlot {
	size := uint32(NumOfBlockPerSlot * NumOfThreadPerBlock)
	return &rpcThreadSlot{
		isRunning: false,
		threads:   make([]*rpcThread, size, size),
	}
}

func (p *rpcThreadSlot) start(processor *rpcProcessor) bool {
	p.Lock()
	defer p.Unlock()

	if !p.isRunning {
		p.gcFinish = make(chan bool)
		p.cacheArray = make(chan *rpcThreadArray, NumOfBlockPerSlot)
		p.emptyArray = make(chan *rpcThreadArray, NumOfBlockPerSlot)
		for i := 0; i < NumOfBlockPerSlot; i++ {
			threadArray := &rpcThreadArray{}
			for n := 0; n < NumOfThreadPerBlock; n++ {
				thread := newThread(processor)
				p.threads[i*NumOfThreadPerBlock+n] = thread
				threadArray[n] = thread
				thread.start()
			}
			p.cacheArray <- threadArray
		}
		go p.gc()
		p.isRunning = true
		return true
	} else {
		return false
	}
}

func (p *rpcThreadSlot) stop() bool {
	p.Lock()
	defer p.Unlock()

	if p.isRunning {
		close(p.cacheArray)
		close(p.emptyArray)
		p.isRunning = false
		<-p.gcFinish
		for i := 0; i < len(p.threads); i++ {
			p.threads[i].stop()
			p.threads[i] = nil
		}
		return true
	} else {
		return false
	}
}

func (p *rpcThreadSlot) gc() {
	gIndex := 0
	threadArray := <-p.emptyArray
	arrIndex := 0
	totalThreads := len(p.threads)
	for p.isRunning {
		for i := 0; i < NumOfThreadGCSwipe; i++ {
			gIndex = (gIndex + 1) % totalThreads
			if p.threads[gIndex].execNS == -1 {
				if p.threads[gIndex].toFree() {
					threadArray[arrIndex] = p.threads[gIndex]
					arrIndex++
					if arrIndex == NumOfThreadPerBlock {
						p.cacheArray <- threadArray
						threadArray = <-p.emptyArray
						arrIndex = 0
						if threadArray == nil {
							break
						}
					}
				}
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	p.gcFinish <- true
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
type rpcProcessor struct {
	isRunning    bool
	logger       *core.Logger
	slots        []*rpcThreadSlot
	maxNodeDepth uint64
	maxCallDepth uint64
	sync.Mutex
}

func newProcessor(
	logger *core.Logger,
	maxNodeDepth uint,
	maxCallDepth uint,
) *rpcProcessor {
	numOfSlots := uint32(runtime.NumCPU() * NumOfSlotPerCore)
	if numOfSlots < NumOfMinSlot {
		numOfSlots = NumOfMinSlot
	}
	if numOfSlots > NumOfMaxSlot {
		numOfSlots = NumOfMaxSlot
	}

	ret := &rpcProcessor{
		isRunning:    false,
		logger:       logger,
		slots:        make([]*rpcThreadSlot, numOfSlots, numOfSlots),
		maxNodeDepth: uint64(maxNodeDepth),
		maxCallDepth: uint64(maxCallDepth),
	}

	return ret
}

func (p *rpcProcessor) start() bool {
	p.Lock()
	defer p.Unlock()

	if !p.isRunning {
		for i := 0; i < len(p.slots); i++ {
			p.slots[i] = newThreadSlot()
			p.slots[i].start(p)
		}
		p.isRunning = true
		return true
	} else {
		return false
	}
}

func (p *rpcProcessor) stop() bool {
	p.Lock()
	defer p.Unlock()

	if p.isRunning {
		for i := 0; i < len(p.slots); i++ {
			p.slots[i].stop()
			p.slots[i] = nil
		}
		p.isRunning = false
		return true
	} else {
		return false
	}
}

func (p *rpcProcessor) put(
	stream *core.RPCStream,
	goroutineFixedRand *rand.Rand,
) bool {
	// get a random uint32
	randInt := 0
	if goroutineFixedRand == nil {
		randInt = int(core.GetRandUint32())
	} else {
		randInt = goroutineFixedRand.Int()
	}
	// put stream in a random slot
	slot := p.slots[randInt%len(p.slots)]
	if slot != nil {
		slot.put(stream)
		return true
	} else {
		return false
	}
}

func rpcProcessorProfile() {
	stream := core.NewRPCStream()
	pools := newProcessor(core.NewLogger(), 16, 16)
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
