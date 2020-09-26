package websocket

import (
	sysError "errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"math"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

type fakeNetListener struct{}

func (p fakeNetListener) Accept() (net.Conn, error) {
	return nil, nil
}
func (p fakeNetListener) Close() error {
	return sysError.New("test error")
}
func (p fakeNetListener) Addr() net.Addr {
	return nil
}

func testHelperStreamConn(
	runOnServer func(core.IServerAdapter, core.IStreamConn),
	runOnClient func(core.IClientAdapter, core.IStreamConn),
) []*base.Error {
	ret := make([]*base.Error, 0)
	lock := &sync.Mutex{}
	fnOnError := func(err *base.Error) {
		lock.Lock()
		defer lock.Unlock()
		ret = append(ret, err)
	}

	if runOnServer == nil {
		runOnServer = func(server core.IServerAdapter, conn core.IStreamConn) {
			for {
				if _, err := conn.ReadStream(
					30*time.Millisecond,
					math.MaxInt64,
				); err != nil {
					return
				}
			}
		}
	}

	waitCH := make(chan bool)
	serverAdapter := NewWebsocketServerAdapter("127.0.0.1:12345")
	clientAdapter := NewWebsocketClientAdapter("ws://127.0.0.1:12345")
	go func() {
		serverAdapter.Open(
			func(conn core.IStreamConn, _ net.Addr) {
				runOnServer(serverAdapter, conn)
				waitCH <- true
			},
			func(_ uint64, err *base.Error) {
				fnOnError(err)
			},
		)
	}()

	for !base.IsTCPPortOccupied(12345) {
		time.Sleep(time.Millisecond)
	}

	go func() {
		clientAdapter.Open(func(conn core.IStreamConn) {
			runOnClient(clientAdapter, conn)
		}, fnOnError)
		serverAdapter.Close(func(u uint64, err *base.Error) {
			fnOnError(err)
		})
		waitCH <- true
	}()

	for i := 0; i < 2; i++ {
		<-waitCH
	}

	return ret
}

func getFieldByPointer(ptr interface{}, fileName string) unsafe.Pointer {
	val := reflect.Indirect(reflect.ValueOf(ptr))
	return unsafe.Pointer(val.FieldByName(fileName).UnsafeAddr())
}

func makeConnFDError(conn *websocket.Conn) {
	// Network file descriptor.
	type netFD struct{}
	netConnPtr := (*net.Conn)(getFieldByPointer(conn, "conn"))
	fdPointer := (**netFD)(getFieldByPointer(*netConnPtr, "fd"))
	*fdPointer = nil
}

func TestConvertToError(t *testing.T) {
	t.Run("err is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(convertToError(nil, errors.ErrRuntimeGeneral)).IsNil()
	})

	t.Run("err is websocket CloseNormalClosure", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(convertToError(
			&websocket.CloseError{Code: websocket.CloseNormalClosure},
			errors.ErrRuntimeGeneral,
		)).Equal(errors.ErrStreamConnIsClosed)
	})

	t.Run("err is others", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(convertToError(sysError.New("error"), errors.ErrStreamConnIsClosed)).
			Equal(errors.ErrStreamConnIsClosed.AddDebug("error"))
		assert(convertToError(
			&websocket.CloseError{Code: websocket.CloseAbnormalClosure},
			errors.ErrRuntimeGeneral,
		)).Equal(errors.ErrRuntimeGeneral.AddDebug(
			"websocket: close 1006 (abnormal closure)",
		))
	})
}

func TestWebsocketStreamConn(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(webSocketStreamConnClosed).Equal(int32(0))
		assert(webSocketStreamConnRunning).Equal(int32(1))
		assert(webSocketStreamConnClosing).Equal(int32(2))
		assert(webSocketStreamConnCanClose).Equal(int32(3))
	})
}

func TestNewWebsocketStreamConn(t *testing.T) {
	t.Run("conn is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(newWebsocketStreamConn(nil)).IsNil()
	})

	t.Run("conn is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		wsConn := &websocket.Conn{}
		v1 := newWebsocketStreamConn(wsConn)
		assert(v1.status).Equal(webSocketStreamConnRunning)
		assert(v1.reading).Equal(int32(0))
		assert(v1.writing).Equal(int32(0))
		assert(cap(v1.closeCH)).Equal(1)
		assert(v1.wsConn).Equal(wsConn)
		assert(fmt.Sprintf("%p", wsConn.CloseHandler())).
			Equal(fmt.Sprintf("%p", v1.onCloseMessage))
	})
}

func TestWebsocketStreamConn_writeMessage(t *testing.T) {
	t.Run("writeMessage failed", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			nil,
			func(client core.IClientAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				makeConnFDError(testConn.wsConn)
				assert(testConn.writeMessage(
					websocket.BinaryMessage,
					[]byte("hello"),
					10*time.Millisecond,
				)).Equal(errors.ErrWebsocketStreamConnWSConnWriteMessage.AddDebug(
					"invalid argument",
				))
			},
		)).Equal([]*base.Error{})
	})

	t.Run("writeMessage success", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			nil,
			func(client core.IClientAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				assert(testConn.writeMessage(
					websocket.BinaryMessage,
					[]byte("hello"),
					10*time.Millisecond,
				)).IsNil()
			},
		)).Equal([]*base.Error{})
	})
}

func TestWebsocketStreamConn_onCloseMessage(t *testing.T) {
	t.Run("webSocketStreamConnRunning", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			nil,
			func(client core.IClientAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				atomic.StoreInt32(&testConn.status, webSocketStreamConnRunning)
				assert(testConn.onCloseMessage(
					websocket.CloseNormalClosure,
					"",
				)).IsNil()
				assert(atomic.LoadInt32(&testConn.status)).
					Equal(webSocketStreamConnCanClose)
			},
		)).Equal([]*base.Error{})
	})

	t.Run("webSocketStreamConnClosing", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			nil,
			func(client core.IClientAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				testConn.closeCH = make(chan bool, 1)
				atomic.StoreInt32(&testConn.status, webSocketStreamConnClosing)
				assert(testConn.onCloseMessage(
					websocket.CloseNormalClosure,
					"",
				)).IsNil()
				assert(atomic.LoadInt32(&testConn.status)).
					Equal(webSocketStreamConnCanClose)
				assert(<-testConn.closeCH).Equal(true)
				testConn.closeCH = nil
			},
		)).Equal([]*base.Error{})
	})

	t.Run("webSocketStreamConnCanClose", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			nil,
			func(client core.IClientAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				atomic.StoreInt32(&testConn.status, webSocketStreamConnCanClose)
				assert(testConn.onCloseMessage(
					websocket.CloseNormalClosure,
					"",
				)).IsNil()
				assert(atomic.LoadInt32(&testConn.status)).
					Equal(webSocketStreamConnCanClose)
			},
		)).Equal([]*base.Error{})
	})

	t.Run("webSocketStreamConnClosed", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			nil,
			func(client core.IClientAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				atomic.StoreInt32(&testConn.status, webSocketStreamConnClosed)
				assert(testConn.onCloseMessage(
					websocket.CloseNormalClosure,
					"",
				)).IsNil()
				assert(atomic.LoadInt32(&testConn.status)).
					Equal(webSocketStreamConnClosed)
			},
		)).Equal([]*base.Error{})
	})
}

func TestWebsocketStreamConn_ReadStream(t *testing.T) {
	t.Run("status is not webSocketStreamConnRunning", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			nil,
			func(client core.IClientAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				atomic.StoreInt32(&testConn.status, webSocketStreamConnCanClose)
				assert(conn.ReadStream(time.Second, 999999)).
					Equal(nil, errors.ErrStreamConnIsClosed)
				assert(atomic.LoadInt32(&testConn.reading)).Equal(int32(0))
			},
		)).Equal([]*base.Error{})
	})

	t.Run("SetReadDeadline error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testHelperStreamConn(
			nil,
			func(client core.IClientAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				makeConnFDError(testConn.wsConn)
				assert(atomic.LoadInt32(&testConn.reading)).Equal(int32(0))
				assert(testConn.ReadStream(time.Second, 999999)).Equal(
					nil,
					errors.ErrWebsocketStreamConnWSConnSetReadDeadline.AddDebug(
						"invalid argument",
					),
				)
			},
		)
	})

	t.Run("ReadMessage timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			nil,
			func(client core.IClientAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				stream, err := testConn.ReadStream(-time.Second, 999999)
				assert(atomic.LoadInt32(&testConn.reading)).Equal(int32(0))
				assert(stream).IsNil()
				if err != nil {
					assert(strings.HasPrefix(
						err.Error(),
						errors.ErrWebsocketStreamConnWSConnReadMessage.Error(),
					)).IsTrue()
					assert(strings.Contains(err.GetMessage(), "timeout")).IsTrue()
				}
			},
		)).Equal([]*base.Error{})
	})

	t.Run("type is websocket.TextMessage", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			func(server core.IServerAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				_ = testConn.wsConn.WriteMessage(websocket.TextMessage, []byte("hello"))
				for {
					if _, e := conn.ReadStream(20*time.Millisecond, 999999); e != nil {
						return
					}
				}
			},
			func(client core.IClientAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				assert(testConn.ReadStream(time.Second, 999999)).Equal(
					nil, errors.ErrWebsocketStreamConnDataIsNotBinary,
				)
				assert(atomic.LoadInt32(&testConn.reading)).Equal(int32(0))
			},
		)).Equal([]*base.Error{})
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			func(server core.IServerAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				_ = testConn.wsConn.WriteMessage(websocket.BinaryMessage, []byte(
					"hello-world-hello-world-hello-world-hello-world-",
				))
				for {
					if _, e := conn.ReadStream(20*time.Millisecond, 999999); e != nil {
						return
					}
				}
			},
			func(client core.IClientAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				stream, err := testConn.ReadStream(20*time.Millisecond, 999999)
				assert(atomic.LoadInt32(&testConn.reading)).Equal(int32(0))
				assert(stream).IsNotNil()
				assert(err).IsNil()
				assert(string(stream.GetBufferUnsafe())).Equal(
					"hello-world-hello-world-hello-world-hello-world-",
				)
			},
		)).Equal([]*base.Error{})
	})
}

func TestWebsocketStreamConn_WriteStream(t *testing.T) {
	t.Run("stream is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			nil,
			func(client core.IClientAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				err := testConn.WriteStream(nil, time.Second)
				assert(err).Equal(errors.ErrWebsocketStreamConnStreamIsNil)
			},
		)).Equal([]*base.Error{})
	})

	t.Run("status is not webSocketStreamConnRunning", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			nil,
			func(client core.IClientAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				atomic.StoreInt32(&testConn.status, webSocketStreamConnClosed)
				assert(testConn.WriteStream(core.NewStream(), time.Second)).
					Equal(errors.ErrStreamConnIsClosed)
			},
		)).Equal([]*base.Error{})
	})

	t.Run("timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			nil,
			func(client core.IClientAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				err := testConn.WriteStream(core.NewStream(), -time.Second)
				assert(err).IsNotNil()
				assert(strings.HasPrefix(
					err.Error(),
					errors.ErrWebsocketStreamConnWSConnWriteMessage.Error(),
				)).IsTrue()
				assert(strings.Contains(err.GetMessage(), "timeout")).IsTrue()
			},
		)).Equal([]*base.Error{})
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			nil,
			func(client core.IClientAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				assert(testConn.WriteStream(core.NewStream(), time.Second)).IsNil()
			},
		)).Equal([]*base.Error{})
	})
}

func TestWebsocketStreamConn_Close(t *testing.T) {
	t.Run("Running => Closing no wait", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			nil,
			func(client core.IClientAdapter, conn core.IStreamConn) {
				assert(conn.Close()).IsNil()
			},
		)).Equal([]*base.Error{})
	})

	t.Run("Running => Closing wait ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			nil,
			func(client core.IClientAdapter, conn core.IStreamConn) {
				go func() {
					_, _ = conn.ReadStream(10*time.Second, 999999)
				}()
				time.Sleep(10 * time.Millisecond)
				assert(conn.Close()).IsNil()
			},
		)).Equal([]*base.Error{})
	})

	t.Run("Running => Closing  wait failed", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			func(server core.IServerAdapter, conn core.IStreamConn) {
				time.Sleep(2500 * time.Millisecond)
			},
			func(client core.IClientAdapter, conn core.IStreamConn) {
				go func() {
					_, _ = conn.ReadStream(10*time.Second, 999999)
				}()
				time.Sleep(10 * time.Millisecond)
				assert(conn.Close()).IsNil()
			},
		)).Equal([]*base.Error{})
	})

	t.Run("Running => Closing Error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			nil,
			func(client core.IClientAdapter, conn core.IStreamConn) {
				makeConnFDError(conn.(*websocketStreamConn).wsConn)
				assert(conn.Close()).Equal(
					errors.ErrWebsocketStreamConnWSConnClose.AddDebug("invalid argument"),
				)
			},
		)).Equal([]*base.Error{})
	})

	t.Run("CanClose => Closed", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			nil,
			func(client core.IClientAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				atomic.StoreInt32(&testConn.status, webSocketStreamConnCanClose)
				testConn.closeCH <- true
				assert(conn.Close()).IsNil()
			},
		)).Equal([]*base.Error{})
	})

	t.Run("CanClose => Closed Error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			nil,
			func(client core.IClientAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				atomic.StoreInt32(&testConn.status, webSocketStreamConnCanClose)
				testConn.closeCH <- true
				makeConnFDError(conn.(*websocketStreamConn).wsConn)
				assert(conn.Close()).Equal(
					errors.ErrWebsocketStreamConnWSConnClose.AddDebug("invalid argument"),
				)
			},
		)).Equal([]*base.Error{})
	})

	t.Run("Closed => Closed", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testHelperStreamConn(
			nil,
			func(client core.IClientAdapter, conn core.IStreamConn) {
				testConn := conn.(*websocketStreamConn)
				atomic.StoreInt32(&testConn.status, webSocketStreamConnClosed)
				assert(conn.Close()).IsNil()
			},
		)).Equal([]*base.Error{})
	})
}

func TestNewWebsocketServerAdapter(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(NewWebsocketServerAdapter("addrString")).Equal(
			&websocketServerAdapter{
				addr:     "addrString",
				wsServer: nil,
			},
		)
	})
}

func TestWsServerAdapter_Open(t *testing.T) {
	t.Run("onError is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(base.RunWithCatchPanic(func() {
			NewWebsocketServerAdapter("t").Open(
				func(core.IStreamConn, net.Addr) {},
				nil,
			)
		})).Equal("onError is nil")
	})

	t.Run("onConnRun is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(base.RunWithCatchPanic(func() {
			NewWebsocketServerAdapter("test").Open(nil, func(uint64, *base.Error) {})
		})).Equal("onConnRun is nil")
	})

	t.Run("it is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		serverAdapter := NewWebsocketServerAdapter("test").(*websocketServerAdapter)
		serverAdapter.SetRunning(func() {})
		waitCH := make(chan *base.Error, 1)
		serverAdapter.Open(
			func(conn core.IStreamConn, _ net.Addr) {},
			func(_ uint64, e *base.Error) {
				waitCH <- e
			},
		)
		assert(<-waitCH).Equal(errors.ErrWebsocketServerAdapterAlreadyRunning)
	})

	t.Run("error addr", func(t *testing.T) {
		assert := base.NewAssert(t)
		serverAdapter := NewWebsocketServerAdapter("error-addr")
		waitCH := make(chan *base.Error, 1)
		serverAdapter.Open(
			func(conn core.IStreamConn, _ net.Addr) {},
			func(_ uint64, e *base.Error) {
				waitCH <- e
			},
		)
		err := <-waitCH
		assert(strings.HasPrefix(
			err.Error(),
			errors.ErrWebsocketServerAdapterWSServerListenAndServe.Error(),
		))
		assert(strings.Contains(err.GetMessage(), "error-addr")).IsTrue()
	})

	t.Run("conn upgrade error", func(t *testing.T) {
		assert := base.NewAssert(t)
		serverAdapter := NewWebsocketServerAdapter(
			"127.0.0.1:12345",
		).(*websocketServerAdapter)

		go func() {
			time.Sleep(30 * time.Millisecond)
			_, _ = http.Get("http://127.0.0.1:12345")
			serverAdapter.Close(func(_ uint64, e *base.Error) {
				assert().Fail("error run here")
			})
		}()

		retCH := make(chan *base.Error, 1)
		serverAdapter.Open(
			func(conn core.IStreamConn, _ net.Addr) {},
			func(_ uint64, e *base.Error) {
				retCH <- e
			},
		)

		assert(<-retCH).Equal(errors.ErrWebsocketServerAdapterUpgrade)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		serverAdapter := NewWebsocketServerAdapter(
			"127.0.0.1:12345",
		).(*websocketServerAdapter)

		go func() {
			time.Sleep(10 * time.Millisecond)
			NewWebsocketClientAdapter("ws://127.0.0.1:12345").Open(
				func(conn core.IStreamConn) {
					// empty
				}, func(e *base.Error) {
					assert().Fail("error run here")
				},
			)
			serverAdapter.Close(func(_ uint64, e *base.Error) {
				assert().Fail("error run here")
			})
		}()

		serverAdapter.Open(
			func(conn core.IStreamConn, _ net.Addr) {
				time.Sleep(30 * time.Millisecond)
			},
			func(_ uint64, e *base.Error) {
				assert().Fail("error run here")
			},
		)
	})
}

func TestWsServerAdapter_Close(t *testing.T) {
	t.Run("onError is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(base.RunWithCatchPanic(func() {
			NewWebsocketServerAdapter("test").Close(nil)
		})).Equal("onError is nil")
	})

	t.Run("SetClosing is false", func(t *testing.T) {
		assert := base.NewAssert(t)
		NewWebsocketServerAdapter("test").Close(func(_ uint64, err *base.Error) {
			assert(strings.HasPrefix(
				err.Error(),
				errors.ErrWebsocketServerAdapterNotRunning.Error(),
			)).IsTrue()
		})
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(base.RunWithCatchPanic(func() {
			serverAdapter := NewWebsocketServerAdapter("127.0.0.1:12345")
			go func() {
				serverAdapter.Open(
					func(conn core.IStreamConn, _ net.Addr) {},
					func(_ uint64, err *base.Error) {
						assert().Fail("error run here")
					},
				)
			}()
			time.Sleep(10 * time.Millisecond)
			serverAdapter.Close(func(_ uint64, err *base.Error) {
				assert().Fail("error run here")
			})
		})).IsNil()
	})

	t.Run("server Close error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(base.RunWithCatchPanic(func() {
			serverAdapter := NewWebsocketServerAdapter("127.0.0.1:12345")
			go func() {
				serverAdapter.Open(
					func(conn core.IStreamConn, _ net.Addr) {},
					func(_ uint64, e *base.Error) {
						assert().Fail("error run here")
					},
				)
			}()

			time.Sleep(20 * time.Millisecond)

			mutex := (*sync.Mutex)(getFieldByPointer(serverAdapter, "mutex"))
			mutex.Lock()
			// make fake error
			wsServer := serverAdapter.(*websocketServerAdapter).wsServer
			httpServerMuPointer := (*sync.Mutex)(getFieldByPointer(wsServer, "mu"))
			listenersPtr := (*map[*net.Listener]struct{})(getFieldByPointer(
				wsServer,
				"listeners",
			))
			fakeListener := net.Listener(fakeNetListener{})
			httpServerMuPointer.Lock()
			*listenersPtr = map[*net.Listener]struct{}{
				&fakeListener: {},
			}
			httpServerMuPointer.Unlock()
			mutex.Unlock()

			errCount := 0
			serverAdapter.Close(func(_ uint64, err *base.Error) {
				if errCount == 0 {
					assert(err).Equal(
						errors.ErrWebsocketServerAdapterWSServerClose.
							AddDebug("test error"),
					)
				} else if errCount == 1 {
					assert(err).Equal(errors.ErrWebsocketServerAdapterCloseTimeout)
				} else {
					assert().Fail("error run here")
				}
				errCount++
			})
		})).IsNil()
	})
}

func TestNewWebsocketClientAdapter(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(NewWebsocketClientAdapter("addrString").(*websocketClientAdapter)).
			Equal(&websocketClientAdapter{
				connectString: "addrString",
				conn:          nil,
			})
	})
}

func TestWsClientAdapter_Open(t *testing.T) {
	t.Run("onError is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(base.RunWithCatchPanic(func() {
			NewWebsocketClientAdapter("test").Open(
				func(conn core.IStreamConn) {},
				nil,
			)
		})).Equal("onError is nil")
	})

	t.Run("onConnRun is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(base.RunWithCatchPanic(func() {
			NewWebsocketClientAdapter("test").Open(nil, func(e *base.Error) {})
		})).Equal("onConnRun is nil")
	})

	t.Run("dial addr error", func(t *testing.T) {
		assert := base.NewAssert(t)
		clientAdapter := NewWebsocketClientAdapter("ws://test")
		retCH := make(chan *base.Error, 1)
		clientAdapter.Open(
			func(conn core.IStreamConn) {},
			func(e *base.Error) {
				retCH <- e
			},
		)
		err := <-retCH
		assert(strings.HasPrefix(
			err.Error(),
			errors.ErrWebsocketClientAdapterDial.Error(),
		)).IsTrue()
		assert(strings.Contains(err.Error(), "dial tcp")).IsTrue()
	})

	t.Run("client already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		_ = base.RunWithSubscribePanic(func() {
			serverAdapter := NewWebsocketServerAdapter("127.0.0.1:12345")
			go func() {
				serverAdapter.Open(
					func(conn core.IStreamConn, _ net.Addr) {},
					func(_ uint64, e *base.Error) {},
				)
			}()
			time.Sleep(10 * time.Millisecond)
			clientAdapter := NewWebsocketClientAdapter(
				"ws://127.0.0.1:12345",
			).(*websocketClientAdapter)
			clientAdapter.SetRunning(func() {})
			clientAdapter.SetClosing(func(ch chan bool) {})
			waitCH := make(chan *base.Error, 1)
			clientAdapter.Open(
				func(conn core.IStreamConn) {},
				func(e *base.Error) {
					waitCH <- e
				},
			)
			assert(<-waitCH).Equal(errors.ErrWebsocketClientAdapterAlreadyRunning)
			serverAdapter.Close(func(_ uint64, e *base.Error) {})
		})
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		serverAdapter := NewWebsocketServerAdapter("127.0.0.1:12345")
		go func() {
			serverAdapter.Open(
				func(conn core.IStreamConn, _ net.Addr) {},
				func(_ uint64, e *base.Error) {},
			)
		}()
		time.Sleep(100 * time.Millisecond)
		clientAdapter := NewWebsocketClientAdapter(
			"ws://127.0.0.1:12345",
		).(*websocketClientAdapter)
		clientAdapter.Open(
			func(conn core.IStreamConn) {},
			func(e *base.Error) {
				assert().Fail("error run here")
			},
		)
		clientAdapter.Close(func(e *base.Error) {})
		serverAdapter.Close(func(_ uint64, e *base.Error) {})
	})
}

func TestWsClientAdapter_Close(t *testing.T) {
	t.Run("onError is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(base.RunWithCatchPanic(func() {
			NewWebsocketClientAdapter("test").Close(nil)
		})).Equal("onError is nil")
	})

	t.Run("SetClosing is false", func(t *testing.T) {
		assert := base.NewAssert(t)
		NewWebsocketClientAdapter("test").Close(func(err *base.Error) {
			assert(err).Equal(errors.ErrWebsocketClientAdapterNotRunning)
		})
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		serverAdapter := NewWebsocketServerAdapter("127.0.0.1:12345")
		go func() {
			serverAdapter.Open(
				func(conn core.IStreamConn, _ net.Addr) {},
				func(_ uint64, e *base.Error) {},
			)
		}()

		clientAdapter := NewWebsocketClientAdapter("ws://127.0.0.1:12345")
		go func() {
			time.Sleep(100 * time.Millisecond)
			clientAdapter.Open(
				func(conn core.IStreamConn) {
					time.Sleep(time.Second)
				},
				func(e *base.Error) {},
			)
		}()

		time.Sleep(200 * time.Millisecond)
		clientAdapter.Close(func(e *base.Error) {
			assert().Fail("error run here")
		})

		serverAdapter.Close(func(_ uint64, e *base.Error) {})
	})

	t.Run("conn Close error", func(t *testing.T) {
		assert := base.NewAssert(t)
		serverAdapter := NewWebsocketServerAdapter("127.0.0.1:12345")
		go func() {
			serverAdapter.Open(
				func(conn core.IStreamConn, _ net.Addr) {},
				func(_ uint64, e *base.Error) {},
			)
		}()

		clientAdapter := NewWebsocketClientAdapter("ws://127.0.0.1:12345")
		go func() {
			time.Sleep(10 * time.Millisecond)
			clientAdapter.Open(
				func(conn core.IStreamConn) {
					time.Sleep(time.Second)
				},
				func(e *base.Error) {},
			)
		}()

		time.Sleep(20 * time.Millisecond)

		mutex := (*sync.Mutex)(getFieldByPointer(clientAdapter, "mutex"))
		mutex.Lock()
		makeConnFDError(
			clientAdapter.(*websocketClientAdapter).
				conn.(*websocketStreamConn).wsConn,
		)
		mutex.Unlock()
		clientAdapter.Close(func(err *base.Error) {
			assert(err).Equal(
				errors.ErrWebsocketStreamConnWSConnClose.AddDebug("invalid argument"),
			)
		})

		serverAdapter.Close(func(_ uint64, e *base.Error) {})
	})

	t.Run("conn Close timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		serverAdapter := NewWebsocketServerAdapter("127.0.0.1:12345")
		go func() {
			serverAdapter.Open(
				func(conn core.IStreamConn, _ net.Addr) {},
				func(_ uint64, e *base.Error) {},
			)
		}()

		clientAdapter := NewWebsocketClientAdapter("ws://127.0.0.1:12345")
		go func() {
			time.Sleep(10 * time.Millisecond)
			clientAdapter.Open(
				func(conn core.IStreamConn) {
					time.Sleep(6 * time.Second)
				},
				func(e *base.Error) {},
			)
		}()

		time.Sleep(20 * time.Millisecond)
		clientAdapter.Close(func(err *base.Error) {
			assert(err).Equal(errors.ErrWebsocketClientAdapterCloseTimeout)
		})

		serverAdapter.Close(func(_ uint64, e *base.Error) {})
	})
}
