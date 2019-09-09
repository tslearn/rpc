package core

import (
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	timeNowPointer = (unsafe.Pointer)(nil)
)

func init() {
	done := make(chan bool)
	tick := time.NewTicker(4 * time.Millisecond)
	fnTimer := func(t time.Time) {
		now := &timeNow{
			timeNS:        t.UnixNano(),
			timeISOString: ConvertToIsoDateString(t),
		}
		atomic.StorePointer(&timeNowPointer, unsafe.Pointer(now))
	}

	fnTimer(time.Now())

	// New go routine for timer
	go func() {
		for {
			select {
			case t := <-tick.C:
				fnTimer(t)
			case <-done:
				return
			}
		}
	}()
}

type timeNow struct {
	timeNS        int64
	timeISOString string
}

// TimeNowNS get now nanoseconds from 1970-01-01
func TimeNowNS() int64 {
	ret := atomic.LoadPointer(&timeNowPointer)
	if ret != nil {
		return (*timeNow)(ret).timeNS
	}
	return time.Now().UnixNano()
}

// TimeNowMS get now milliseconds from 1970-01-01
func TimeNowMS() int64 {
	return TimeNowNS() / int64(time.Millisecond)
}

// TimeNowMS get now iso string like this: 2019-09-09T09:47:16.180+08:00
func TimeNowISOString() string {
	ret := atomic.LoadPointer(&timeNowPointer)
	if ret != nil {
		return (*timeNow)(ret).timeISOString
	}
	return ConvertToIsoDateString(time.Now())
}

// TimeSpanFrom get time.Duration from fromNS
func TimeSpanFrom(startNS int64) time.Duration {
	return time.Duration(TimeNowNS() - startNS)
}

// TimeSpanBetween get time.Duration between startNS and endNS
func TimeSpanBetween(startNS int64, endNS int64) time.Duration {
	return time.Duration(endNS - startNS)
}
