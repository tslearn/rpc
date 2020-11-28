package reuseport

import (
	"fmt"
	"testing"
)

func Test_Debug(t *testing.T) {
	fmt.Println(tcpReusablePort("tcp4", "0.0.0.0:8080", true))
	fmt.Println(tcpReusablePort("tcp4", "0.0.0.0:8080", true))

	fmt.Println(tcpReusablePort("tcp6", "[::]:1024", true))
	fmt.Println(tcpReusablePort("tcp6", "[::]:1024", true))
}
