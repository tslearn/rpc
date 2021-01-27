package client

import (
	"crypto/tls"
	"fmt"
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
	rBufSize         int
	wBufSize         int
	heartbeat        time.Duration
	heartbeatTimeout time.Duration
	requestInterval  time.Duration
}

// Client ...
type Client struct {
	config          *Config
	sessionString   string
	adapter         *adapter.Adapter
	conn            *adapter.StreamConn
	preSendHead     *SendItem
	preSendTail     *SendItem
	channels        []Channel
	lastCheckTimeNS int64
	lastPingTimeNS  int64
	orcManager      *base.ORCManager
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
		config:          &Config{},
		sessionString:   "",
		adapter:         nil,
		conn:            nil,
		preSendHead:     nil,
		preSendTail:     nil,
		channels:        nil,
		lastCheckTimeNS: base.TimeNow().UnixNano(),
		orcManager:      base.NewORCManager(),
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
		ret.orcManager.Run(func(isRunning func() bool) bool {
			for isRunning() {
				time.Sleep(80 * time.Millisecond)
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

// Close ...
func (p *Client) Close() bool {
	return p.orcManager.Close(func() bool {
		return p.adapter.Close()
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
	p.conn = nil
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
				p.config.requestInterval = time.Duration(requestInterval)

				p.channels = make([]Channel, numOfChannels)
				for i := 0; i < len(p.channels); i++ {
					(&p.channels[i]).sequence = uint64(i)
					(&p.channels[i]).item = nil
				}
			} else {
				// config and channels have already initialized. so ignore this
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
	p.onError(err)
	if streamConn != nil {
		streamConn.Close()
	}
}

func (p *Client) tryToSendPing(nowNS int64) {
	if p.conn == nil {
		return
	} else if nowNS-p.lastPingTimeNS < int64(p.config.heartbeat) {
		return
	} else {
		// Send Ping
		p.lastPingTimeNS = nowNS
		stream := core.NewStream()
		stream.SetCallbackID(0)
		stream.WriteInt64(core.ControlStreamPing)
		p.conn.WriteStreamAndRelease(stream)
	}
}

func (p *Client) tryToTimeout(nowNS int64) {
	if nowNS-p.lastCheckTimeNS > 800*int64(time.Millisecond) {
		p.lastCheckTimeNS = nowNS

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
			if p.conn.IsActive(nowNS, p.config.heartbeatTimeout) {
				p.conn.Close()
			}
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
	item := newSendItem()
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
