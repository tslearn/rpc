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

var (
	vRPCArray  RPCArray
	vRPCMap    rpcMap
	vRPCBytes  rpcBytes
	vRPCString rpcString

	readTypeString string
	readTypeBytes  []byte

	pContext Context
	pReturn  Return
)

func getArgumentsErrorPosition(fn reflect.Value) int {
	if fn.Type().NumIn() < 1 {
		return 0
	}

	if fn.Type().In(0) != reflect.ValueOf(pContext).Type() {
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
		default:
			if argType == reflect.ValueOf(vRPCString).Type() ||
				argType == reflect.ValueOf(vRPCBytes).Type() ||
				argType == reflect.ValueOf(vRPCArray).Type() ||
				argType == reflect.ValueOf(vRPCMap).Type() {
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
		reflectFn.Type().In(0) != reflect.ValueOf(pContext).Type() {
		return "", false
	}

	if reflectFn.Type().NumOut() != 1 ||
		reflectFn.Type().Out(0) != reflect.ValueOf(pReturn).Type() {
		return "", false
	}

	ret := ""
	for i := 1; i < reflectFn.Type().NumIn(); i++ {
		argType := reflectFn.Type().In(i)

		if argType == reflect.ValueOf(vRPCArray).Type() {
			ret += "A"
		} else if argType == reflect.ValueOf(vRPCMap).Type() {
			ret += "M"
		} else if argType == reflect.ValueOf(vRPCBytes).Type() {
			ret += "X"
		} else if argType == reflect.ValueOf(vRPCString).Type() {
			ret += "S"
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

func toRPCArray(val []interface{}, ctx *rpcContext) RPCArray {
	if val == nil {
		return nilRPCArray
	}

	ret := newRPCArray(ctx)
	for i := 0; i < len(val); i++ {
		ret.Append(val[i])
	}
	return ret
}

func toRPCMap(val map[string]interface{}, ctx *rpcContext) rpcMap {
	if val == nil {
		return nilRPCMap
	}

	ret := newRPCMap(ctx)
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
	l := rpcString{}
	r := rpcString{}

	ctx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}

	switch left.(type) {
	case string:
		l = rpcString{ctx: ctx, status: rpcStatusAllocated, bytes: ([]byte)(left.(string))}
	case rpcString:
		l = left.(rpcString)
	}

	switch right.(type) {
	case string:
		r = rpcString{ctx: ctx, status: rpcStatusAllocated, bytes: ([]byte)(right.(string))}
	case rpcString:
		r = right.(rpcString)
	default:
		return false
	}

	if l.ctx == nil && r.ctx == nil &&
		l.status == rpcStatusError && r.status == rpcStatusError &&
		len(l.bytes) == 0 && len(r.bytes) == 0 {
		return true
	}

	return l.OK() && r.OK() &&
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
	l := rpcBytes{}
	r := rpcBytes{}

	ctx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}

	switch left.(type) {
	case []byte:
		l = rpcBytes{ctx: ctx, status: rpcStatusAllocated, bytes: left.([]byte)}
	case rpcBytes:
		l = left.(rpcBytes)
	}
	switch right.(type) {
	case []byte:
		r = rpcBytes{ctx: ctx, status: rpcStatusAllocated, bytes: right.([]byte)}
	case rpcBytes:
		r = right.(rpcBytes)
	default:
		return false
	}

	if l.ctx == nil && r.ctx == nil &&
		l.status == rpcStatusError && r.status == rpcStatusError &&
		len(l.bytes) == 0 && len(r.bytes) == 0 {
		return true
	}

	return l.OK() && r.OK() &&
		l.status == r.status && equalBytes(l.bytes, r.bytes)
}

func equalRPCArray(left interface{}, right interface{}) bool {
	l := left.(RPCArray)
	r, ok := right.(RPCArray)
	if !ok {
		return false
	}
	if l.ctx == nil && l.in == nil && r.ctx == nil && r.in == nil {
		return true
	}
	if !l.ok() || !r.ok() {
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
	l := left.(rpcMap)
	r, ok := right.(rpcMap)
	if !ok {
		return false
	}
	if l.ctx == nil && l.in == nil && r.ctx == nil && r.in == nil {
		return true
	}
	if !l.ok() || !r.ok() {
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
	case rpcString:
		return equalRPCString(left, right)
	case []byte:
		return equalRPCBytes(left, right)
	case rpcBytes:
		return equalRPCBytes(left, right)
	case RPCArray:
		return equalRPCArray(left, right)
	case rpcMap:
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
	case rpcString:
		l, ok = left.(rpcString).ToString()
	}
	if !ok {
		return -1
	}
	switch right.(type) {
	case string:
		r, ok = right.(string)
	case rpcString:
		r, ok = right.(rpcString).ToString()
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
	case rpcBytes:
		l, ok = left.(rpcBytes).ToBytes()
	}
	if !ok {
		return -1
	}
	switch right.(type) {
	case []byte:
		r, ok = right.([]byte)
	case rpcBytes:
		r, ok = right.(rpcBytes).ToBytes()
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
	case rpcString:
		return containsRPCString(left, right)
	case []byte:
		return containsRPCBytes(left, right)
	case rpcBytes:
		return containsRPCBytes(left, right)
	case RPCArray:
		return containsRPCArray(left, right)
	}
	return -1
}
