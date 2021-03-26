package router

import (
	"github.com/rpccloud/rpc/internal/rpc"
	"testing"
	"time"
)

func TestDebug(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		hub := rpc.NewTestStreamHub()

		server := NewServer("127.0.0.1:8080", nil, hub)
		go func() {
			server.Open()
			server.Run()
		}()

		client := NewClient("127.0.0.1:8080", nil)
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
