package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

func main() {
	sum := int64(0)

	go func() {
		for i := 0; i < 10000; i++ {
			atomic.AddInt64(&sum, 1)
			//sum = sum + 1
			time.Sleep(time.Microsecond)
		}
	}()

	go func() {
		for {
			if atomic.LoadInt64(&sum) == 10000 {
				fmt.Println("finish)")
				break
			}
		}
	}()

	time.Sleep(100 * time.Second)

	fmt.Println(sum)
}
