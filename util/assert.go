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
type Assert struct {
	t        interface{ Fail() }
	hookFail func()
	args     []interface{}
}

// NewAssert create new assert class
func NewAssert(t *testing.T) func(args ...interface{}) *Assert {
	return func(args ...interface{}) *Assert {
		return &Assert{
			t:        t,
			hookFail: nil,
			args:     args,
		}
	}
}

// Fail ...
func (p *Assert) Fail() {
	if p.hookFail != nil {
		p.hookFail()
	} else {
		if _, file, line, ok := runtime.Caller(2); ok {
			fmt.Printf("%s:%d\n", file, line)
		}
		p.t.Fail()
	}
}

// Equals ...
func (p *Assert) Equals(args ...interface{}) {
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
func (p *Assert) IsNil() {
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
func (p *Assert) IsNotNil() {
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
func (p *Assert) IsTrue() {
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
func (p *Assert) IsFalse() {
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
