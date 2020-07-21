package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
)

func GoSafe(fn func()) {
	go func() {
		defer func() {
			if v := recover(); v != nil {
				fmt.Println(v)
				fmt.Println(string(debug.Stack()))
			}
		}()

		fn()
	}()
}

func StartServer() {
	http.HandleFunc("/", HelloServer)
	http.ListenAndServe(":8080", nil)
}

func Bad() {
	time.Sleep(3 * time.Second)
	panic("NONO")
}

func RunWithPanicCatch(fn func()) (ret interface{}) {
	defer func() {
		ret = recover()
	}()

	fn()
	return
}

func main() {
	go StartServer()

	go func() {
		RunWithPanicCatch(func() {
			time.Sleep(3 * time.Second)
			panic("NONO")
		})
	}()
	GoSafe(Bad)

	for {
		time.Sleep(time.Second)
	}
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}
