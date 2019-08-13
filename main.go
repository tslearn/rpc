package main

import (
	"fmt"
	"github.com/huandu/go-tls/g"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var gmap = sync.Map{}
var count = uint64(0)

func routine(idx int) {
	for i := 0; i < 10; i++ {
		ptr, ok := gmap.Load(idx)
		if ok {
			if ptr.(uintptr) != uintptr(g.G()) {
				fmt.Println("Error")
			} else {
				if v := atomic.AddUint64(&count, 1); v%1000000 == 0 {
					fmt.Println("OK", count)
				}
			}
		} else {
			gmap.Store(idx, uintptr(g.G()))
		}
		time.Sleep(time.Second)
	}
}

func main() {
	mu := sync.Mutex{}

	runtime.GC()

	mu.Lock()

	for n := 0; n < 1; n++ {
		go routine(n)
	}

	time.Sleep(100 * time.Second)

	mu.Unlock()
}
