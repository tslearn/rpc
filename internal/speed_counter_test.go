package internal

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestNewSpeedCounter(t *testing.T) {
	assert := NewAssert(t)

	sc := NewSpeedCounter()
	assert(sc.total).Equals(int64(0))
	assert(sc.lastCount).Equals(int64(0))
	assert(time.Now().Sub(sc.lastTime) < 10*time.Millisecond).IsTrue()
	assert(time.Now().Sub(sc.lastTime) > -10*time.Millisecond).IsTrue()
}

func TestSpeedCounter_Add(t *testing.T) {
	assert := NewAssert(t)
	sc := NewSpeedCounter()
	sc.Add(5)
	assert(sc.Total()).Equals(int64(5))
	sc.Add(-5)
	assert(sc.Total()).Equals(int64(0))
}

func TestSpeedCounter_Total(t *testing.T) {
	assert := NewAssert(t)
	sc := NewSpeedCounter()
	sc.Add(5)
	sc.Add(10)
	assert(sc.Total()).Equals(int64(15))
}

func TestSpeedCounter_CalculateSpeed(t *testing.T) {
	assert := NewAssert(t)

	// Add 100 and calculate
	sc := NewSpeedCounter()
	sc.Add(100)
	time.Sleep(100 * time.Millisecond)
	speed, duration := sc.CalculateSpeed(TimeNow())
	assert(speed > 700 && speed < 1100).IsTrue()
	assert(duration > 90*time.Millisecond)
	assert(duration < 110*time.Millisecond)

	// Add -100 and calculate
	sc = NewSpeedCounter()
	sc.Add(-100)
	time.Sleep(100 * time.Millisecond)
	speed, duration = sc.CalculateSpeed(TimeNow())
	assert(speed > -1100 && speed < -700).IsTrue()
	assert(duration > 90*time.Millisecond)
	assert(duration < 110*time.Millisecond)

	// time interval is zero
	now := TimeNow()
	sc = NewSpeedCounter()
	sc.CalculateSpeed(now)
	assert(sc.CalculateSpeed(now)).Equals(int64(0), time.Duration(0))
}

func TestSpeedCounter_CalculateSpeed_Parallels(t *testing.T) {
	sc := NewSpeedCounter()

	start := int32(1)
	fnCounter := func() {
		for atomic.LoadInt32(&start) > 0 {
			sc.Add(1)
			time.Sleep(time.Millisecond)
		}
	}
	fnCalculate := func() {
		for atomic.LoadInt32(&start) > 0 {
			sc.CalculateSpeed(TimeNow())
			time.Sleep(time.Millisecond)
		}
	}

	for i := 0; i < 20; i++ {
		go fnCounter()
		go fnCalculate()
	}

	time.Sleep(300 * time.Millisecond)
	atomic.StoreInt32(&start, 0)
}
