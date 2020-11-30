package server

import (
	"testing"
)

func Test_Debug(t *testing.T) {

	waitCH := make(chan bool)
	pCount := 10

	for i := 0; i < pCount; i++ {
		go func() {
			acceptor := NewAcceptor("tcp4", "0.0.0.0:8080")
			err := acceptor.Listen()
			if err != nil {
				panic(err)
			}
			_ = acceptor.Serve()
			waitCH <- true
		}()
	}

	for i := 0; i < pCount; i++ {
		<-waitCH
	}
}
