package main

import "time"

func main() {
	for i := 0; i < 1000000; i++ {
		go func() {
			time.Sleep(1000 * time.Second)
		}()
	}

	time.Sleep(1000 * time.Second)
}
