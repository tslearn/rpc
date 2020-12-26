package gateway

import (
	"fmt"
	"sync"

	"github.com/rpccloud/rpc/internal/adapter/common"
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
	id       uint64
	security string
	conn     *common.StreamConn
	gateway  *GateWay
	channels []Channel
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
	ret.gateway = gateway
	ret.channels = make([]Channel, gateway.GetConfig().numOfChannels)
	return ret
}

// Initialized ...
func (p *Session) Initialized(conn *common.StreamConn) {
	p.Lock()
	p.conn = conn
	p.conn.SetReceiver(p)
	p.Unlock()

	config := p.gateway.config
	stream := core.NewStream()
	stream.WriteInt64(core.ControlStreamConnectResponse)
	stream.WriteString(fmt.Sprintf("%d-%s", p.id, p.security))
	stream.Write(config.numOfChannels)
	stream.Write(config.transLimit)
	stream.Write(int64(config.heartbeat))
	stream.Write(int64(config.timeoutRW))
	stream.Write(int64(config.clientRequestTimeout))
	stream.Write(int64(config.clientRequestInterval))
	p.conn.WriteStreamAndRelease(stream)
}

// StreamIn ...
func (p *Session) StreamIn(stream *core.Stream) {
	p.Lock()
	defer p.Unlock()

	keepStream := false

	defer func() {
		if !keepStream {
			stream.Release()
		}
	}()

	cbID := stream.GetCallbackID()
	config := p.gateway.GetConfig()

	if cbID > 0 {
		stream.SetSessionID(p.id)
		channel := p.channels[cbID%uint64(len(p.channels))]
		if accepted, backStream := channel.In(cbID); !accepted {
			// ignore accepted
		} else if backStream != nil {
			p.conn.WriteStreamAndRelease(backStream.Clone())
			return
		} else {
			keepStream = true
			if err := p.gateway.slot.SendStream(stream); err != nil {

			}
			return
		}
	} else if kind, err := stream.ReadInt64(); err != nil {
		return err
	} else if kind == core.ControlStreamPing {
		p.checkTimeout()
		// Send Pong
		stream.SetWritePosToBodyStart()
		stream.WriteInt64(core.ControlStreamPong)
		return p.conn.WriteStream(stream, config.writeTimeout)
	} else {
		return errors.ErrStream
	}
}

// StreamOut ...
func (p *Session) StreamOut(stream *core.Stream) *base.Error {
	p.Lock()
	defer p.Unlock()
	cbID := stream.GetCallbackID()

	// record stream
	if cbID > 0 {
		if ok := p.channels[cbID%uint64(len(p.channels))].Out(stream); !ok {
			return nil
		}
	}

	// write stream
	return p.conn.WriteStream(stream, p.gateway.GetConfig().writeTimeout)
}

func (p *Session) checkTimeout() {
	nowNS := base.TimeNow().UnixNano()
	for i := 0; i < len(p.channels); i++ {
		p.channels[i].Timeout(nowNS, int64(p.gateway.GetConfig().cacheTimeout))
	}
}

// Release ...
func (p *Session) Release() {
	p.Lock()
	defer p.Unlock()
	for i := 0; i < len(p.channels); i++ {
		p.channels[i].Clean()
	}
	sessionCache.Put(p)
}

// OnConnOpen ...
func (p *Session) OnConnOpen(streamConn *common.StreamConn) {
	// Route to gateway
	p.gateway.OnConnOpen(streamConn)
}

// OnConnError ...
func (p *Session) OnConnError(streamConn *common.StreamConn, err *base.Error) {
	// Route to gateway
	p.gateway.OnConnError(streamConn, err)
}

// OnConnClose ...
func (p *Session) OnConnClose(streamConn *common.StreamConn) {

}

// OnConnReadStream ...
func (p *Session) OnConnReadStream(
	streamConn *common.StreamConn,
	stream *core.Stream,
) {

}
