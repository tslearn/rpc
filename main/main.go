package main

import "fmt"

func AAA() {
	defer func() {
		fmt.Println("defer AAA")
	}()

	panic("err")
}

func main() {
	AAA()
}
