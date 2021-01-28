package client

import (
	"fmt"
	"github.com/rpccloud/rpc"
	"testing"
	"time"

	"github.com/rpccloud/rpc/internal/server"
)

func TestClient_Debug(t *testing.T) {
	userService := rpc.NewService().
		On("SayHello", func(rt rpc.Runtime, name rpc.String) rpc.Return {
			time.Sleep(2 * time.Second)
			return rt.Reply("hello " + name)
		})

	rpcServer := server.NewServer().Listen("ws", "0.0.0.0:28888", nil)
	rpcServer.AddService("user", userService, nil)

	go func() {
		rpcServer.SetNumOfThreads(1024).Serve()
	}()

	time.Sleep(300 * time.Millisecond)
	rpcClient := newClient("ws", "0.0.0.0:28888", nil, 1200, 1200)
	for i := 0; i < 10; i++ {
		go func() {
			fmt.Println(
				rpcClient.SendMessage(8*time.Second, "#.user:SayHello", "kitty"),
			)
		}()
	}

	// haha, close it
	time.Sleep(time.Second)
	rpcClient.conn.Close()

	time.Sleep(10 * time.Second)

	rpcClient.Close()
	rpcServer.Close()
}
