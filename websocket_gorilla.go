package rpcc

import (
	"github.com/gorilla/websocket"
	"github.com/tslearn/rpcc/internal"
	"github.com/tslearn/rpcc/util"
	"net/http"
	"time"
)

// Begin ***** webSocketConn ***** //
type webSocketConn websocket.Conn

func (p *webSocketConn) ReadStream(
	timeout time.Duration,
	readLimit int64,
) (Stream, Error) {
	if conn := (*websocket.Conn)(p); conn == nil {
		return nil, internal.NewRPCError(
			"webSocketConn: ReadStream: nil object",
		)
	} else {
		conn.SetReadLimit(readLimit)
		if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return nil, internal.NewRPCError(
				util.ConcatString("webSocketConn: ReadStream: ", err.Error()),
			)
		} else if mt, message, err := conn.ReadMessage(); err != nil {
			return nil, internal.NewRPCError(
				util.ConcatString("webSocketConn: ReadStream: ", err.Error()),
			)
		} else if mt != websocket.BinaryMessage {
			return nil, internal.NewRPCError(
				"webSocketConn: ReadStream: unsupported protocol",
			)
		} else if len(message) < internal.StreamBodyPos {
			return nil, internal.NewRPCError(
				"webSocketConn: ReadStream: stream data error",
			)
		} else {
			stream := internal.NewRPCStream()
			stream.SetWritePos(0)
			stream.PutBytes(message)
			return stream, nil
		}
	}
}

func (p *webSocketConn) WriteStream(
	stream Stream,
	timeout time.Duration,
	writeLimit int64,
) Error {
	if conn := (*websocket.Conn)(p); conn == nil {
		return internal.NewRPCError(
			"webSocketConn: WriteStream: nil object",
		)
	} else if stream == nil {
		return internal.NewRPCError(
			"webSocketConn: WriteStream: stream is nil",
		)
	} else if writeLimit > 0 && writeLimit < int64(stream.GetWritePos()) {
		return internal.NewRPCError(
			"webSocketConn: WriteStream: stream data overflow",
		)
	} else if err := conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return internal.NewRPCError(
			util.ConcatString("webSocketConn: WriteStream: ", err.Error()),
		)
	} else if err := conn.WriteMessage(
		websocket.BinaryMessage,
		stream.GetBufferUnsafe(),
	); err != nil {
		return internal.NewRPCError(
			util.ConcatString("webSocketConn: WriteStream: ", err.Error()),
		)
	} else {
		return nil
	}
}

func (p *webSocketConn) Close() Error {
	if conn := (*websocket.Conn)(p); conn == nil {
		return internal.NewRPCError(
			"webSocketConn: Close: nil object",
		)
	} else if err := conn.Close(); err != nil {
		return internal.NewRPCError(
			util.ConcatString("webSocketConn: Close: ", err.Error()),
		)
	} else {
		return nil
	}
}

// End ***** webSocketConn ***** //

// Begin ***** WebSocketServerEndPoint ***** //
var (
	wsUpgradeManager = websocket.Upgrader{
		ReadBufferSize:    2048,
		WriteBufferSize:   2048,
		EnableCompression: true,
	}
)

type WebSocketServerEndPoint struct {
	addr       string
	path       string
	closeCH    chan bool
	httpServer *http.Server
	util.AutoLock
}

func NewWebSocketServerEndPoint(addr string, path string) IEndPoint {
	if path == "" || path[0] != '/' {
		path = "/" + path
	}

	return &WebSocketServerEndPoint{
		addr:       addr,
		path:       path,
		closeCH:    nil,
		httpServer: nil,
	}
}

// Open it must be none block
func (p *WebSocketServerEndPoint) Open(
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
			err = internal.NewRPCError(
				"WebSocketServerEndPoint: Open: onConnRun is nil",
			)
			return false
		} else if p.httpServer != nil {
			err = internal.NewRPCError(
				"WebSocketServerEndPoint: Open: it has already been opened",
			)
			return false
		} else {
			mux := http.NewServeMux()
			mux.HandleFunc(p.path, func(w http.ResponseWriter, req *http.Request) {
				if conn, err := wsUpgradeManager.Upgrade(w, req, nil); err != nil {
					onError(internal.NewRPCError(
						util.ConcatString("WebSocketServerEndPoint: Open: ", err.Error()),
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
					onError(internal.NewRPCError(
						util.ConcatString("WebSocketServerEndPoint: Open: ", e.Error()),
					))
				}
			}()
			return true
		}
	}).(bool)
}

func (p *WebSocketServerEndPoint) Close(onError func(Error)) bool {
	err := Error(nil)
	defer func() {
		if onError != nil && err != nil {
			onError(err)
		}
	}()

	closeCH := p.CallWithLock(func() interface{} {
		if p.closeCH == nil {
			err = internal.NewRPCError(
				"WebSocketServerEndPoint: Close: has not been opened",
			)
			return nil
		} else if e := p.httpServer.Close(); e != nil {
			err = internal.NewRPCError(
				util.ConcatString("WebSocketServerEndPoint: Close: ", e.Error()),
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
			err = internal.NewRPCError(
				"WebSocketServerEndPoint: Close: can not close in 10 seconds",
			)
			return false
		}
	}
}

func (p *WebSocketServerEndPoint) IsRunning() bool {
	return p.CallWithLock(func() interface{} {
		return p.closeCH != nil
	}).(bool)
}

func (p *WebSocketServerEndPoint) ConnectString() string {
	return "ws://" + p.addr + p.path
}

// End ***** WebSocketServerEndPoint ***** //

// Begin ***** WebSocketClientEndPoint ***** //
type WebSocketClientEndPoint struct {
	conn          *webSocketConn
	closeCH       chan bool
	connectString string
	util.AutoLock
}

func NewWebSocketClientEndPoint(connectString string) IEndPoint {
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
			err = internal.NewRPCError(
				"WebSocketClientEndPoint: Open: onConnRun is nil",
			)
			return false
		} else if p.conn != nil {
			err = internal.NewRPCError(
				"WebSocketClientEndPoint: Open: it has already been opened",
			)
			return false
		} else if wsConn, _, err := websocket.DefaultDialer.Dial(
			p.connectString,
			nil,
		); err != nil {
			err = internal.NewRPCError(
				util.ConcatString("WebSocketClientEndPoint: Open: ", err.Error()),
			)
			return false
		} else if wsConn == nil {
			err = internal.NewRPCError(
				"WebSocketClientEndPoint: Open: wsConn is nil",
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
			err = internal.NewRPCError(
				"WebSocketClientEndPoint: Close: has not been opened",
			)
			return nil
		} else if e := p.conn.Close(); e != nil {
			err = internal.NewRPCError(
				util.ConcatString("WebSocketClientEndPoint: Close: ", e.Error()),
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
			err = internal.NewRPCError(
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
