package client

import (
	"crypto/tls"
	"sync"
	"time"

	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
)

// Config ...
type Config struct {
	numOfChannels    int
	transLimit       int
	heartbeat        time.Duration
	heartbeatTimeout time.Duration
}

// Client ...
type Client struct {
	config         *Config
	sessionString  string
	adapter        *adapter.Adapter
	conn           *adapter.StreamConn
	preSendHead    *SendItem
	preSendTail    *SendItem
	channels       []Channel
	lastPingTimeNS int64
	orcManager     *base.ORCManager
	onError        func(err *base.Error)
	sync.Mutex
}

func newClient(
	network string,
	addr string,
	tlsConfig *tls.Config,
	rBufSize int,
	wBufSize int,
	onError func(err *base.Error),
) *Client {
	ret := &Client{
		config:        &Config{},
		sessionString: "",
		adapter:       nil,
		conn:          nil,
		preSendHead:   nil,
		preSendTail:   nil,
		channels:      nil,
		orcManager:    base.NewORCManager(),
		onError:       onError,
	}

	// init adapter
	clientAdapter := adapter.NewClientAdapter(
		network, addr, tlsConfig, rBufSize, wBufSize, ret,
	)
	clientAdapter.Open()
	go func() {
		clientAdapter.Run()
	}()
	ret.adapter = clientAdapter

	ret.orcManager.Open(func() bool {
		return true
	})

	go func() {
		ret.orcManager.Run(func(isRunning func() bool) bool {
			for isRunning() {
				time.Sleep(time.Second)
				nowNS := base.TimeNow().UnixNano()
				ret.Lock()
				ret.tryToTimeout(nowNS)
				ret.tryToDeliverPreSendMessages()
				ret.tryToSendPing(nowNS)
				ret.Unlock()
			}

			return true
		})
	}()

	return ret
}

func (p *Client) tryToSendPing(nowNS int64) {
	if p.conn == nil || nowNS-p.lastPingTimeNS < int64(p.config.heartbeat) {
		return
	}

	// Send Ping
	p.lastPingTimeNS = nowNS
	stream := core.NewStream()
	stream.SetCallbackID(0)
	stream.WriteInt64(core.ControlStreamPing)
	p.conn.WriteStreamAndRelease(stream)
}

func (p *Client) tryToTimeout(nowNS int64) {
	// sweep pre send list
	preValidItem := (*SendItem)(nil)
	item := p.preSendHead
	for item != nil {
		if item.CheckTime(nowNS) {
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

			item.next = nil
			item = nextItem
		} else {
			preValidItem = item
			item = item.next
		}
	}

	// sweep the channels
	for i := 0; i < len(p.channels); i++ {
		(&p.channels[i]).CheckTime(nowNS)
	}

	// check conn timeout
	if p.conn != nil {
		if !p.conn.IsActive(nowNS, p.config.heartbeatTimeout) {
			p.conn.Close()
		}
	}
}

func (p *Client) tryToDeliverPreSendMessages() {
	if p.conn == nil || p.channels == nil {
		return
	}

	findFree := 0
	channelSize := len(p.channels)

	for findFree < channelSize && p.preSendHead != nil {
		// find a free channel
		for p.channels[findFree].item != nil {
			findFree++
		}

		if findFree < channelSize {
			// remove sendItem from linked list
			item := p.preSendHead
			if item == p.preSendTail {
				p.preSendHead = nil
				p.preSendTail = nil
			} else {
				p.preSendHead = p.preSendHead.next
			}
			item.next = nil

			(&p.channels[findFree]).Use(item, len(p.channels))
			p.conn.WriteStreamAndRelease(item.sendStream.Clone())
		}
	}
}

// SendMessage ...
func (p *Client) SendMessage(
	timeout time.Duration,
	target string,
	args ...interface{},
) (interface{}, *base.Error) {
	item := NewSendItem()
	defer item.Release()

	item.timeoutNS = int64(timeout)

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

	p.tryToDeliverPreSendMessages()
	p.Unlock()

	// wait for response
	retStream := <-item.returnCH
	defer retStream.Release()

	return core.ParseResponseStream(retStream)
}

// Close ...
func (p *Client) Close() bool {
	return p.orcManager.Close(func() bool {
		return p.adapter.Close()
	}, func() {
		p.adapter = nil
	})
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

// OnConnReadStream ...
func (p *Client) OnConnReadStream(
	streamConn *adapter.StreamConn,
	stream *core.Stream,
) {
	p.Lock()
	defer p.Unlock()

	callbackID := stream.GetCallbackID()

	if p.conn == nil {
		p.conn = streamConn

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

				p.channels = make([]Channel, numOfChannels)
				for i := 0; i < len(p.channels); i++ {
					(&p.channels[i]).sequence = uint64(i)
					(&p.channels[i]).item = nil
				}
			} else {
				// try to resend channel message
				for i := 0; i < len(p.channels); i++ {
					if item := (&p.channels[i]).item; item != nil {
						p.conn.WriteStreamAndRelease(item.sendStream.Clone())
					}
				}
			}

			p.lastPingTimeNS = base.TimeNow().UnixNano()
		}

		stream.Release()
	} else {
		if callbackID == 0 {
			if kind, err := stream.ReadInt64(); err != nil {
				p.OnConnError(streamConn, err)
			} else if kind != core.ControlStreamPong {
				p.OnConnError(streamConn, errors.ErrStream)
			} else {
				// ignore
			}
			stream.Release()
		} else if p.channels != nil {
			channel := &p.channels[callbackID%uint64(len(p.channels))]
			channel.Free(stream)
		} else {
			// ignore
			stream.Release()
		}
	}
}

// OnConnError ...
func (p *Client) OnConnError(streamConn *adapter.StreamConn, err *base.Error) {
	if p.onError != nil {
		p.onError(err)
	}

	if streamConn != nil {
		streamConn.Close()
	}
}

// OnConnClose ...
func (p *Client) OnConnClose(_ *adapter.StreamConn) {
	p.Lock()
	defer p.Unlock()
	p.conn = nil
}
