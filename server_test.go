package rpcc

import (
	"fmt"
	"testing"
	"time"
)

func TestServer_Debug(t *testing.T) {
	// err := common.RPCError(nil)
	fmt.Printf("test %s", 3*time.Second)
}

func TestServer_Basic(t *testing.T) {
	// assert := common.NewAssert(t)
	server := NewServer(true, 1024, 32, nil).AddAdapter(
		NewWebSocketServerAdapter("127.0.0.1:8080", "test"),
	)
	server.AddService("user", NewService().
		Reply("sayHello", true, func(ctx Context, name string) Return {
			time.Sleep(20 * time.Second)
			return ctx.OK("hello " + name)
		}),
	)

	server.Open()

	// client
	go func() {
		client := NewClient(
			NewWebSocketClientEndPoint("ws://127.0.0.1:8080/test"),
		)

		_ = client.Open()

		for i := 0; i < 100; i++ {
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