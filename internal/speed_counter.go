package internal

import (
	"sync/atomic"
	"time"
)

// SpeedCounter count the speed
type SpeedCounter struct {
	total     int64
	lastCount int64
	lastTime  time.Time
	Lock
}

// NewSpeedCounter create a SpeedCounter
func NewSpeedCounter() *SpeedCounter {
	return &SpeedCounter{
		total:     0,
		lastCount: 0,
		lastTime:  time.Now(),
	}
}

// Add count n times
func (p *SpeedCounter) Add(n int64) int64 {
	return atomic.AddInt64(&p.total, n)
}

// Total get the total count
func (p *SpeedCounter) Total() int64 {
	return atomic.LoadInt64(&p.total)
}

// CalculateSpeed ...
func (p *SpeedCounter) CalculateSpeed() int64 {
	return p.CallWithLock(func() interface{} {
		now := time.Now()
		deltaTime := now.Sub(p.lastTime)
		if deltaTime <= 0 {
			return int64(0)
		}
		deltaCount := atomic.LoadInt64(&p.total) - p.lastCount
		p.lastCount += deltaCount
		p.lastTime = now
		return (deltaCount * int64(time.Second)) / int64(deltaTime)
	}).(int64)
}
