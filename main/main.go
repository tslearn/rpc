package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

func main() {
	value := uint64(1)

	go func() {
		for i := 0; i < 100000; i++ {
			value++
			//sum = sum + 1
			time.Sleep(time.Microsecond)
		}
	}()

	go func() {
		for {
			if atomic.LoadUint64(&value) >= 100000 {
				fmt.Println("finish")
				break
			}

			time.Sleep(time.Microsecond)
		}
	}()

	time.Sleep(100 * time.Second)
}
