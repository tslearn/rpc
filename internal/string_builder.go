package internal

import (
	"sync"
)

var stringBuilderCache = &sync.Pool{
	New: func() interface{} {
		return &StringBuilder{
			buffer: make([]byte, 0, 4096),
		}
	},
}

// StringBuilder high performance string builder
type StringBuilder struct {
	buffer []byte
}

// NewStringBuilder create a string builder
func NewStringBuilder() *StringBuilder {
	return stringBuilderCache.Get().(*StringBuilder)
}

// Reset Reset the builder
func (p *StringBuilder) Reset() {
	if cap(p.buffer) == 4096 {
		p.buffer = p.buffer[:0]
	} else {
		p.buffer = make([]byte, 0, 4096)
	}
}

// Release Release the builder
func (p *StringBuilder) Release() {
	p.Reset()
	stringBuilderCache.Put(p)
}

// AppendByte append byte to the buffer
func (p *StringBuilder) AppendByte(byte byte) {
	p.buffer = append(p.buffer, byte)
}

// AppendBytes append bytes to the buffer
func (p *StringBuilder) AppendBytes(bytes []byte) {
	p.buffer = append(p.buffer, bytes...)
}

// AppendString append a string to string builder
func (p *StringBuilder) AppendString(str string) {
	p.buffer = append(p.buffer, str...)
}

// Merge write a string builder to current string builder
func (p *StringBuilder) Merge(builder *StringBuilder) {
	if builder == nil {
		return
	}
	p.buffer = append(p.buffer, builder.buffer...)
}

// IsEmpty return is the string builder is empty
func (p *StringBuilder) IsEmpty() bool {
	return len(p.buffer) == 0
}

// String get the builder string
func (p *StringBuilder) String() string {
	return string(p.buffer)
}
