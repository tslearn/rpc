package client

import (
	"fmt"
	"testing"
	"time"

	"github.com/rpccloud/rpc/internal/server"
)

func TestClient_Debug(t *testing.T) {
	rpcServer := server.NewServer().Listen("ws", "0.0.0.0:28888", nil)

	go func() {
		rpcServer.SetNumOfThreads(1024).Serve()
	}()

	time.Sleep(500 * time.Millisecond)
	rpcClient := newClient("ws", "0.0.0.0:28888", nil, 1200, 1200)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			fmt.Println(rpcClient.SendMessage(2*time.Second, "#.test:SayHello", idx))
		}(i)
	}
	time.Sleep(4 * time.Second)
	rpcClient.Close()
	rpcServer.Close()
}
