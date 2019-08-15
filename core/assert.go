package core

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"unsafe"
)

type failedInterface interface {
	Fail()
}

var reportFail = func(p *Assert) {
	_, file, line, _ := runtime.Caller(2)
	fmt.Printf("%s:%d\n", file, line)
	p.t.Fail()
}

func assertIsNil(val interface{}) (ret bool) {
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

func assertRPCArrayEqual(left rpcArray, right rpcArray) bool {
	if left.ctx == nil && left.in == nil && right.ctx == nil && right.in == nil {
		return true
	}
	if !left.ok() || !right.ok() {
		return false
	}
	if left.Size() != right.Size() {
		return false
	}

	for i := 0; i < left.Size(); i++ {
		lv, ok := left.Get(i)
		if !ok {
			return false
		}
		rv, ok := right.Get(i)
		if !ok {
			return false
		}
		if !assertEquals(lv, rv) {
			return false
		}
	}
	return true
}

func assertRPCMapEqual(left rpcMap, right rpcMap) bool {
	if left.ctx == nil && left.in == nil && right.ctx == nil && right.in == nil {
		return true
	}
	if !left.ok() || !right.ok() {
		return false
	}

	lKeys := left.Keys()
	rKeys := right.Keys()

	if len(lKeys) != len(rKeys) {
		return false
	}

	for _, lKey := range lKeys {
		lv, ok := left.Get(lKey)
		if !ok {
			return false
		}
		rv, ok := right.Get(lKey)
		if !ok {
			return false
		}
		if !assertEquals(lv, rv) {
			return false
		}
	}

	return true
}

func assertEquals(left interface{}, right interface{}) bool {
	leftNil := assertIsNil(left)
	rightNil := assertIsNil(right)

	if leftNil {
		return rightNil
	}

	if rightNil {
		return false
	}

	switch left.(type) {
	case []byte:
		rBytes, ok := right.([]byte)
		if !ok {
			return false
		}
		return equalBytes(left.([]byte), rBytes)
	case rpcArray:
		rArray, ok := right.(rpcArray)
		if !ok {
			return false
		}
		return assertRPCArrayEqual(left.(rpcArray), rArray)
	case rpcMap:
		rMap, ok := right.(rpcMap)
		if !ok {
			return false
		}
		return assertRPCMapEqual(left.(rpcMap), rMap)
	case *rpcError:
		rError, ok := right.(*rpcError)
		if !ok {
			return false
		}
		return left.(*rpcError).Error() == rError.Error()
	default:
		return left == right
	}
}

func contains(left interface{}, right interface{}) bool {
	switch left.(type) {
	case string:
		rString, ok := right.(string)
		if !ok {
			return false
		}
		return strings.Contains(left.(string), rString)
	case []byte:
		rBytes, ok := right.([]byte)
		if !ok {
			return false
		}
		return bytes.Contains(left.([]byte), rBytes)
	case rpcArray:
		l := left.(rpcArray)
		for i := 0; i < l.Size(); i++ {
			lv, ok := l.Get(i)
			if !ok {
				return false
			}
			if assertEquals(lv, right) {
				return true
			}
		}
		return false
	}
	return false
}

// Assert class
type Assert struct {
	t    failedInterface
	args []interface{}
}

// NewAssert create new assert class
func NewAssert(t *testing.T) func(args ...interface{}) *Assert {
	return func(args ...interface{}) *Assert {
		return &Assert{
			t:    t,
			args: args,
		}
	}
}

// Equals ...
func (p *Assert) Equals(args ...interface{}) {
	if len(p.args) < 1 {
		reportFail(p)
		return
	}

	if len(p.args) != len(args) {
		reportFail(p)
		return
	}

	for i := 0; i < len(p.args); i++ {
		if !assertEquals(p.args[i], args[i]) {
			reportFail(p)
			return
		}
	}
}

// Contains ...
func (p *Assert) Contains(val interface{}) {
	if len(p.args) < 1 {
		reportFail(p)
		return
	}

	for i := 0; i < len(p.args); i++ {
		if !contains(p.args[i], val) {
			reportFail(p)
			return
		}
	}
}

// IsNil ...
func (p *Assert) IsNil() {
	if len(p.args) < 1 {
		reportFail(p)
		return
	}

	for i := 0; i < len(p.args); i++ {
		if !assertIsNil(p.args[i]) {
			reportFail(p)
			return
		}
	}
}

// IsNotNil ...
func (p *Assert) IsNotNil() {
	if len(p.args) < 1 {
		reportFail(p)
		return
	}

	for i := 0; i < len(p.args); i++ {
		if assertIsNil(p.args[i]) {
			reportFail(p)
			return
		}
	}
}

// IsTrue ...
func (p *Assert) IsTrue() {
	if len(p.args) < 1 {
		reportFail(p)
		return
	}

	for i := 0; i < len(p.args); i++ {
		if p.args[i] != true {
			reportFail(p)
			return
		}
	}
}

// IsFalse ...
func (p *Assert) IsFalse() {
	if len(p.args) < 1 {
		reportFail(p)
		return
	}

	for i := 0; i < len(p.args); i++ {
		if p.args[i] != false {
			reportFail(p)
			return
		}
	}
}
