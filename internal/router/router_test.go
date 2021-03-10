package router

import (
	"fmt"
	"net"
	"testing"
)

func TestDebug(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		ln1, e := net.Listen("tcp", "127.0.0.1:50000")
		fmt.Println(ln1, e)
		_ = ln1.Close()
		ln2, e := net.Listen("tcp", "127.0.0.1:50000")
		fmt.Println(ln2, e)
	})
}
