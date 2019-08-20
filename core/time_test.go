package core

import (
	"testing"
	"time"
)

func TestTimeNowNS(t *testing.T) {
	assert := newAssert(t)

	for i := 0; i < 500000; i++ {
		nowNS := TimeNowNS()
		assert(time.Now().UnixNano()-nowNS < int64(20*time.Millisecond)).IsTrue()
		assert(time.Now().UnixNano()-nowNS > int64(-20*time.Millisecond)).IsTrue()
	}

	for i := 0; i < 500; i++ {
		time.Sleep(time.Millisecond)
		nowNS := TimeNowNS()
		assert(time.Now().UnixNano()-nowNS < int64(10*time.Millisecond)).IsTrue()
		assert(time.Now().UnixNano()-nowNS > int64(-10*time.Millisecond)).IsTrue()
	}
}

func TestTimeNowMS(t *testing.T) {
	assert := newAssert(t)
	nowMS := TimeNowMS()
	assert(time.Now().UnixNano()-nowMS*int64(time.Millisecond) < int64(5*time.Millisecond)).IsTrue()
}

func TestTimeSpanFrom(t *testing.T) {
	assert := newAssert(t)
	ns := TimeNowNS()
	time.Sleep(50 * time.Millisecond)
	dur := TimeSpanFrom(ns)
	assert(int64(dur) > int64(40*time.Millisecond)).IsTrue()
	assert(int64(dur) < int64(60*time.Millisecond)).IsTrue()
}

func TestTimeSpanBetween(t *testing.T) {
	assert := newAssert(t)
	start := TimeNowNS()
	time.Sleep(50 * time.Millisecond)
	dur := TimeSpanBetween(start, TimeNowNS())
	assert(int64(dur) > int64(40*time.Millisecond)).IsTrue()
	assert(int64(dur) < int64(60*time.Millisecond)).IsTrue()
}
