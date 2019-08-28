package core

import (
	"testing"
	"time"
)

func BenchmarkNewWebSocketServer(b *testing.B) {
	time.Sleep(2 * time.Second)

	client := (*WebSocketClient)(nil)
	for client == nil {
		client = NewWebSocketClient(
			"ws://127.0.0.1:12345/ws",
			16000,
			128*1024,
		)
	}
	b.ReportAllocs()
	b.N = 10000
	b.SetParallelism(500)
	b.ResetTimer()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, _ = client.SendMessage(
			"$.user:sayHello",
			"tianshuo",
		)
	}

	b.StopTimer()
	_ = client.Close()
}
