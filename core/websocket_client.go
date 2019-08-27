package core

import (
	"encoding/binary"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
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

type websocketClientCallback struct {
	id     uint32
	timeNS int64
	ch     chan bool
	stream *rpcStream
}

// WebSocketClient is implement of INetClient via web socket
type WebSocketClient struct {
	conn       *websocket.Conn
	closeChan  chan bool
	readyState int64
	sync.Map
	seed uint32
	sync.Mutex
}

// NewWebSocketClient create a WebSocketClient, and connect to url
func NewWebSocketClient(
	url string,
	timeoutMS uint64,
	readSizeLimit uint64,
) *WebSocketClient {
	client := &WebSocketClient{
		conn:       nil,
		closeChan:  make(chan bool, 1),
		readyState: wsClientConnecting,
		seed:       1,
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
			if err := conn.Close(); err != nil {
				client.onError(conn, NewRPCErrorByError(err))
			}
			atomic.StoreInt64(&client.readyState, wsClientClosed)
			client.onClose(conn)
			client.closeChan <- true
		}()

		conn.SetReadLimit(int64(readSizeLimit))
		timeoutNS := int64(timeoutMS) * int64(time.Millisecond)

		for {
			// set next read dead line
			nextTimeoutNS := TimeNowNS() + timeoutNS
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
				client.onError(conn, NewRPCError(
					"WebSocketClient: unknown message type",
				))
				return
			}
		}
	}()

	return client
}

// IsOpen returns true when the WebSocketClient is linked, otherwise false
func (p *WebSocketClient) IsOpen() bool {
	return atomic.LoadInt64(&p.readyState) == wsClientOpen
}

// SendBinary send byte array to the remote server
func (p *WebSocketClient) send(data []byte) *rpcError {
	if atomic.LoadInt64(&p.readyState) != wsClientOpen {
		err := NewRPCError("WebSocketClient: connection is not opened")
		p.onError(p.conn, err)
		return err
	}

	p.Lock()
	err := p.conn.WriteMessage(websocket.BinaryMessage, data)
	p.Unlock()

	if err != nil {
		ret := NewRPCErrorByError(err)
		p.onError(p.conn, ret)
		return ret
	}

	return nil
}

func (p *WebSocketClient) registerCallback() *websocketClientCallback {
	ret := (*websocketClientCallback)(nil)
	p.Lock()
	for {
		p.seed += 1
		if p.seed == math.MaxUint32 {
			p.seed = 1
		}
		if _, ok := p.Load(p.seed); !ok {
			ret = &websocketClientCallback{
				id:     p.seed,
				timeNS: TimeNowNS(),
				ch:     make(chan bool),
				stream: newRPCStream(),
			}
			p.Store(ret.id, ret)
			break
		}
	}
	p.Unlock()
	return ret
}

func (p *WebSocketClient) unregisterCallback(key uint32) bool {
	if _, ok := p.Load(key); ok {
		p.Delete(key)
		return true
	} else {
		return false
	}
}

func (p *WebSocketClient) getCallbackByID(key uint32) *websocketClientCallback {
	if v, ok := p.Load(key); ok {
		return v.(*websocketClientCallback)
	} else {
		return nil
	}
}

func (p *WebSocketClient) SendMessage(
	target string,
	args ...interface{},
) (interface{}, *rpcError) {
	callback := p.registerCallback()
	defer p.unregisterCallback(callback.id)

	stream := callback.stream
	// set client callback id
	stream.setClientCallbackID(p.seed)
	// write target
	stream.WriteString(target)
	// write depth
	stream.WriteUint64(0)
	// write from
	stream.WriteString("@")

	for i := 0; i < len(args); i++ {
		if stream.Write(args[i]) != RPCStreamWriteOK {
			return nil, NewRPCError("args not supported")
		}
	}

	if err := p.send(stream.getBufferUnsafe()); err != nil {
		return nil, err
	}

	if response := <-callback.ch; !response {
		return nil, NewRPCError("timeout")
	}

	success, ok := stream.ReadBool()
	if !ok {
		return nil, NewRPCError("data format error")
	}

	if success {
		if ret, ok := stream.Read(); ok {
			return ret, nil
		} else {
			return nil, NewRPCError("data format error")
		}
	} else {
		message, ok := stream.ReadString()
		if !ok {
			return nil, NewRPCError("data format error")
		}
		debug, ok := stream.ReadString()
		if !ok {
			return nil, NewRPCError("data format error")
		}
		return nil, NewRPCErrorByDebug(message, debug)
	}
}

// Close close the WebSocketClient
func (p *WebSocketClient) Close() *rpcError {
	p.Lock()
	defer p.Unlock()

	if atomic.LoadInt64(&p.readyState) != wsClientOpen {
		err := NewRPCError(
			"WebSocketClient: connection is not opened",
		)
		p.onError(p.conn, err)
		return err
	}

	if err := p.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	); err != nil {
		ret := NewRPCErrorByError(err)
		p.onError(p.conn, ret)
		return ret
	}

	atomic.StoreInt64(&p.readyState, wsClientClosing)

	select {
	case <-p.closeChan:
		return nil
	case <-time.After(2000 * time.Millisecond):
		err := NewRPCError(
			"WebSocketClient: close timeout",
		)
		p.onError(p.conn, err)
		return err
	}
}

func (p *WebSocketClient) onOpen(conn *websocket.Conn) {
	//fmt.Println("client onOpen", conn)
}

func (p *WebSocketClient) onError(conn *websocket.Conn, err *rpcError) {
	//fmt.Println("client onError", conn, err)
}

func (p *WebSocketClient) onClose(conn *websocket.Conn) {
	//fmt.Println("client onClose", conn)
}

func (p *WebSocketClient) onBinary(conn *websocket.Conn, bytes []byte) {
	if len(bytes) > 5 {
		b := bytes[1:5]
		clientID := binary.LittleEndian.Uint32(b)
		if cbItem := p.getCallbackByID(clientID); cbItem != nil {
			stream := cbItem.stream
			stream.setWritePosUnsafe(0)
			stream.putBytes(bytes)
			cbItem.ch <- true
		}
	}
}
