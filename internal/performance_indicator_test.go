package internal

import (
	"testing"
	"time"
)

func TestNewPerformanceIndicator(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	pi1 := newPerformanceIndicator()
	assert(pi1).IsNotNil()
	assert(pi1.lastTotal).Equals(int64(0))
	assert(len(pi1.successArray)).Equals(8)
	for i := 0; i < len(pi1.successArray); i++ {
		assert(pi1.successArray[i]).Equals(int64(0))
	}
	assert(len(pi1.failArray)).Equals(8)
	for i := 0; i < len(pi1.failArray); i++ {
		assert(pi1.failArray[i]).Equals(int64(0))
	}
	assert(TimeNow().Sub(pi1.lastTime) < 20*time.Millisecond).IsTrue()
	assert(TimeNow().Sub(pi1.lastTime) > -20*time.Millisecond).IsTrue()
}

func TestRpcPerformanceIndicator_Count(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	pi := newPerformanceIndicator()
	for i := 0; i < 2000; i++ {
		pi.Count(time.Duration(i)*time.Millisecond, true)
		pi.Count(time.Duration(i)*time.Millisecond, false)
	}
	assert(pi.successArray).
		Equals([8]int64{5, 15, 30, 50, 100, 300, 500, 1000})
	assert(pi.failArray).
		Equals([8]int64{5, 15, 30, 50, 100, 300, 500, 1000})
}

func TestRpcPerformanceIndicator_Calculate(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	pi1 := newPerformanceIndicator()
	for i := 0; i < 1500; i++ {
		go func(idx int) {
			for k := 0; k < 100; k++ {
				pi1.Count(time.Duration(idx)*time.Millisecond, true)
				pi1.Count(time.Duration(idx)*time.Millisecond, false)
				time.Sleep(time.Millisecond)
			}
		}(i)
	}
	time.Sleep(200 * time.Millisecond)
	assert(pi1.Calculate(pi1.lastTime.Add(time.Second))).
		Equals(int64(300000), time.Second)

	// Test(2)
	pi2 := newPerformanceIndicator()
	pi2.lastTotal = 3
	assert(pi2.Calculate(pi2.lastTime.Add(time.Second))).
		Equals(int64(0), time.Duration(0))

	// Test(3)
	pi3 := newPerformanceIndicator()
	assert(pi3.Calculate(pi3.lastTime.Add(-time.Second))).
		Equals(int64(0), time.Duration(0))
}
