package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tslearn/rpcc/g"
)

var gmap = sync.Map{}
var count = uint64(0)

const loop = 1000000

func testEE() (ret []byte) {
	ret = make([]byte, 800, 800)
	for i := 0; i < 800; i++ {
		ret[i] = 12
	}
	return
}

func testDD() (ret []byte) {
	ret = testEE()

	if len(ret) == 8 {
		for i := 0; i < 8; i++ {
			if ret[i] != 12 {
				fmt.Println("NO")
			}
		}
	}
	return ret
}

func testCC() (ret []byte) {
	ret = testDD()
	if len(ret) == 3 {
		fmt.Println("NO")
	}
	return
}

func routine(idx int) {
	for i := 0; i < 100; i++ {
		testCC()
		ptr, ok := gmap.Load(idx)
		if ok {
			if ptr.(uintptr) != uintptr(g.G()) {
				fmt.Println("Error")
			} else {
				if v := atomic.AddUint64(&count, 1); v%loop == 0 {
					fmt.Println("OK", v)
				}
			}
		} else {
			gmap.Store(idx, uintptr(g.G()))
		}
		time.Sleep(time.Second)
	}
}

func main() {
	runtime.GC()

	for n := 0; n < loop; n++ {
		go routine(n)
	}

	for {
		time.Sleep(time.Millisecond)
	}
}
