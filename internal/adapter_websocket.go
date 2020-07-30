package internal

import (
	"github.com/gorilla/websocket"
	"net/http"
	"runtime/debug"
	"sync/atomic"
	"time"
	"unsafe"
)

type webSocketConn websocket.Conn

func (p *webSocketConn) ReadStream(
	timeout time.Duration,
	readLimit int64,
) (*Stream, Error) {
	if conn := (*websocket.Conn)(p); conn == nil {
		return nil,
			NewKernelPanic("object is nil").AddDebug(string(debug.Stack()))
	} else {
		conn.SetReadLimit(readLimit)
		if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return nil, NewTransportError(err.Error())
		} else if mt, message, err := conn.ReadMessage(); err != nil {
			return nil, NewTransportError(err.Error())
		} else if mt != websocket.BinaryMessage {
			return nil, NewTransportError("unsupported websocket protocol")
		} else {
			stream := NewStream()
			stream.SetWritePos(0)
			stream.PutBytes(message)
			return stream, nil
		}
	}
}

func (p *webSocketConn) WriteStream(
	stream *Stream,
	timeout time.Duration,
) Error {
	if conn := (*websocket.Conn)(p); conn == nil {
		return NewKernelPanic("object is nil").
			AddDebug(string(debug.Stack()))
	} else if stream == nil {
		return NewKernelPanic("stream is nil").
			AddDebug(string(debug.Stack()))
	} else if err := conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return NewTransportError(err.Error())
	} else if err := conn.WriteMessage(
		websocket.BinaryMessage,
		stream.GetBufferUnsafe(),
	); err != nil {
		return NewTransportError(err.Error())
	} else {
		return nil
	}
}

func (p *webSocketConn) Close() Error {
	if conn := (*websocket.Conn)(p); conn == nil {
		return NewKernelPanic("object is nil").
			AddDebug(string(debug.Stack()))
	} else if err := conn.Close(); err != nil {
		return NewTransportError(err.Error())
	} else {
		return nil
	}
}

var (
	wsUpgradeManager = websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
	}
)

type WebSocketServerAdapter struct {
	addr     string
	wsServer unsafe.Pointer
}

func NewWebSocketServerAdapter(addr string) IAdapter {
	return &WebSocketServerAdapter{
		addr:     addr,
		wsServer: nil,
	}
}

// Open ...
func (p *WebSocketServerAdapter) Open(
	onConnRun func(IStreamConn),
	onError func(Error),
) {
	if onError == nil {
		panic("onError is nil")
	} else if onConnRun == nil {
		onError(NewKernelPanic(
			"onConnRun is nil",
		).AddDebug(string(debug.Stack())))
	} else {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			if conn, err := wsUpgradeManager.Upgrade(w, req, nil); err != nil {
				onError(NewTransportError(err.Error()))
			} else {
				onConnRun((*webSocketConn)(conn))
			}
		})
		wsServer := &http.Server{
			Addr:    p.addr,
			Handler: mux,
		}
		if !atomic.CompareAndSwapPointer(
			&p.wsServer,
			nil,
			unsafe.Pointer(wsServer),
		) {
			onError(NewKernelPanic(
				"it is already running",
			).AddDebug(string(debug.Stack())))
		} else {
			defer func() {
				atomic.StorePointer(&p.wsServer, nil)
			}()

			if e := wsServer.ListenAndServe(); e != nil && e != http.ErrServerClosed {
				onError(NewRuntimePanic(e.Error()))
			}
		}
	}
}

// Close ...
func (p *WebSocketServerAdapter) Close(onError func(Error)) {
	if onError == nil {
		panic("onError is nil")
	} else if server := atomic.LoadPointer(&p.wsServer); server == nil {
		onError(NewRuntimePanic("it is not running"))
	} else if e := (*http.Server)(server).Close(); e != nil {
		onError(NewRuntimePanic(e.Error()).AddDebug(string(debug.Stack())))
	} else {
		count := 200
		for count > 0 {
			if atomic.CompareAndSwapPointer(&p.wsServer, server, server) {
				time.Sleep(100 * time.Millisecond)
				count -= 1
			} else {
				return
			}
		}
		onError(NewRuntimePanic(
			"can not close within 20 seconds",
		).AddDebug(string(debug.Stack())))
	}
}

type WebSocketClientEndPoint struct {
	conn          unsafe.Pointer
	connectString string
}

func NewWebSocketClientEndPoint(connectString string) IAdapter {
	return &WebSocketClientEndPoint{
		conn:          nil,
		connectString: connectString,
	}
}

func (p *WebSocketClientEndPoint) Open(
	onConnRun func(IStreamConn),
	onError func(Error),
) {
	if onError == nil {
		panic("onError is nil")
	} else if onConnRun == nil {
		onError(NewKernelPanic(
			"onConnRun is nil",
		).AddDebug(string(debug.Stack())))
	} else if conn, _, err := websocket.DefaultDialer.Dial(
		p.connectString,
		nil,
	); err != nil {
		onError(NewRuntimePanic(err.Error()))
	} else if !atomic.CompareAndSwapPointer(&p.conn, nil, unsafe.Pointer(conn)) {
		onError(NewKernelPanic("it is already running"))
	} else {
		defer func() {
			atomic.StorePointer(&p.conn, nil)
		}()
		onConnRun((*webSocketConn)(conn))
	}
}

func (p *WebSocketClientEndPoint) Close(onError func(Error)) {
	if onError == nil {
		panic("onError is nil")
	} else if conn := atomic.LoadPointer(&p.conn); conn == nil {
		onError(NewRuntimePanic("it is not running"))
	} else if e := (*websocket.Conn)(conn).Close(); e != nil {
		onError(NewRuntimePanic(e.Error()))
	} else {
		count := 200
		for count > 0 {
			if atomic.CompareAndSwapPointer(&p.conn, conn, conn) {
				time.Sleep(100 * time.Millisecond)
				count -= 1
			} else {
				return
			}
		}
		onError(NewRuntimePanic(
			"can not close within 20 seconds",
		).AddDebug(string(debug.Stack())))
	}
}
