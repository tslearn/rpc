package main

import (
	"fmt"
	"github.com/tslearn/rpcc/internal"
	"time"
)

func main() {
	fmt.Println(internal.CurrentGoroutineID())

	for i := 0; i < 10; i++ {
		go func() {
			fmt.Println(internal.CurrentGoroutineID())
		}()
	}

	time.Sleep(time.Second)
}
