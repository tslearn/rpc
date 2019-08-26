package core

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// The connection is connecting
	wsClientConnecting = int64(0)
	// The connection is open and ready to communicate.
	wsClientOpen = int64(1)
	// The connection is in the process of closing.
	wsClientClosing = int64(2)
	// The connection is closed or couldn't be opened.
	wsClientClosed = int64(3)
)

// WebSocketClient is implement of INetClient via web socket
type WebSocketClient struct {
	conn          *websocket.Conn
	closeChan     chan bool
	readTimeoutNS uint64
	readyState    int64
	sync.Mutex
}

// NewWebSocketClient create a WebSocketClient, and connect to url
func NewWebSocketClient(url string) *WebSocketClient {
	client := &WebSocketClient{
		conn:          nil,
		closeChan:     make(chan bool, 1),
		readTimeoutNS: 16 * uint64(time.Second),
		readyState:    wsClientConnecting,
	}

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		client.onError(nil, NewRPCErrorByError(err))
		return nil
	}

	client.conn = conn
	client.readyState = wsClientOpen
	client.onOpen(conn)

	go func() {
		defer func() {
			err := client.conn.Close()
			if err != nil {
				client.onError(client.conn, NewRPCErrorByError(err))
			}
			atomic.StoreInt64(&client.readyState, wsClientClosed)
			client.onClose(client.conn)
			client.closeChan <- true
		}()

		conn.SetReadLimit(64 * 1024)

		for {
			// set next read dead line
			nextTimeoutNS := TimeNowNS() +
				int64(atomic.LoadUint64(&client.readTimeoutNS))
			if err := conn.SetReadDeadline(time.Unix(
				nextTimeoutNS/int64(time.Second),
				nextTimeoutNS%int64(time.Second),
			)); err != nil {
				client.onError(conn, NewRPCErrorByError(err))
				return
			}

			// read message
			mt, message, err := conn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					client.onError(conn, NewRPCErrorByError(err))
				}
				return
			}
			switch mt {
			case websocket.BinaryMessage:
				client.onBinary(conn, message)
				break
			case websocket.CloseMessage:
				return
			default:
				client.onError(conn, NewRPCError("unknown message type"))
				return
			}
		}
	}()

	return client
}

// SetReadLimit set WebSocketClient read limit in byte
func (p *WebSocketClient) SetReadSizeLimit(readSizeLimit uint64) {
	p.conn.SetReadLimit(int64(readSizeLimit))
}

// SetReadTimeoutMS set WebSocketClient timeout in millisecond
func (p *WebSocketClient) SetReadTimeoutMS(readTimeoutMS uint64) {
	atomic.StoreUint64(&p.readTimeoutNS, readTimeoutMS*uint64(time.Millisecond))
}

// IsOpen returns true when the WebSocketClient is linked, otherwise false
func (p *WebSocketClient) IsOpen() bool {
	return atomic.LoadInt64(&p.readyState) == wsClientOpen
}

// SendBinary send byte array to the remote server
func (p *WebSocketClient) SendBinary(data []byte) *rpcError {
	return p.send(websocket.BinaryMessage, data)
}

func (p *WebSocketClient) send(messageType int, data []byte) *rpcError {
	if atomic.LoadInt64(&p.readyState) != wsClientOpen {
		err := NewRPCError(
			"websocket-client: send error, connection is not opened",
		)
		p.onError(p.conn, err)
		return err
	}

	p.Lock()
	defer p.Unlock()
	err := p.conn.WriteMessage(messageType, data)

	if err != nil {
		ret := NewRPCErrorByError(err)
		p.onError(p.conn, ret)
		return ret
	}

	return nil
}

// Close close the WebSocketClient
func (p *WebSocketClient) Close() *rpcError {
	p.Lock()
	defer p.Unlock()

	if atomic.LoadInt64(&p.readyState) != wsClientOpen {
		err := NewRPCError(
			"websocket-client: connection is not opened",
		)
		p.onError(p.conn, err)
		return err
	}

	err := p.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)

	if err != nil {
		ret := NewRPCErrorByError(err)
		p.onError(p.conn, ret)
		return ret
	}

	atomic.StoreInt64(&p.readyState, wsClientClosing)

	select {
	case <-p.closeChan:
		return nil
	case <-time.After(1200 * time.Millisecond):
		err := NewRPCError(
			"websocket-client: close timeout",
		)
		p.onError(p.conn, err)
		return err
	}
}

func (p *WebSocketClient) onOpen(conn *websocket.Conn) {
	fmt.Println("client onOpen", conn)
}

func (p *WebSocketClient) onError(conn *websocket.Conn, err *rpcError) {
	fmt.Println("client onError", conn, err)
}

func (p *WebSocketClient) onClose(conn *websocket.Conn) {
	fmt.Println("client onClose", conn)
}

func (p *WebSocketClient) onBinary(conn *websocket.Conn, bytes []byte) {
	fmt.Println("client onBinary", conn, bytes)
}
