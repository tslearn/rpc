package gateway

import (
	"github.com/rpccloud/rpc/internal"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
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
	config   *SessionConfig
	slot     internal.IStreamRouterSlot
	channels []Channel
	sync.Mutex
}

func newSession(
	id uint64,
	config *SessionConfig,
	slot internal.IStreamRouterSlot,
) *Session {
	ret := sessionCache.Get().(*Session)
	ret.id = id
	ret.security = base.GetRandString(32)
	ret.conn = nil
	ret.config = config
	ret.slot = slot
	ret.channels = make([]Channel, config.NumOfChannels())
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

	keepStream := false

	defer func() {
		if !keepStream {
			stream.Release()
		}
	}()

	cbID := stream.GetCallbackID()

	if cbID > 0 {
		stream.SetSessionID(p.id)
		if ok, retStream := p.channels[cbID%uint64(len(p.channels))].In(cbID); !ok {
			return nil
		} else if retStream != nil {
			return p.conn.WriteStream(retStream, p.config.WriteTimeout())
		} else {
			keepStream = true
			return p.slot.SendStream(stream)
		}
	} else if kind, err := stream.ReadInt64(); err != nil {
		return err
	} else if kind == core.ControlStreamPing {
		// Send Pong
		stream.SetWritePosToBodyStart()
		stream.WriteInt64(core.ControlStreamPong)
		return p.conn.WriteStream(stream, p.config.WriteTimeout())
	} else {
		return errors.ErrStream
	}
}

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
	return p.conn.WriteStream(stream, p.config.WriteTimeout())
}

func (p *Session) checkTimeout() {
	nowNS := base.TimeNow().UnixNano()
	for i := 0; i < len(p.channels); i++ {
		p.channels[i].Timeout(nowNS, int64(p.config.cacheTimeout))
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
