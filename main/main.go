package main

import (
	"time"
)

func main() {
	mp := make(map[int]int)

	mp[0] = 0

	go func() {
		for i := 1; i < 100000; i++ {
			mp[i] = 0
			//sum = sum + 1
			time.Sleep(time.Microsecond)
		}
	}()

	go func() {
		for {
			mp[0] = mp[0] + 1
			//sum = sum + 1
			time.Sleep(time.Microsecond)

			if mp[0] > 100000 {
				break
			}
		}
	}()

	time.Sleep(100 * time.Second)
}
