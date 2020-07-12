package util

import (
	"sync/atomic"
	"time"
)

// SpeedCounter count the speed
type SpeedCounter struct {
	total     int64
	lastCount int64
	lastNS    int64
	AutoLock
}

// NewSpeedCounter create a SpeedCounter
func NewSpeedCounter() *SpeedCounter {
	return &SpeedCounter{
		total:     0,
		lastCount: 0,
		lastNS:    TimeNowNS(),
	}
}

// Add count n times
func (p *SpeedCounter) Add(n int64) {
	atomic.AddInt64(&p.total, n)
}

// Total get the total count
func (p *SpeedCounter) Total() int64 {
	return atomic.LoadInt64(&p.total)
}

// Calculate get the speed and reinitialized the SpeedCounter
func (p *SpeedCounter) Calculate() int64 {
	return p.CallWithLock(func() interface{} {
		deltaNS := TimeNowNS() - p.lastNS
		if deltaNS <= 0 {
			return int64(0)
		}
		deltaCount := atomic.LoadInt64(&p.total) - p.lastCount
		p.lastCount += deltaCount
		p.lastNS += deltaNS
		return (deltaCount * int64(time.Second)) / deltaNS
	}).(int64)
}
