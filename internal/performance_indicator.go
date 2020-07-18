package internal

import (
	"sync/atomic"
	"time"
)

// rpcPerformanceIndicator ...
type rpcPerformanceIndicator struct {
	failed       int64
	successArray [10]int64
	lastTotal    int64
	lastTime     time.Time
	Lock
}

// newPerformanceIndicator ...
func newPerformanceIndicator() *rpcPerformanceIndicator {
	return &rpcPerformanceIndicator{
		failed:       0,
		successArray: [10]int64{},
		lastTotal:    0,
		lastTime:     TimeNow(),
	}
}

// Calculate ...
func (p *rpcPerformanceIndicator) Calculate(
	now time.Time,
) (speed int64, duration time.Duration) {
	p.DoWithLock(func() {
		// calculate total called
		total := atomic.LoadInt64(&p.failed)
		for i := 0; i < 10; i++ {
			total += atomic.LoadInt64(&p.successArray[i])
		}
		deltaCount := total - p.lastTotal
		deltaTime := now.Sub(p.lastTime)

		if deltaTime <= 0 {
			speed = -1
			duration = time.Duration(0)
		} else if deltaCount < 0 {
			speed = -1
			duration = time.Duration(0)
		} else {
			p.lastTime = now
			p.lastTotal = total
			speed = (deltaCount * int64(time.Second)) / int64(deltaTime)
			duration = deltaTime
		}
	})

	return
}

// Count ...
func (p *rpcPerformanceIndicator) Count(
	duration time.Duration,
	successful bool,
) {
	if successful {
		if duration < 2*time.Millisecond {
			atomic.AddInt64(&p.successArray[0], 1)
		} else if duration < 6*time.Millisecond {
			atomic.AddInt64(&p.successArray[1], 1)
		} else if duration < 10*time.Millisecond {
			atomic.AddInt64(&p.successArray[2], 1)
		} else if duration < 20*time.Millisecond {
			atomic.AddInt64(&p.successArray[3], 1)
		} else if duration < 40*time.Millisecond {
			atomic.AddInt64(&p.successArray[4], 1)
		} else if duration < 60*time.Millisecond {
			atomic.AddInt64(&p.successArray[5], 1)
		} else if duration < 100*time.Millisecond {
			atomic.AddInt64(&p.successArray[6], 1)
		} else if duration < 300*time.Millisecond {
			atomic.AddInt64(&p.successArray[7], 1)
		} else if duration < 1000*time.Millisecond {
			atomic.AddInt64(&p.successArray[8], 1)
		} else {
			atomic.AddInt64(&p.successArray[9], 1)
		}
	} else {
		atomic.AddInt64(&p.failed, 1)
	}
}
