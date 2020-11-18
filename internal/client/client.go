package client

import (
	"fmt"
	"github.com/rpccloud/rpc/internal"
	"github.com/rpccloud/rpc/internal/adapter/websocket"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"github.com/rpccloud/rpc/internal/gateway"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	clientStatusClosed  = int32(0)
	clientStatusClosing = int32(0)
	clientStatusRunning = int32(1)
)

// Client ...
type Client struct {
	status               int32
	closeCH              chan bool
	connectString        string
	sessionString        string
	conn                 internal.IStreamConn
	preSendHead          *SendItem
	preSendTail          *SendItem
	channels             []Channel
	freeChannels         chan int
	config               gateway.SessionConfig
	lastTimeoutCheckTime time.Time
	sync.Mutex
}

func newClient(connectString string) (*Client, *base.Error) {
	adapter := (internal.IClientAdapter)(nil)

	if urlInfo, e := url.Parse(connectString); e != nil {
		return nil, errors.ErrClientConnectString.AddDebug(e.Error())
	} else if urlInfo.Scheme == "ws" || urlInfo.Scheme == "wss" {
		adapter = websocket.NewWebsocketClientAdapter(connectString)
	} else {
		return nil, errors.ErrClientConnectString.AddDebug(
			fmt.Sprintf("unsupported scheme %s", urlInfo.Scheme),
		)
	}

	ret := &Client{
		status:               clientStatusRunning,
		closeCH:              make(chan bool, 1),
		connectString:        connectString,
		sessionString:        "",
		conn:                 nil,
		preSendHead:          nil,
		preSendTail:          nil,
		config:               gateway.SessionConfig{},
		channels:             nil,
		freeChannels:         nil,
		lastTimeoutCheckTime: base.TimeNow(),
	}

	go func() {
		for atomic.LoadInt32(&ret.status) == clientStatusRunning {
			adapter.Open(ret.onConnRun, ret.onError)

			if atomic.LoadInt32(&ret.status) == clientStatusRunning {
				time.Sleep(time.Second)
			}
		}
	}()

	go func() {
		for atomic.LoadInt32(&ret.status) == clientStatusRunning {
			now := base.TimeNow()
			ret.tryToTimeout(now)
			for ret.tryToDeliverPreSendMessage() {
				// loop until failed
			}
			time.Sleep(100 * time.Millisecond)
		}

		atomic.StoreInt32(&ret.status, clientStatusClosed)
		ret.closeCH <- true
	}()

	return ret, nil
}

func (p *Client) onError(err *base.Error) {
	fmt.Println("client onError: ", err)
}

// Close ...
func (p *Client) Close() bool {
	if !atomic.CompareAndSwapInt32(
		&p.status,
		clientStatusRunning,
		clientStatusClosing,
	) {
		p.onError(errors.ErrClientNotRunning)
		return false
	}

	err := func() *base.Error {
		p.Lock()
		defer p.Unlock()

		if p.conn != nil {
			return p.conn.Close()
		} else {
			return nil
		}
	}()

	if err != nil {
		p.onError(errors.ErrClientNotRunning)
		return false
	}

	<-p.closeCH
	return true
}

func (p *Client) initConn(conn internal.IStreamConn) *base.Error {
	sendStream := core.NewStream()
	backStream := (*core.Stream)(nil)

	defer func() {
		sendStream.Release()
		if backStream != nil {
			backStream.Release()
		}
	}()

	sendStream.SetCallbackID(0)
	sendStream.WriteInt64(core.ControlStreamConnectRequest)
	sendStream.WriteString(p.sessionString)

	if err := conn.WriteStream(sendStream, 3*time.Second); err != nil {
		return err
	} else if backStream, err = conn.ReadStream(3*time.Second, 0); err != nil {
		return err
	} else if backStream.GetCallbackID() != 0 {
		return errors.ErrStream
	} else if kind, err := backStream.ReadInt64(); err != nil {
		return err
	} else if kind != core.ControlStreamConnectResponse {
		return errors.ErrStream
	} else if sessionString, err := backStream.ReadString(); err != nil {
		return err
	} else if config, err := gateway.ReadSessionConfig(backStream); err != nil {
		return err
	} else if !backStream.IsReadFinish() {
		return errors.ErrStream
	} else {
		p.Lock()
		defer p.Unlock()

		if sessionString != p.sessionString {
			// new session
			p.sessionString = sessionString
			p.config = config
			p.channels = make([]Channel, config.NumOfChannels())
			p.freeChannels = make(chan int, config.NumOfChannels())
			for i := 0; i < len(p.channels); i++ {
				p.channels[i].seq = uint64(config.NumOfChannels()) + uint64(i)
				p.channels[i].item = nil
				p.freeChannels <- i
			}
		} else if !p.config.Equals(&config) {
			// old session, but config changes
			p.sessionString = ""
			return errors.ErrClientConfigChanges
		} else {
			// old session
		}

		return nil
	}
}

func (p *Client) onConnRun(conn internal.IStreamConn) {
	// init conn
	if err := p.initConn(conn); err != nil {
		p.onError(err)
		return
	}

	// set the conn
	p.Lock()
	p.conn = conn
	p.Unlock()

	err := (*base.Error)(nil)

	// clear the conn when finish
	defer func() {
		if err != nil {
			p.onError(err)
		}

		p.Lock()
		p.conn = nil
		p.Unlock()

		if err := conn.Close(); err != nil {
			p.onError(err)
		}
	}()

	// receive messages
	for atomic.LoadInt32(&p.status) == clientStatusRunning {
		if stream, e := conn.ReadStream(
			p.config.ReadTimeout(),
			p.config.TransLimit(),
		); e != nil {
			if e != errors.ErrStreamConnIsClosed {
				err = e
			}
			return
		} else if callbackID := stream.GetCallbackID(); callbackID > 0 {
			p.Lock()
			// TODO
			// if v, ok := p.sendMap[callbackID]; ok {
			//  if v.Return(stream) {
			//    delete(p.sendMap, callbackID)
			//  }
			//}
			p.Unlock()
		} else {
			// broadcast message is not supported now
			err = errors.ErrStream
			return
		}
	}
}

func (p *Client) tryToTimeout(now time.Time) {
	p.Lock()
	defer p.Unlock()

	if now.Sub(p.lastTimeoutCheckTime) > 800*time.Millisecond {
		p.lastTimeoutCheckTime = now

		// sweep pre send list
		preValidItem := (*SendItem)(nil)
		item := p.preSendHead
		for item != nil {
			if now.Sub(item.startTime) > item.timeout && item.Timeout() {
				nextItem := item.next

				if preValidItem == nil {
					p.preSendHead = nextItem
				} else {
					preValidItem.next = nextItem
				}

				if item == p.preSendTail {
					p.preSendTail = preValidItem
					if p.preSendTail != nil {
						p.preSendTail.next = nil
					}
				}
				item = nextItem
			} else {
				preValidItem = item
				item = item.next
			}
		}

		// TODO
		// sweep send map
		//for key, value := range p.sendMap {
		//  if now.Sub(value.startTime) > value.timeout && value.Timeout() {
		//    delete(p.sendMap, key)
		//  }
		//}
	}
}

func (p *Client) tryToDeliverPreSendMessage() bool {
	p.Lock()
	defer p.Unlock()

	if atomic.LoadInt32(&p.status) != clientStatusRunning { // not running
		return false
	} else if p.conn == nil { // not connected
		return false
	} else if p.preSendHead == nil { // preSend queue is empty
		return false
	} else if p.channels == nil {
		return false
	} else {
		select {
		case channelID := <-p.freeChannels:
			channel := p.channels[channelID]

			// get and set the send item
			item := p.preSendHead
			if item == p.preSendTail {
				p.preSendHead = nil
				p.preSendTail = nil
			} else {
				p.preSendHead = p.preSendHead.next
			}

			item.id = atomic.LoadUint64(&channel.seq)
			item.next = nil
			item.sendStream.SetCallbackID(item.id)
			atomic.StorePointer(&channel.item, (unsafe.Pointer)(item))

			// try to send
			if err := p.conn.WriteStream(
				item.sendStream,
				p.config.WriteTimeout(),
			); err != nil {
				p.onError(err)
				return false
			}

			item.sendTime = base.TimeNow()
			return true
		default:
			return false
		}
	}
}

// SendMessage ...
func (p *Client) SendMessage(
	timeout time.Duration,
	target string,
	args ...interface{},
) (interface{}, *base.Error) {
	item := newSendItem()
	defer item.Release()

	item.startTime = base.TimeNow()
	item.timeout = timeout

	// set depth
	item.sendStream.SetDepth(0)
	// write target
	item.sendStream.WriteString(target)
	// write from
	item.sendStream.WriteString("@")
	// write args
	for i := 0; i < len(args); i++ {
		if reason := item.sendStream.Write(args[i]); reason != core.StreamWriteOK {
			return nil, errors.ErrUnsupportedValue.AddDebug(reason)
		}
	}

	// add item to the list tail
	p.Lock()
	if p.preSendTail == nil {
		p.preSendHead = item
		p.preSendTail = item
	} else {
		p.preSendTail.next = item
		p.preSendTail = item
	}
	p.Unlock()

	// wait for response
	return core.ParseResponseStream(<-item.returnCH)
}
