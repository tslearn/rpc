package internal

import (
	"errors"
	"fmt"
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
	contextType = reflect.ValueOf(Context{}).Type()
	returnType  = reflect.ValueOf(emptyReturn).Type()
	boolType    = reflect.ValueOf(true).Type()
	int64Type   = reflect.ValueOf(int64(0)).Type()
	uint64Type  = reflect.ValueOf(uint64(0)).Type()
	float64Type = reflect.ValueOf(float64(0)).Type()
	stringType  = reflect.ValueOf("").Type()
	bytesType   = reflect.ValueOf(Bytes{}).Type()
	arrayType   = reflect.ValueOf(Array{}).Type()
	mapType     = reflect.ValueOf(Map{}).Type()

	seedInt64 = int64(10000)

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

	goroutinePrefix = "goroutine "
	base64String    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789" +
		"+/"
)

type timeInfo struct {
	time          time.Time
	timeISOString string
}

type rpcFuncMeta struct {
	name       string
	body       string
	identifier string
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

func isNil(val interface{}) (ret bool) {
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

func getFuncKind(fn reflect.Value) (string, error) {
	if fn.Kind() != reflect.Func {
		return "", errors.New("handler must be a function")
	} else if fn.Type().NumIn() < 1 ||
		fn.Type().In(0) != reflect.ValueOf(Context{}).Type() {
		return "", fmt.Errorf(
			"handler 1st argument type must be %s",
			convertTypeToString(contextType),
		)
	} else if fn.Type().NumOut() != 1 ||
		fn.Type().Out(0) != reflect.ValueOf(emptyReturn).Type() {
		return "", fmt.Errorf(
			"handler return type must be %s",
			convertTypeToString(returnType),
		)
	} else {
		sb := NewStringBuilder()
		defer sb.Release()

		for i := 1; i < fn.Type().NumIn(); i++ {
			switch fn.Type().In(i) {
			case bytesType:
				sb.AppendByte('X')
			case arrayType:
				sb.AppendByte('A')
			case mapType:
				sb.AppendByte('M')
			case int64Type:
				sb.AppendByte('I')
			case uint64Type:
				sb.AppendByte('U')
			case boolType:
				sb.AppendByte('B')
			case float64Type:
				sb.AppendByte('F')
			case stringType:
				sb.AppendByte('S')
			default:
				return "", fmt.Errorf(
					"handler %s argument type %s is not supported",
					ConvertOrdinalToString(1+uint(i)),
					fn.Type().In(i),
				)
			}
		}

		return sb.String(), nil
	}
}

func convertTypeToString(reflectType reflect.Type) string {
	switch reflectType {
	case nil:
		return "<nil>"
	case contextType:
		return "rpc.Context"
	case returnType:
		return "rpc.Return"
	case bytesType:
		return "rpc.Bytes"
	case arrayType:
		return "rpc.Array"
	case mapType:
		return "rpc.Map"
	case boolType:
		return "rpc.Bool"
	case int64Type:
		return "rpc.Int64"
	case uint64Type:
		return "rpc.Uint64"
	case float64Type:
		return "rpc.Float64"
	case stringType:
		return "rpc.String"
	default:
		return reflectType.String()
	}
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
