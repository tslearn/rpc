package internal

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

type Assert interface {
	Fail()
	Equals(args ...interface{})
	IsNil()
	IsNotNil()
	IsTrue()
	IsFalse()
}

// NewAssert create new assert class
func NewAssert(t *testing.T) func(args ...interface{}) Assert {
	return func(args ...interface{}) Assert {
		return &rpcAssert{
			t:    t,
			args: args,
		}
	}
}

// rpcAssert ...
type rpcAssert struct {
	t    interface{ Fail() }
	args []interface{}
}

// Fail ...
func (p *rpcAssert) Fail() {
	if _, file, line, ok := runtime.Caller(2); ok {
		fmt.Printf("%s:%d\n", file, line)
	}
	p.t.Fail()
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
