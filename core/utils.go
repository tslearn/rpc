package core

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

func getArgumentsErrorPosition(fn reflect.Value) int {
	if fn.Type().NumIn() < 1 {
		return 0
	}

	if fn.Type().In(0) != reflect.ValueOf(nilContext).Type() {
		return 0
	}

	for i := 1; i < fn.Type().NumIn(); i++ {
		argType := fn.Type().In(i)
		switch argType.Kind() {
		case reflect.Uint64:
			continue
		case reflect.Int64:
			continue
		case reflect.Float64:
			continue
		case reflect.Bool:
			continue
		case reflect.String:
			continue
		default:
			if argType == reflect.ValueOf(emptyBytes).Type() ||
				argType == reflect.ValueOf(nilRPCArray).Type() ||
				argType == reflect.ValueOf(nilRPCMap).Type() {
				continue
			}
			return i
		}
	}

	return -1
}

func getFuncKind(fn interface{}) (string, bool) {
	reflectFn := reflect.ValueOf(fn)

	if reflectFn.Type().NumIn() < 1 ||
		reflectFn.Type().In(0) != reflect.ValueOf(nilContext).Type() {
		return "", false
	}

	if reflectFn.Type().NumOut() != 1 ||
		reflectFn.Type().Out(0) != reflect.ValueOf(nilReturn).Type() {
		return "", false
	}

	ret := ""
	for i := 1; i < reflectFn.Type().NumIn(); i++ {
		argType := reflectFn.Type().In(i)

		if argType == reflect.ValueOf(nilRPCArray).Type() {
			ret += "A"
		} else if argType == reflect.ValueOf(nilRPCMap).Type() {
			ret += "M"
		} else if argType == reflect.ValueOf(emptyBytes).Type() {
			ret += "X"
		} else {
			switch argType.Kind() {
			case reflect.Int64:
				ret += "I"
			case reflect.Uint64:
				ret += "U"
			case reflect.Bool:
				ret += "B"
			case reflect.Float64:
				ret += "F"
			case reflect.String:
				ret += "S"
			default:
				return "", false
			}
		}
	}
	return ret, true
}

// GetStackString reports the call stack information
func GetStackString(skip uint) string {
	sb := NewStringBuilder()

	idx := uint(1)
	pc, file, line, _ := runtime.Caller(int(skip + idx))

	first := true
	for pc != 0 {
		fn := runtime.FuncForPC(pc)

		if first {
			first = false
			sb.AppendFormat("-%02d %s: %s:%d", idx, fn.Name(), file, line)
		} else {
			sb.AppendFormat("\n-%02d %s: %s:%d", idx, fn.Name(), file, line)
		}

		idx++
		pc, file, line, _ = runtime.Caller(int(skip + idx))
	}

	ret := sb.String()
	sb.Release()
	return ret
}

// FindLinesByPrefix find the lines start with prefix string
func FindLinesByPrefix(debug string, prefix string) []string {
	ret := make([]string, 0, 0)
	dbgArr := strings.Split(debug, "\n")
	for _, v := range dbgArr {
		if strings.HasPrefix(strings.TrimSpace(v), strings.TrimSpace(prefix)) {
			ret = append(ret, v)
		}
	}
	return ret
}

// GetByteArrayDebugString get the debug string of []byte
func GetByteArrayDebugString(bs []byte) string {
	sb := stringBuilderPool.Get().(*StringBuilder)
	first := true
	for i := 0; i < len(bs); i++ {
		if i%16 == 0 {
			if first {
				first = false
				sb.AppendFormat("%04d: ", i)
			} else {
				sb.AppendFormat("\n%04d: ", i)
			}
		}
		sb.AppendFormat("0x%02x ", bs[i])
	}
	ret := sb.String()
	sb.Release()
	return ret
}

// GetURLBySchemeHostPortAndPath get the url by connection parameters
func GetURLBySchemeHostPortAndPath(scheme string, host string, port uint16, path string) string {
	if len(scheme) > 0 && len(host) > 0 {
		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}
		return fmt.Sprintf("%s://%s:%d/%s", scheme, host, port, path)
	}

	return ""
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

// GetObjectFieldPointer get field pointer by reflection
// this function always used in testing
func GetObjectFieldPointer(objPointer interface{}, fileName string) unsafe.Pointer {
	pointerVal := reflect.ValueOf(objPointer)
	val := reflect.Indirect(pointerVal)
	member := val.FieldByName(fileName)
	return unsafe.Pointer(member.UnsafeAddr())
}

// AddPrefixPerLine ...
func AddPrefixPerLine(origin string, prefix string) string {
	var buf bytes.Buffer
	arr := strings.Split(origin, "\n")
	first := true
	for _, v := range arr {
		if first {
			first = false
		} else {
			buf.WriteString("\n")
		}

		buf.WriteString(fmt.Sprintf("%s%s", prefix, v))
	}
	return buf.String()
}
