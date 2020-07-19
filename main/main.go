package main

import "fmt"

func AAA() {
	defer func() {
		fmt.Println("defer AAA")
	}()

	panic("err")
}

func MMM(v uint64) {
	fmt.Printf("%d", v)
}

func In(v interface{}) {
	fmt.Printf("%d", v.(uint64))
}

func main() {
	// MMM(uint32(10))

	In(10)
}
