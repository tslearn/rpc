package rpc

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
	userService := NewService().
		Reply("SayHello", func(ctx Context, userName string) Return {
			return ctx.OK("hello " + userName)
		})

	server := NewServer().
		AddService("user", userService).
		ListenWebSocket("127.0.0.1:8080")

	go func() {
		server.Serve()
	}()

	time.Sleep(time.Second)
	client, err := Dial("ws://127.0.0.1:8080/")
	if err != nil {
		panic(err)
	}

	for i := 0; i < 50; i++ {
		go func(idx int) {
			fmt.Println(client.sendMessage(
				5*time.Second,
				"#.user:SayHello",
				fmt.Sprintf("ts%d", idx),
			))
		}(i)

		time.Sleep(30 * time.Millisecond)
	}

	time.Sleep(3 * time.Second)
	_ = client.Close()

	server.Close()
}
