package internal

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestNewSpeedCounter(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	sc1 := NewSpeedCounter()
	assert(sc1).IsNotNil()
	assert(sc1.total).Equals(int64(0))
	assert(sc1.lastCount).Equals(int64(0))
	assert(time.Now().Sub(sc1.lastTime) < 10*time.Millisecond).IsTrue()
	assert(time.Now().Sub(sc1.lastTime) > -10*time.Millisecond).IsTrue()
}

func TestSpeedCounter_Count(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	sc1 := NewSpeedCounter()
	waitCH := make(chan bool)
	for i := 0; i < 10000; i++ {
		go func() {
			sc1.Count()
			waitCH <- true
		}()
	}
	for i := 0; i < 10000; i++ {
		<-waitCH
	}
	assert(sc1.Total()).Equals(int64(10000))
}

func TestSpeedCounter_Total(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	sc1 := NewSpeedCounter()
	waitCH := make(chan bool)
	for i := 0; i < 10000; i++ {
		go func() {
			sc1.Count()
			waitCH <- true
		}()
	}
	for i := 0; i < 10000; i++ {
		<-waitCH
	}
	assert(sc1.Total()).Equals(int64(10000))
}

func TestSpeedCounter_CalculateSpeed(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	sc1 := NewSpeedCounter()

	for n := 0; n < 100; n++ {
		waitCH := make(chan bool)
		for i := 0; i < 10000; i++ {
			go func() {
				sc1.Count()
				waitCH <- true
			}()
		}
		for i := 0; i < 10000; i++ {
			<-waitCH
		}
		assert(sc1.CalculateSpeed(sc1.lastTime)).
			Equals(int64(0), time.Duration(0))
		assert(sc1.CalculateSpeed(sc1.lastTime.Add(time.Second))).
			Equals(int64(10000), time.Second)
	}

	// Test(1)
	sc2 := NewSpeedCounter()
	atomic.StoreInt64(&sc2.lastCount, 10)
	assert(sc2.CalculateSpeed(sc2.lastTime.Add(time.Second))).
		Equals(int64(0), time.Duration(0))
}
