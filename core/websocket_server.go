package core

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var wsUpgradeManager = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
}

// WebSocketServer is implement of INetServer via web socket
type WebSocketServer struct {
	processor     *rpcProcessor
	logger        *Logger
	startNS       int64
	readSizeLimit int64
	readTimeoutNS int64
	httpServer    *http.Server
	closeChan     chan bool
	seed          int64
	sync.Mutex
}

// NewWebSocketServer create a WebSocketClient
func NewWebSocketServer() *WebSocketServer {
	logger := NewLogger()
	processor := newProcessor(
		logger,
		32,
		32,
		func(stream *rpcStream, success bool) {

		},
	)
	return &WebSocketServer{
		processor:     processor,
		logger:        logger,
		startNS:       0,
		readSizeLimit: 64 * 1024,
		readTimeoutNS: 60 * int64(time.Second),
		httpServer:    nil,
		closeChan:     make(chan bool, 1),
		seed:          1,
	}
}

// Open make the WebSocketServer start serve
func (p *WebSocketServer) Start(
	host string,
	port uint16,
	path string,
) (ret *rpcError) {
	timeNS := TimeNowNS()
	if atomic.CompareAndSwapInt64(&p.startNS, 0, timeNS) {
		defer func() {
			atomic.CompareAndSwapInt64(&p.startNS, timeNS, 0)
			p.logger.Infof(
				"websocket-server: stopped",
			)
			p.closeChan <- true
		}()

		p.logger.Infof(
			"websocket-server: start at %s",
			GetURLBySchemeHostPortAndPath("ws", host, port, path),
		)
		serverMux := http.NewServeMux()
		serverMux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
			if req != nil && req.Header != nil {
				req.Header.Del("Origin")
			}
			conn, err := wsUpgradeManager.Upgrade(w, req, nil)
			if err != nil {
				p.logger.Errorf(
					"websocket-server: %s",
					err.Error(),
				)
				return
			}

			conn.SetReadLimit(atomic.LoadInt64(&p.readSizeLimit))
			p.onOpen(conn)

			defer func() {
				err := conn.Close()
				if err != nil {
					ret = NewRPCErrorByError(err)
					p.onError(conn, ret)
				}
				p.onClose(conn)
			}()

			timeoutNS := int64(0)
			readTimeoutNS := atomic.LoadInt64(&p.readTimeoutNS)

			for {
				nowNS := TimeNowNS()

				if readTimeoutNS > 0 {
					timeoutNS = nowNS + readTimeoutNS
				} else {
					timeoutNS = nowNS + 3600*int64(time.Second)
				}
				if conn.SetReadDeadline(
					time.Unix(timeoutNS/int64(time.Second), timeoutNS%int64(time.Second)),
				) != nil {
					p.onError(conn, NewRPCErrorByError(err))
					return
				}

				mt, message, err := conn.ReadMessage()
				if err != nil {
					if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
						p.onError(conn, NewRPCErrorByError(err))
					}
					return
				}

				switch mt {
				case websocket.BinaryMessage:
					p.onBinary(conn, message)
				case websocket.CloseMessage:
					return
				default:
					p.onError(conn, NewRPCError("unknown message type"))
					return
				}
			}
		})
		p.Lock()
		p.httpServer = &http.Server{
			Addr:    fmt.Sprintf("%s:%d", host, port),
			Handler: serverMux,
		}
		p.Unlock()
		return NewRPCErrorByError(p.httpServer.ListenAndServe())
	}

	return NewRPCError(
		"websocket-server: has already been opened",
	)
}

// Close make the WebSocketServer stop serve
func (p *WebSocketServer) Close() *rpcError {
	p.Lock()
	defer p.Unlock()

	if atomic.LoadInt64(&p.startNS) > 0 {
		err := NewRPCErrorByError(p.httpServer.Close())
		<-p.closeChan
		return err
	}

	return NewRPCError(
		"websocket-server: close error, it is not opened",
	)
}

func (p *WebSocketServer) onOpen(conn *websocket.Conn) {
	fmt.Println("server conn onOpen", conn)
}

func (p *WebSocketServer) onError(conn *websocket.Conn, err *rpcError) {
	fmt.Println("server conn onError", conn, err)
}

func (p *WebSocketServer) onClose(conn *websocket.Conn) {
	fmt.Println("server conn onClose", conn)
}

func (p *WebSocketServer) onBinary(conn *websocket.Conn, bytes []byte) {
	fmt.Println("server conn onBinary length ", len(bytes))
}

// GetLogger get WebSocketServer logger
func (p *WebSocketServer) GetLogger() *Logger {
	return p.logger
}

// SetReadSizeLimit set WebSocketServer read limit in byte
func (p *WebSocketServer) SetReadSizeLimit(readLimit uint64) {
	atomic.StoreInt64(&p.readSizeLimit, int64(readLimit))
}

// SetReadTimeoutMS set WebSocketServer timeout in millisecond
func (p *WebSocketServer) SetReadTimeoutMS(readTimeoutMS uint64) {
	atomic.StoreInt64(
		&p.readTimeoutNS,
		int64(readTimeoutMS)*int64(time.Millisecond),
	)
}
