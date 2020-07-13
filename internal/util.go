package internal

import (
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

var (
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

func getFuncKind(fn interface{}) (string, bool) {
	if fn == nil {
		return "", false
	} else if reflectFn := reflect.ValueOf(fn); reflectFn.Kind() != reflect.Func {
		return "", false
	} else if reflectFn.Type().NumIn() < 1 ||
		reflectFn.Type().In(0) != reflect.ValueOf(nilContext).Type() {
		return "", false
	} else if reflectFn.Type().NumOut() != 1 ||
		reflectFn.Type().Out(0) != reflect.ValueOf(nilReturn).Type() {
		return "", false
	} else {
		sb := NewStringBuilder()
		defer sb.Release()

		for i := 1; i < reflectFn.Type().NumIn(); i++ {
			switch reflectFn.Type().In(i) {
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
				return "", false
			}
		}

		return sb.String(), true
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
		return "rpc.Int"
	case uint64Type:
		return "rpc.Uint"
	case float64Type:
		return "rpc.Float"
	case stringType:
		return "rpc.String"
	default:
		return reflectType.String()
	}
}

func getArgumentsErrorPosition(fn reflect.Value) int {
	if fn.Type().NumIn() < 1 {
		return 0
	} else if fn.Type().In(0) != reflect.ValueOf(nilContext).Type() {
		return 0
	} else {
		for i := 1; i < fn.Type().NumIn(); i++ {
			switch fn.Type().In(i) {
			case bytesType:
				continue
			case arrayType:
				continue
			case mapType:
				continue
			case int64Type:
				continue
			case uint64Type:
				continue
			case boolType:
				continue
			case float64Type:
				continue
			case stringType:
				continue
			default:
				return i
			}
		}
		return -1
	}
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

// FindLinesByPrefix find the lines start with prefix string
func FindLinesByPrefix(text string, prefix string) []string {
	ret := make([]string, 0, 0)
	for _, v := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimSpace(v), strings.TrimSpace(prefix)) {
			ret = append(ret, v)
		}
	}
	return ret
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

// GetStackString reports the call stack information
func GetStackString(skip uint) string {
	sb := NewStringBuilder()
	defer sb.Release()

	for idx := 1; ; idx++ {
		if pc, file, line, _ := runtime.Caller(int(skip) + idx); pc == 0 {
			break
		} else {
			if fn := runtime.FuncForPC(pc); fn != nil {
				if !sb.IsEmpty() {
					sb.AppendByte('\n')
				}

				sb.AppendByte('-')
				sb.AppendBytes(intToStringCache2[idx%100])
				sb.AppendByte(' ')
				sb.AppendString(fn.Name())
				sb.AppendString(": ")
				sb.AppendString(file)
				sb.AppendByte(':')
				sb.AppendString(strconv.Itoa(line))
			}
		}
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
