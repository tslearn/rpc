package rpcc

import (
	"github.com/gorilla/websocket"
	"github.com/tslearn/rpcc/util"
	"net/http"
	"time"
)

// Begin ***** webSocketConn ***** //
type webSocketConn websocket.Conn

func (p *webSocketConn) ReadStream(
	timeout time.Duration,
	readLimit int64,
) (*common.RPCStream, common.RPCError) {
	if conn := (*websocket.Conn)(p); conn == nil {
		return nil, common.NewRPCError(
			"webSocketConn: ReadStream: nil object",
		)
	} else {
		conn.SetReadLimit(readLimit)
		if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return nil, common.NewRPCError(
				common.ConcatString("webSocketConn: ReadStream: ", err.Error()),
			)
		} else if mt, message, err := conn.ReadMessage(); err != nil {
			return nil, common.NewRPCError(
				common.ConcatString("webSocketConn: ReadStream: ", err.Error()),
			)
		} else if mt != websocket.BinaryMessage {
			return nil, common.NewRPCError(
				"webSocketConn: ReadStream: unsupported protocol",
			)
		} else if len(message) < common.StreamBodyPos {
			return nil, common.NewRPCError(
				"webSocketConn: ReadStream: stream data error",
			)
		} else {
			stream := common.NewRPCStream()
			stream.SetWritePos(0)
			stream.PutBytes(message)
			return stream, nil
		}
	}
}

func (p *webSocketConn) WriteStream(
	stream *common.RPCStream,
	timeout time.Duration,
	writeLimit int64,
) common.RPCError {
	if conn := (*websocket.Conn)(p); conn == nil {
		return common.NewRPCError(
			"webSocketConn: WriteStream: nil object",
		)
	} else if stream == nil {
		return common.NewRPCError(
			"webSocketConn: WriteStream: stream is nil",
		)
	} else if writeLimit > 0 && writeLimit < int64(stream.GetWritePos()) {
		return common.NewRPCError(
			"webSocketConn: WriteStream: stream data overflow",
		)
	} else if err := conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return common.NewRPCError(
			common.ConcatString("webSocketConn: WriteStream: ", err.Error()),
		)
	} else if err := conn.WriteMessage(
		websocket.BinaryMessage,
		stream.GetBufferUnsafe(),
	); err != nil {
		return common.NewRPCError(
			common.ConcatString("webSocketConn: WriteStream: ", err.Error()),
		)
	} else {
		return nil
	}
}

func (p *webSocketConn) Close() common.RPCError {
	if conn := (*websocket.Conn)(p); conn == nil {
		return common.NewRPCError(
			"webSocketConn: Close: nil object",
		)
	} else if err := conn.Close(); err != nil {
		return common.NewRPCError(
			common.ConcatString("webSocketConn: Close: ", err.Error()),
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
	common.AutoLock
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
	onError func(common.RPCError),
) bool {
	err := common.RPCError(nil)
	defer func() {
		if onError != nil && err != nil {
			onError(err)
		}
	}()

	return p.CallWithLock(func() interface{} {
		if onConnRun == nil {
			err = common.NewRPCError(
				"WebSocketServerEndPoint: Open: onConnRun is nil",
			)
			return false
		} else if p.httpServer != nil {
			err = common.NewRPCError(
				"WebSocketServerEndPoint: Open: it has already been opened",
			)
			return false
		} else {
			mux := http.NewServeMux()
			mux.HandleFunc(p.path, func(w http.ResponseWriter, req *http.Request) {
				if conn, err := wsUpgradeManager.Upgrade(w, req, nil); err != nil {
					onError(common.NewRPCError(
						common.ConcatString("WebSocketServerEndPoint: Open: ", err.Error()),
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
					onError(common.NewRPCError(
						common.ConcatString("WebSocketServerEndPoint: Open: ", e.Error()),
					))
				}
			}()
			return true
		}
	}).(bool)
}

func (p *WebSocketServerEndPoint) Close(onError func(common.RPCError)) bool {
	err := common.RPCError(nil)
	defer func() {
		if onError != nil && err != nil {
			onError(err)
		}
	}()

	closeCH := p.CallWithLock(func() interface{} {
		if p.closeCH == nil {
			err = common.NewRPCError(
				"WebSocketServerEndPoint: Close: has not been opened",
			)
			return nil
		} else if e := p.httpServer.Close(); e != nil {
			err = common.NewRPCError(
				common.ConcatString("WebSocketServerEndPoint: Close: ", e.Error()),
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
			err = common.NewRPCError(
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
	common.AutoLock
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
	onError func(common.RPCError),
) bool {
	err := common.RPCError(nil)
	defer func() {
		if onError != nil && err != nil {
			onError(err)
		}
	}()

	return p.CallWithLock(func() interface{} {
		if onConnRun == nil {
			err = common.NewRPCError(
				"WebSocketClientEndPoint: Open: onConnRun is nil",
			)
			return false
		} else if p.conn != nil {
			err = common.NewRPCError(
				"WebSocketClientEndPoint: Open: it has already been opened",
			)
			return false
		} else if wsConn, _, err := websocket.DefaultDialer.Dial(
			p.connectString,
			nil,
		); err != nil {
			err = common.NewRPCError(
				common.ConcatString("WebSocketClientEndPoint: Open: ", err.Error()),
			)
			return false
		} else if wsConn == nil {
			err = common.NewRPCError(
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

func (p *WebSocketClientEndPoint) Close(onError func(common.RPCError)) bool {
	err := common.RPCError(nil)
	defer func() {
		if onError != nil && err != nil {
			onError(err)
		}
	}()

	closeCH := p.CallWithLock(func() interface{} {
		if p.closeCH == nil {
			err = common.NewRPCError(
				"WebSocketClientEndPoint: Close: has not been opened",
			)
			return nil
		} else if e := p.conn.Close(); e != nil {
			err = common.NewRPCError(
				common.ConcatString("WebSocketClientEndPoint: Close: ", e.Error()),
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
			err = common.NewRPCError(
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
