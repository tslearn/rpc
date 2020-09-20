package adapter

import (
	"github.com/gorilla/websocket"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"net"
	"net/http"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

const webSocketStreamConnClosed = int32(0)
const webSocketStreamConnRunning = int32(1)
const webSocketStreamConnClosing = int32(2)
const webSocketStreamConnCanClose = int32(3)

func toTransportError(err error) *base.Error {
	if err == nil {
		return nil
	} else if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		return base.TransportWarnStreamConnIsClosed
	} else {
		return base.TransportWarn.AddDebug(err.Error())
	}
}

type webSocketStreamConn struct {
	status  int32
	reading int32
	writing int32
	closeCH chan bool
	conn    *websocket.Conn
	sync.Mutex
}

func newWebSocketStreamConn(conn *websocket.Conn) *webSocketStreamConn {
	if conn == nil {
		return nil
	}

	ret := &webSocketStreamConn{
		status:  webSocketStreamConnRunning,
		reading: 0,
		writing: 0,
		closeCH: make(chan bool, 1),
		conn:    conn,
	}
	conn.SetCloseHandler(ret.onCloseMessage)

	return ret
}

func (p *webSocketStreamConn) writeMessage(
	messageType int,
	data []byte,
	timeout time.Duration,
) *base.Error {
	p.Lock()
	defer p.Unlock()
	_ = p.conn.SetWriteDeadline(base.TimeNow().Add(timeout))
	return toTransportError(p.conn.WriteMessage(messageType, data))
}

func (p *webSocketStreamConn) onCloseMessage(code int, _ string) error {
	if atomic.CompareAndSwapInt32(
		&p.status,
		webSocketStreamConnRunning,
		webSocketStreamConnCanClose,
	) {
		_ = p.writeMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(code, ""),
			time.Second,
		)
		return nil
	} else if atomic.CompareAndSwapInt32(
		&p.status,
		webSocketStreamConnClosing,
		webSocketStreamConnCanClose,
	) {
		p.closeCH <- true
		return nil
	} else {
		return nil
	}
}

func (p *webSocketStreamConn) ReadStream(
	timeout time.Duration,
	readLimit int64,
) (*core.Stream, *base.Error) {
	atomic.StoreInt32(&p.reading, 1)
	defer atomic.StoreInt32(&p.reading, 0)

	p.conn.SetReadLimit(readLimit)
	if !atomic.CompareAndSwapInt32(
		&p.status,
		webSocketStreamConnRunning,
		webSocketStreamConnRunning,
	) {
		return nil, base.TransportWarnStreamConnIsClosed
	} else if e := p.conn.SetReadDeadline(base.TimeNow().Add(timeout)); e != nil {
		return nil, toTransportError(e)
	} else if mt, message, e := p.conn.ReadMessage(); e != nil {
		return nil, toTransportError(e)
	} else if mt != websocket.BinaryMessage {
		return nil, base.SecurityWarnWebsocketDataNotBinary
	} else {
		stream := core.NewStream()
		stream.PutBytesTo(message, 0)
		return stream, nil
	}
}

func (p *webSocketStreamConn) WriteStream(
	stream *core.Stream,
	timeout time.Duration,
) *base.Error {
	atomic.StoreInt32(&p.writing, 1)
	defer atomic.StoreInt32(&p.writing, 0)

	if stream == nil {
		return base.KernelFatal.
			AddDebug("stream is nil").
			AddDebug(string(debug.Stack()))
	} else if !atomic.CompareAndSwapInt32(
		&p.status,
		webSocketStreamConnRunning,
		webSocketStreamConnRunning,
	) {
		return base.TransportWarnStreamConnIsClosed
	} else {
		return p.writeMessage(
			websocket.BinaryMessage,
			stream.GetBufferUnsafe(),
			timeout,
		)
	}
}

func (p *webSocketStreamConn) Close() *base.Error {
	if atomic.CompareAndSwapInt32(
		&p.status,
		webSocketStreamConnRunning,
		webSocketStreamConnClosing,
	) {
		defer atomic.StoreInt32(&p.status, webSocketStreamConnClosed)

		// 1. send close message to peer
		_ = p.writeMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			time.Second,
		)

		// 2. if it is reading or writing now,
		//    wait for peer confirm close message (within 2 seconds)
		if atomic.LoadInt32(&p.reading) > 0 || atomic.LoadInt32(&p.writing) > 0 {
			select {
			case <-p.closeCH:
				// p.conn has already been closed gracefully. do not close it again!
				return nil
			case <-time.After(time.Second):
			}
		}

		// 3. close and return
		return toTransportError(p.conn.Close())
	} else if atomic.CompareAndSwapInt32(
		&p.status,
		webSocketStreamConnCanClose,
		webSocketStreamConnClosed,
	) {
		// 1. close and return
		return toTransportError(p.conn.Close())
	} else {
		return nil
	}
}

var (
	wsUpgradeManager = websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type wsServerAdapter struct {
	addr     string
	wsServer *http.Server
	base.StatusManager
}

// NewWebSocketServerAdapter ...
func NewWebSocketServerAdapter(addr string) core.IServerAdapter {
	return &wsServerAdapter{
		addr:     addr,
		wsServer: nil,
	}
}

// Open ...
func (p *wsServerAdapter) Open(
	onConnRun func(core.IStreamConn, net.Addr),
	onError func(uint64, *base.Error),
) {
	if onError == nil {
		panic("onError is nil")
	} else if onConnRun == nil {
		panic("onConnRun is nil")
	} else if !p.SetRunning(func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			if conn, err := wsUpgradeManager.Upgrade(w, req, nil); err != nil {
				onError(0, base.SecurityWarnWebsocketUpgradeError)
			} else {
				streamConn := newWebSocketStreamConn(conn)
				onConnRun(streamConn, conn.RemoteAddr())
			}
		})
		p.wsServer = &http.Server{
			Addr:    p.addr,
			Handler: mux,
		}
	}) {
		onError(
			0,
			base.KernelFatal.
				AddDebug("it is already running").
				AddDebug(string(debug.Stack())),
		)
	} else {
		if e := p.wsServer.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			onError(0, base.RuntimeFatal.AddDebug(e.Error()))
		}
		p.SetClosing(nil)
		p.SetClosed(func() {
			p.wsServer = nil
		})
	}
}

// Close ...
func (p *wsServerAdapter) Close(onError func(uint64, *base.Error)) {
	waitCH := chan bool(nil)
	if onError == nil {
		panic("onError is nil")
	} else if !p.SetClosing(func(ch chan bool) {
		waitCH = ch
		if e := p.wsServer.Close(); e != nil {
			onError(0, base.RuntimeError.AddDebug(e.Error()))
		}
	}) {
		onError(
			0,
			base.KernelFatal.
				AddDebug("it is not running").
				AddDebug(string(debug.Stack())),
		)
	} else {
		select {
		case <-waitCH:
		case <-time.After(5 * time.Second):
			onError(
				0,
				base.RuntimeError.
					AddDebug("it cannot be closed within 5 seconds").
					AddDebug(string(debug.Stack())),
			)
		}
	}
}

type wsClientAdapter struct {
	conn          core.IStreamConn
	connectString string
	base.StatusManager
}

// NewWebSocketClientAdapter ...
func NewWebSocketClientAdapter(connectString string) core.IClientAdapter {
	return &wsClientAdapter{
		conn:          nil,
		connectString: connectString,
	}
}

func (p *wsClientAdapter) Open(
	onConnRun func(core.IStreamConn),
	onError func(*base.Error),
) {
	if onError == nil {
		panic("onError is nil")
	} else if onConnRun == nil {
		panic("onConnRun is nil")
	} else if conn, _, err := websocket.DefaultDialer.Dial(
		p.connectString,
		nil,
	); err != nil {
		onError(base.RuntimeError.AddDebug(err.Error()))
	} else {
		streamConn := newWebSocketStreamConn(conn)
		if !p.SetRunning(func() {
			p.conn = streamConn
		}) {
			_ = conn.Close()
			onError(
				base.KernelFatal.
					AddDebug("it is already running").
					AddDebug(string(debug.Stack())),
			)
		} else {
			onConnRun(streamConn)
			p.SetClosing(nil)
			p.SetClosed(func() {
				p.conn = nil
			})
		}
	}
}

func (p *wsClientAdapter) Close(onError func(*base.Error)) {
	waitCH := chan bool(nil)

	if onError == nil {
		panic("onError is nil")
	} else if !p.SetClosing(func(ch chan bool) {
		waitCH = ch
		if e := p.conn.Close(); e != nil {
			onError(base.RuntimeError.AddDebug(e.Error()))
		}
	}) {
		onError(
			base.KernelFatal.
				AddDebug("it is not running").
				AddDebug(string(debug.Stack())),
		)
	} else {
		select {
		case <-waitCH:
		case <-time.After(5 * time.Second):
			onError(
				base.RuntimeError.
					AddDebug("it cannot be closed within 5 seconds").
					AddDebug(string(debug.Stack())),
			)
		}
	}
}
