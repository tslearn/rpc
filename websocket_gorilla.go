package rpc

import (
	"github.com/gorilla/websocket"
	"github.com/rpccloud/rpc/internal"
	"net/http"
	"runtime/debug"
	"time"
)

// Begin ***** webSocketConn ***** //
type webSocketConn websocket.Conn

func (p *webSocketConn) ReadStream(
	timeout time.Duration,
	readLimit int64,
) (*Stream, Error) {
	if conn := (*websocket.Conn)(p); conn == nil {
		return nil, internal.NewKernelPanic(
			"rpc: object nil object",
		).AddDebug(string(debug.Stack()))
	} else {
		conn.SetReadLimit(readLimit)
		if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return nil, internal.NewTransportError(
				internal.ConcatString("rpc: ", err.Error()),
			)
		} else if mt, message, err := conn.ReadMessage(); err != nil {
			return nil, internal.NewTransportError(
				internal.ConcatString("rpc: ", err.Error()),
			)
		} else if mt != websocket.BinaryMessage {
			return nil, internal.NewTransportError(
				"rpc: unsupported protocol",
			)
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
	writeLimit int64,
) Error {
	if conn := (*websocket.Conn)(p); conn == nil {
		return internal.NewTransportError(
			"webSocketConn: WriteStream: nil object",
		)
	} else if stream == nil {
		return internal.NewTransportError(
			"webSocketConn: WriteStream: stream is nil",
		)
	} else if writeLimit > 0 && writeLimit < int64(stream.GetWritePos()) {
		return internal.NewTransportError(
			"webSocketConn: WriteStream: stream data overflow",
		)
	} else if err := conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return internal.NewTransportError(
			internal.ConcatString("webSocketConn: WriteStream: ", err.Error()),
		)
	} else if err := conn.WriteMessage(
		websocket.BinaryMessage,
		stream.GetBufferUnsafe(),
	); err != nil {
		return internal.NewTransportError(
			internal.ConcatString("webSocketConn: WriteStream: ", err.Error()),
		)
	} else {
		return nil
	}
}

func (p *webSocketConn) Close() Error {
	if conn := (*websocket.Conn)(p); conn == nil {
		return internal.NewTransportError(
			"webSocketConn: Close: nil object",
		)
	} else if err := conn.Close(); err != nil {
		return internal.NewTransportError(
			internal.ConcatString("webSocketConn: Close: ", err.Error()),
		)
	} else {
		return nil
	}
}

// End ***** webSocketConn ***** //

// Begin ***** WebSocketServerAdapter ***** //
var (
	wsUpgradeManager = websocket.Upgrader{
		ReadBufferSize:    2048,
		WriteBufferSize:   2048,
		EnableCompression: true,
	}
)

type WebSocketServerAdapter struct {
	addr       string
	path       string
	closeCH    chan bool
	httpServer *http.Server
	internal.Lock
}

func NewWebSocketServerAdapter(addr string, path string) IAdapter {
	if path == "" || path[0] != '/' {
		path = "/" + path
	}

	return &WebSocketServerAdapter{
		addr:       addr,
		path:       path,
		closeCH:    nil,
		httpServer: nil,
	}
}

// Start it must be none block
func (p *WebSocketServerAdapter) Open(
	onConnRun func(IStreamConn),
	onError func(Error),
) bool {
	err := Error(nil)
	defer func() {
		if onError != nil && err != nil {
			onError(err)
		}
	}()

	return p.CallWithLock(func() interface{} {
		if onConnRun == nil {
			err = internal.NewTransportError(
				"WebSocketServerAdapter: Start: onConnRun is nil",
			)
			return false
		} else if p.httpServer != nil {
			err = internal.NewTransportError(
				"WebSocketServerAdapter: Start: it has already been opened",
			)
			return false
		} else {
			mux := http.NewServeMux()
			mux.HandleFunc(p.path, func(w http.ResponseWriter, req *http.Request) {
				if conn, err := wsUpgradeManager.Upgrade(w, req, nil); err != nil {
					onError(internal.NewTransportError(
						internal.ConcatString("WebSocketServerAdapter: Start: ", err.Error()),
					))
				} else {
					onConnRun((*webSocketConn)(conn))
				}
			})

			closeCH := make(chan bool, 1)
			p.closeCH = closeCH
			p.httpServer = &http.Server{
				Addr:    p.addr,
				Handler: mux,
			}
			// start sub routine
			go func() {
				defer func() {
					p.DoWithLock(func() {
						p.closeCH = nil
						p.httpServer = nil
						closeCH <- true
					})
				}()
				if e := p.httpServer.ListenAndServe(); e != nil && e != http.ErrServerClosed {
					onError(internal.NewTransportError(
						internal.ConcatString("WebSocketServerAdapter: Start: ", e.Error()),
					))
				}
			}()
			return true
		}
	}).(bool)
}

func (p *WebSocketServerAdapter) Close(onError func(Error)) bool {
	err := Error(nil)
	defer func() {
		if onError != nil && err != nil {
			onError(err)
		}
	}()

	closeCH := p.CallWithLock(func() interface{} {
		if p.closeCH == nil {
			err = internal.NewTransportError(
				"WebSocketServerAdapter: Close: has not been opened",
			)
			return nil
		} else if e := p.httpServer.Close(); e != nil {
			err = internal.NewTransportError(
				internal.ConcatString("WebSocketServerAdapter: Close: ", e.Error()),
			)
			return nil
		} else {
			ret := p.closeCH
			p.closeCH = nil
			return ret
		}
	})

	if closeCH == nil {
		return false
	} else {
		select {
		case <-closeCH.(chan bool):
			return true
		case <-time.After(10 * time.Second):
			err = internal.NewTransportError(
				"WebSocketServerAdapter: Close: can not close in 10 seconds",
			)
			return false
		}
	}
}

func (p *WebSocketServerAdapter) IsRunning() bool {
	return p.CallWithLock(func() interface{} {
		return p.closeCH != nil
	}).(bool)
}

func (p *WebSocketServerAdapter) ConnectString() string {
	return "ws://" + p.addr + p.path
}

// End ***** WebSocketServerAdapter ***** //

// Begin ***** WebSocketClientEndPoint ***** //
type WebSocketClientEndPoint struct {
	conn          *webSocketConn
	closeCH       chan bool
	connectString string
	internal.Lock
}

func NewWebSocketClientEndPoint(connectString string) IAdapter {
	return &WebSocketClientEndPoint{
		conn:          nil,
		closeCH:       nil,
		connectString: connectString,
	}
}

func (p *WebSocketClientEndPoint) Open(
	onConnRun func(IStreamConn),
	onError func(Error),
) bool {
	err := Error(nil)
	defer func() {
		if onError != nil && err != nil {
			onError(err)
		}
	}()

	return p.CallWithLock(func() interface{} {
		if onConnRun == nil {
			err = internal.NewTransportError(
				"WebSocketClientEndPoint: Start: onConnRun is nil",
			)
			return false
		} else if p.conn != nil {
			err = internal.NewTransportError(
				"WebSocketClientEndPoint: Start: it has already been opened",
			)
			return false
		} else if wsConn, _, err := websocket.DefaultDialer.Dial(
			p.connectString,
			nil,
		); err != nil {
			err = internal.NewTransportError(
				internal.ConcatString("WebSocketClientEndPoint: Start: ", err.Error()),
			)
			return false
		} else if wsConn == nil {
			err = internal.NewTransportError(
				"WebSocketClientEndPoint: Start: wsConn is nil",
			)
			return false
		} else {
			closeCH := make(chan bool, 1)
			p.closeCH = closeCH
			p.conn = (*webSocketConn)(wsConn)
			go func() {
				defer func() {
					p.DoWithLock(func() {
						p.closeCH = nil
						p.conn = nil
						closeCH <- true
					})
				}()
				onConnRun(p.conn)
			}()
			return true
		}
	}).(bool)
}

func (p *WebSocketClientEndPoint) Close(onError func(Error)) bool {
	err := Error(nil)
	defer func() {
		if onError != nil && err != nil {
			onError(err)
		}
	}()

	closeCH := p.CallWithLock(func() interface{} {
		if p.closeCH == nil {
			err = internal.NewTransportError(
				"WebSocketClientEndPoint: Close: has not been opened",
			)
			return nil
		} else if e := p.conn.Close(); e != nil {
			err = internal.NewTransportError(
				internal.ConcatString("WebSocketClientEndPoint: Close: ", e.Error()),
			)
			return nil
		} else {
			ret := p.closeCH
			p.closeCH = nil
			return ret
		}
	})

	if closeCH == nil {
		return false
	} else {
		select {
		case <-closeCH.(chan bool):
			return true
		case <-time.After(10 * time.Second):
			err = internal.NewTransportError(
				"WebSocketClientEndPoint: Close: can not close in 10 seconds",
			)
			return false
		}
	}
}

func (p *WebSocketClientEndPoint) IsRunning() bool {
	return p.CallWithLock(func() interface{} {
		return p.closeCH != nil
	}).(bool)
}

func (p *WebSocketClientEndPoint) ConnectString() string {
	return p.connectString
}

// End ***** WebSocketClientEndPoint ***** //
