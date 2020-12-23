package client

// FreeChannelStack ...
type FreeChannelStack struct {
	items []int
	pos   int
}

// NewFreeChannelStack ...
func NewFreeChannelStack(size int) *FreeChannelStack {
	return &FreeChannelStack{
		items: make([]int, size),
		pos:   0,
	}
}

// Push ...
func (p *FreeChannelStack) Push(v int) bool {
	if p.pos < len(p.items) {
		p.items[p.pos] = v
		p.pos++
		return true
	}

	return false
}

// Pop ...
func (p *FreeChannelStack) Pop() (int, bool) {
	if p.pos > 0 {
		p.pos--
		return p.items[p.pos], true
	}

	return 0, false
}
