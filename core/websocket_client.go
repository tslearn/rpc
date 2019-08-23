package core

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
	"sync/atomic"
	"time"
)

const (
	wsClientConnecting = int64(0) // The connection is not yet open.
	wsClientOpen       = int64(1) // The connection is open and ready to communicate.
	wsClientClosing    = int64(2) // The connection is in the process of closing.
	wsClientClosed     = int64(3) // The connection is closed or couldn't be opened.
)

// WebSocketClient is implement of INetClient via web socket
type WebSocketClient struct {
	readTimeoutNS int64
	//readSizeLimit     int64
	conn       *websocket.Conn
	closeChan  chan bool
	readyState int64
	sync.Mutex
}

// NewWebSocketClient create a WebSocketClient, and connect to url
func NewWebSocketClient(url string) *WebSocketClient {
	client := &WebSocketClient{
		readTimeoutNS: 16 * int64(time.Second),
		conn:          nil,
		closeChan:     make(chan bool, 1),
		readyState:    wsClientConnecting,
	}

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		rpcErr := WrapSystemError(err)
		client.onConnError(nil, rpcErr)
		return nil
	}

	client.conn = conn
	atomic.StoreInt64(&client.readyState, wsClientOpen)
	client.onConnOpen(conn)

	go func() {
		defer func() {
			err := client.conn.Close()
			if err != nil {
				client.onConnError(client.conn, WrapSystemError(err))
			}
			atomic.StoreInt64(&client.readyState, wsClientClosed)
			client.onConnClose(client.conn)
			client.closeChan <- true
		}()

		conn.SetReadLimit(64 * 1024)

		for {
			readTimeoutNS := atomic.LoadInt64(&client.readTimeoutNS)
			if err := client.setReadTimeout(readTimeoutNS); err != nil {
				client.onConnError(client.conn, err)
				return
			}

			// deal message
			mt, message, err := client.conn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					client.onConnError(client.conn, WrapSystemError(err))
				}
				return
			}
			switch mt {
			case websocket.BinaryMessage:
				client.onConnBinary(client.conn, message)
				break
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
func (p *WebSocketClient) SetReadTimeoutMS(readTimeoutMS uint64) *rpcError {
	timeoutNS := int64(readTimeoutMS) * int64(time.Millisecond)
	atomic.StoreInt64(&p.readTimeoutNS, timeoutNS)
	return p.setReadTimeout(timeoutNS)
}

func (p *WebSocketClient) setReadTimeout(readTimeoutNS int64) *rpcError {
	if p.conn == nil {
		return NewRPCError(
			"websocket-client: conn is nil",
		)
	}

	nextTimeoutNS := int64(0)
	nowNS := TimeNowNS()

	if readTimeoutNS > 0 {
		nextTimeoutNS = nowNS + int64(readTimeoutNS)
	} else {
		nextTimeoutNS = nowNS + 3600*int64(time.Second)
	}

	return WrapSystemError(p.conn.SetReadDeadline(
		time.Unix(nextTimeoutNS/int64(time.Second), nextTimeoutNS%int64(time.Second)),
	))
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
		p.onConnError(p.conn, err)
		return err
	}

	p.Lock()
	defer p.Unlock()
	err := p.conn.WriteMessage(messageType, data)

	if err != nil {
		ret := WrapSystemError(err)
		p.onConnError(p.conn, ret)
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
		p.onConnError(p.conn, err)
		return err
	}

	err := p.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)

	if err != nil {
		ret := WrapSystemError(err)
		p.onConnError(p.conn, ret)
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
		p.onConnError(p.conn, err)
		return err
	}
}

func (p *WebSocketClient) onConnOpen(conn *websocket.Conn) {
	fmt.Println("client onOpen", conn)
}

func (p *WebSocketClient) onConnError(conn *websocket.Conn, err *rpcError) {
	fmt.Println("client onError", conn, err)
}

func (p *WebSocketClient) onConnClose(conn *websocket.Conn) {
	fmt.Println("client onClose", conn)
}

func (p *WebSocketClient) onConnBinary(conn *websocket.Conn, bytes []byte) {
	fmt.Println("client onBinary", conn, bytes)
}
