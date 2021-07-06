package base

import (
	"os"
	"sync"
)

// SyncPoolDebug should only works on debug mode, when release it,
// please replace it with sync.Pool

// SyncPool ...
// type SyncPool = SyncPoolDebug

// SyncPool ...
type SyncPool = sync.Pool

var (
	safePoolDebugMutex = sync.Mutex{}
	safePoolDebugMap   = map[interface{}]bool{}
)

// SyncPoolDebug ...
type SyncPoolDebug struct {
	pool sync.Pool
	New  func() interface{}
}

// Put ...
func (p *SyncPoolDebug) Put(value interface{}) {
	safePoolDebugMutex.Lock()
	defer safePoolDebugMutex.Unlock()

	if value == nil {
		panic("value is nil")
	}

	if _, ok := safePoolDebugMap[value]; ok {
		delete(safePoolDebugMap, value)
		p.pool.Put(value)
	} else {
		panic("check failed")
	}
}

// Get ...
func (p *SyncPoolDebug) Get() interface{} {
	safePoolDebugMutex.Lock()
	defer safePoolDebugMutex.Unlock()

	if p.pool.New == nil {
		p.pool.New = p.New
		_, _ = os.Stdout.WriteString(
			"Warn: SyncPool is in debug mode, which may slow down the program",
		)
	}

	x := p.pool.Get()

	if _, ok := safePoolDebugMap[x]; !ok {
		safePoolDebugMap[x] = true
	} else {
		panic("check failed")
	}

	return x
}
