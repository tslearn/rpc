package client

import (
	"crypto/tls"
	"fmt"
	"github.com/rpccloud/rpc/internal/adapter"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rpccloud/rpc/internal"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"github.com/rpccloud/rpc/internal/gateway"
)

const (
	clientStatusClosed     = int32(0)
	clientStatusPreClosing = int32(1)
	clientStatusClosing    = int32(2)
	clientStatusRunning    = int32(3)
)

// Client ...
type Client struct {
	sessionString        string
	adapter              *adapter.RunnableService
	conn                 *adapter.StreamConn
	preSendHead          *SendItem
	preSendTail          *SendItem
	channels             []Channel
	freeChannels         *FreeChannelStack
	lastTimeoutCheckTime time.Time
	lastControlSendTime  time.Time
}

func newClient(
	network string,
	addr string,
	tlsConfig *tls.Config,
	rBufSize int,
	wBufSize int,
) *Client {
	ret := &Client{
		status:        clientStatusRunning,
		sessionString: "",

		preSendHead:          nil,
		preSendTail:          nil,
		channels:             nil,
		freeChannels:         nil,
		lastTimeoutCheckTime: base.TimeNow(),
		lastControlSendTime:  base.TimeNow(),
	}

	adapter := adapter.NewClientAdapter(network, addr, tlsConfig, rBufSize, wBufSize, ret)

	if adapter == nil {
		return nil
	}

	go func() {
		for {
			switch atomic.LoadInt32(&ret.status) {
			case clientStatusRunning:
				time.Sleep(time.Second)
				adapter.Open()
				time.Sleep(time.Second)
			case clientStatusPreClosing:
				time.Sleep(20 * time.Millisecond)
			case clientStatusClosing:
				atomic.StoreInt32(&ret.status, clientStatusClosed)
				return
			}
		}
	}()

	go func() {
		for {
			switch atomic.LoadInt32(&ret.status) {
			case clientStatusRunning:
				now := base.TimeNow()
				ret.tryToTimeout(now)
				ret.tryToDeliverPreSendMessages()
				ret.tryToSendPing(now)
				time.Sleep(80 * time.Millisecond)
			case clientStatusPreClosing:
				adapter.Close()
				atomic.StoreInt32(&ret.status, clientStatusClosing)
				return
			default:
				panic("internal error")
			}
		}
	}()

	return ret
}

// Close ...
func (p *Client) Close() bool {
	if !atomic.CompareAndSwapInt32(
		&p.status,
		clientStatusRunning,
		clientStatusPreClosing,
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

func (p *Client) onError(err *base.Error) {
	fmt.Println("client onError: ", err)
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
	} else if backStream, err = conn.ReadStream(3*time.Second, 1024); err != nil {
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
			p.freeChannels = NewFreeChannelStack(int(config.NumOfChannels()))
			for i := 0; i < len(p.channels); i++ {
				p.channels[i].id = i
				p.channels[i].client = p
				p.channels[i].seq = uint64(config.NumOfChannels()) + uint64(i)
				p.channels[i].item = nil
				p.freeChannels.Push(i)
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

func (p *Client) setConn(conn internal.IStreamConn) {
	p.Lock()
	defer p.Unlock()
	p.conn = conn
}

func (p *Client) onCallbackStream(stream *core.Stream, callbackID uint64) {
	p.Lock()
	defer p.Unlock()

	if p.channels != nil {
		if chSize := uint64(len(p.channels)); chSize > 0 {
			p.channels[callbackID%chSize].OnCallbackStream(stream)
		}
	}
}

func (p *Client) onConnRun(conn internal.IStreamConn) {
	// init conn
	if err := p.initConn(conn); err != nil {
		p.onError(err)
		return
	}

	err := (*base.Error)(nil)
	p.setConn(conn)

	defer func() {
		p.setConn(nil)

		if err != nil {
			p.onError(err)
		}

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
			p.onCallbackStream(stream, callbackID)
		} else if kind, e := stream.ReadInt64(); e != nil {
			err = e
			return
		} else if kind == core.ControlStreamPong {
			// ignore
		} else {
			// broadcast message is not supported now
			err = errors.ErrStream
			return
		}
	}
}

func (p *Client) tryToSendPing(now time.Time) {
	p.Lock()
	defer p.Unlock()

	deltaTime := now.Sub(p.lastControlSendTime)

	if p.conn == nil {
		return
	} else if deltaTime < p.config.Heartbeat() {
		return
	} else {
		// Send Ping
		p.lastControlSendTime = now
		sendStream := core.NewStream()
		defer sendStream.Release()
		sendStream.SetCallbackID(0)
		sendStream.WriteInt64(core.ControlStreamPing)
		if err := p.conn.WriteStream(
			sendStream,
			p.config.WriteTimeout(),
		); err != nil {
			p.onError(err)
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
			if item.CheckAndTimeout(now) {
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

		// sweep the channels
		for i := 0; i < len(p.channels); i++ {
			p.channels[i].OnTimeout(now)
		}
	}
}

func (p *Client) tryToDeliverPreSendMessages() {
	p.Lock()
	defer p.Unlock()

	for p.tryToDeliverPreSendOneMessage() {
	}
}

func (p *Client) tryToDeliverPreSendOneMessage() bool {
	if atomic.LoadInt32(&p.status) != clientStatusRunning { // not running
		return false
	} else if p.conn == nil { // not connected
		return false
	} else if p.preSendHead == nil { // preSend queue is empty
		return false
	} else if p.channels == nil {
		return false
	} else if channelID, ok := p.freeChannels.Pop(); !ok {
		return false
	} else {
		channel := &p.channels[channelID]

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
		channel.item = item

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

	p.tryToDeliverPreSendMessages()

	// wait for response
	retStream := <-item.returnCH
	defer retStream.Release()
	return core.ParseResponseStream(retStream)
}
