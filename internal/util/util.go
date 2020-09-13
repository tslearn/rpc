package util

import (
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	seedInt64              = int64(10000)
	timeNowPointer         = (unsafe.Pointer)(nil)
	timeCacheFailedCounter = NewSpeedCounter()
	defaultISODateBuffer   = []byte{
		0x30, 0x30, 0x30, 0x30, 0x2D, 0x30, 0x30, 0x2D, 0x30, 0x30,
		0x54, 0x30, 0x30, 0x3A, 0x30, 0x30, 0x3A, 0x30, 0x30, 0x2E,
		0x30, 0x30, 0x30, 0x2B, 0x30, 0x30, 0x3A, 0x30, 0x30,
	}
	intToStringCache2 = make([][]byte, 100)
	intToStringCache3 = make([][]byte, 1000)
	intToStringCache4 = make([][]byte, 10000)

	base64String = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789" +
		"+/"
)

type timeInfo struct {
	time          time.Time
	timeISOString string
}

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
}

func IsNil(val interface{}) (ret bool) {
	defer func() {
		if e := recover(); e != nil {
			ret = false
		}
	}()

	if val == nil {
		return true
	}

	return reflect.ValueOf(val).IsNil()
}

func MinInt(v1 int, v2 int) int {
	if v1 < v2 {
		return v1
	}

	return v2
}

func MaxInt(v1 int, v2 int) int {
	if v1 < v2 {
		return v2
	}

	return v1
}

func StringToBytesUnsafe(s string) (ret []byte) {
	bytesHeader := (*reflect.SliceHeader)(unsafe.Pointer(&ret))
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bytesHeader.Len = stringHeader.Len
	bytesHeader.Cap = stringHeader.Len
	bytesHeader.Data = stringHeader.Data
	return
}

func BytesToStringUnsafe(bytes []byte) (ret string) {
	bytesHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&ret))
	stringHeader.Len = bytesHeader.Len
	stringHeader.Data = bytesHeader.Data
	return
}

func IsUTF8Bytes(bytes []byte) bool {
	idx := 0
	length := len(bytes)

	for idx < length {
		c := bytes[idx]
		if c < 128 {
			idx++
		} else if c < 224 {
			if (idx+2 > length) ||
				(bytes[idx+1]&0xC0 != 0x80) {
				return false
			}
			idx += 2
		} else if c < 240 {
			if (idx+3 > length) ||
				(bytes[idx+1]&0xC0 != 0x80) ||
				(bytes[idx+2]&0xC0 != 0x80) {
				return false
			}
			idx += 3
		} else if c < 248 {
			if (idx+4 > length) ||
				(bytes[idx+1]&0xC0 != 0x80) ||
				(bytes[idx+2]&0xC0 != 0x80) ||
				(bytes[idx+3]&0xC0 != 0x80) {
				return false
			}
			idx += 4
		} else {
			return false
		}
	}

	return idx == length
}

func runStoreTime() {
	defer atomic.StorePointer(&timeNowPointer, nil)

	for i := 0; i < 800; i++ {
		now := time.Now()
		atomic.StorePointer(&timeNowPointer, unsafe.Pointer(&timeInfo{
			time:          now,
			timeISOString: ConvertToIsoDateString(now),
		}))
		time.Sleep(1 * time.Millisecond)
	}
}

func onTimeCacheFailed() {
	if timeCacheFailedCounter.Count()%10000 == 0 {
		if speed, _ := timeCacheFailedCounter.Calculate(
			TimeNow(),
		); speed > 10000 {
			if atomic.LoadPointer(&timeNowPointer) == nil {
				go runStoreTime()
			}
		}
	}
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
	if item := (*timeInfo)(atomic.LoadPointer(&timeNowPointer)); item != nil {
		return item.time
	}

	onTimeCacheFailed()
	return time.Now()
}

// TimeNowISOString get now iso string like this: 2019-09-09T09:47:16.180+08:00
func TimeNowISOString() string {
	if item := (*timeInfo)(atomic.LoadPointer(&timeNowPointer)); item != nil {
		return item.timeISOString
	}

	onTimeCacheFailed()
	return ConvertToIsoDateString(time.Now())
}

// GetSeed get int64 seed, it is goroutine safety
func GetSeed() int64 {
	return atomic.AddInt64(&seedInt64, 1)
}

// GetRandString get random string
func GetRandString(strLen int) string {
	sb := NewStringBuilder()
	defer sb.Release()

	for strLen > 0 {
		rand64 := rand.Uint64()
		for used := 0; used < 10 && strLen > 0; used++ {
			sb.AppendByte(base64String[rand64%64])
			rand64 = rand64 / 64
			strLen--
		}
	}
	return sb.String()
}

// AddPrefixPerLine ...
func AddPrefixPerLine(text string, prefix string) string {
	sb := NewStringBuilder()
	defer sb.Release()

	first := true
	array := strings.Split(text, "\n")
	for idx, v := range array {
		if first {
			first = false
		} else {
			sb.AppendByte('\n')
		}

		if v != "" || idx == 0 || idx != len(array)-1 {
			sb.AppendString(prefix)
			sb.AppendString(v)
		}
	}
	return sb.String()
}

// ConcatString ...
func ConcatString(args ...string) string {
	sb := NewStringBuilder()
	defer sb.Release()
	for _, v := range args {
		sb.AppendString(v)
	}
	return sb.String()
}

// GetFileLine ...
func GetFileLine(skip uint) string {
	return AddFileLine("", skip+1)
}

// AddFileLine ...
func AddFileLine(header string, skip uint) string {
	sb := NewStringBuilder()
	defer sb.Release()

	if _, file, line, ok := runtime.Caller(int(skip) + 1); ok && line > 0 {
		if header != "" {
			sb.AppendString(header)
			sb.AppendByte(' ')
		}

		sb.AppendString(file)
		sb.AppendByte(':')
		sb.AppendString(strconv.Itoa(line))
	} else {
		sb.AppendString(header)
	}

	return sb.String()
}

// ConvertOrdinalToString ...
func ConvertOrdinalToString(n uint) string {
	if n == 0 {
		return ""
	}

	switch n {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	case 3:
		return "3rd"
	default:
		return strconv.Itoa(int(n)) + "th"
	}
}
