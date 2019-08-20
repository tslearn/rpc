package core

import (
	"fmt"
	"runtime"
	"testing"
)

var reportFail = func(p *rpcAssert) {
	_, file, line, _ := runtime.Caller(2)
	fmt.Printf("%s:%d\n", file, line)
	p.t.Fail()
}

// rpcAssert class
type rpcAssert struct {
	t    interface{ Fail() }
	args []interface{}
}

// newAssert create new assert class
func newAssert(t *testing.T) func(args ...interface{}) *rpcAssert {
	return func(args ...interface{}) *rpcAssert {
		return &rpcAssert{
			t:    t,
			args: args,
		}
	}
}

// Equals ...
func (p *rpcAssert) Equals(args ...interface{}) {
	if len(p.args) < 1 {
		reportFail(p)
		return
	}

	if len(p.args) != len(args) {
		reportFail(p)
		return
	}

	for i := 0; i < len(p.args); i++ {
		if !rpcEquals(p.args[i], args[i]) {
			reportFail(p)
			return
		}
	}
}

// Contains ...
func (p *rpcAssert) Contains(val interface{}) {
	if len(p.args) < 1 {
		reportFail(p)
		return
	}

	for i := 0; i < len(p.args); i++ {
		if !rpcContains(p.args[i], val) {
			reportFail(p)
			return
		}
	}
}

// IsNil ...
func (p *rpcAssert) IsNil() {
	if len(p.args) < 1 {
		reportFail(p)
		return
	}

	for i := 0; i < len(p.args); i++ {
		if !rpcIsNil(p.args[i]) {
			reportFail(p)
			return
		}
	}
}

// IsNotNil ...
func (p *rpcAssert) IsNotNil() {
	if len(p.args) < 1 {
		reportFail(p)
		return
	}

	for i := 0; i < len(p.args); i++ {
		if rpcIsNil(p.args[i]) {
			reportFail(p)
			return
		}
	}
}

// IsTrue ...
func (p *rpcAssert) IsTrue() {
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
func (p *rpcAssert) IsFalse() {
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
