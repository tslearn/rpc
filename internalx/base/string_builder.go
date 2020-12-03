package base

const stringBuilderBufferSize = 512

var stringBuilderCache = &SyncPool{
	New: func() interface{} {
		return &StringBuilder{
			buffer: make([]byte, 0, stringBuilderBufferSize),
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

// Reset reset the builder
func (p *StringBuilder) Reset() {
	if cap(p.buffer) == stringBuilderBufferSize {
		p.buffer = p.buffer[:0]
	} else {
		p.buffer = make([]byte, 0, stringBuilderBufferSize)
	}
}

// Release release the builder
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
