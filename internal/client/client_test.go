package client

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/server"
	"testing"
	"time"
)

func TestClient_Debug(t *testing.T) {
	rpcServer := server.NewServer().ListenWebSocket("0.0.0.0:28888")

	go func() {
		rpcServer.Serve()
	}()

	time.Sleep(2000 * time.Millisecond)

	rpcClient, err := newClient("ws://0.0.0.0:28888")

	if err != nil {
		panic(err)
	}

	fmt.Println(rpcClient.SendMessage(20*time.Second, "#.test:SayHello"))

	rpcServer.Close()
}
