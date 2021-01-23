package gateway

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/adapter"
	"sync"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
)

var sessionCache = &sync.Pool{
	New: func() interface{} {
		return &Session{}
	},
}

// Session ...
type Session struct {
	id         uint64
	security   string
	conn       *adapter.StreamConn
	idleTimeNS int64
	gateway    *GateWay
	channels   []Channel
	prev       *Session
	next       *Session
	sync.Mutex
}

func newSession(
	id uint64,
	gateway *GateWay,
) *Session {
	ret := sessionCache.Get().(*Session)
	ret.id = id
	ret.security = base.GetRandString(32)
	ret.conn = nil
	ret.idleTimeNS = base.TimeNow().UnixNano()
	ret.gateway = gateway
	ret.channels = make([]Channel, gateway.GetConfig().numOfChannels)
	return ret
}

func (p *Session) Close() {
	p.Lock()
	defer p.Unlock()

	if p.conn != nil {
		p.conn.Close()
	}

	if p.channels != nil {
		for i := 0; i < len(p.channels); i++ {
			p.channels[i].Clean()
		}
	}
}

func (p *Session) TimeCheck(nowNS int64) {
	p.Lock()
	defer p.Unlock()

	if p.conn != nil {
		// conn timeout
		if !p.conn.IsActive(nowNS, p.gateway.config.heartbeatTimeout) {
			p.conn.Close()
		}
	} else {
		// session timeout
		if nowNS-p.idleTimeNS > int64(p.gateway.config.serverSessionTimeout) {
			p.Close()
			p.gateway.Remove(p.id)
		}
	}

	// channel timeout
	for i := 0; i < len(p.channels); i++ {
		p.channels[i].TimeCheck(
			nowNS,
			int64(p.gateway.config.serverCacheTimeout),
		)
	}
}

// WriteStreamAndRelease ...
func (p *Session) OutStream(stream *core.Stream) {
	p.Lock()
	defer p.Unlock()

	// record stream
	channel := &p.channels[stream.GetCallbackID()%uint64(len(p.channels))]
	if channel.Out(stream) {
		p.conn.WriteStreamAndRelease(stream.Clone())
	} else {
		stream.Release()
	}
}

// Release ...
func (p *Session) Release() {
	p.Lock()
	defer p.Unlock()
	for i := 0; i < len(p.channels); i++ {
		(&p.channels[i]).Clean()
	}
	sessionCache.Put(p)
}

// OnConnOpen ...
func (p *Session) OnConnOpen(streamConn *adapter.StreamConn) {
	p.Lock()
	p.conn = streamConn
	p.conn.SetReceiver(p)
	p.Unlock()

	config := p.gateway.config
	stream := core.NewStream()
	stream.WriteInt64(core.ControlStreamConnectResponse)
	stream.WriteString(fmt.Sprintf("%d-%s", p.id, p.security))
	stream.Write(config.numOfChannels)
	stream.Write(config.transLimit)
	stream.Write(int64(config.heartbeat))
	stream.Write(int64(config.heartbeatTimeout))
	stream.Write(int64(config.clientRequestInterval))
	p.conn.WriteStreamAndRelease(stream)
}

// OnConnError ...
func (p *Session) OnConnError(streamConn *adapter.StreamConn, err *base.Error) {
	// Route to gateway
	p.gateway.OnConnError(streamConn, err)
}

// OnConnClose ...
func (p *Session) OnConnClose(_ *adapter.StreamConn) {
	//p.gateway.OnConnClose(streamConn)
	p.Lock()
	p.conn = nil
	p.idleTimeNS = base.TimeNow().UnixNano()
	p.Unlock()
}

// OnConnReadStream ...
func (p *Session) OnConnReadStream(
	streamConn *adapter.StreamConn,
	stream *core.Stream,
) {
	p.Lock()
	defer p.Unlock()

	keepStream := false

	defer func() {
		if !keepStream {
			stream.Release()
		}
	}()

	cbID := stream.GetCallbackID()

	if cbID > 0 {
		stream.SetGatewayID(p.gateway.id)
		stream.SetSessionID(p.id)
		channel := &p.channels[cbID%uint64(len(p.channels))]
		if accepted, backStream := channel.In(cbID); accepted {
			keepStream = true
			if err := p.gateway.routerSender.SendStreamToRouter(stream); err != nil {
				p.OnConnError(streamConn, err)
			}
		} else if backStream != nil {
			// do not release the backStream, so we need to clone it
			p.conn.WriteStreamAndRelease(backStream.Clone())
		} else {
			// ignore
		}
	} else if kind, err := stream.ReadInt64(); err != nil {
		p.OnConnError(streamConn, err)
	} else if kind == core.ControlStreamPing {
		keepStream = true
		stream.SetWritePosToBodyStart()
		stream.WriteInt64(core.ControlStreamPong)
		p.conn.WriteStreamAndRelease(stream)
	} else {
		p.OnConnError(streamConn, errors.ErrStream)
	}
}
