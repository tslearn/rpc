package client

import (
	"crypto/tls"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
)

// Client ...
type Client struct {
	config         *Config
	sessionString  string
	adapter        *adapter.Adapter
	streamConn     *adapter.StreamConn
	preSendHead    *SendItem
	preSendTail    *SendItem
	channels       []Channel
	freeChannels   *FreeChannelStack
	lastCheckTime  time.Time
	lastActiveTime time.Time
	lastPingTime   time.Time
	orcManager     *base.ORCManager
	sync.Mutex
}

func newClient(
	network string,
	addr string,
	tlsConfig *tls.Config,
	rBufSize int,
	wBufSize int,
) *Client {
	ret := &Client{
		config:         &Config{},
		sessionString:  "",
		adapter:        nil,
		streamConn:     nil,
		preSendHead:    nil,
		preSendTail:    nil,
		channels:       nil,
		freeChannels:   nil,
		lastCheckTime:  base.TimeNow(),
		lastActiveTime: base.TimeNow(),
		orcManager:     base.NewORCManager(),
	}
	ret.config.rBufSize = rBufSize
	ret.config.wBufSize = wBufSize
	clientAdapter := adapter.NewClientAdapter(
		network, addr, tlsConfig, rBufSize, wBufSize, ret,
	)

	// Start the adapter
	clientAdapter.Open()
	go func() {
		clientAdapter.Run()
	}()

	ret.adapter = clientAdapter

	// Start the client (send the messages)
	ret.orcManager.Open(func() bool {
		return true
	})

	go func() {
		ret.orcManager.Run(func(isRunning func() bool) {
			for isRunning() {
				time.Sleep(80 * time.Millisecond)
				now := base.TimeNow()
				ret.Lock()
				ret.tryToTimeout(now)
				ret.tryToDeliverPreSendMessages()
				ret.tryToSendPing(now)
				ret.Unlock()
			}
		})
	}()

	return ret
}

// Close ...
func (p *Client) Close() bool {
	return p.orcManager.Close(func() {
		p.adapter.Close()
	}, func() {
		p.adapter = nil
	})
}

func (p *Client) onError(err *base.Error) {
	fmt.Println("client onError: ", err)
}

// OnConnOpen ...
func (p *Client) OnConnOpen(streamConn *adapter.StreamConn) {
	p.Lock()
	defer p.Unlock()

	stream := core.NewStream()
	stream.SetCallbackID(0)
	stream.WriteInt64(core.ControlStreamConnectRequest)
	stream.WriteString(p.sessionString)
	streamConn.WriteStreamAndRelease(stream)
}

// OnConnClose ...
func (p *Client) OnConnClose(_ *adapter.StreamConn) {
	p.Lock()
	defer p.Unlock()
	p.streamConn = nil
}

// OnConnReadStream ...
func (p *Client) OnConnReadStream(
	streamConn *adapter.StreamConn,
	stream *core.Stream,
) {
	p.Lock()
	defer p.Unlock()

	callbackID := stream.GetCallbackID()

	if p.streamConn == nil {
		p.streamConn = streamConn

		if callbackID != 0 {
			p.OnConnError(streamConn, errors.ErrStream)
		} else if kind, err := stream.ReadInt64(); err != nil {
			p.OnConnError(streamConn, err)
		} else if kind != core.ControlStreamConnectResponse {
			p.OnConnError(streamConn, errors.ErrStream)
		} else if sessionString, err := stream.ReadString(); err != nil {
			p.OnConnError(streamConn, err)
		} else if numOfChannels, err := stream.ReadInt64(); err != nil {
			p.OnConnError(streamConn, err)
		} else if transLimit, err := stream.ReadInt64(); err != nil {
			p.OnConnError(streamConn, err)
		} else if heartbeat, err := stream.ReadInt64(); err != nil {
			p.OnConnError(streamConn, err)
		} else if heartbeatTimeout, err := stream.ReadInt64(); err != nil {
			p.OnConnError(streamConn, err)
		} else if requestTimeout, err := stream.ReadInt64(); err != nil {
			p.OnConnError(streamConn, err)
		} else if requestInterval, err := stream.ReadInt64(); err != nil {
			p.OnConnError(streamConn, err)
		} else if !stream.IsReadFinish() {
			p.OnConnError(streamConn, errors.ErrStream)
		} else {
			if sessionString != p.sessionString {
				// new session
				p.sessionString = sessionString

				// update config
				p.config.numOfChannels = int(numOfChannels)
				p.config.transLimit = int(transLimit)
				p.config.heartbeat = time.Duration(heartbeat)
				p.config.heartbeatTimeout = time.Duration(heartbeatTimeout)
				p.config.requestTimeout = time.Duration(requestTimeout)
				p.config.requestInterval = time.Duration(requestInterval)

				numOfChannels := p.config.numOfChannels
				p.channels = make([]Channel, numOfChannels)
				p.freeChannels = NewFreeChannelStack(numOfChannels)
				for i := 0; i < len(p.channels); i++ {
					p.channels[i].id = i
					p.channels[i].client = p
					p.channels[i].seq = uint64(numOfChannels) + uint64(i)
					p.channels[i].item = nil
					p.freeChannels.Push(i)
				}
			} else {
				// config and channels have already initialized. so ignore this
			}

			now := base.TimeNow()
			p.lastActiveTime = now
			p.lastPingTime = now
		}

		stream.Release()
	} else {
		if callbackID == 0 {
			if kind, err := stream.ReadInt64(); err != nil {
				p.OnConnError(streamConn, err)
			} else if kind != core.ControlStreamPong {
				p.OnConnError(streamConn, errors.ErrStream)
			} else {
				p.lastActiveTime = base.TimeNow()
			}
			stream.Release()
		} else if p.channels != nil {
			channel := &p.channels[callbackID%uint64(len(p.channels))]
			channel.ReceiveStream(stream)
		} else {
			// ignore
			stream.Release()
		}
	}
}

// OnConnError ...
func (p *Client) OnConnError(streamConn *adapter.StreamConn, err *base.Error) {
	p.onError(err)
	streamConn.Close()
}

func (p *Client) tryToSendPing(now time.Time) {
	if p.streamConn == nil {
		return
	} else if now.Sub(p.lastPingTime) < p.config.heartbeat {
		return
	} else {
		// Send Ping
		p.lastPingTime = now
		stream := core.NewStream()
		stream.SetCallbackID(0)
		stream.WriteInt64(core.ControlStreamPing)
		p.streamConn.WriteStreamAndRelease(stream)
	}
}

func (p *Client) tryToTimeout(now time.Time) {
	if now.Sub(p.lastCheckTime) > 800*time.Millisecond {
		p.lastCheckTime = now

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
			(&p.channels[i]).OnTimeout(now)
		}
	}
}

func (p *Client) tryToDeliverPreSendMessages() {
	for {
		if p.streamConn == nil { // not running
			return
		} else if p.preSendHead == nil { // preSend queue is empty
			return
		} else if p.channels == nil {
			return
		} else if channelID, ok := p.freeChannels.Pop(); !ok {
			return
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

			p.streamConn.WriteStreamAndRelease(item.sendStream.Clone())
			item.sendTime = base.TimeNow()
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
