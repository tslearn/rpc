package rpc

import (
	"github.com/gorilla/websocket"
	"github.com/rpccloud/rpc/internal"
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
			internal.NewKernelPanic("object is nil").AddDebug(string(debug.Stack()))
	} else {
		conn.SetReadLimit(readLimit)
		if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return nil, internal.NewTransportError(err.Error())
		} else if mt, message, err := conn.ReadMessage(); err != nil {
			return nil, internal.NewTransportError(err.Error())
		} else if mt != websocket.BinaryMessage {
			return nil, internal.NewTransportError("unsupported websocket protocol")
		} else {
			stream := internal.NewStream()
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
		return internal.NewKernelPanic("object is nil").
			AddDebug(string(debug.Stack()))
	} else if stream == nil {
		return internal.NewKernelPanic("stream is nil").
			AddDebug(string(debug.Stack()))
	} else if err := conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return internal.NewTransportError(err.Error())
	} else if err := conn.WriteMessage(
		websocket.BinaryMessage,
		stream.GetBufferUnsafe(),
	); err != nil {
		return internal.NewTransportError(err.Error())
	} else {
		return nil
	}
}

func (p *webSocketConn) Close() Error {
	if conn := (*websocket.Conn)(p); conn == nil {
		return internal.NewKernelPanic("object is nil").
			AddDebug(string(debug.Stack()))
	} else if err := conn.Close(); err != nil {
		return internal.NewTransportError(err.Error())
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
	path     string
	wsServer unsafe.Pointer
}

func NewWebSocketServerAdapter(addr string, path string) IAdapter {
	if path == "" || path[0] != '/' {
		path = "/" + path
	}

	return &WebSocketServerAdapter{
		addr:     addr,
		path:     path,
		wsServer: nil,
	}
}

// Open ...
func (p *WebSocketServerAdapter) Open(
	onConnRun func(IStreamConn),
	onConnError func(Error),
) Error {
	if onConnRun == nil {
		return internal.NewKernelPanic("onConnRun is nil").
			AddDebug(string(debug.Stack()))
	} else if onConnError == nil {
		return internal.NewKernelPanic("onConnError is nil").
			AddDebug(string(debug.Stack()))
	} else {
		mux := http.NewServeMux()
		mux.HandleFunc(p.path, func(w http.ResponseWriter, req *http.Request) {
			if conn, err := wsUpgradeManager.Upgrade(w, req, nil); err != nil {
				onConnError(internal.NewTransportError(err.Error()))
			} else {
				onConnRun((*webSocketConn)(conn))
			}
		})
		httpServer := &http.Server{
			Addr:    p.addr,
			Handler: mux,
		}

		if !atomic.CompareAndSwapPointer(
			&p.wsServer,
			nil,
			unsafe.Pointer(httpServer),
		) {
			return internal.NewKernelPanic("it has already been opened").
				AddDebug(string(debug.Stack()))
		}

		defer func() {
			atomic.StorePointer(&p.wsServer, nil)
		}()

		if e := httpServer.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			return internal.NewRuntimePanic(e.Error())
		}

		return nil
	}
}

// Close ...
func (p *WebSocketServerAdapter) Close() Error {
	if server := (*http.Server)(atomic.LoadPointer(&p.wsServer)); server == nil {
		return internal.NewKernelPanic("it has not been opened").
			AddDebug(string(debug.Stack()))
	} else if e := server.Close(); e != nil {
		return internal.NewRuntimePanic(e.Error())
	} else {
		return nil
	}
}

func (p *WebSocketServerAdapter) IsRunning() bool {
	return atomic.LoadPointer(&p.wsServer) != nil
}

func (p *WebSocketServerAdapter) ConnectString() string {
	return "ws://" + p.addr + p.path
}

type WebSocketClientEndPoint struct {
	conn          unsafe.Pointer // *webSocketConn
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
	onConnError func(Error),
) Error {
	if onConnRun == nil {
		return internal.NewKernelPanic("onConnRun is nil").
			AddDebug(string(debug.Stack()))
	} else if onConnError != nil {
		return internal.NewKernelPanic("onConnError is not nil").
			AddDebug(string(debug.Stack()))
	} else if conn, _, err := websocket.DefaultDialer.Dial(
		p.connectString,
		nil,
	); err != nil {
		return internal.NewRuntimePanic(err.Error())
	} else if !atomic.CompareAndSwapPointer(&p.conn, nil, unsafe.Pointer(conn)) {
		return internal.NewKernelPanic("it has already been opened")
	} else {
		defer func() {
			atomic.StorePointer(&p.conn, nil)
		}()
		onConnRun((*webSocketConn)(conn))
		return nil
	}
}

func (p *WebSocketClientEndPoint) Close() Error {
	if conn := (*websocket.Conn)(atomic.LoadPointer(&p.conn)); conn == nil {
		return internal.NewKernelPanic("it has not been opened").
			AddDebug(string(debug.Stack()))
	} else if e := conn.Close(); e != nil {
		return internal.NewRuntimePanic(e.Error())
	} else {
		return nil
	}
}

func (p *WebSocketClientEndPoint) IsRunning() bool {
	return atomic.LoadPointer(&p.conn) != nil
}

func (p *WebSocketClientEndPoint) ConnectString() string {
	return p.connectString
}
