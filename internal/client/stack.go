package client

type FreeChannelStack struct {
	items []int
	pos   int
}

func NewFreeChannelStack(size int) *FreeChannelStack {
	return &FreeChannelStack{
		items: make([]int, size),
		pos:   0,
	}
}

func (p *FreeChannelStack) Push(v int) bool {
	if p.pos < len(p.items) {
		p.items[p.pos] = v
		p.pos++
		return true
	}

	return false
}

func (p *FreeChannelStack) Pop() (int, bool) {
	if p.pos > 0 {
		p.pos--
		return p.items[p.pos], true
	}

	return 0, false
}
