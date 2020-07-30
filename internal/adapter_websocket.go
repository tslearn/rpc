package internal

import (
	"github.com/gorilla/websocket"
	"net/http"
	"runtime/debug"
	"time"
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
		return NewKernelPanic("object is nil").AddDebug(string(debug.Stack()))
	} else if stream == nil {
		return NewKernelPanic("stream is nil").AddDebug(string(debug.Stack()))
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
		return NewKernelPanic("object is nil").AddDebug(string(debug.Stack()))
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

type wsServerAdapter struct {
	addr     string
	wsServer *http.Server
	StatusManager
}

func NewWebSocketServerAdapter(addr string) IAdapter {
	return &wsServerAdapter{
		addr:     addr,
		wsServer: nil,
	}
}

// Open ...
func (p *wsServerAdapter) Open(
	onConnRun func(IStreamConn),
	onError func(Error),
) {
	if onError == nil {
		panic("onError is nil")
	} else if onConnRun == nil {
		panic("onConnRun is nil")
	} else if !p.SetRunning(func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			if conn, err := wsUpgradeManager.Upgrade(w, req, nil); err != nil {
				onError(NewTransportError(err.Error()))
			} else {
				onConnRun((*webSocketConn)(conn))
			}
		})
		p.wsServer = &http.Server{
			Addr:    p.addr,
			Handler: mux,
		}
	}) {
		onError(NewKernelPanic(
			"it is already running",
		).AddDebug(string(debug.Stack())))
	} else {
		if e := p.wsServer.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			onError(NewRuntimePanic(e.Error()))
		}
		p.SetClosing(nil)
		p.SetClosed(func() {
			p.wsServer = nil
		})
	}
}

// Close ...
func (p *wsServerAdapter) Close(onError func(Error)) {
	waitCH := chan struct{}(nil)
	if onError == nil {
		panic("onError is nil")
	} else if !p.SetClosing(func(ch chan struct{}) {
		waitCH = ch
		if e := p.wsServer.Close(); e != nil {
			onError(NewRuntimePanic(e.Error()))
		}
	}) {
		onError(NewKernelPanic(
			"it is not running",
		).AddDebug(string(debug.Stack())))
	} else {
		select {
		case <-waitCH:
		case <-time.After(20 * time.Second):
			onError(NewRuntimePanic(
				"can not close within 20 seconds",
			).AddDebug(string(debug.Stack())))
		}
	}
}

type wsClientAdapter struct {
	conn          *websocket.Conn
	connectString string
	StatusManager
}

func NewWebSocketClientEndPoint(connectString string) IAdapter {
	return &wsClientAdapter{
		conn:          nil,
		connectString: connectString,
	}
}

func (p *wsClientAdapter) Open(
	onConnRun func(IStreamConn),
	onError func(Error),
) {
	if onError == nil {
		panic("onError is nil")
	} else if onConnRun == nil {
		panic("onConnRun is nil")
	} else if conn, _, err := websocket.DefaultDialer.Dial(
		p.connectString,
		nil,
	); err != nil {
		onError(NewRuntimePanic(err.Error()))
	} else if !p.SetRunning(func() {
		p.conn = conn
	}) {
		_ = conn.Close()
		onError(NewKernelPanic(
			"it is already running",
		).AddDebug(string(debug.Stack())))
	} else {
		onConnRun((*webSocketConn)(conn))
		p.SetClosing(nil)
		p.SetClosed(func() {
			p.conn = nil
		})
	}
}

func (p *wsClientAdapter) Close(onError func(Error)) {
	waitCH := chan struct{}(nil)

	if onError == nil {
		panic("onError is nil")
	} else if !p.SetClosing(func(ch chan struct{}) {
		waitCH = ch
		if e := p.conn.Close(); e != nil {
			onError(NewRuntimePanic(e.Error()))
		}
	}) {
		onError(NewKernelPanic(
			"it is not running",
		).AddDebug(string(debug.Stack())))
	} else {
		select {
		case <-waitCH:
		case <-time.After(20 * time.Second):
			onError(NewRuntimePanic(
				"can not close within 20 seconds",
			).AddDebug(string(debug.Stack())))
		}
	}
}
