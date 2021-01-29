package client

import (
	"crypto/tls"
	"fmt"
	"github.com/rpccloud/rpc"
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/rpccloud/rpc/internal/server"
)

func TestClient_Debug(t *testing.T) {
	userService := rpc.NewService().
		On("SayHello", func(rt rpc.Runtime, name rpc.String) rpc.Return {
			time.Sleep(2 * time.Second)
			return rt.Reply("hello " + name)
		})

	rpcServer := server.NewServer().Listen("ws", "0.0.0.0:28888", nil)
	rpcServer.AddService("user", userService, nil)

	go func() {
		rpcServer.SetNumOfThreads(1024).Serve()
	}()

	time.Sleep(300 * time.Millisecond)
	rpcClient := newClient(
		"ws", "0.0.0.0:28888", nil, 1200, 1200, func(err *base.Error) {},
	)
	for i := 0; i < 10; i++ {
		go func() {
			fmt.Println(
				rpcClient.SendMessage(8*time.Second, "#.user:SayHello", "kitty"),
			)
		}()
	}

	// haha, close it
	time.Sleep(time.Second)
	rpcClient.conn.Close()

	time.Sleep(10 * time.Second)

	rpcClient.Close()
	rpcServer.Close()
}

func getTestServer() *server.Server {
	userService := rpc.NewService().
		On("SayHello", func(rt rpc.Runtime, name rpc.String) rpc.Return {
			time.Sleep(2 * time.Second)
			return rt.Reply("hello " + name)
		}).
		On("Sleep", func(rt rpc.Runtime, timeNS int64) rpc.Return {
			time.Sleep(time.Duration(timeNS))
			return rt.Reply(nil)
		})

	rpcServer := server.NewServer().Listen("tcp", "0.0.0.0:8765", nil)
	rpcServer.AddService("user", userService, nil)

	go func() {
		rpcServer.SetNumOfThreads(1024).Serve()
	}()

	time.Sleep(100 * time.Millisecond)

	return rpcServer
}

func TestNewClient(t *testing.T) {
	type TestAdapter struct {
		isClient   bool
		network    string
		addr       string
		tlsConfig  *tls.Config
		rBufSize   int
		wBufSize   int
		receiver   adapter.IReceiver
		service    base.IORCService
		orcManager *base.ORCManager
	}

	type TestORCManager struct {
		sequence     uint64
		isWaitChange bool
		mu           sync.Mutex
		cond         sync.Cond
	}

	t.Run("test", func(t *testing.T) {
		testServer := getTestServer()
		defer testServer.Close()

		assert := base.NewAssert(t)
		onError := func(err *base.Error) {}
		v := newClient("tcp", "127.0.0.1:8765", nil, 1024, 2048, onError)

		for v.conn == nil {
			time.Sleep(10 * time.Millisecond)
		}

		assert(v.config).Equal(&Config{
			numOfChannels:    32,
			transLimit:       4 * 1024 * 1024,
			heartbeat:        4 * time.Second,
			heartbeatTimeout: 8 * time.Second,
		})
		assert(len(v.sessionString) >= 34).IsTrue()
		testAdapter := (*TestAdapter)(unsafe.Pointer(v.adapter))
		assert(testAdapter.isClient).IsTrue()
		assert(testAdapter.network).Equal("tcp")
		assert(testAdapter.addr).Equal("127.0.0.1:8765")
		assert(testAdapter.tlsConfig).Equal(nil)
		assert(testAdapter.rBufSize).Equal(1024)
		assert(testAdapter.wBufSize).Equal(2048)
		assert(testAdapter.receiver).Equal(v)
		assert(testAdapter.service).IsNotNil()
		adapterOrcManager := (*TestORCManager)(
			unsafe.Pointer(testAdapter.orcManager),
		)
		// orcStatusReady | orcLockBit = 1 | 1 << 2 = 5
		assert(adapterOrcManager.sequence % 8).Equal(uint64(5))
		assert(v.preSendHead).IsNil()
		assert(v.preSendTail).IsNil()
		assert(len(v.channels)).Equal(32)
		assert(v.lastPingTimeNS > 0).IsTrue()
		// orcStatusReady | orcLockBit = 1 | 1 << 2 = 5
		assert((*TestORCManager)(unsafe.Pointer(v.orcManager)).sequence % 8).
			Equal(uint64(5))
		assert(v.onError).IsNotNil()

		// check tryLoop
		_, err := v.SendMessage(
			500*time.Millisecond,
			"#.user:Sleep",
			int64(2*time.Second),
		)
		assert(err).Equal(errors.ErrClientTimeout)
		v.Close()
	})
}
