package util

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

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

// Assert ...
type Assert = *rpcAssert

type rpcAssert struct {
	t      interface{ Fail() }
	hookFN func()
	args   []interface{}
}

// NewAssert create new assert class
func NewAssert(t *testing.T) func(args ...interface{}) Assert {
	return func(args ...interface{}) *rpcAssert {
		return &rpcAssert{
			t:      t,
			hookFN: nil,
			args:   args,
		}
	}
}

// Fail ...
func (p *rpcAssert) Fail() {
	if p.hookFN != nil {
		p.hookFN()
	} else {
		_, file, line, _ := runtime.Caller(2)
		fmt.Printf("%s:%d\n", file, line)
		p.t.Fail()
	}
}

// Equals ...
func (p *rpcAssert) Equals(args ...interface{}) {
	if len(p.args) < 1 {
		p.Fail()
	} else if len(p.args) != len(args) {
		p.Fail()
	} else {
		for i := 0; i < len(p.args); i++ {
			if !reflect.DeepEqual(p.args[i], args[i]) {
				p.Fail()
			}
		}
	}
}

// IsNil ...
func (p *rpcAssert) IsNil() {
	if len(p.args) < 1 {
		p.Fail()
	} else {
		for i := 0; i < len(p.args); i++ {
			if !isNil(p.args[i]) {
				p.Fail()
			}
		}
	}
}

// IsNotNil ...
func (p *rpcAssert) IsNotNil() {
	if len(p.args) < 1 {
		p.Fail()
	} else {
		for i := 0; i < len(p.args); i++ {
			if isNil(p.args[i]) {
				p.Fail()
			}
		}
	}
}

// IsTrue ...
func (p *rpcAssert) IsTrue() {
	if len(p.args) < 1 {
		p.Fail()
	} else {
		for i := 0; i < len(p.args); i++ {
			if p.args[i] != true {
				p.Fail()
			}
		}
	}
}

// IsFalse ...
func (p *rpcAssert) IsFalse() {
	if len(p.args) < 1 {
		p.Fail()
	} else {
		for i := 0; i < len(p.args); i++ {
			if p.args[i] != false {
				p.Fail()
			}
		}
	}
}
