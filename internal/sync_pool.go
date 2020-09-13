package internal

import (
	"log"
	"sync"
)

// safePoolDebug should only works on Debug mode, when release it,
// please replace it with sync.Pool
// type SyncPool = sync.Pool
type SyncPool = SyncPoolDebug

var (
	safePoolDebugMap = sync.Map{}
)

type SyncPoolDebug struct {
	pool sync.Pool
	New  func() interface{}
}

func (p *SyncPoolDebug) Put(x interface{}) {
	if _, ok := safePoolDebugMap.Load(x); ok {
		safePoolDebugMap.Delete(x)
		p.pool.Put(x)
	} else {
		panic("Put unmanaged object in pool, or may be put twice")
	}
}

func (p *SyncPoolDebug) Get() interface{} {
	if p.pool.New == nil {
		p.pool.New = p.New
		log.Printf(
			"Warn: SyncPool is in debug mode, which may slow down the program",
		)
	}

	x := p.pool.Get()

	if _, loaded := safePoolDebugMap.LoadOrStore(x, true); loaded {
		panic("Get allocated object in pool")
	}

	return x
}
