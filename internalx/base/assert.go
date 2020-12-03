package base

import (
	"fmt"
	"reflect"
)

// Assert ...
type Assert interface {
	Fail(reason string)
	Equal(args ...interface{})
	IsNil()
	IsNotNil()
	IsTrue()
	IsFalse()
}

// NewAssert create new assert class
func NewAssert(t interface{ Fail() }) func(args ...interface{}) Assert {
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

func (p *rpcAssert) fail(reason string) {
	Log(fmt.Sprintf("\t%s\n\t%s\n", reason, GetFileLine(2)))
	p.t.Fail()
}

// Fail ...
func (p *rpcAssert) Fail(reason string) {
	p.fail(reason)
}

// Equals ...
func (p *rpcAssert) Equal(args ...interface{}) {
	if len(p.args) < 1 {
		p.fail("arguments is empty")
	} else if len(p.args) != len(args) {
		p.fail("arguments length not match")
	} else {
		for i := 0; i < len(p.args); i++ {
			if !reflect.DeepEqual(p.args[i], args[i]) {
				if !IsNil(p.args[i]) || !IsNil(args[i]) {
					p.fail(fmt.Sprintf(
						"%s argment does not equal\n\twant:\n%s\n\tgot:\n%s",
						ConvertOrdinalToString(uint(i+1)),
						AddPrefixPerLine(fmt.Sprintf("%T(%v)", args[i], args[i]), "\t"),
						AddPrefixPerLine(fmt.Sprintf("%T(%v)", p.args[i], p.args[i]), "\t"),
					))
				}
			}
		}
	}
}

// IsNil ...
func (p *rpcAssert) IsNil() {
	if len(p.args) < 1 {
		p.fail("arguments is empty")
	} else {
		for i := 0; i < len(p.args); i++ {
			if !IsNil(p.args[i]) {
				p.fail(fmt.Sprintf(
					"%s argument is not nil",
					ConvertOrdinalToString(uint(i+1)),
				))
			}
		}
	}
}

// IsNotNil ...
func (p *rpcAssert) IsNotNil() {
	if len(p.args) < 1 {
		p.fail("arguments is empty")
	} else {
		for i := 0; i < len(p.args); i++ {
			if IsNil(p.args[i]) {
				p.fail(fmt.Sprintf(
					"%s argument is nil",
					ConvertOrdinalToString(uint(i+1)),
				))
			}
		}
	}
}

// IsTrue ...
func (p *rpcAssert) IsTrue() {
	if len(p.args) < 1 {
		p.fail("arguments is empty")
	} else {
		for i := 0; i < len(p.args); i++ {
			if p.args[i] != true {
				p.fail(fmt.Sprintf(
					"%s argument is not true",
					ConvertOrdinalToString(uint(i+1)),
				))
			}
		}
	}
}

// IsFalse ...
func (p *rpcAssert) IsFalse() {
	if len(p.args) < 1 {
		p.fail("arguments is empty")
	} else {
		for i := 0; i < len(p.args); i++ {
			if p.args[i] != false {
				p.fail(fmt.Sprintf(
					"%s argument is not false",
					ConvertOrdinalToString(uint(i+1)),
				))
			}
		}
	}
}
