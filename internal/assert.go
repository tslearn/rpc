package internal

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

// RPCAssert ...
type RPCAssert struct {
	t        interface{ Fail() }
	hookFail func()
	args     []interface{}
}

// NewRPCAssert create new assert class
func NewRPCAssert(t *testing.T) func(args ...interface{}) *RPCAssert {
	return func(args ...interface{}) *RPCAssert {
		return &RPCAssert{
			t:        t,
			hookFail: nil,
			args:     args,
		}
	}
}

// Fail ...
func (p *RPCAssert) Fail() {
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
func (p *RPCAssert) Equals(args ...interface{}) {
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
func (p *RPCAssert) IsNil() {
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
func (p *RPCAssert) IsNotNil() {
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
func (p *RPCAssert) IsTrue() {
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
func (p *RPCAssert) IsFalse() {
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
