package core

import (
	"sync"
	"sync/atomic"
	"time"
)

type indicator struct {
	failed      int64
	performance [10]int64
	origins     *sync.Map
	qps         int64
	lastTotal   int64
	lastNS      int64
	sync.Mutex
}

func newIndicator() *indicator {
	return &indicator{}
}

func (p *indicator) setRecordOrigin(isRecord bool) {
	p.Lock()
	defer p.Unlock()
	if isRecord {
		p.origins = &sync.Map{}
	} else {
		p.origins = nil
	}
}

// onTimer this func must be called in a fixed thread
func (p *indicator) onTimer(nowNS int64) {
	// calculate total called
	total := p.failed
	for i := 0; i < 10; i++ {
		total += p.performance[i]
	}

	// calculate delta count and time
	deltaCount := total - p.lastTotal
	deltaNS := nowNS - p.lastNS

	// if delta count and time is acceptable, then update qps
	if deltaNS > 0 && deltaCount >= 0 {
		p.qps = (deltaCount * int64(time.Second)) / deltaNS
		p.lastTotal += deltaCount
		p.lastNS += deltaNS
	}
}

func (p *indicator) count(
	duration time.Duration,
	origin string,
	successful bool,
) {
	if successful {
		if duration < 2*time.Millisecond {
			atomic.AddInt64(&p.performance[0], 1)
		} else if duration < 6*time.Millisecond {
			atomic.AddInt64(&p.performance[1], 1)
		} else if duration < 10*time.Millisecond {
			atomic.AddInt64(&p.performance[2], 1)
		} else if duration < 20*time.Millisecond {
			atomic.AddInt64(&p.performance[3], 1)
		} else if duration < 40*time.Millisecond {
			atomic.AddInt64(&p.performance[4], 1)
		} else if duration < 60*time.Millisecond {
			atomic.AddInt64(&p.performance[5], 1)
		} else if duration < 100*time.Millisecond {
			atomic.AddInt64(&p.performance[6], 1)
		} else if duration < 300*time.Millisecond {
			atomic.AddInt64(&p.performance[7], 1)
		} else if duration < 1000*time.Millisecond {
			atomic.AddInt64(&p.performance[8], 1)
		} else {
			atomic.AddInt64(&p.performance[9], 1)
		}
	} else {
		atomic.AddInt64(&p.failed, 1)
	}

	if originMap := p.origins; originMap != nil {
		if _, ok := p.origins.Load(origin); !ok {
			p.origins.Store(origin, true)
		}
	}
}
