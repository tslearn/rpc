package util

import (
	"fmt"
	"sync"
)

var stringBuilderPool = &sync.Pool{
	New: func() interface{} {
		return &rpcStringBuilder{
			buffer: make([]byte, 0, 4096),
		}
	},
}

// StringBuilder high performance string builder
type StringBuilder = *rpcStringBuilder
type rpcStringBuilder struct {
	buffer []byte
}

// NewStringBuilder create a string builder
func NewStringBuilder() StringBuilder {
	return stringBuilderPool.Get().(*rpcStringBuilder)
}

// Reset Reset the builder
func (p *rpcStringBuilder) Reset() {
	if cap(p.buffer) == 4096 {
		p.buffer = p.buffer[:0]
	} else {
		p.buffer = make([]byte, 0, 4096)
	}
}

// Release Release the builder
func (p *rpcStringBuilder) Release() {
	p.Reset()
	stringBuilderPool.Put(p)
}

// AppendByte append byte to the buffer
func (p *rpcStringBuilder) AppendByte(byte byte) {
	p.buffer = append(p.buffer, byte)
}

// AppendBytes append bytes to the buffer
func (p *rpcStringBuilder) AppendBytes(bytes []byte) {
	p.buffer = append(p.buffer, bytes...)
}

// AppendString append a string to string builder
func (p *rpcStringBuilder) AppendString(str string) {
	p.buffer = append(p.buffer, str...)
}

// AppendFormat append a formatted string to string builder
func (p *rpcStringBuilder) AppendFormat(format string, a ...interface{}) {
	p.AppendString(fmt.Sprintf(format, a...))
}

// Merge write a string builder to current string builder
func (p *rpcStringBuilder) Merge(builder StringBuilder) {
	if builder == nil {
		return
	}
	p.buffer = append(p.buffer, builder.buffer...)
}

// IsEmpty return is the string builder is empty
func (p *rpcStringBuilder) IsEmpty() bool {
	return len(p.buffer) == 0
}

// String get the builder string
func (p *rpcStringBuilder) String() string {
	return string(p.buffer)
}
