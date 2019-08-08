package core

import (
	"math/rand"
	"sync/atomic"
	"time"
)

const (
	randCacheSize = 1000001
	base64String  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789" +
		"+/"
)

var (
	randPos         = uint64(0)
	randCacheUint32 = make([]uint32, randCacheSize, randCacheSize)
)

// GetRandUint32 get Uint32 rand and fast,
// this is approximate rand but high performance
func GetRandUint32() uint32 {
	return randCacheUint32[atomic.AddUint64(&randPos, 1)%randCacheSize]
}

// GetRandString get random string
func GetRandString(strLen int) string {
	sb := NewStringBuilder()
	for i := 0; i < strLen; i++ {
		sb.AppendByte(base64String[rand.Uint32()%64])
	}
	ret := sb.String()
	sb.Release()
	return ret
}

func init() {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < randCacheSize; i++ {
		randCacheUint32[i] = rand.Uint32()
	}
}
