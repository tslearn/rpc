package core

import (
	"github.com/gorilla/websocket"
	"math"
	"testing"
	"time"
	"unsafe"
)

func TestNewWebSocketServer(t *testing.T) {
	assert := newAssert(t)

	server := NewWebSocketServer()
	assert(server).IsNotNil()
	assert(server.processor).IsNotNil()
	assert(server.logger).IsNotNil()
	assert(server.isRunning).IsFalse()
	assert(server.readSizeLimit).Equals(uint64(64 * 1024))
	assert(server.readTimeoutNS).Equals(60 * uint64(time.Second))
	assert(server.httpServer).IsNil()
	assert(len(server.closeChan)).Equals(0)
	assert(cap(server.closeChan)).Equals(1)
	assert(server.seed).Equals(uint32(1))
}

func TestWebSocketServer_registerConn(t *testing.T) {
	assert := newAssert(t)
	server := NewWebSocketServer()

	wsConn1 := (*websocket.Conn)(unsafe.Pointer(uintptr(343451)))
	assert(server.registerConn(wsConn1)).Equals(uint32(2))

	server.seed = 1
	wsConn2 := (*websocket.Conn)(unsafe.Pointer(uintptr(343452)))
	assert(server.registerConn(wsConn2)).Equals(uint32(3))

	server.seed = math.MaxUint32 - 2
	wsConn3 := (*websocket.Conn)(unsafe.Pointer(uintptr(343453)))
	assert(server.registerConn(wsConn3)).Equals(uint32(math.MaxUint32 - 1))

	server.seed = math.MaxUint32 - 2
	wsConn4 := (*websocket.Conn)(unsafe.Pointer(uintptr(343454)))
	assert(server.registerConn(wsConn4)).Equals(uint32(1))
}

func TestWebSocketServer_unregisterConn(t *testing.T) {
	assert := newAssert(t)
	server := NewWebSocketServer()

	for i := uint32(650000); i < 652399; i++ {
		server.seed = i - 1
		wsConn := (*websocket.Conn)(unsafe.Pointer(uintptr(i)))
		assert(server.registerConn(wsConn)).Equals(i)
	}

	for i := uint32(650000); i < 652399; i++ {
		assert(server.unregisterConn(i)).IsTrue()
		assert(server.unregisterConn(i - 10000)).IsFalse()
	}
}

func TestWebSocketServer_getConnByID(t *testing.T) {
	assert := newAssert(t)
	server := NewWebSocketServer()

	for i := uint32(650000); i < 652399; i++ {
		server.seed = i - 1
		wsConn := (*websocket.Conn)(unsafe.Pointer(uintptr(i)))
		assert(server.registerConn(wsConn)).Equals(i)
	}

	for i := uint32(650000); i < 652399; i++ {
		assert(server.getConnByID(i).wsConn).
			Equals((*websocket.Conn)(unsafe.Pointer(uintptr(i))))
		assert(server.getConnByID(i - 10000)).IsNil()
	}
}

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
