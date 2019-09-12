package core

import (
	"fmt"
	"github.com/gorilla/websocket"
	"math"
	"net/http"
	"strconv"
	"strings"
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
)

type wsServerConn struct {
	id         uint32
	wsConn     *websocket.Conn
	connIndex  uint32
	security   string
	deadlineNS int64
	streamCH   chan *rpcStream
	sequence   uint32
	sync.Mutex
}

func (p *wsServerConn) getSequence() uint32 {
	ret := uint32(0)
	p.Lock()
	ret = p.sequence
	p.Unlock()
	return ret
}

func (p *wsServerConn) setSequence(from uint32, to uint32) bool {
	ret := false
	p.Lock()
	if p.sequence == from && from != to {
		p.sequence = to
		ret = true
	}
	p.Unlock()
	return ret
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
	server := &WebSocketServer{
		processor:     nil,
		logger:        NewLogger(),
		status:        wsServerClosed,
		readSizeLimit: 64 * 1024,
		readTimeoutNS: 60 * uint64(time.Second),
		httpServer:    nil,
		seed:          1,
	}

	server.processor = newProcessor(
		server.logger,
		32,
		32,
		func(stream *rpcStream, success bool) {
			if serverConn := server.getConnByID(
				stream.getClientConnInfo(),
			); serverConn != nil {
				serverConn.streamCH <- stream
			}
		},
	)
	return server
}

func (p *WebSocketServer) serverConnWriteRoutine(serverConn *wsServerConn) {
	ch := serverConn.streamCH
	for stream := <-ch; stream != nil; stream = <-ch {
		stream.setClientConnInfo(0)
		for serverConn.security != "" {
			if wsConn := serverConn.wsConn; wsConn != nil {
				if err := serverConn.wsConn.WriteMessage(
					websocket.BinaryMessage,
					stream.getBufferUnsafe(),
				); err == nil {
					stream.Release()
					break
				} else {
					p.onError(serverConn, err.Error())
				}
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func (p *WebSocketServer) registerConn(
	wsConn *websocket.Conn,
	id uint32,
	security string,
) *wsServerConn {
	// id and security is ok
	if v, ok := p.Load(id); ok {
		serverConn, ok := v.(*wsServerConn)
		if ok && serverConn != nil && serverConn.security == security {
			serverConn.wsConn = wsConn
			return serverConn
		}
	}

	p.Lock()
	ret := (*wsServerConn)(nil)
	for {
		p.seed += 1
		if p.seed == math.MaxUint32 {
			p.seed = 1
		}

		id = p.seed
		if _, ok := p.Load(id); !ok {
			ret = &wsServerConn{
				id:         id,
				sequence:   1,
				security:   GetRandString(32),
				wsConn:     wsConn,
				connIndex:  0,
				deadlineNS: 0,
				streamCH:   make(chan *rpcStream, 64),
			}
			p.Store(id, ret)
			go p.serverConnWriteRoutine(ret)
			break
		}
	}
	p.Unlock()
	return ret
}

func (p *WebSocketServer) unregisterConn(id uint32) bool {
	if serverConn, ok := p.Load(id); ok {
		serverConn.(*wsServerConn).wsConn = nil
		serverConn.(*wsServerConn).deadlineNS = TimeNowNS() +
			25*int64(time.Second)
		return true
	} else {
		return false
	}
}

func (p *WebSocketServer) swipeConn() {
	for atomic.LoadInt32(&p.status) != wsServerClosed {
		nowNS := TimeNowNS()
		p.Range(func(key, value interface{}) bool {
			v, ok := value.(*wsServerConn)
			if ok && v != nil {
				if v.deadlineNS > 0 && v.deadlineNS < nowNS {
					p.Delete(key)
					v.wsConn = nil
					v.security = ""
					close(v.streamCH)
					v.streamCH = nil
				}
			}
			return true
		})

		time.Sleep(500 * time.Millisecond)
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
		p.logger.Error(err.String())
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
			p.logger.Error(err.String())
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

			connID := uint32(0)
			connSecurity := ""
			keys, ok := req.URL.Query()["conn"]
			if ok && len(keys) == 1 {
				arr := strings.Split(keys[0], "#")
				if len(arr) == 2 {
					if parseID, err := strconv.ParseUint(arr[0], 10, 64); err == nil {
						connID = uint32(parseID)
						connSecurity = arr[1]
					}
				}
			}

			wsConn, err := wsUpgradeManager.Upgrade(w, req, nil)
			if err != nil {
				p.logger.Errorf("WebSocketServer: %s", err.Error())
				return
			}

			serverConn := p.registerConn(wsConn, connID, connSecurity)

			if err := wsConn.WriteMessage(
				websocket.BinaryMessage,
				[]byte(fmt.Sprintf("%d#%s", serverConn.id, serverConn.security)),
			); err != nil {
				p.logger.Errorf("WebSocketServer: %s", err.Error())
				return
			}

			wsConn.SetReadLimit(int64(atomic.LoadUint64(&p.readSizeLimit)))

			p.onOpen(serverConn)

			defer func() {
				err := wsConn.Close()
				if err != nil {
					p.onError(serverConn, err.Error())
				}
				p.onClose(serverConn)
				p.unregisterConn(serverConn.id)
			}()

			for {
				nextTimeoutNS := TimeNowNS() +
					int64(atomic.LoadUint64(&p.readTimeoutNS))
				if err := wsConn.SetReadDeadline(time.Unix(
					nextTimeoutNS/int64(time.Second),
					nextTimeoutNS%int64(time.Second),
				)); err != nil {
					p.onError(serverConn, err.Error())
					return
				}

				mt, message, err := wsConn.ReadMessage()
				if err != nil {
					if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
						p.onError(serverConn, err.Error())
					}
					return
				}
				switch mt {
				case websocket.BinaryMessage:
					stream := newRPCStream()
					stream.setWritePosUnsafe(0)
					stream.putBytes(message)

					serverSequence := stream.getClientConnInfo()
					callbackID := stream.getClientCallbackID()

					// this is system instructions
					if callbackID == 0 {

					} else { // this is rpc callback function
						if serverConn.setSequence(serverSequence, callbackID) {
							stream.setClientConnInfo(serverConn.id)
							p.onStream(serverConn, stream)
						} else {
							stream.Release()
							p.onError(serverConn, "server sequence error")
							return
						}
					}
				default:
					p.onError(serverConn, "unknown message type")
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

		time.AfterFunc(250*time.Millisecond, func() {
			atomic.CompareAndSwapInt32(&p.status, wsServerOpening, wsServerOpened)
		})

		go p.swipeConn()

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

func (p *WebSocketServer) onOpen(serverConn *wsServerConn) {
	p.logger.Infof("WebSocketServerConn[%d]: opened", serverConn.id)
}

func (p *WebSocketServer) onError(serverConn *wsServerConn, msg string) {
	p.logger.Warnf("WebSocketServerConn[%d]: %s", serverConn.id, msg)
}

func (p *WebSocketServer) onClose(serverConn *wsServerConn) {
	p.logger.Infof("WebSocketServerConn[%d]: closed", serverConn.id)
}

func (p *WebSocketServer) onStream(
	serverConn *wsServerConn,
	stream *rpcStream,
) {
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
