package router

import (
	"testing"
	"time"

	"github.com/rpccloud/rpc/internal/rpc"
)

func TestDebug(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		server := NewServer("127.0.0.1:8080", nil, rpc.NewTestStreamReceiver())
		go func() {
			server.Open()
			server.Run()
		}()

		client, _ := NewClient("127.0.0.1:8080", nil, rpc.NewTestStreamReceiver())
		//go func() {
		//	client.Open()
		//	client.Run()
		//}()

		stream := rpc.NewStream()
		client.SendStream(stream)

		time.Sleep(time.Second)
	})
}

func BenchmarkRouter_AddSlot(b *testing.B) {
	ch := make(chan bool)

	for i := 0; i < b.N; i++ {
		select {
		case <-ch:
		case <-time.After(10 * time.Microsecond):
		}
	}
}
