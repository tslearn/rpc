package atomic

import (
	"fmt"
	"sync/atomic"
	"testing"
)

func TestDebug(t *testing.T) {
	sum := int64(0)
	waitCH := make(chan bool)

	go func() {
		for i := 0; i < 10000; i++ {
			if atomic.LoadInt64(&sum) == -1 {
				panic("error")
			}
		}
		waitCH <- true
	}()

	go func() {
		for i := 0; i < 10000; i++ {
			if sum == -1 {
				panic("error")
			}
			atomic.AddInt64(&sum, 1)
		}
		waitCH <- true
	}()

	<-waitCH
	<-waitCH

	fmt.Println(sum)
}
