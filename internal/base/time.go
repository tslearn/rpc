package base

import (
	"math/rand"
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	gTimeMaster          = newTimeMaster()
	defaultISODateBuffer = []byte{
		0x30, 0x30, 0x30, 0x30, 0x2D, 0x30, 0x30, 0x2D, 0x30, 0x30,
		0x54, 0x30, 0x30, 0x3A, 0x30, 0x30, 0x3A, 0x30, 0x30, 0x2E,
		0x30, 0x30, 0x30, 0x2B, 0x30, 0x30, 0x3A, 0x30, 0x30,
	}
	intToStringCache2 = make([][]byte, 100)
	intToStringCache3 = make([][]byte, 1000)
	intToStringCache4 = make([][]byte, 10000)
)

func init() {
	rand.Seed(time.Now().UnixNano())
	charToASCII := [10]byte{
		0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39,
	}
	for i := 0; i < 100; i++ {
		intToStringCache2[i] = []byte{
			charToASCII[(i/10)%10],
			charToASCII[i%10],
		}
	}
	for i := 0; i < 1000; i++ {
		intToStringCache3[i] = []byte{
			charToASCII[(i/100)%10],
			charToASCII[(i/10)%10],
			charToASCII[i%10],
		}
	}
	for i := 0; i < 10000; i++ {
		intToStringCache4[i] = []byte{
			charToASCII[(i/1000)%10],
			charToASCII[(i/100)%10],
			charToASCII[(i/10)%10],
			charToASCII[i%10],
		}
	}

	go func() {
		gTimeMaster.Run()
	}()
}

type timeInfo struct {
	time          time.Time
	timeISOString string
}

const timeMasterStatusNone = int32(0)
const timeMasterStatusRunning = int32(1)
const timeMasterStatusClosed = int32(2)

type timeMaster struct {
	status           int32
	timeNowPointer   unsafe.Pointer
	timeSpeedCounter *SpeedCounter
}

func newTimeMaster() *timeMaster {
	return &timeMaster{
		timeNowPointer:   nil,
		timeSpeedCounter: NewSpeedCounter(),
	}
}

func (p *timeMaster) Run() bool {
	if atomic.CompareAndSwapInt32(
		&p.status,
		timeMasterStatusNone,
		timeMasterStatusRunning,
	) {
		for atomic.CompareAndSwapInt32(
			&p.status,
			timeMasterStatusRunning,
			timeMasterStatusRunning,
		) {
			speed, _ := p.timeSpeedCounter.Calculate(time.Now())
			if speed > 10000 {
				for i := 0; i < 100; i++ {
					now := time.Now()
					atomic.StorePointer(&p.timeNowPointer, unsafe.Pointer(&timeInfo{
						time:          now,
						timeISOString: ConvertToIsoDateString(now),
					}))
					time.Sleep(time.Millisecond)
				}
			} else {
				atomic.StorePointer(&p.timeNowPointer, nil)
				time.Sleep(100 * time.Millisecond)
			}
		}
		return true
	}

	return false
}

func (p *timeMaster) Close() bool {
	return atomic.CompareAndSwapInt32(
		&p.status,
		timeMasterStatusRunning,
		timeMasterStatusClosed,
	)
}

func (p *timeMaster) Count() {
	p.timeSpeedCounter.Count()
}

func (p *timeMaster) TimeNow() time.Time {
	p.Count()

	if item := (*timeInfo)(atomic.LoadPointer(&p.timeNowPointer)); item != nil {
		return item.time
	}

	return time.Now()
}

func (p *timeMaster) TimeNowISOString() string {
	p.Count()

	if item := (*timeInfo)(atomic.LoadPointer(&p.timeNowPointer)); item != nil {
		return item.timeISOString
	}

	return ConvertToIsoDateString(time.Now())
}

// ConvertToIsoDateString convert time.Time to iso string
// return format "2019-09-09T09:47:16.180+08:00"
func ConvertToIsoDateString(date time.Time) string {
	buf := make([]byte, 29)
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

// TimeNow get now nanoseconds from 1970-01-01
func TimeNow() time.Time {
	return gTimeMaster.TimeNow()
}

// TimeNowISOString get now iso string like this: 2019-09-09T09:47:16.180+08:00
func TimeNowISOString() string {
	return gTimeMaster.TimeNowISOString()
}

// IsTimeApproximatelyEqual ...
func IsTimeApproximatelyEqual(t1 time.Time, t2 time.Time) bool {
	delta := t1.Sub(t2)
	return delta < 500*time.Millisecond && delta > -500*time.Millisecond
}
