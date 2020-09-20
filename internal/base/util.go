package base

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"unsafe"
)

var (
	seedInt64    = int64(10000)
	base64String = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789" +
		"+/"
)

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

// IsTCPPortOccupied ...
func IsTCPPortOccupied(port uint16) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 0)
	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
	}()
	return err == nil
}

func ReadFromFile(filePath string) (string, *Error) {
	ret, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", KernelFatal.AddDebug(err.Error())
	}

	// for windows, remove \r
	return strings.Replace(string(ret), "\r", "", -1), nil
}
