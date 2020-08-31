package internal

import "sync"

// safePoolDebug should only works on Debug mode, when release it,
// please replace it with sync.Pool
type SafePool = safePoolDebug

var (
	safePoolDebugAllocMap   = make(map[interface{}]bool)
	safePoolDebugAllocMutex = &sync.Mutex{}
)

type safePoolDebug struct {
	pool sync.Pool
	New  func() interface{}
}

func (p *safePoolDebug) Put(x interface{}) {
	safePoolDebugAllocMutex.Lock()
	defer safePoolDebugAllocMutex.Unlock()

	if _, ok := safePoolDebugAllocMap[x]; ok {
		delete(safePoolDebugAllocMap, x)
		p.pool.Put(x)
	} else {
		panic("Put unmanaged object in pool, or may be put twice")
	}
}

func (p *safePoolDebug) Get() interface{} {
	safePoolDebugAllocMutex.Lock()
	defer safePoolDebugAllocMutex.Unlock()

	if p.pool.New == nil {
		p.pool.New = p.New
	}

	x := p.pool.Get()

	if _, ok := safePoolDebugAllocMap[x]; ok {
		panic("Get allocated object in pool")
	} else {
		safePoolDebugAllocMap[x] = true
	}

	return x
}
