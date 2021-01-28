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

	time.Sleep(300 * time.Millisecond)
	rpcClient := newClient("ws", "0.0.0.0:28888", nil, 1200, 1200)
	for i := 0; i < 10; i++ {
		fmt.Println(rpcClient.SendMessage(8*time.Second, "#.test:SayHello", i))
	}
	rpcClient.Close()
	rpcServer.Close()
}
