package core

import (
	"fmt"
	"runtime"
	"testing"
)

type failedInterface interface {
	Fail()
}

var reportFail = func(p *Assert) {
	_, file, line, _ := runtime.Caller(2)
	fmt.Printf("%s:%d\n", file, line)
	p.t.Fail()
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
		if !equals(p.args[i], args[i]) {
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
		if contains(p.args[i], val) != 1 {
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
		if !isNil(p.args[i]) {
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
		if isNil(p.args[i]) {
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
