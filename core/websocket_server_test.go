package core

import (
	"github.com/gorilla/websocket"
	"math"
	"net"
	"net/http"
	"strings"
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

func TestWebSocketServer_AddService(t *testing.T) {
	assert := newAssert(t)
	server := NewWebSocketServer()

	service := newServiceMeta().
		Echo("sayHello", true, func(
			ctx *rpcContext,
			name string,
		) *rpcReturn {
			return ctx.OK("hello " + name)
		})

	logErrorCH := make(chan string, 100)
	logSubscription := server.GetLogger().Subscribe()
	logSubscription.Error = func(msg string) {
		logErrorCH <- msg
	}

	server.AddService("user", service)
	assert(len(logErrorCH)).Equals(0)
	server.AddService("user", service)
	assert(len(logErrorCH)).Equals(1)
	assert(<-logErrorCH).Contains("AddService name user is duplicated")
}

func TestWebSocketServer_StartBackground(t *testing.T) {
	assert := newAssert(t)
	server := NewWebSocketServer()
	logInfoCH := make(chan string, 100)
	logErrorCH := make(chan string, 100)
	logSubscription := server.GetLogger().Subscribe()
	logSubscription.Info = func(msg string) {
		logInfoCH <- msg
	}
	logSubscription.Error = func(msg string) {
		logErrorCH <- msg
	}

	server.StartBackground("0.0.0.0", 55555, "/ws")
	assert(<-logInfoCH).
		Contains("WebSocketServer: start at ws://0.0.0.0:55555/ws")

	server.StartBackground("0.0.0.0", 63333, "/ws")
	assert(<-logErrorCH).Contains("WebSocketServer: has already been started")

	_ = server.Close()
}

func TestWebSocketServer_Start_Close(t *testing.T) {
	assert := newAssert(t)
	server := NewWebSocketServer()
	logInfoCH := make(chan string, 100)
	logErrorCH := make(chan string, 100)
	logSubscription := server.GetLogger().Subscribe()
	logSubscription.Info = func(msg string) {
		logInfoCH <- msg
	}
	logSubscription.Error = func(msg string) {
		logErrorCH <- msg
	}

	// ok
	go func() {
		time.Sleep(200 * time.Millisecond)
		assert(server.isRunning).IsTrue()
		assert(server.processor.isRunning).IsTrue()
		_ = server.Close()
	}()
	assert(server.Start("0.0.0.0", 55555, "/ws")).IsNil()
	assert(server.isRunning).IsFalse()
	assert(server.processor.isRunning).IsFalse()
	assert(<-logInfoCH).
		Contains("WebSocketServer: start at ws://0.0.0.0:55555/ws")
	assert(<-logInfoCH).Contains("WebSocketServer: stopped")

	// error host
	err1 := server.Start("this is wrong", 55555, "/ws")
	assert(err1.GetMessage()).
		Equals("listen tcp: lookup this is wrong: no such host")
	assert(err1.GetDebug()).Equals("")

	// not closed correctly
	err2 := server.Start("0.0.0.0", 55555, "/ws")
	assert(err2.GetMessage()).
		Equals("WebSocketServer: has not been closed correctly")
	_ = server.Close()

	// error port is used
	l, _ := net.Listen("tcp", "0.0.0.0:55555")
	err3 := server.Start("0.0.0.0", 55555, "/ws")
	assert(err3.GetMessage()).
		Equals("listen tcp 0.0.0.0:55555: bind: address already in use")
	_ = server.Close()
	_ = l.Close()
}

func TestWebSocketServer_Start_HandleFunc(t *testing.T) {
	server := NewWebSocketServer().
		AddService("user", newServiceMeta().
			Echo("sayHello", true, func(
				ctx *rpcContext,
				name string,
			) *rpcReturn {
				return ctx.OK("hello " + name)
			}))
	server.StartBackground("0.0.0.0", 12345, "/")

	assert := newAssert(t)
	logInfoCH := make(chan string, 100)
	logErrorCH := make(chan string, 100)
	logWarnCH := make(chan string, 100)
	logSubscription := server.GetLogger().Subscribe()
	logSubscription.Info = func(msg string) {
		logInfoCH <- msg
	}
	logSubscription.Warning = func(msg string) {
		logWarnCH <- msg
	}
	logSubscription.Error = func(msg string) {
		logErrorCH <- msg
	}

	// ok
	client0 := NewWebSocketClient(
		"ws://127.0.0.1:12345/",
		16000,
		128*1024,
	)
	assert(client0.SendMessage("$.user:sayHello", "tianshuo")).
		Equals("hello tianshuo", nil)
	_ = client0.Close()
	assert(<-logInfoCH).Contains("WebSocketServerConn[2]: opened")
	assert(<-logInfoCH).Contains("WebSocketServerConn[2]: closed")

	// upgrade error
	_, _ = http.Get("http://127.0.0.1:12345")
	assert(<-logErrorCH).
		Contains("websocket: the client is not using the websocket protocol")

	// SetReadTimeoutMS 0
	server.SetReadTimeoutMS(0)
	client1 := NewWebSocketClient(
		"ws://127.0.0.1:12345/",
		16000,
		128*1024,
	)
	assert(<-logInfoCH).Contains("WebSocketServerConn[3]: opened")
	openTime := TimeNowMS()
	assert(<-logInfoCH).Contains("WebSocketServerConn[3]: closed")
	closeTime := TimeNowMS()
	assert(<-logWarnCH).Contains("i/o timeout")
	assert(closeTime-openTime < 20).IsTrue()
	_ = client1.Close()

	// SetReadTimeoutMS 100
	server.SetReadTimeoutMS(100)
	client2 := NewWebSocketClient(
		"ws://127.0.0.1:12345/",
		16000,
		128*1024,
	)
	assert(<-logInfoCH).Contains("WebSocketServerConn[4]: opened")
	openTime = TimeNowMS()
	assert(<-logInfoCH).Contains("WebSocketServerConn[4]: closed")
	closeTime = TimeNowMS()
	assert(<-logWarnCH).Contains("i/o timeout")
	assert(closeTime-openTime >= 100).IsTrue()
	assert(closeTime-openTime < 120).IsTrue()
	_ = client2.Close()

	// SetReadTimeoutMS error
	server.GetLogger().Subscribe().Info = func(msg string) {
		if strings.Contains(msg, "WebSocketServerConn[5]: opened") {
			conn := server.getConnByID(5).wsConn
			netConnPtr := (*net.Conn)(GetObjectFieldPointer(conn, "conn"))
			fdPointer := (**int)(GetObjectFieldPointer(*netConnPtr, "fd"))
			*fdPointer = nil
		}
	}

	client3 := NewWebSocketClient(
		"ws://127.0.0.1:12345/",
		16000,
		128*1024,
	)
	assert(<-logInfoCH).Contains("WebSocketServerConn[5]: opened")
	assert(<-logWarnCH).Contains("WebSocketServerConn[5]: invalid argument")
	assert(<-logWarnCH).Contains("WebSocketServerConn[5]: invalid argument")
	assert(<-logInfoCH).Contains("WebSocketServerConn[5]: closed")
	_ = client3.Close()

	assert(len(logInfoCH)).Equals(0)
	assert(len(logWarnCH)).Equals(0)
	assert(len(logErrorCH)).Equals(0)

	_ = server.Close()
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
