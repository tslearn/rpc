package util

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestConvertToIsoDateString(t *testing.T) {
	assert := NewAssert(t)
	start, _ := time.Parse(
		"2006-01-02T15:04:05.999Z07:00",
		"0001-01-01T00:00:00+00:00",
	)

	for i := 0; i < 1000000; i++ {
		parseTime, err := time.Parse(
			"2006-01-02T15:04:05.999Z07:00",
			ConvertToIsoDateString(start),
		)
		assert(err).IsNil()
		assert(parseTime.UnixNano()).Equals(start.UnixNano())
		start = start.Add(271099197000000)
	}

	largeTime, _ := time.Parse(
		"2006-01-02T15:04:05.999Z07:00",
		"9998-01-01T00:00:00+00:00",
	)
	largeTime = largeTime.Add(1000000 * time.Hour)
	assert(ConvertToIsoDateString(largeTime)).
		Equals("9999-01-30T16:00:00.000+00:00")

	time1, _ := time.Parse(
		"2006-01-02T15:04:05.999Z07:00",
		"2222-12-22T11:11:11.333-11:59",
	)
	assert(ConvertToIsoDateString(time1)).
		Equals("2222-12-22T11:11:11.333-11:59")

	time2, _ := time.Parse(
		"2006-01-02T15:04:05.999Z07:00",
		"2222-12-22T11:11:11.333+11:59",
	)
	assert(ConvertToIsoDateString(time2)).
		Equals("2222-12-22T11:11:11.333+11:59")

	time3, _ := time.Parse(
		"2006-01-02T15:04:05.999Z07:00",
		"2222-12-22T11:11:11.333+00:00",
	)
	assert(ConvertToIsoDateString(time3)).
		Equals("2222-12-22T11:11:11.333+00:00")

	time4, _ := time.Parse(
		"2006-01-02T15:04:05.999Z07:00",
		"2222-12-22T11:11:11.333-00:00",
	)
	assert(ConvertToIsoDateString(time4)).
		Equals("2222-12-22T11:11:11.333+00:00")
}

func TestTimeNowNS(t *testing.T) {
	assert := NewAssert(t)

	for i := 0; i < 500000; i++ {
		nowNS := TimeNowNS()
		assert(time.Now().UnixNano()-nowNS < int64(20*time.Millisecond)).IsTrue()
		assert(time.Now().UnixNano()-nowNS > int64(-20*time.Millisecond)).IsTrue()
	}

	for i := 0; i < 500; i++ {
		nowNS := TimeNowNS()
		assert(time.Now().UnixNano()-nowNS < int64(10*time.Millisecond)).IsTrue()
		assert(time.Now().UnixNano()-nowNS > int64(-10*time.Millisecond)).IsTrue()
		time.Sleep(time.Millisecond)
	}

	// hack timeNowPointer to nil
	atomic.StorePointer(&timeNowPointer, nil)
	for i := 0; i < 500; i++ {
		nowNS := TimeNowNS()
		assert(time.Now().UnixNano()-nowNS < int64(10*time.Millisecond)).IsTrue()
		assert(time.Now().UnixNano()-nowNS > int64(-10*time.Millisecond)).IsTrue()
		time.Sleep(time.Millisecond)
	}
}

func TestTimeNowMS(t *testing.T) {
	assert := NewAssert(t)
	nowNS := TimeNowMS() * int64(time.Millisecond)
	assert(time.Now().UnixNano()-nowNS < int64(10*time.Millisecond)).IsTrue()
	assert(time.Now().UnixNano()-nowNS > int64(-10*time.Millisecond)).IsTrue()
}

func TestTimeNowISOString(t *testing.T) {
	assert := NewAssert(t)

	for i := 0; i < 500; i++ {
		if nowNS, err := time.Parse(
			"2006-01-02T15:04:05.999Z07:00",
			TimeNowISOString(),
		); err == nil {
			assert(
				time.Now().UnixNano()-nowNS.UnixNano() < int64(10*time.Millisecond),
			).IsTrue()
			assert(
				time.Now().UnixNano()-nowNS.UnixNano() > int64(-10*time.Millisecond),
			).IsTrue()
		} else {
			assert().Fail()
		}
		time.Sleep(time.Millisecond)
	}

	// hack timeNowPointer to nil
	atomic.StorePointer(&timeNowPointer, nil)
	for i := 0; i < 500; i++ {
		if nowNS, err := time.Parse(
			"2006-01-02T15:04:05.999Z07:00",
			TimeNowISOString(),
		); err == nil {
			assert(
				time.Now().UnixNano()-nowNS.UnixNano() < int64(10*time.Millisecond),
			).IsTrue()
			assert(
				time.Now().UnixNano()-nowNS.UnixNano() > int64(-10*time.Millisecond),
			).IsTrue()
		} else {
			assert().Fail()
		}
		time.Sleep(time.Millisecond)
	}
}

func TestTimeSpanFrom(t *testing.T) {
	assert := NewAssert(t)
	ns := TimeNowNS()
	time.Sleep(50 * time.Millisecond)
	dur := TimeSpanFrom(ns)
	assert(int64(dur) > int64(40*time.Millisecond)).IsTrue()
	assert(int64(dur) < int64(60*time.Millisecond)).IsTrue()
}

func TestTimeSpanBetween(t *testing.T) {
	assert := NewAssert(t)
	start := TimeNowNS()
	time.Sleep(50 * time.Millisecond)
	dur := TimeSpanBetween(start, TimeNowNS())
	assert(int64(dur) > int64(40*time.Millisecond)).IsTrue()
	assert(int64(dur) < int64(60*time.Millisecond)).IsTrue()
}
