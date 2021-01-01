package client

import (
	"fmt"
	"testing"
	"time"

	"github.com/rpccloud/rpc/internal/server"
)

func TestClient_Debug(t *testing.T) {
	rpcServer := server.NewServer().Listen("tcp", "0.0.0.0:28888", nil)

	go func() {
		rpcServer.SetNumOfThreads(1024).Serve()
	}()

	time.Sleep(500 * time.Millisecond)
	rpcClient := newClient("tcp", "0.0.0.0:28888", nil, 1200, 1200)
	for i := 0; i < 1; i++ {
		fmt.Println(rpcClient.SendMessage(8*time.Second, "#.test:SayHello", i))
	}

	time.Sleep(15 * time.Second)
	rpcClient.Close()
	rpcServer.Close()
}
