package client

import (
	"crypto/tls"
	"github.com/rpccloud/rpc"
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/rpccloud/rpc/internal/server"
)

type testNetConn struct {
	writeCH   chan []byte
	isRunning bool
}

func newTestNetConn() *testNetConn {
	return &testNetConn{
		isRunning: true,
		writeCH:   make(chan []byte, 1024),
	}
}

func (p *testNetConn) Read(_ []byte) (n int, err error) {
	panic("not implemented")
}

func (p *testNetConn) Write(b []byte) (n int, err error) {
	buf := make([]byte, len(b))
	copy(buf, b)
	p.writeCH <- buf
	return len(b), nil
}

func (p *testNetConn) Close() error {
	p.isRunning = false
	return nil
}

func (p *testNetConn) LocalAddr() net.Addr {
	panic("not implemented")
}

func (p *testNetConn) RemoteAddr() net.Addr {
	panic("not implemented")
}

func (p *testNetConn) SetDeadline(_ time.Time) error {
	panic("not implemented")
}

func (p *testNetConn) SetReadDeadline(_ time.Time) error {
	panic("not implemented")
}

func (p *testNetConn) SetWriteDeadline(_ time.Time) error {
	panic("not implemented")
}

func getTestServer() *server.Server {
	userService := rpc.NewService().
		On("SayHello", func(rt rpc.Runtime, name rpc.String) rpc.Return {
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

func TestClient_tryToSendPing(t *testing.T) {
	t.Run("p.conn == nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{}

		v.tryToSendPing(1)
		assert(v.lastPingTimeNS).Equal(int64(0))
	})

	t.Run("do not need to ping", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{
			lastPingTimeNS: base.TimeNow().UnixNano(),
			config:         &Config{heartbeat: time.Second},
		}
		netConn := newTestNetConn()
		syncConn := adapter.NewClientSyncConn(netConn, 1200, 1200)
		streamConn := adapter.NewStreamConn(syncConn, v)
		syncConn.SetNext(streamConn)
		v.conn = streamConn

		v.tryToSendPing(base.TimeNow().UnixNano())
		assert(len(netConn.writeCH)).Equal(0)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{
			lastPingTimeNS: 10000,
			config:         &Config{heartbeat: time.Second},
		}
		netConn := newTestNetConn()
		syncConn := adapter.NewClientSyncConn(netConn, 1200, 1200)
		streamConn := adapter.NewStreamConn(syncConn, v)
		syncConn.SetNext(streamConn)
		v.conn = streamConn

		v.tryToSendPing(base.TimeNow().UnixNano())

		stream := core.NewStream()
		stream.PutBytesTo(<-netConn.writeCH, 0)
		assert(stream.ReadInt64()).Equal(int64(core.ControlStreamPing), nil)
		assert(stream.IsReadFinish()).IsTrue()
		assert(stream.CheckStream()).IsTrue()
	})
}

func TestClient_tryToTimeout(t *testing.T) {
	fnTest := func(totalItems int, timeoutItems int) bool {
		v := &Client{
			config: &Config{heartbeatTimeout: 9 * time.Millisecond},
		}

		if totalItems < 0 || totalItems < timeoutItems {
			panic("error")
		}

		beforeData := make([]*SendItem, totalItems)
		afterData := make([]*SendItem, totalItems)
		nowNS := base.TimeNow().UnixNano()

		for i := 0; i < totalItems; i++ {
			item := NewSendItem(int64(time.Second))
			item.startTimeNS = nowNS

			if v.preSendTail == nil {
				v.preSendHead = item
				v.preSendTail = item
			} else {
				v.preSendTail.next = item
				v.preSendTail = item
			}

			beforeData[i] = item
			afterData[i] = item
		}

		for i := 0; i < timeoutItems; i++ {
			rand.Seed(time.Now().UnixNano())
			idx := rand.Int() % len(afterData)
			afterData[idx].timeoutNS = 0
			afterData = append(afterData[:idx], afterData[idx+1:]...)
		}

		fnCheck := func(c *Client, arr []*SendItem) bool {
			if len(arr) == 0 {
				return c.preSendHead == nil && c.preSendTail == nil
			}

			if c.preSendHead != arr[0] || c.preSendTail != arr[len(arr)-1] {
				return false
			}

			if c.preSendTail.next != nil {
				return false
			}

			for i := 0; i < len(arr)-1; i++ {
				if arr[i].next != arr[i+1] {
					return false
				}
			}

			return true
		}

		if !fnCheck(v, beforeData) {
			return false
		}

		v.tryToTimeout(nowNS + int64(500*time.Millisecond))
		return fnCheck(v, afterData)
	}

	t.Run("check if the channels has been swept", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{
			lastPingTimeNS: 10000,
			config:         &Config{heartbeatTimeout: 9 * time.Millisecond},
			channels:       make([]Channel, 1),
		}
		item := NewSendItem(int64(5 * time.Millisecond))
		v.channels[0].Use(item, 1)

		v.tryToTimeout(item.sendTimeNS + int64(4*time.Millisecond))
		assert(v.channels[0].sequence).Equal(uint64(1))
		assert(v.channels[0].item).IsNotNil()

		v.tryToTimeout(item.sendTimeNS + int64(10*time.Millisecond))
		assert(v.channels[0].sequence).Equal(uint64(1))
		assert(v.channels[0].item).IsNil()
	})

	t.Run("check if the conn has been swept", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{
			lastPingTimeNS: 10000,
			config:         &Config{heartbeatTimeout: 9 * time.Millisecond},
			channels:       make([]Channel, 1),
		}

		// conn is nil
		v.tryToTimeout(base.TimeNow().UnixNano() + int64(4*time.Millisecond))
		assert(v.conn).IsNil()

		// set conn
		netConn := newTestNetConn()
		syncConn := adapter.NewClientSyncConn(netConn, 1200, 1200)
		streamConn := adapter.NewStreamConn(syncConn, v)
		syncConn.SetNext(streamConn)
		v.conn = streamConn

		// conn is active
		v.tryToTimeout(base.TimeNow().UnixNano() + int64(4*time.Millisecond))
		assert(netConn.isRunning).IsTrue()

		// conn is not active
		v.tryToTimeout(base.TimeNow().UnixNano() + int64(20*time.Millisecond))
		assert(netConn.isRunning).IsFalse()
	})

	t.Run("item timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		for n := 0; n < 10; n++ {
			for i := 0; i < 10; i++ {
				for j := 0; j <= i; j++ {
					assert(fnTest(i, j)).IsTrue()
				}
			}
		}
	})
}

func TestClient_tryToDeliverPreSendMessages(t *testing.T) {
	fnTest := func(totalPreItems int, chSize int, chFree int) bool {
		if chSize < chFree {
			panic("error")
		}

		v := &Client{
			lastPingTimeNS: 10000,
			config:         &Config{heartbeatTimeout: 9 * time.Millisecond},
			channels:       make([]Channel, chSize),
		}

		netConn := newTestNetConn()
		syncConn := adapter.NewClientSyncConn(netConn, 1200, 1200)
		streamConn := adapter.NewStreamConn(syncConn, v)
		syncConn.SetNext(streamConn)
		v.conn = streamConn
		chFreeArr := make([]int, chSize)

		for i := 0; i < len(v.channels); i++ {
			(&v.channels[i]).sequence = uint64(i)
			(&v.channels[i]).item = nil
			chFreeArr[i] = i
		}

		itemsArray := make([]*SendItem, totalPreItems)
		for i := 0; i < totalPreItems; i++ {
			itemsArray[i] = NewSendItem(int64(time.Second))
			if v.preSendHead == nil {
				v.preSendHead = itemsArray[i]
				v.preSendTail = itemsArray[i]
			} else {
				v.preSendTail.next = itemsArray[i]
				v.preSendTail = itemsArray[i]
			}
		}

		fnCheck := func(c *Client, arr []*SendItem) bool {
			if len(arr) == 0 {
				return c.preSendHead == nil && c.preSendTail == nil
			}

			if c.preSendHead != arr[0] || c.preSendTail != arr[len(arr)-1] {
				return false
			}

			if c.preSendTail.next != nil {
				return false
			}

			for i := 0; i < len(arr)-1; i++ {
				if arr[i].next != arr[i+1] {
					return false
				}
			}

			return true
		}

		if !fnCheck(v, itemsArray) {
			panic("error")
		}

		for len(chFreeArr) > chFree {
			rand.Seed(base.TimeNow().UnixNano())
			idx := rand.Int() % len(chFreeArr)
			(&v.channels[chFreeArr[idx]]).item = NewSendItem(int64(time.Second))
			chFreeArr = append(chFreeArr[:idx], chFreeArr[idx+1:]...)
		}

		v.tryToDeliverPreSendMessages()

		return fnCheck(v, itemsArray[base.MinInt(len(itemsArray), chFree):])
	}

	t.Run("p.conn == nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{
			lastPingTimeNS: 10000,
			config:         &Config{heartbeatTimeout: 9 * time.Millisecond},
			channels:       make([]Channel, 1),
			preSendHead:    NewSendItem(0),
		}
		v.tryToDeliverPreSendMessages()
		assert(v.preSendHead).IsNotNil()
	})

	t.Run("p.channel == nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{
			lastPingTimeNS: 10000,
			config:         &Config{heartbeatTimeout: 9 * time.Millisecond},
			conn:           adapter.NewStreamConn(nil, nil),
			preSendHead:    NewSendItem(0),
		}
		v.tryToDeliverPreSendMessages()
		assert(v.preSendHead).IsNotNil()
	})

	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for n := 16; n <= 32; n++ {
			for i := 0; i < 10; i++ {
				for j := 0; j <= n; j++ {
					assert(fnTest(i, n, j)).IsTrue()
				}
			}
		}
	})
}

func TestClient_SendMessage(t *testing.T) {
	t.Run("args error", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{
			channels: make([]Channel, 0),
		}

		assert(v.SendMessage(time.Second, "#.user:SayHello", make(chan bool))).
			Equal(nil, errors.ErrUnsupportedValue.AddDebug(
				"value type(chan bool) is not supported",
			))
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		rpcServer := getTestServer()
		defer rpcServer.Close()

		rpcClient := newClient(
			"tcp", "0.0.0.0:8765", nil, 1200, 1200, func(err *base.Error) {},
		)

		waitCH := make(chan []interface{})
		for i := 0; i < 300; i++ {
			go func() {
				v, err := rpcClient.SendMessage(
					3*time.Second,
					"#.user:SayHello",
					"kitty",
				)
				waitCH <- []interface{}{v, err}
			}()
		}

		for i := 0; i < 300; i++ {
			assert(<-waitCH...).Equal("hello kitty", nil)
		}

		rpcClient.Close()
	})
}

func TestClient_Close(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		onError := func(err *base.Error) {}
		v := newClient("tcp", "127.0.0.1:1234", nil, 1200, 1200, onError)
		assert(v.adapter).IsNotNil()
		assert(v.Close()).IsTrue()
		assert(v.adapter).IsNil()
	})
}

func TestClient_OnConnOpen(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{sessionString: "123456"}
		netConn := newTestNetConn()
		syncConn := adapter.NewClientSyncConn(netConn, 1200, 1200)
		streamConn := adapter.NewStreamConn(syncConn, v)
		syncConn.SetNext(streamConn)
		v.conn = streamConn

		v.OnConnOpen(streamConn)

		stream := core.NewStream()
		stream.PutBytesTo(<-netConn.writeCH, 0)
		assert(stream.ReadInt64()).
			Equal(int64(core.ControlStreamConnectRequest), nil)
		assert(stream.ReadString()).Equal("123456", nil)
	})
}

func TestClient_OnConnReadStream(t *testing.T) {
	fnTestClient := func() (*Client, *adapter.StreamConn, *testNetConn) {
		v := &Client{config: &Config{}}
		netConn := newTestNetConn()
		syncConn := adapter.NewClientSyncConn(netConn, 1200, 1200)
		streamConn := adapter.NewStreamConn(syncConn, v)
		syncConn.SetNext(streamConn)
		return v, streamConn, netConn
	}

	t.Run("p.conn == nil, stream.callbackID == 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		stream := core.NewStream()
		stream.SetCallbackID(12)
		v, streamConn, _ := fnTestClient()
		v.onError = func(e *base.Error) { err = e }
		v.OnConnReadStream(streamConn, stream)
		assert(err).Equal(errors.ErrStream)
	})

	t.Run("read kind error", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		stream := core.NewStream()
		stream.SetCallbackID(0)
		v, streamConn, _ := fnTestClient()
		v.onError = func(e *base.Error) { err = e }
		v.OnConnReadStream(streamConn, stream)
		assert(err).Equal(errors.ErrStream)
	})

	t.Run("kind != core.ControlStreamConnectResponse", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		stream := core.NewStream()
		stream.SetCallbackID(0)
		stream.WriteInt64(5432)
		v, streamConn, _ := fnTestClient()
		v.onError = func(e *base.Error) { err = e }
		v.OnConnReadStream(streamConn, stream)
		assert(err).Equal(errors.ErrStream)
	})

	t.Run("read sessionString error", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		stream := core.NewStream()
		stream.SetCallbackID(0)
		stream.WriteInt64(int64(core.ControlStreamConnectResponse))
		v, streamConn, _ := fnTestClient()
		v.onError = func(e *base.Error) { err = e }
		v.OnConnReadStream(streamConn, stream)
		assert(err).Equal(errors.ErrStream)
	})

	t.Run("read numOfChannels error", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		stream := core.NewStream()
		stream.SetCallbackID(0)
		stream.WriteInt64(int64(core.ControlStreamConnectResponse))
		stream.WriteString("12-87654321876543218765432187654321")
		v, streamConn, _ := fnTestClient()
		v.onError = func(e *base.Error) { err = e }
		v.OnConnReadStream(streamConn, stream)
		assert(err).Equal(errors.ErrStream)
	})

	t.Run("numOfChannels config error", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		stream := core.NewStream()
		stream.SetCallbackID(0)
		stream.WriteInt64(int64(core.ControlStreamConnectResponse))
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(0)
		v, streamConn, _ := fnTestClient()
		v.onError = func(e *base.Error) { err = e }
		v.OnConnReadStream(streamConn, stream)
		assert(err).Equal(errors.ErrClientConfig)
	})

	t.Run("read transLimit error", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		stream := core.NewStream()
		stream.SetCallbackID(0)
		stream.WriteInt64(int64(core.ControlStreamConnectResponse))
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(32)
		v, streamConn, _ := fnTestClient()
		v.onError = func(e *base.Error) { err = e }
		v.OnConnReadStream(streamConn, stream)
		assert(err).Equal(errors.ErrStream)
	})

	t.Run("transLimit config error", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		stream := core.NewStream()
		stream.SetCallbackID(0)
		stream.WriteInt64(int64(core.ControlStreamConnectResponse))
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(32)
		stream.WriteInt64(0)
		v, streamConn, _ := fnTestClient()
		v.onError = func(e *base.Error) { err = e }
		v.OnConnReadStream(streamConn, stream)
		assert(err).Equal(errors.ErrClientConfig)
	})

	t.Run("read heartbeat error", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		stream := core.NewStream()
		stream.SetCallbackID(0)
		stream.WriteInt64(int64(core.ControlStreamConnectResponse))
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(32)
		stream.WriteInt64(4 * 1024 * 1024)
		v, streamConn, _ := fnTestClient()
		v.onError = func(e *base.Error) { err = e }
		v.OnConnReadStream(streamConn, stream)
		assert(err).Equal(errors.ErrStream)
	})

	t.Run("heartbeat config error", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		stream := core.NewStream()
		stream.SetCallbackID(0)
		stream.WriteInt64(int64(core.ControlStreamConnectResponse))
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(32)
		stream.WriteInt64(4 * 1024 * 1024)
		stream.WriteInt64(0)
		v, streamConn, _ := fnTestClient()
		v.onError = func(e *base.Error) { err = e }
		v.OnConnReadStream(streamConn, stream)
		assert(err).Equal(errors.ErrClientConfig)
	})

	t.Run("read heartbeatTimeout error", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		stream := core.NewStream()
		stream.SetCallbackID(0)
		stream.WriteInt64(int64(core.ControlStreamConnectResponse))
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(32)
		stream.WriteInt64(4 * 1024 * 1024)
		stream.WriteInt64(int64(time.Second))
		v, streamConn, _ := fnTestClient()
		v.onError = func(e *base.Error) { err = e }
		v.OnConnReadStream(streamConn, stream)
		assert(err).Equal(errors.ErrStream)
	})

	t.Run("heartbeatTimeout config error", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		stream := core.NewStream()
		stream.SetCallbackID(0)
		stream.WriteInt64(int64(core.ControlStreamConnectResponse))
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(32)
		stream.WriteInt64(4 * 1024 * 1024)
		stream.WriteInt64(int64(time.Second))
		stream.WriteInt64(0)
		v, streamConn, _ := fnTestClient()
		v.onError = func(e *base.Error) { err = e }
		v.OnConnReadStream(streamConn, stream)
		assert(err).Equal(errors.ErrClientConfig)
	})

	t.Run("stream is not finish", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		stream := core.NewStream()
		stream.SetCallbackID(0)
		stream.WriteInt64(int64(core.ControlStreamConnectResponse))
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(32)
		stream.WriteInt64(4 * 1024 * 1024)
		stream.WriteInt64(int64(time.Second))
		stream.WriteInt64(int64(2 * time.Second))
		stream.WriteBool(false)
		v, streamConn, _ := fnTestClient()
		v.onError = func(e *base.Error) { err = e }
		v.OnConnReadStream(streamConn, stream)
		assert(err).Equal(errors.ErrStream)
	})

	t.Run("ok, sessionString != p.sessionString", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		stream := core.NewStream()
		stream.SetCallbackID(0)
		stream.WriteInt64(int64(core.ControlStreamConnectResponse))
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(32)
		stream.WriteInt64(4 * 1024 * 1024)
		stream.WriteInt64(int64(time.Second))
		stream.WriteInt64(int64(2 * time.Second))
		v, streamConn, _ := fnTestClient()
		v.onError = func(e *base.Error) { err = e }
		v.OnConnReadStream(streamConn, stream)
		assert(err).IsNil()
		assert(v.sessionString).Equal("12-87654321876543218765432187654321")

		assert(v.config.numOfChannels).Equal(32)
		assert(v.config.transLimit).Equal(4 * 1024 * 1024)
		assert(v.config.heartbeat).Equal(1 * time.Second)
		assert(v.config.heartbeatTimeout).Equal(2 * time.Second)
		for i := 0; i < 32; i++ {
			assert(v.channels[i].sequence).Equal(uint64(i))
			assert(v.channels[i].item).IsNil()
		}
		assert(v.lastPingTimeNS > 0).IsTrue()
	})

	t.Run("ok, sessionString == p.sessionString", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := core.NewStream()
		stream.SetCallbackID(0)
		stream.WriteInt64(int64(core.ControlStreamConnectResponse))
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(32)
		stream.WriteInt64(4 * 1024 * 1024)
		stream.WriteInt64(int64(time.Second))
		stream.WriteInt64(int64(2 * time.Second))
		v, streamConn, netConn := fnTestClient()
		v.channels = make([]Channel, 32)
		for i := 0; i < 32; i++ {
			(&v.channels[i]).sequence = uint64(i)
			(&v.channels[i]).Use(NewSendItem(0), 32)
		}

		v.sessionString = "12-87654321876543218765432187654321"
		v.OnConnReadStream(streamConn, stream)
		assert(len(netConn.writeCH)).Equal(32)
		assert(v.lastPingTimeNS > 0).IsTrue()
	})

	t.Run("p.conn != nil, callbackID == 0, 01", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		stream := core.NewStream()
		v, streamConn, _ := fnTestClient()
		v.conn = streamConn
		v.onError = func(e *base.Error) { err = e }
		v.OnConnReadStream(streamConn, stream)
		assert(err).Equal(errors.ErrStream)
	})

	t.Run("p.conn != nil, callbackID == 0, 02", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		stream := core.NewStream()
		stream.WriteInt64(5432)
		v, streamConn, _ := fnTestClient()
		v.conn = streamConn
		v.onError = func(e *base.Error) { err = e }
		v.OnConnReadStream(streamConn, stream)
		assert(err).Equal(errors.ErrStream)
	})

	t.Run("p.conn != nil, callbackID == 0, 03", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		stream := core.NewStream()
		stream.WriteInt64(int64(core.ControlStreamPong))
		v, streamConn, _ := fnTestClient()
		v.conn = streamConn
		v.onError = func(e *base.Error) { err = e }
		v.OnConnReadStream(streamConn, stream)
		assert(err).IsNil()
	})

	t.Run("p.conn != nil, callbackID != 0, 01", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := core.NewStream()
		stream.SetCallbackID(17 + 32)
		stream.WriteInt64(int64(core.ControlStreamPong))
		v, streamConn, _ := fnTestClient()
		v.conn = streamConn
		v.channels = make([]Channel, 32)
		(&v.channels[17]).sequence = 17
		(&v.channels[17]).Use(NewSendItem(0), 32)
		v.OnConnReadStream(streamConn, stream)
		assert(v.channels[17].item).IsNil()
	})

	t.Run("p.conn != nil, callbackID != 0, 02", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		stream := core.NewStream()
		stream.SetCallbackID(17)
		stream.WriteInt64(int64(core.ControlStreamPong))
		v, streamConn, _ := fnTestClient()
		v.conn = streamConn
		v.channels = make([]Channel, 32)
		v.onError = func(e *base.Error) { err = e }
		(&v.channels[17]).sequence = 17
		(&v.channels[17]).Use(NewSendItem(0), 32)
		v.OnConnReadStream(streamConn, stream)
		assert(v.channels[17].item).IsNotNil()
		assert(err).Equal(errors.ErrStream)
	})

	t.Run("p.conn != nil, callbackID != 0, 03", func(t *testing.T) {
		stream := core.NewStream()
		stream.SetCallbackID(17 + 32)
		stream.WriteInt64(int64(core.ControlStreamPong))
		v, streamConn, _ := fnTestClient()
		v.conn = streamConn
		v.OnConnReadStream(streamConn, stream)
	})
}

func TestClient_OnConnError(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		v := &Client{sessionString: "123456", onError: func(e *base.Error) {
			err = e
		}}
		netConn := newTestNetConn()
		syncConn := adapter.NewClientSyncConn(netConn, 1200, 1200)
		streamConn := adapter.NewStreamConn(syncConn, v)
		syncConn.SetNext(streamConn)
		v.conn = streamConn
		v.OnConnError(streamConn, errors.ErrStream)
		assert(err).Equal(errors.ErrStream)
		assert(netConn.isRunning).IsFalse()
	})
}

func TestClient_OnConnClose(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{}
		netConn := newTestNetConn()
		syncConn := adapter.NewClientSyncConn(netConn, 1200, 1200)
		streamConn := adapter.NewStreamConn(syncConn, v)
		syncConn.SetNext(streamConn)
		v.conn = streamConn
		v.OnConnClose(streamConn)
		assert(v.conn).IsNil()
	})
}
