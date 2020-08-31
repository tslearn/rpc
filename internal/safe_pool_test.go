package internal

import (
	"fmt"
	"testing"
)

func TestSafePoolDebug(t *testing.T) {
	pool := &SafePool{
		New: func() interface{} {
			ret := make([]byte, 512)
			return &ret
		},
	}

	v1 := pool.Get()
	v2 := pool.Get()

	fmt.Println(safePoolDebugAllocMap)

	pool.Put(v1)
	pool.Put(v2)

	fmt.Println(safePoolDebugAllocMap)

}
