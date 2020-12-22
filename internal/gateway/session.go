package gateway

import (
	"sync"

	"github.com/rpccloud/rpc/internal"
	"github.com/rpccloud/rpc/internal/adapter"
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
	conn     internal.IStreamConn
	gateway  *GateWay
	slot     internal.IStreamRouterSlot
	channels []Channel
	sync.Mutex
}

func newSession(
	id uint64,
	gateway *GateWay,
	slot internal.IStreamRouterSlot,
) *Session {
	ret := sessionCache.Get().(*Session)
	ret.id = id
	ret.security = base.GetRandString(32)
	ret.conn = nil
	ret.gateway = gateway
	ret.slot = slot
	ret.channels = make([]Channel, gateway.GetConfig().NumOfChannels())
	return ret
}

// SetConn ...
func (p *Session) SetConn(conn internal.IStreamConn) {
	p.Lock()
	defer p.Unlock()
	p.conn = conn
}

// StreamIn ...
func (p *Session) StreamIn(stream *core.Stream) *base.Error {
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
		if ok, retStream := p.channels[cbID%uint64(len(p.channels))].In(cbID); !ok {
			return nil
		} else if retStream != nil {
			return p.conn.WriteStream(retStream, config.WriteTimeout())
		} else {
			keepStream = true
			return p.slot.SendStream(stream)
		}
	} else if kind, err := stream.ReadInt64(); err != nil {
		return err
	} else if kind == core.ControlStreamPing {
		p.checkTimeout()
		// Send Pong
		stream.SetWritePosToBodyStart()
		stream.WriteInt64(core.ControlStreamPong)
		return p.conn.WriteStream(stream, config.WriteTimeout())
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
	return p.conn.WriteStream(stream, p.gateway.GetConfig().WriteTimeout())
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
func (p *Session) OnConnOpen(streamConn *adapter.StreamConn) {
	p.gateway.OnConnOpen(streamConn)
}

// OnConnClose ...
func (p *Session) OnConnClose(streamConn *adapter.StreamConn) {
	// p.gateway.OnConnClose(streamConn)
}

// OnConnReadStream ...
func (p *Session) OnConnReadStream(streamConn *adapter.StreamConn, stream *core.Stream) {

}

// OnConnError ...
func (p *Session) OnConnError(streamConn *adapter.StreamConn, err *base.Error) {
	p.gateway.onError(p.id, err)
}
