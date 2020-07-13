package internal

import "testing"

func TestGetSeed(t *testing.T) {
	assert := NewRPCAssert(t)
	seed := GetSeed()
	assert(seed > 10000).IsTrue()

	for i := int64(0); i < 1000; i++ {
		assert(GetSeed()).Equals(seed + 1 + i)
	}
}
