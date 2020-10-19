package rpc

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestServer_Debug(t *testing.T) {
	// err := common.RPCError(nil)
	fmt.Printf("test %s", 3*time.Second)
}

func TestServer_Basic(t *testing.T) {
	seed := uint64(0)
	userService := NewService().
		On("SayHello", func(ctx Runtime, userName string) Return {
			if ret, err := ctx.Call("#.user:GetIndex"); err != nil {
				return ctx.Error(err)
			} else if idx, ok := ret.(uint64); !ok {
				return ctx.Error(errors.New("#.user:GetIndex return type error"))
			} else {
				return ctx.OK(idx)
			}
		}).
		On("GetIndex", func(ctx Runtime) Return {
			return ctx.OK(atomic.AddUint64(&seed, 1))
		})

	server := NewServer().SetNumOfThreads(1).
		AddService("user", userService, nil).
		ListenWebSocket("127.0.0.1:8080")

	go func() {
		server.Serve()
	}()

	time.Sleep(1 * time.Second)
	client, err := Dial("ws://127.0.0.1:8080/")
	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		go func(idx int) {
			fmt.Println(client.SendMessage(
				5*time.Second,
				"#.user:SayHello",
				fmt.Sprintf("ts%d", idx),
			))
		}(i)
		// fmt.Println(client.SendMessage(
		//   5*time.Second,
		//   "#.user:SayHello",
		//   fmt.Sprintf("ts"),
		// ))
		time.Sleep(30 * time.Millisecond)
	}

	_ = client.Close()

	time.Sleep(2 * time.Second)
	server.Close()
}
