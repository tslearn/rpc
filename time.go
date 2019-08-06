package common

import (
	"sync/atomic"
	"time"
)

// there are two mode for time api: low mode or high mode
// mode are switch automatically by the speed of api called
// if currentNano equals zero, it represent low mode,
// low mode accuracy is decided by system
// if currentNano bigger than zero, it represent high mode,
// high mode accuracy is 1ms

var (
	timeCurrentNano  = int64(0)
	timeSpeedCounter = NewSpeedCounter()
)

func init() {
	// New go routine for timer
	go func() {
		timeCount := int64(0)

		for true {
			if atomic.LoadInt64(&timeCurrentNano) > 0 {
				atomic.StoreInt64(&timeCurrentNano, time.Now().UnixNano())
				time.Sleep(890 * time.Microsecond)
				timeCount++
			} else {
				time.Sleep(150 * time.Millisecond)
				timeCount = 0
			}

			if timeCount%200 == 0 {
				speed := timeSpeedCounter.Calculate()
				if speed > 10000 {
					atomic.StoreInt64(&timeCurrentNano, time.Now().UnixNano())
				} else {
					atomic.StoreInt64(&timeCurrentNano, 0)
				}
			}
		}
	}()
}

// TimeNowNS get now nanoseconds from 1970-01-01
func TimeNowNS() int64 {
	timeSpeedCounter.Add(1)
	ret := atomic.LoadInt64(&timeCurrentNano)
	if ret > 0 {
		return ret
	}
	return time.Now().UnixNano()
}

// TimeNowMS get now milliseconds from 1970-01-01
func TimeNowMS() int64 {
	return TimeNowNS() / int64(time.Millisecond)
}

// TimeSpanFrom get time.Duration from fromNS
func TimeSpanFrom(startNS int64) time.Duration {
	return time.Duration(TimeNowNS() - startNS)
}

// TimeSpanBetween get time.Duration between startNS and endNS
func TimeSpanBetween(startNS int64, endNS int64) time.Duration {
	return time.Duration(endNS - startNS)
}
