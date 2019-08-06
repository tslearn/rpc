package common

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

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

func toRPCArray(val []interface{}) RPCArray {
	if val == nil {
		return nilRPCArray
	}

	ret := newRPCArray(nil)
	for i := 0; i < len(val); i++ {
		ret.Append(val[i])
	}
	return ret
}

func toRPCMap(val map[string]interface{}) RPCMap {
	if val == nil {
		return nilRPCMap
	}

	ret := newRPCMap(nil)
	for name, value := range val {
		ret.Set(name, value)
	}
	return ret
}

func isNil(val interface{}) (ret bool) {
	if val == nil {
		return true
	}

	switch val.(type) {
	case unsafe.Pointer:
		return val.(unsafe.Pointer) == nil
	case uintptr:
		return val.(uintptr) == 0
	}

	defer func() {
		if e := recover(); e != nil {
			ret = false
		}
	}()

	rv := reflect.ValueOf(val)
	return rv.IsNil()
}

func equalRPCString(left interface{}, right interface{}) bool {
	l := RPCString{}
	r := RPCString{}

	switch left.(type) {
	case string:
		l = RPCString{status: rpcStatusAllocated, bytes: ([]byte)(left.(string))}
	case RPCString:
		l = left.(RPCString)
	}

	switch right.(type) {
	case string:
		r = RPCString{status: rpcStatusAllocated, bytes: ([]byte)(right.(string))}
	case RPCString:
		r = right.(RPCString)
	default:
		return false
	}

	if l.OK() && r.OK() {
		return equalBytes(l.bytes, r.bytes)
	}

	return l.pub.OK() && r.pub.OK() &&
		l.status == r.status && equalBytes(l.bytes, r.bytes)
}

func equalBytes(left []byte, right []byte) bool {
	if left == nil && right == nil {
		return true
	} else if left == nil {
		return false
	} else if right == nil {
		return false
	} else {
		if len(left) != len(right) {
			return false
		}
		for i := 0; i < len(left); i++ {
			if left[i] != right[i] {
				return false
			}
		}
		return true
	}
}

func equalRPCBytes(left interface{}, right interface{}) bool {
	l := RPCBytes{}
	r := RPCBytes{}

	switch left.(type) {
	case []byte:
		l = RPCBytes{status: rpcStatusAllocated, bytes: left.([]byte)}
	case RPCBytes:
		l = left.(RPCBytes)
	}
	switch right.(type) {
	case []byte:
		r = RPCBytes{status: rpcStatusAllocated, bytes: right.([]byte)}
	case RPCBytes:
		r = right.(RPCBytes)
	default:
		return false
	}

	if l.OK() && r.OK() {
		return equalBytes(l.bytes, r.bytes)
	}

	return l.pub.OK() && r.pub.OK() &&
		l.status == r.status && equalBytes(l.bytes, r.bytes)
}

func equalRPCArray(left interface{}, right interface{}) bool {
	l := left.(RPCArray)
	r, ok := right.(RPCArray)
	if !ok {
		return false
	}
	if l.pub == nil && l.in == nil && r.pub == nil && r.in == nil {
		return true
	}
	if !l.OK() || !r.OK() {
		return false
	}
	if l.Size() != r.Size() {
		return false
	}

	for i := 0; i < l.Size(); i++ {
		lv, ok := l.Get(i)
		if !ok {
			return false
		}
		rv, ok := r.Get(i)
		if !ok {
			return false
		}
		if !equals(lv, rv) {
			return false
		}
	}
	return true
}

func equalRPCMap(left interface{}, right interface{}) bool {
	l := left.(RPCMap)
	r, ok := right.(RPCMap)
	if !ok {
		return false
	}
	if l.pub == nil && l.in == nil && r.pub == nil && r.in == nil {
		return true
	}
	if !l.OK() || !r.OK() {
		return false
	}

	lKeys := l.Keys()
	rKeys := r.Keys()

	if len(lKeys) != len(rKeys) {
		return false
	}

	for _, lKey := range lKeys {
		lv, ok := l.Get(lKey)
		if !ok {
			return false
		}
		rv, ok := r.Get(lKey)
		if !ok {
			return false
		}
		if !equals(lv, rv) {
			return false
		}
	}

	return true
}

func equals(left interface{}, right interface{}) bool {
	leftNil := isNil(left)
	rightNil := isNil(right)

	if leftNil {
		return rightNil
	}

	if rightNil {
		return false
	}

	switch left.(type) {
	case string:
		return equalRPCString(left, right)
	case RPCString:
		return equalRPCString(left, right)
	case []byte:
		return equalRPCBytes(left, right)
	case RPCBytes:
		return equalRPCBytes(left, right)
	case RPCArray:
		return equalRPCArray(left, right)
	case RPCMap:
		return equalRPCMap(left, right)
	case RPCError:
		lError := left.(RPCError)
		rError, ok := right.(RPCError)
		if !ok {
			return false
		}
		return lError.Error() == rError.Error()
	default:
		return left == right
	}
}

func containsRPCString(left interface{}, right interface{}) int {
	l := ""
	r := ""
	ok := false
	switch left.(type) {
	case string:
		l, ok = left.(string)
	case RPCString:
		l, ok = left.(RPCString).ToString()
	}
	if !ok {
		return -1
	}
	switch right.(type) {
	case string:
		r, ok = right.(string)
	case RPCString:
		r, ok = right.(RPCString).ToString()
	default:
		return -1
	}
	if !ok {
		return -1
	}

	if strings.Contains(l, r) {
		return 1
	}
	return 0
}

func containsRPCBytes(left interface{}, right interface{}) int {
	l := []byte(nil)
	r := []byte(nil)
	ok := false

	switch left.(type) {
	case []byte:
		l, ok = left.([]byte)
	case RPCBytes:
		l, ok = left.(RPCBytes).ToBytes()
	}
	if !ok {
		return -1
	}
	switch right.(type) {
	case []byte:
		r, ok = right.([]byte)
	case RPCBytes:
		r, ok = right.(RPCBytes).ToBytes()
	default:
		return -1
	}
	if !ok {
		return -1
	}

	if len(r) == 0 {
		return 1
	}
	if len(l) < len(r) {
		return 0
	}
	for i := 0; i <= len(l)-len(r); i++ {
		k := 0
		for k < len(r) && l[i+k] == r[k] {
			k++
		}
		if k == len(r) {
			return 1
		}
	}
	return 0
}

func containsRPCArray(left interface{}, right interface{}) int {
	l := left.(RPCArray)

	for i := 0; i < l.Size(); i++ {
		lv, ok := l.Get(i)
		if !ok {
			return 0
		}

		if equals(lv, right) {
			return 1
		}
	}
	return 0
}

func contains(left interface{}, right interface{}) int {
	switch left.(type) {
	case string:
		return containsRPCString(left, right)
	case RPCString:
		return containsRPCString(left, right)
	case []byte:
		return containsRPCBytes(left, right)
	case RPCBytes:
		return containsRPCBytes(left, right)
	case RPCArray:
		return containsRPCArray(left, right)
	}
	return -1
}
