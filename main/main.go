package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan bool, 1)

	go func() {
		time.Sleep(time.Second)
		ch <- true
	}()

	select {
	case <-ch:
		fmt.Println("A")
	case <-time.After(3 * time.Second):
		fmt.Println("B")
	}

	time.Sleep(10 * time.Second)
	//go func() {
	//	for i := 0; i < 100000; i++ {
	//		value++
	//		//sum = sum + 1
	//		time.Sleep(time.Microsecond)
	//	}
	//}()
	//
	//go func() {
	//	for {
	//		if atomic.LoadUint64(&value) >= 100000 {
	//			fmt.Println("finish")
	//			break
	//		}
	//
	//		time.Sleep(time.Microsecond)
	//	}
	//}()
	//
	//time.Sleep(100 * time.Second)
}
