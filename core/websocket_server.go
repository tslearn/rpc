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
	readSizeLimit uint64
	readTimeoutNS uint64
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
		readTimeoutNS: 60 * uint64(time.Second),
		httpServer:    nil,
		closeChan:     make(chan bool, 1),
		seed:          1,
	}
}

func (p *WebSocketServer) AddService(
	name string,
	serviceMeta *rpcServiceMeta,
) *WebSocketServer {
	err := p.processor.AddService(name, serviceMeta)
	if err != nil {
		p.logger.Error(err)
	}
	return p
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
			p.processor.stop()
			p.logger.Infof(
				"WebSocketServer: stopped",
			)
			p.closeChan <- true
		}()

		p.logger.Infof(
			"WebSocketServer: start at %s",
			GetURLBySchemeHostPortAndPath("ws", host, port, path),
		)
		p.processor.start()
		serverMux := http.NewServeMux()
		serverMux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
			if req != nil && req.Header != nil {
				req.Header.Del("Origin")
			}
			conn, err := wsUpgradeManager.Upgrade(w, req, nil)
			if err != nil {
				p.logger.Errorf(
					"WebSocketServer: %s",
					err.Error(),
				)
				return
			}

			conn.SetReadLimit(int64(atomic.LoadUint64(&p.readSizeLimit)))
			p.onOpen(conn)

			defer func() {
				err := conn.Close()
				if err != nil {
					ret = NewRPCErrorByError(err)
					p.onError(conn, ret)
				}
				p.onClose(conn)
			}()

			for {
				nextTimeoutNS := TimeNowNS() +
					int64(atomic.LoadUint64(&p.readTimeoutNS))
				if err := conn.SetReadDeadline(time.Unix(
					nextTimeoutNS/int64(time.Second),
					nextTimeoutNS%int64(time.Second),
				)); err != nil {
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
					p.onError(conn, NewRPCError(
						"WebSocketServer: unknown message type",
					))
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
		"WebSocketServer: has already been opened",
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
		"WebSocketServer: close error, it is not opened",
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
	stream := newRPCStream()
	stream.setWritePosUnsafe(0)
	stream.putBytes(bytes)
	p.processor.put(stream)
}

// GetLogger get WebSocketServer logger
func (p *WebSocketServer) GetLogger() *Logger {
	return p.logger
}

// SetReadSizeLimit set WebSocketServer read limit in byte
func (p *WebSocketServer) SetReadSizeLimit(readLimit uint64) {
	atomic.StoreUint64(&p.readSizeLimit, readLimit)
}

// SetReadTimeoutMS set WebSocketServer timeout in millisecond
func (p *WebSocketServer) SetReadTimeoutMS(readTimeoutMS uint64) {
	atomic.StoreUint64(&p.readTimeoutNS, readTimeoutMS*uint64(time.Millisecond))
}
