package util

import (
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	timeNowPointer         = (unsafe.Pointer)(nil)
	timeCacheFailedCounter = NewSpeedCounter()
	defaultISODateBuffer   = []byte{
		0x30, 0x30, 0x30, 0x30, 0x2D, 0x30, 0x30, 0x2D, 0x30, 0x30,
		0x54, 0x30, 0x30, 0x3A, 0x30, 0x30, 0x3A, 0x30, 0x30, 0x2E,
		0x30, 0x30, 0x30, 0x2B, 0x30, 0x30, 0x3A, 0x30, 0x30,
	}
	intToStringCache2 = make([][]byte, 100, 100)
	intToStringCache3 = make([][]byte, 1000, 1000)
	intToStringCache4 = make([][]byte, 10000, 10000)
)

func init() {
	charToASCII := [10]byte{
		0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39,
	}
	for i := 0; i < 100; i++ {
		for j := 0; j < 2; j++ {
			intToStringCache2[i] = []byte{
				charToASCII[(i/10)%10],
				charToASCII[i%10],
			}
		}
	}
	for i := 0; i < 1000; i++ {
		for j := 0; j < 3; j++ {
			intToStringCache3[i] = []byte{
				charToASCII[(i/100)%10],
				charToASCII[(i/10)%10],
				charToASCII[i%10],
			}
		}
	}
	for i := 0; i < 10000; i++ {
		for j := 0; j < 4; j++ {
			intToStringCache4[i] = []byte{
				charToASCII[(i/1000)%10],
				charToASCII[(i/100)%10],
				charToASCII[(i/10)%10],
				charToASCII[i%10],
			}
		}
	}
}

func runStoreTime() {
	defer atomic.StorePointer(&timeNowPointer, nil)

	for i := 0; i < 800; i++ {
		now := time.Now()
		atomic.StorePointer(&timeNowPointer, unsafe.Pointer(&timeInfo{
			timeNS:        now.UnixNano(),
			timeISOString: ConvertToIsoDateString(now),
		}))
		time.Sleep(time.Millisecond)
	}
}

func onCacheFailed() {
	if timeCacheFailedCounter.Add(1)%10000 == 0 {
		if timeCacheFailedCounter.CalculateSpeed() > 10000 {
			now := time.Now()
			if atomic.CompareAndSwapPointer(
				&timeNowPointer,
				nil,
				unsafe.Pointer(&timeInfo{
					timeNS:        now.UnixNano(),
					timeISOString: ConvertToIsoDateString(now),
				}),
			) {
				go runStoreTime()
			}
		}
	}
}

type timeInfo struct {
	timeNS        int64
	timeISOString string
}

// ConvertToIsoDateString convert time.Time to iso string
// return format "2019-09-09T09:47:16.180+08:00"
func ConvertToIsoDateString(date time.Time) string {
	buf := make([]byte, 29, 29)
	// copy template
	copy(buf, defaultISODateBuffer)
	// copy year
	year := date.Year()
	if year <= 0 {
		year = 0
	}
	if year >= 9999 {
		year = 9999
	}
	copy(buf, intToStringCache4[year])
	// copy month
	copy(buf[5:], intToStringCache2[date.Month()%100])
	// copy date
	copy(buf[8:], intToStringCache2[date.Day()%100])
	// copy hour
	copy(buf[11:], intToStringCache2[date.Hour()%100])
	// copy minute
	copy(buf[14:], intToStringCache2[date.Minute()%100])
	// copy second
	copy(buf[17:], intToStringCache2[date.Second()%100])
	// copy ms
	copy(buf[20:], intToStringCache3[(date.Nanosecond()/1000000)%1000])
	// copy timezone
	_, offsetSecond := date.Zone()
	if offsetSecond < 0 {
		buf[23] = '-'
		offsetSecond = -offsetSecond
	}
	copy(buf[24:], intToStringCache2[(offsetSecond/3600)%100])
	copy(buf[27:], intToStringCache2[(offsetSecond%3600)/60])
	return string(buf)
}

// TimeNowNS get now nanoseconds from 1970-01-01
func TimeNowNS() int64 {
	if item := (*timeInfo)(atomic.LoadPointer(&timeNowPointer)); item != nil {
		return item.timeNS
	}

	onCacheFailed()
	return time.Now().UnixNano()
}

// TimeNowMS get now milliseconds from 1970-01-01
func TimeNowMS() int64 {
	return TimeNowNS() / int64(time.Millisecond)
}

// TimeNowISOString get now iso string like this: 2019-09-09T09:47:16.180+08:00
func TimeNowISOString() string {
	if item := (*timeInfo)(atomic.LoadPointer(&timeNowPointer)); item != nil {
		return item.timeISOString
	}

	onCacheFailed()
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
