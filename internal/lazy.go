package internal

type lazyArrayItem = int
type lazyMapItem struct {
	key string
	pos int
}

type lazyMemory struct {
	arrayBuffer []lazyArrayItem
	mapBuffer   []lazyMapItem
	arrayPos    int
	mapPos      int
}

func newLazyMemory() *lazyMemory {
	return &lazyMemory{
		arrayBuffer: make([]lazyArrayItem, 128, 128),
		arrayPos:    0,
		mapBuffer:   make([]lazyMapItem, 32, 32),
		mapPos:      0,
	}
}

func (p *lazyMemory) Reset() {
	p.arrayPos = 0
	p.mapPos = 0
}

func (p *lazyMemory) AllowArrayItem(size int) []lazyArrayItem {
	if end := p.arrayPos + size; end > 128 {
		return make([]lazyArrayItem, size, size)
	} else {
		p.arrayPos = end
		return p.arrayBuffer[end-size : end]
	}
}

func (p *lazyMemory) AllowMapItem(size int) []lazyMapItem {
	if end := p.mapPos + size; end > 32 {
		return make([]lazyMapItem, size, size)
	} else {
		p.mapPos = end
		return p.mapBuffer[end-size : end]
	}
}
