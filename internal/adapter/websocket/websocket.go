package websocket

import (
	"github.com/gorilla/websocket"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const webSocketStreamConnClosed = int32(0)
const webSocketStreamConnRunning = int32(1)
const webSocketStreamConnClosing = int32(2)
const webSocketStreamConnCanClose = int32(3)

func convertToError(err error, template *base.Error) *base.Error {
	if err == nil {
		return nil
	} else if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		return errors.ErrStreamConnIsClosed
	} else {
		return template.AddDebug(err.Error())
	}
}

type websocketStreamConn struct {
	status  int32
	reading int32
	writing int32
	closeCH chan bool
	wsConn  *websocket.Conn
	sync.Mutex
}

func newWebsocketStreamConn(conn *websocket.Conn) *websocketStreamConn {
	if conn == nil {
		return nil
	}

	ret := &websocketStreamConn{
		status:  webSocketStreamConnRunning,
		reading: 0,
		writing: 0,
		closeCH: make(chan bool, 1),
		wsConn:  conn,
	}
	conn.SetCloseHandler(ret.onCloseMessage)

	return ret
}

func (p *websocketStreamConn) writeMessage(
	messageType int,
	data []byte,
	timeout time.Duration,
) *base.Error {
	p.Lock()
	defer p.Unlock()
	_ = p.wsConn.SetWriteDeadline(base.TimeNow().Add(timeout))
	return convertToError(
		p.wsConn.WriteMessage(messageType, data),
		errors.ErrWebsocketStreamConnWSConnWriteMessage,
	)
}

func (p *websocketStreamConn) onCloseMessage(code int, _ string) error {
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

func (p *websocketStreamConn) ReadStream(
	timeout time.Duration,
	readLimit int64,
) (*core.Stream, *base.Error) {
	atomic.StoreInt32(&p.reading, 1)
	defer atomic.StoreInt32(&p.reading, 0)

	p.wsConn.SetReadLimit(readLimit)
	if !atomic.CompareAndSwapInt32(
		&p.status,
		webSocketStreamConnRunning,
		webSocketStreamConnRunning,
	) {
		return nil, errors.ErrStreamConnIsClosed
	} else if e := p.wsConn.SetReadDeadline(
		base.TimeNow().Add(timeout),
	); e != nil {
		return nil, convertToError(
			e,
			errors.ErrWebsocketStreamConnWSConnSetReadDeadline,
		)
	} else if mt, message, e := p.wsConn.ReadMessage(); e != nil {
		return nil, convertToError(
			e,
			errors.ErrWebsocketStreamConnWSConnReadMessage,
		)
	} else if mt != websocket.BinaryMessage {
		return nil, errors.ErrWebsocketStreamConnDataIsNotBinary
	} else {
		stream := core.NewStream()
		stream.PutBytesTo(message, 0)
		return stream, nil
	}
}

func (p *websocketStreamConn) WriteStream(
	stream *core.Stream,
	timeout time.Duration,
) *base.Error {
	atomic.StoreInt32(&p.writing, 1)
	defer atomic.StoreInt32(&p.writing, 0)

	if stream == nil {
		return errors.ErrWebsocketStreamConnStreamIsNil
	} else if !atomic.CompareAndSwapInt32(
		&p.status,
		webSocketStreamConnRunning,
		webSocketStreamConnRunning,
	) {
		return errors.ErrStreamConnIsClosed
	} else {
		return p.writeMessage(
			websocket.BinaryMessage,
			stream.GetBufferUnsafe(),
			timeout,
		)
	}
}

func (p *websocketStreamConn) Close() *base.Error {
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
			case <-time.After(2 * time.Second):
			}
		}

		// 3. close and return
		return convertToError(p.wsConn.Close(), errors.ErrWebsocketStreamConnWSConnClose)
	} else if atomic.CompareAndSwapInt32(
		&p.status,
		webSocketStreamConnCanClose,
		webSocketStreamConnClosed,
	) {
		// 1. close and return
		return convertToError(p.wsConn.Close(), errors.ErrWebsocketStreamConnWSConnClose)
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

type websocketServerAdapter struct {
	addr     string
	wsServer *http.Server
	base.StatusManager
}

// NewWebsocketServerAdapter ...
func NewWebsocketServerAdapter(addr string) core.IServerAdapter {
	return &websocketServerAdapter{
		addr:     addr,
		wsServer: nil,
	}
}

// Open ...
func (p *websocketServerAdapter) Open(
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
				onError(0, errors.ErrWebsocketServerAdapterUpgrade)
			} else {
				onConnRun(newWebsocketStreamConn(conn), conn.RemoteAddr())
			}
		})
		p.wsServer = &http.Server{
			Addr:    p.addr,
			Handler: mux,
		}
	}) {
		onError(0, errors.ErrWebsocketServerAdapterAlreadyRunning)
	} else {
		if e := p.wsServer.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			onError(
				0,
				errors.ErrWebsocketServerAdapterWSServerListenAndServe.
					AddDebug(e.Error()),
			)
		}
		p.SetClosing(nil)
		p.SetClosed(func() {
			p.wsServer = nil
		})
	}
}

// Close ...
func (p *websocketServerAdapter) Close(onError func(uint64, *base.Error)) {
	waitCH := chan bool(nil)
	if onError == nil {
		panic("onError is nil")
	} else if !p.SetClosing(func(ch chan bool) {
		waitCH = ch
		if e := p.wsServer.Close(); e != nil {
			onError(
				0,
				errors.ErrWebsocketServerAdapterWSServerClose.
					AddDebug(e.Error()),
			)
		}
	}) {
		onError(0, errors.ErrWebsocketServerAdapterNotRunning)
	} else {
		select {
		case <-waitCH:
		case <-time.After(5 * time.Second):
			onError(0, errors.ErrWebsocketServerAdapterCloseTimeout)
		}
	}
}

type websocketClientAdapter struct {
	conn          core.IStreamConn
	connectString string
	base.StatusManager
}

// NewWebsocketClientAdapter ...
func NewWebsocketClientAdapter(connectString string) core.IClientAdapter {
	return &websocketClientAdapter{
		conn:          nil,
		connectString: connectString,
	}
}

func (p *websocketClientAdapter) Open(
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
		onError(errors.ErrWebsocketClientAdapterDial.AddDebug(err.Error()))
	} else {
		streamConn := newWebsocketStreamConn(conn)
		if !p.SetRunning(func() {
			p.conn = streamConn
		}) {
			_ = conn.Close()
			onError(errors.ErrWebsocketClientAdapterAlreadyRunning)
		} else {
			onConnRun(streamConn)
			p.SetClosing(nil)
			p.SetClosed(func() {
				p.conn = nil
			})
		}
	}
}

func (p *websocketClientAdapter) Close(onError func(*base.Error)) {
	waitCH := chan bool(nil)

	if onError == nil {
		panic("onError is nil")
	} else if !p.SetClosing(func(ch chan bool) {
		waitCH = ch
		if e := p.conn.Close(); e != nil {
			onError(e)
		}
	}) {
		onError(errors.ErrWebsocketClientAdapterNotRunning)
	} else {
		select {
		case <-waitCH:
		case <-time.After(5 * time.Second):
			onError(errors.ErrWebsocketClientAdapterCloseTimeout)
		}
	}
}
