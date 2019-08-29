package core

import (
	"fmt"
	"github.com/gorilla/websocket"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	wsServerOpening    = int32(1)
	wsServerOpened     = int32(2)
	wsServerClosing    = int32(3)
	wsServerDidClosing = int32(4)
	wsServerClosed     = int32(5)
)

var (
	wsUpgradeManager = websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
	}

	wsServerConnCache = sync.Pool{
		New: func() interface{} {
			return &wsServerConn{
				wsConn: nil,
			}
		},
	}
)

type wsServerConn struct {
	wsConn *websocket.Conn
	sync.Mutex
}

func (p *wsServerConn) send(data []byte) error {
	p.Lock()
	err := p.wsConn.WriteMessage(websocket.BinaryMessage, data)
	p.Unlock()
	return err
}

// WebSocketServer is implement of INetServer via web socket
type WebSocketServer struct {
	processor     *rpcProcessor
	logger        *Logger
	status        int32
	readSizeLimit uint64
	readTimeoutNS uint64
	httpServer    *http.Server
	seed          uint32
	sync.Map
	sync.Mutex
}

// NewWebSocketServer create a WebSocketClient
func NewWebSocketServer() *WebSocketServer {
	logger := NewLogger()
	server := &WebSocketServer{
		processor:     nil,
		logger:        logger,
		status:        wsServerClosed,
		readSizeLimit: 64 * 1024,
		readTimeoutNS: 60 * uint64(time.Second),
		httpServer:    nil,
		seed:          1,
	}

	server.processor = newProcessor(
		logger,
		32,
		32,
		func(stream *rpcStream, success bool) {
			connID := stream.getClientConnID()
			if conn := server.getConnByID(connID); conn != nil {
				stream.setClientConnID(0)
				if err := conn.send(stream.getBufferUnsafe()); err != nil {
					server.onError(connID, err.Error())
				}
			}
		},
	)
	return server
}

func (p *WebSocketServer) registerConn(wsConn *websocket.Conn) uint32 {
	id := uint32(0)
	p.Lock()
	for {
		p.seed += 1
		if p.seed == math.MaxUint32 {
			p.seed = 1
		}

		id = p.seed
		if _, ok := p.Load(id); !ok {
			serverConn := wsServerConnCache.Get().(*wsServerConn)
			serverConn.wsConn = wsConn
			p.Store(id, serverConn)
			break
		}
	}
	p.Unlock()
	return id
}

func (p *WebSocketServer) unregisterConn(id uint32) bool {
	if serverConn, ok := p.Load(id); ok {
		p.Delete(id)
		serverConn.(*wsServerConn).wsConn = nil
		wsServerConnCache.Put(serverConn)
		return true
	} else {
		return false
	}
}

func (p *WebSocketServer) getConnByID(id uint32) *wsServerConn {
	if v, ok := p.Load(id); ok {
		return v.(*wsServerConn)
	} else {
		return nil
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

func (p *WebSocketServer) StartBackground(
	host string,
	port uint16,
	path string,
) {
	wait := true
	go func() {
		err := p.Start(host, port, path)
		if err != nil {
			p.logger.Error(err)
		}
		wait = false
	}()

	for wait {
		if atomic.LoadInt32(&p.status) == wsServerOpened {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// Start make the WebSocketServer start serve
func (p *WebSocketServer) Start(
	host string,
	port uint16,
	path string,
) (ret *rpcError) {
	if atomic.CompareAndSwapInt32(&p.status, wsServerClosed, wsServerOpening) {
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
			wsConn, err := wsUpgradeManager.Upgrade(w, req, nil)
			if err != nil {
				p.logger.Errorf("WebSocketServer: %s", err.Error())
				return
			}

			wsConn.SetReadLimit(int64(atomic.LoadUint64(&p.readSizeLimit)))
			connID := p.registerConn(wsConn)
			p.onOpen(connID)

			defer func() {
				err := wsConn.Close()
				if err != nil {
					p.onError(connID, err.Error())
				}
				p.onClose(connID)
				p.unregisterConn(connID)
			}()

			for {
				nextTimeoutNS := TimeNowNS() +
					int64(atomic.LoadUint64(&p.readTimeoutNS))
				if err := wsConn.SetReadDeadline(time.Unix(
					nextTimeoutNS/int64(time.Second),
					nextTimeoutNS%int64(time.Second),
				)); err != nil {
					p.onError(connID, err.Error())
					return
				}

				mt, message, err := wsConn.ReadMessage()
				if err != nil {
					if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
						p.onError(connID, err.Error())
					}
					return
				}
				switch mt {
				case websocket.BinaryMessage:
					p.onBinary(connID, message)
				default:
					p.onError(connID, "unknown message type")
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

		time.AfterFunc(300*time.Millisecond, func() {
			atomic.CompareAndSwapInt32(&p.status, wsServerOpening, wsServerOpened)
		})
		ret := p.httpServer.ListenAndServe()
		time.Sleep(400 * time.Millisecond)

		p.Lock()
		p.httpServer = nil
		p.Unlock()

		p.processor.stop()
		p.logger.Infof(
			"WebSocketServer: stopped",
		)

		if atomic.LoadInt32(&p.status) == wsServerClosing {
			atomic.StoreInt32(&p.status, wsServerDidClosing)
		} else {
			atomic.StoreInt32(&p.status, wsServerClosed)
		}

		if ret != nil && ret.Error() == "http: Server closed" {
			return nil
		} else {
			return NewRPCErrorByError(ret)
		}
	} else {
		return NewRPCError(
			"WebSocketServer: has already been started",
		)
	}
}

// Close make the WebSocketServer stop serve
func (p *WebSocketServer) Close() *rpcError {
	if atomic.CompareAndSwapInt32(&p.status, wsServerOpened, wsServerClosing) {
		err := NewRPCErrorByError(p.httpServer.Close())
		for !atomic.CompareAndSwapInt32(
			&p.status,
			wsServerDidClosing,
			wsServerClosed,
		) {
			time.Sleep(20 * time.Millisecond)
		}
		return err
	} else {
		return NewRPCError(
			"WebSocketServer: close error, it is not opened",
		)
	}
}

func (p *WebSocketServer) onOpen(connID uint32) {
	p.logger.Infof("WebSocketServerConn[%d]: opened", connID)
}

func (p *WebSocketServer) onError(connID uint32, msg string) {
	p.logger.Warningf("WebSocketServerConn[%d]: %s", connID, msg)
}

func (p *WebSocketServer) onClose(connID uint32) {
	p.logger.Infof("WebSocketServerConn[%d]: closed", connID)
}

func (p *WebSocketServer) onBinary(connID uint32, bytes []byte) {
	stream := newRPCStream()
	stream.setWritePosUnsafe(0)
	stream.putBytes(bytes)
	stream.setClientConnID(connID)
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
