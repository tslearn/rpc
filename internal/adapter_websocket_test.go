package internal

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

func testWithStreamConn(
	runOnServer func(IAdapter, IStreamConn),
	runOnClient func(IAdapter, IStreamConn),
) []Error {
	ret := make([]Error, 0)
	lock := NewLock()
	fnOnError := func(err Error) {
		lock.DoWithLock(func() {
			ret = append(ret, err)
		})
	}

	waitCH := make(chan bool)
	serverAdapter := NewWebSocketServerAdapter("127.0.0.1:12345")
	clientAdapter := NewWebSocketClientAdapter("ws://127.0.0.1:12345")
	go func() {
		serverAdapter.Open(func(conn IStreamConn) {
			runOnServer(serverAdapter, conn)
		}, fnOnError)
		waitCH <- true
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		clientAdapter.Open(func(conn IStreamConn) {
			runOnClient(clientAdapter, conn)
		}, fnOnError)
		time.Sleep(100 * time.Millisecond)
		serverAdapter.Close(fnOnError)
		waitCH <- true
	}()

	for i := 0; i < 2; i++ {
		<-waitCH
	}

	return ret
}

func makeConnSetReadDeadlineError(conn *websocket.Conn) {
	fnGetField := func(objPointer interface{}, fileName string) unsafe.Pointer {
		val := reflect.Indirect(reflect.ValueOf(objPointer))
		return unsafe.Pointer(val.FieldByName(fileName).UnsafeAddr())
	}

	// Network file descriptor.
	type netFD struct{}

	netConnPtr := (*net.Conn)(fnGetField(conn, "conn"))
	fdPointer := (**netFD)(fnGetField(*netConnPtr, "fd"))
	*fdPointer = nil
}

func TestToTransportError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(toTransportError(nil)).IsNil()

	// Test(2)
	assert(toTransportError(&websocket.CloseError{
		Code: websocket.CloseNormalClosure,
	})).Equals(ErrTransportStreamConnIsClosed)

	// Test(3)
	assert(toTransportError(&websocket.CloseError{
		Code: websocket.CloseAbnormalClosure,
	})).Equals(NewTransportError("websocket: close 1006 (abnormal closure)"))
}

func TestNewWebSocketStreamConn(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	wsConn := &websocket.Conn{}
	sc1 := newWebSocketStreamConn(wsConn)
	assert(fmt.Sprintf("%p", wsConn.CloseHandler())).
		Equals(fmt.Sprintf("%p", sc1.onCloseMessage))
	assert(sc1.status).Equals(webSocketStreamConnRunning)
	assert(sc1.conn).Equals(wsConn)
	assert(cap(sc1.closeCH)).Equals(1)

	// Test(2)
	assert(newWebSocketStreamConn(nil)).IsNil()
}

func TestWebSocketStreamConn_onCloseMessage(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(testWithStreamConn(
		func(server IAdapter, conn IStreamConn) {
			for {
				if _, err := conn.ReadStream(time.Second, 999999); err != nil {
					return
				}
			}
		},
		func(client IAdapter, conn IStreamConn) {
			testConn := conn.(*webSocketStreamConn)
			atomic.StoreInt32(&testConn.status, webSocketStreamConnRunning)
			// 1. onCloseMessage case
			//    webSocketStreamConnRunning => webSocketStreamConnCanClose
			assert(testConn.onCloseMessage(
				websocket.CloseNormalClosure,
				"",
			)).IsNil()
			assert(atomic.LoadInt32(&testConn.status)).
				Equals(webSocketStreamConnCanClose)
		},
	)).Equals([]Error{})

	// Test(2)
	assert(testWithStreamConn(
		func(server IAdapter, conn IStreamConn) {
			for {
				if _, err := conn.ReadStream(time.Second, 999999); err != nil {
					return
				}
			}
		},
		func(client IAdapter, conn IStreamConn) {
			testConn := conn.(*webSocketStreamConn)
			atomic.StoreInt32(&testConn.status, webSocketStreamConnClosing)
			// 1. onCloseMessage case
			//    webSocketStreamConnClosing => webSocketStreamConnCanClose
			assert(testConn.onCloseMessage(
				websocket.CloseNormalClosure,
				"",
			)).IsNil()
			assert(atomic.LoadInt32(&testConn.status)).
				Equals(webSocketStreamConnCanClose)
		},
	)).Equals([]Error{})

	// Test(3)
	assert(testWithStreamConn(
		func(server IAdapter, conn IStreamConn) {
			for {
				if _, err := conn.ReadStream(time.Second, 999999); err != nil {
					return
				}
			}
		},
		func(client IAdapter, conn IStreamConn) {
			testConn := conn.(*webSocketStreamConn)
			atomic.StoreInt32(&testConn.status, webSocketStreamConnCanClose)
			// 1. onCloseMessage case
			//    webSocketStreamConnCanClose (error status)
			assert(testConn.onCloseMessage(
				websocket.CloseNormalClosure,
				"",
			)).IsNil()
			assert(atomic.LoadInt32(&testConn.status)).
				Equals(webSocketStreamConnCanClose)
		},
	)).Equals([]Error{})

	// Test(4)
	assert(testWithStreamConn(
		func(server IAdapter, conn IStreamConn) {
			for {
				if _, err := conn.ReadStream(time.Second, 999999); err != nil {
					return
				}
			}
		},
		func(client IAdapter, conn IStreamConn) {
			testConn := conn.(*webSocketStreamConn)
			atomic.StoreInt32(&testConn.status, webSocketStreamConnClosed)
			// 1. onCloseMessage case
			//    webSocketStreamConnClosed (error status)
			assert(testConn.onCloseMessage(
				websocket.CloseNormalClosure,
				"",
			)).IsNil()
			assert(atomic.LoadInt32(&testConn.status)).
				Equals(webSocketStreamConnClosed)
		},
	)).Equals([]Error{})
}

func TestWebSocketStreamConn_ReadStream(t *testing.T) {
	assert := NewAssert(t)

	// Test(1) status is not webSocketStreamConnRunning
	assert(testWithStreamConn(
		func(server IAdapter, conn IStreamConn) {
			for {
				if _, err := conn.ReadStream(time.Second, 999999); err != nil {
					return
				}
			}
		},
		func(client IAdapter, conn IStreamConn) {
			testConn := conn.(*webSocketStreamConn)
			atomic.StoreInt32(&testConn.status, webSocketStreamConnCanClose)
			// status is not running
			assert(conn.ReadStream(time.Second, 999999)).
				Equals((*Stream)(nil), ErrTransportStreamConnIsClosed)
			assert(atomic.LoadInt32(&testConn.reading)).Equals(int32(0))
		},
	)).Equals([]Error{})

	// Test(2) SetReadDeadline Error
	testWithStreamConn(
		func(server IAdapter, conn IStreamConn) {
			for {
				if _, err := conn.ReadStream(time.Second, 999999); err != nil {
					return
				}
			}
		},
		func(client IAdapter, conn IStreamConn) {
			testConn := conn.(*webSocketStreamConn)
			// make error
			makeConnSetReadDeadlineError(testConn.conn)
			stream, err := testConn.ReadStream(-time.Second, 999999)
			assert(atomic.LoadInt32(&testConn.reading)).Equals(int32(0))
			assert(stream).IsNil()
			assert(err).IsNotNil()
		},
	)

	// Test(3) ReadMessage timeout
	assert(testWithStreamConn(
		func(server IAdapter, conn IStreamConn) {
			for {
				if _, err := conn.ReadStream(time.Second, 999999); err != nil {
					return
				}
			}
		},
		func(client IAdapter, conn IStreamConn) {
			testConn := conn.(*webSocketStreamConn)
			// ReadMessage timeout
			stream, err := testConn.ReadStream(-time.Second, 999999)
			assert(atomic.LoadInt32(&testConn.reading)).Equals(int32(0))
			assert(stream).IsNil()
			if err != nil {
				assert(strings.Contains(err.GetMessage(), "timeout")).IsTrue()
			}
		},
	)).Equals([]Error{})

	// Test(4) type is websocket.TextMessage
	assert(testWithStreamConn(
		func(server IAdapter, conn IStreamConn) {
			testConn := conn.(*webSocketStreamConn)
			_ = testConn.conn.WriteMessage(websocket.TextMessage, []byte("hello"))
			for {
				if _, err := conn.ReadStream(time.Second, 999999); err != nil {
					return
				}
			}
		},
		func(client IAdapter, conn IStreamConn) {
			testConn := conn.(*webSocketStreamConn)
			// ReadMessage type is websocket.TextMessage
			assert(testConn.ReadStream(-time.Second, 999999)).Equals(
				(*Stream)(nil),
				NewTransportError("unsupported websocket protocol"),
			)
			assert(atomic.LoadInt32(&testConn.reading)).Equals(int32(0))
		},
	)).Equals([]Error{})

	// Test(5) OK
	assert(testWithStreamConn(
		func(server IAdapter, conn IStreamConn) {
			testConn := conn.(*webSocketStreamConn)
			_ = testConn.conn.WriteMessage(websocket.BinaryMessage, []byte("hello"))
			for {
				if _, err := conn.ReadStream(time.Second, 999999); err != nil {
					return
				}
			}
		},
		func(client IAdapter, conn IStreamConn) {
			testConn := conn.(*webSocketStreamConn)
			// ReadMessage type is websocket.TextMessage
			stream, err := testConn.ReadStream(-time.Second, 999999)
			assert(atomic.LoadInt32(&testConn.reading)).Equals(int32(0))
			assert(stream).IsNotNil()
			assert(err).IsNil()
			assert(string(stream.GetBufferUnsafe())).Equals("hello")
		},
	)).Equals([]Error{})
}

func TestWebSocketStreamConn_WriteStream(t *testing.T) {
	assert := NewAssert(t)

	// Test(1) stream is nil
	assert(testWithStreamConn(
		func(server IAdapter, conn IStreamConn) {
			for {
				if _, err := conn.ReadStream(time.Second, 999999); err != nil {
					return
				}
			}
		},
		func(client IAdapter, conn IStreamConn) {
			testConn := conn.(*webSocketStreamConn)
			err := testConn.WriteStream(nil, time.Second)
			assert(err).IsNotNil()
			assert(err.GetMessage()).Equals("stream is nil")
			assert(strings.Contains(err.GetDebug(), "adapter_websocket.go"))
			assert(strings.Contains(err.GetDebug(), "WriteStream"))
		},
	)).Equals([]Error{})

	// Test(2) status is not webSocketStreamConnRunning
	assert(testWithStreamConn(
		func(server IAdapter, conn IStreamConn) {
			for {
				if _, err := conn.ReadStream(time.Second, 999999); err != nil {
					return
				}
			}
		},
		func(client IAdapter, conn IStreamConn) {
			testConn := conn.(*webSocketStreamConn)
			atomic.StoreInt32(&testConn.status, webSocketStreamConnClosed)
			assert(testConn.WriteStream(NewStream(), time.Second)).
				Equals(ErrTransportStreamConnIsClosed)
		},
	)).Equals([]Error{})

	// Test(3) timeout
	assert(testWithStreamConn(
		func(server IAdapter, conn IStreamConn) {
			for {
				if _, err := conn.ReadStream(time.Second, 999999); err != nil {
					return
				}
			}
		},
		func(client IAdapter, conn IStreamConn) {
			testConn := conn.(*webSocketStreamConn)
			err := testConn.WriteStream(NewStream(), -time.Second)
			assert(err).IsNotNil()
			assert(strings.Contains(err.GetMessage(), "timeout")).IsTrue()
		},
	)).Equals([]Error{})

	// Test(4) OK
	assert(testWithStreamConn(
		func(server IAdapter, conn IStreamConn) {
			for {
				if _, err := conn.ReadStream(time.Second, 999999); err != nil {
					return
				}
			}
		},
		func(client IAdapter, conn IStreamConn) {
			testConn := conn.(*webSocketStreamConn)
			assert(testConn.WriteStream(NewStream(), time.Second)).IsNil()
		},
	)).Equals([]Error{})
}

func TestWebSocketStreamConn_Close(t *testing.T) {
	assert := NewAssert(t)
	// Test(1) webSocketStreamConnRunning => webSocketStreamConnClosing no wait
	assert(testWithStreamConn(
		func(server IAdapter, conn IStreamConn) {
			for {
				if _, err := conn.ReadStream(time.Second, 999999); err != nil {
					return
				}
			}
		},
		func(client IAdapter, conn IStreamConn) {
			assert(conn.Close()).IsNil()
		},
	)).Equals([]Error{})

	// Test(2) webSocketStreamConnRunning => webSocketStreamConnClosing wait ok
	assert(testWithStreamConn(
		func(server IAdapter, conn IStreamConn) {
			for {
				if _, err := conn.ReadStream(time.Second, 999999); err != nil {
					return
				}
			}
		},
		func(client IAdapter, conn IStreamConn) {
			go func() {
				_, _ = conn.ReadStream(10*time.Second, 999999)
			}()
			time.Sleep(100 * time.Millisecond)
			assert(conn.Close()).IsNil()
		},
	)).Equals([]Error{})

	// Test(3) webSocketStreamConnRunning => webSocketStreamConnClosing
	// wait failed
	assert(testWithStreamConn(
		func(server IAdapter, conn IStreamConn) {
			time.Sleep(2500 * time.Millisecond)
		},
		func(client IAdapter, conn IStreamConn) {
			go func() {
				_, _ = conn.ReadStream(10*time.Second, 999999)
			}()
			time.Sleep(100 * time.Millisecond)
			assert(conn.Close()).IsNil()
		},
	)).Equals([]Error{})

	// Test(4) webSocketStreamConnCanClose => webSocketStreamConnClosed
	assert(testWithStreamConn(
		func(server IAdapter, conn IStreamConn) {
			for {
				if _, err := conn.ReadStream(time.Second, 999999); err != nil {
					return
				}
			}
		},
		func(client IAdapter, conn IStreamConn) {
			testConn := conn.(*webSocketStreamConn)
			atomic.StoreInt32(&testConn.status, webSocketStreamConnCanClose)
			testConn.closeCH <- true
			assert(conn.Close()).IsNil()
		},
	)).Equals([]Error{})

	// Test(5) webSocketStreamConnClosed => webSocketStreamConnClosed
	assert(testWithStreamConn(
		func(server IAdapter, conn IStreamConn) {
			for {
				if _, err := conn.ReadStream(time.Second, 999999); err != nil {
					return
				}
			}
		},
		func(client IAdapter, conn IStreamConn) {
			testConn := conn.(*webSocketStreamConn)
			atomic.StoreInt32(&testConn.status, webSocketStreamConnClosed)
			assert(conn.Close()).IsNil()
		},
	)).Equals([]Error{})
}

func TestNewWebSocketServerAdapter(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewWebSocketServerAdapter("addrString")).Equals(&wsServerAdapter{
		addr:     "addrString",
		wsServer: nil,
	})
}

func TestWsServerAdapter_Open(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(testRunWithCatchPanic(func() {
		NewWebSocketServerAdapter("test").Open(func(conn IStreamConn) {}, nil)
	})).Equals("onError is nil")

	//assert(err1.GetMessage()).Equals("onError is nil")
	//assert(strings.Contains(err1.GetDebug(), "TestWsServerAdapter_Open"))
	//assert(strings.Contains(err1.GetDebug(), "adapter_websocket.go"))
}

func TestWsServerAdapter_Close(t *testing.T) {

}

//
//func TestNewWebSocketServerEndPoint(t *testing.T) {
//	assert := internal.NewAssert(t)
//
//	server := NewWebSocketServerAdapter("127.0.0.1:20080", "/test")
//
//	for i := 0; i < 2; i++ {
//		assert(
//			server.Open(func(conn IStreamConn) {
//				fmt.Println(conn)
//			}, func(err Error) {
//				fmt.Println(err)
//			}),
//		).IsTrue()
//
//		time.Sleep(2 * time.Second)
//
//		assert(
//			server.Close(func(err Error) {
//				fmt.Println(err)
//			}),
//		).IsTrue()
//	}
//}
//
//func TestNewWebSocketClientEndPoint(t *testing.T) {
//	assert := internal.NewAssert(t)
//	server := NewWebSocketServerAdapter("127.0.0.1:20080", "/test")
//	server.Open(func(conn IStreamConn) {
//		time.Sleep(3 * time.Second)
//	}, func(err Error) {
//		fmt.Println(err)
//	})
//
//	client := NewWebSocketClientAdapter("ws://127.0.0.1:20080/test")
//	assert(
//		client.Open(func(conn IStreamConn) {
//			time.Sleep(3 * time.Second)
//		}, func(err Error) {
//			fmt.Println(err)
//		}),
//	).IsTrue()
//
//	time.Sleep(1 * time.Second)
//
//	assert(
//		client.Close(func(err Error) {
//			fmt.Println(err)
//		}),
//	).IsTrue()
//
//	server.Close(func(err Error) {
//		fmt.Println(err)
//	})
//}
