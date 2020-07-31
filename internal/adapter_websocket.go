package internal

import (
	"github.com/gorilla/websocket"
	"net/http"
	"runtime/debug"
	"sync/atomic"
	"time"
)

const webSocketStreamConnClosed = int32(0)
const webSocketStreamConnRunning = int32(1)
const webSocketStreamConnClosing = int32(2)
const webSocketStreamConnCanClose = int32(2)

type webSocketStreamConn struct {
	status   int32
	canClose chan bool
	conn     *websocket.Conn
}

func newWebSocketStreamConn(conn *websocket.Conn) *webSocketStreamConn {
	ret := &webSocketStreamConn{
		conn:     conn,
		canClose: make(chan bool, 1),
		status:   webSocketStreamConnRunning,
	}
	conn.SetCloseHandler(ret.onCloseMessage)
	return ret
}

func (p *webSocketStreamConn) onCloseMessage(code int, _ string) error {
	if atomic.CompareAndSwapInt32(
		&p.status,
		webSocketStreamConnRunning,
		webSocketStreamConnCanClose,
	) {
		_ = p.conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(code, ""),
			TimeNow().Add(time.Second),
		)
		return nil
	} else if atomic.CompareAndSwapInt32(
		&p.status,
		webSocketStreamConnClosing,
		webSocketStreamConnCanClose,
	) {
		p.canClose <- true
		return nil
	} else {
		return nil
	}
}

func toTransportError(err error) Error {
	if err == nil {
		return nil
	} else if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		return ErrTransportStreamConnIsClosed
	} else {
		return NewTransportError(err.Error())
	}
}

func (p *webSocketStreamConn) ReadStream(
	timeout time.Duration,
	readLimit int64,
) (stream *Stream, err Error) {
	p.conn.SetReadLimit(readLimit)
	if e := p.conn.SetReadDeadline(time.Now().Add(timeout)); e != nil {
		return nil, toTransportError(e)
	} else if mt, message, e := p.conn.ReadMessage(); e != nil {
		return nil, toTransportError(e)
	} else if mt != websocket.BinaryMessage {
		return nil, NewTransportError("unsupported websocket protocol")
	} else {
		stream := NewStream()
		stream.SetWritePos(0)
		stream.PutBytes(message)
		return stream, nil
	}
}

func (p *webSocketStreamConn) WriteStream(
	stream *Stream,
	timeout time.Duration,
) (err Error) {
	if stream == nil {
		return NewKernelPanic("stream is nil").AddDebug(string(debug.Stack()))
	} else if e := p.conn.SetWriteDeadline(time.Now().Add(timeout)); e != nil {
		return toTransportError(e)
	} else if e := p.conn.WriteMessage(
		websocket.BinaryMessage,
		stream.GetBufferUnsafe(),
	); e != nil {
		return toTransportError(e)
	} else {
		return nil
	}
}

func (p *webSocketStreamConn) Close() Error {
	if atomic.CompareAndSwapInt32(
		&p.status,
		webSocketStreamConnRunning,
		webSocketStreamConnClosing,
	) {
		_ = p.conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)

		select {
		case <-p.canClose:
		case <-time.After(2 * time.Second):
		}

		atomic.StoreInt32(&p.status, webSocketStreamConnClosed)
		if e := p.conn.Close(); e != nil {
			return NewTransportError(e.Error())
		}
		return nil
	} else if atomic.CompareAndSwapInt32(
		&p.status,
		webSocketStreamConnCanClose,
		webSocketStreamConnClosed,
	) {
		if e := p.conn.Close(); e != nil {
			return NewTransportError(e.Error())
		}
		return nil
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
				streamConn := newWebSocketStreamConn(conn)
				onConnRun(streamConn)
				if err := streamConn.Close(); err != nil {
					onError(err)
				}
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
	waitCH := chan bool(nil)
	if onError == nil {
		panic("onError is nil")
	} else if !p.SetClosing(func(ch chan bool) {
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
	conn          IStreamConn
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
	} else {
		streamConn := newWebSocketStreamConn(conn)
		if !p.SetRunning(func() {
			p.conn = streamConn
		}) {
			_ = conn.Close()
			onError(NewKernelPanic(
				"it is already running",
			).AddDebug(string(debug.Stack())))
		} else {
			onConnRun(streamConn)
			p.SetClosing(nil)
			p.SetClosed(func() {
				p.conn = nil
			})
			if err := streamConn.Close(); err != nil {
				onError(err)
			}
		}
	}
}

func (p *wsClientAdapter) Close(onError func(Error)) {
	waitCH := chan bool(nil)

	if onError == nil {
		panic("onError is nil")
	} else if !p.SetClosing(func(ch chan bool) {
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
