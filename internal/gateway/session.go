package gateway

import (
	"github.com/rpccloud/rpc/internal"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"sync"
)

var sessionCache = &sync.Pool{
	New: func() interface{} {
		return &Session{}
	},
}

type Session struct {
	id       uint64
	security string
	conn     internal.IStreamConn
	gateway  *GateWay
	channels []Channel
	sync.Mutex
}

func newSession(id uint64, gateway *GateWay) *Session {
	ret := sessionCache.Get().(*Session)
	ret.id = id
	ret.security = base.GetRandString(32)
	ret.conn = nil
	ret.gateway = gateway
	ret.channels = make([]Channel, gateway.config.NumOfChannels())
	return ret
}

func (p *Session) SetConn(conn internal.IStreamConn) {
	p.Lock()
	defer p.Unlock()
	p.conn = conn
}

func (p *Session) StreamIn(stream *core.Stream) *base.Error {
	p.Lock()
	defer p.Unlock()

	cbID := stream.GetCallbackID()

	if cbID > 0 {
		stream.SetSessionID(p.id)
		channel := p.channels[cbID%uint64(len(p.channels))]
		if retStream, err := channel.In(cbID, uint64(len(p.channels))); err != nil {
			return err
		} else if retStream != nil {
			return p.conn.WriteStream(stream, p.gateway.config.WriteTimeout())
		} else {
			return p.gateway.slot.SendStream(stream)
		}
	} else {
		// control message
		return nil
	}
}

func (p *Session) StreamOut(stream *core.Stream) *base.Error {
	p.Lock()
	defer p.Unlock()
	cbID := stream.GetCallbackID()

	// record stream
	if cbID > 0 {
		channel := p.channels[cbID%uint64(len(p.channels))]
		if err := channel.Out(stream); err != nil {
			return err
		}
	}

	// write stream
	return p.conn.WriteStream(stream, p.gateway.config.WriteTimeout())
}

func (p *Session) checkTimeout() {
	nowNS := base.TimeNow().UnixNano()
	for i := 0; i < len(p.channels); i++ {
		p.channels[i].Timeout(nowNS, int64(p.gateway.config.cacheTimeout))
	}
}

func (p *Session) Release() {
	p.Lock()
	defer p.Unlock()
	for i := 0; i < len(p.channels); i++ {
		p.channels[i].Clean()
	}
	sessionCache.Put(p)
}
