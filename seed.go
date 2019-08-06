package common

import "sync/atomic"

var seed = int64(10000)

// GetSeed get int64 seed, it is goroutine safety
func GetSeed() int64 {
	return atomic.AddInt64(&seed, 1)
}
