package gateway

import (
	"github.com/rpccloud/rpc/internal"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"sync"
	"sync/atomic"
	"time"
)

var sessionCache = &sync.Pool{
	New: func() interface{} {
		return &Session{}
	},
}

type SessionIDGenerator interface {
	GetID() (uint64, *base.Error)
}

type SingleGenerator struct {
	id uint64
}

func NewSingleGenerator() *SingleGenerator {
	return &SingleGenerator{
		id: 10000,
	}
}

func (p *SingleGenerator) GetID() (uint64, *base.Error) {
	return atomic.AddUint64(&p.id, 1), nil
}

type SessionConfig struct {
	channels     int64
	transLimit   int64
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func getDefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		channels:     64,
		transLimit:   4 * 1024 * 1024,
		readTimeout:  12 * time.Second,
		writeTimeout: 3 * time.Second,
	}
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
	ret.channels = make([]Channel, gateway.GetSessionConfig().channels)
	return ret
}

func (p *Session) SetConn(conn internal.IStreamConn) {
	p.Lock()
	defer p.Unlock()

	p.conn = conn
}

func (p *Session) StreamIn(stream *core.Stream) *base.Error {
	cbID := stream.GetCallbackID()

	if cbID > 0 {
		channel := p.channels[cbID%uint64(len(p.channels))]
		if retStream, err := channel.In(cbID, uint64(len(p.channels))); err != nil {
			return err
		} else if retStream != nil {
			return p.StreamOut(retStream)
		} else {
			return p.gateway.OnStream(stream)
		}
	} else {
		return nil
	}
}

func (p *Session) StreamOut(stream *core.Stream) *base.Error {
	cbID := stream.GetCallbackID()
	config := p.gateway.GetSessionConfig()

	// record stream
	if cbID > 0 {
		channel := p.channels[cbID%uint64(len(p.channels))]
		if err := channel.Out(stream); err != nil {
			return err
		}
	}

	// write stream
	return func() *base.Error {
		p.Lock()
		defer p.Unlock()
		return p.conn.WriteStream(stream, config.writeTimeout)
	}()
}
