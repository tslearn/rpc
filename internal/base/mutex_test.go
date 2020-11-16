package base

import (
	"testing"
)

func TestMutex_Synchronized(t *testing.T) {
	assert := NewAssert(t)
	sum := 0
	mu := Mutex{}
	wait := make(chan bool)

	for i := 0; i < 1000; i++ {
		go func() {
			for j := 0; j < 10000; j++ {
				mu.Synchronized(func() {
					sum += j
				})
			}
			wait <- true
		}()
	}

	for i := 0; i < 1000; i++ {
		<-wait
	}

	assert(sum).Equal(49995000000)
}
