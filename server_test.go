package rpcc

import (
	"fmt"
	"github.com/tslearn/kzserver/common"
	"testing"
	"time"
)

func TestServer_Debug(t *testing.T) {
	// err := common.RPCError(nil)
	fmt.Printf("test %s", 3*time.Second)
}

func TestServer_Basic(t *testing.T) {
	// assert := common.NewAssert(t)
	server := NewServer(32, nil).
		AddService("user", NewService().
			Echo("sayHello", true, func(
				ctx common.RPCContext,
				name string,
			) common.RPCReturn {
				time.Sleep(1000 * time.Second)
				return ctx.OK("hello " + name)
			}),
		).
		AddEndPoint(
			NewWebSocketServerEndPoint("127.0.0.1:8080", "test"),
		)
	server.Open()

	// client
	go func() {
		client := NewClient(
			NewWebSocketClientEndPoint("ws://127.0.0.1:8080/test"),
		)

		_ = client.Open()

		for i := 0; i < 300; i++ {
			go func(idx int) {
				fmt.Println(client.sendMessage(
					5*time.Second,
					"$.user:sayHello",
					fmt.Sprintf("ts%d", idx),
				))
			}(i)

			time.Sleep(30 * time.Millisecond)
		}

		time.Sleep(10 * time.Second)
		_ = client.Close()
		fmt.Println("Finish")

		server.Close()
	}()

	time.Sleep(100 * time.Second)
}
