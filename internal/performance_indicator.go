package internal

import (
	"sync/atomic"
	"time"
)

// rpcPerformanceIndicator ...
type rpcPerformanceIndicator struct {
	failArray    [8]int64
	successArray [8]int64
	lastTotal    int64
	lastTime     time.Time
	Lock
}

// newPerformanceIndicator ...
func newPerformanceIndicator() *rpcPerformanceIndicator {
	return &rpcPerformanceIndicator{
		failArray:    [8]int64{},
		successArray: [8]int64{},
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
		total := int64(0)
		for i := 0; i < len(p.failArray); i++ {
			total += atomic.LoadInt64(&p.failArray[i])
		}
		for i := 0; i < len(p.successArray); i++ {
			total += atomic.LoadInt64(&p.successArray[i])
		}
		deltaCount := total - p.lastTotal
		deltaTime := now.Sub(p.lastTime)

		if deltaTime <= 0 {
			speed = 0
			duration = time.Duration(0)
		} else if deltaCount < 0 {
			speed = 0
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
func (p *rpcPerformanceIndicator) Count(duration time.Duration, success bool) {
	idx := 0

	if duration < 5*time.Millisecond {
		idx = 0
	} else if duration < 20*time.Millisecond {
		idx = 1
	} else if duration < 50*time.Millisecond {
		idx = 2
	} else if duration < 100*time.Millisecond {
		idx = 3
	} else if duration < 200*time.Millisecond {
		idx = 4
	} else if duration < 500*time.Millisecond {
		idx = 5
	} else if duration < 1000*time.Millisecond {
		idx = 6
	} else {
		idx = 7
	}

	if success {
		atomic.AddInt64(&p.successArray[idx], 1)
	} else {
		atomic.AddInt64(&p.failArray[idx], 1)
	}
}
