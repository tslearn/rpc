package rpcc

import (
	"fmt"
	"github.com/tslearn/kzserver/common"
	"testing"
	"time"
)

func TestNewWebSocketServerEndPoint(t *testing.T) {
	assert := common.NewAssert(t)

	server := NewWebSocketServerEndPoint("127.0.0.1:20080", "/test")

	for i := 0; i < 2; i++ {
		assert(
			server.Open(func(conn IStreamConn) {
				fmt.Println(conn)
			}, func(err common.RPCError) {
				fmt.Println(err)
			}),
		).IsTrue()

		time.Sleep(2 * time.Second)

		assert(
			server.Close(func(err common.RPCError) {
				fmt.Println(err)
			}),
		).IsTrue()
	}
}

func TestNewWebSocketClientEndPoint(t *testing.T) {
	assert := common.NewAssert(t)
	server := NewWebSocketServerEndPoint("127.0.0.1:20080", "/test")
	server.Open(func(conn IStreamConn) {
		time.Sleep(3 * time.Second)
	}, func(err common.RPCError) {
		fmt.Println(err)
	})

	client := NewWebSocketClientEndPoint("ws://127.0.0.1:20080/test")
	assert(
		client.Open(func(conn IStreamConn) {
			time.Sleep(3 * time.Second)
		}, func(err common.RPCError) {
			fmt.Println(err)
		}),
	).IsTrue()

	time.Sleep(1 * time.Second)

	assert(
		client.Close(func(err common.RPCError) {
			fmt.Println(err)
		}),
	).IsTrue()

	server.Close(func(err common.RPCError) {
		fmt.Println(err)
	})
}
