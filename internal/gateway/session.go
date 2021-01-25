package gateway

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/adapter"
	"sync"
	"sync/atomic"

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
	id           uint64
	gateway      *GateWay
	security     string
	conn         *adapter.StreamConn
	channels     []Channel
	activeTimeNS int64
	prev         *Session
	next         *Session
	sync.Mutex
}

func NewSession(id uint64, gateway *GateWay) *Session {
	ret := sessionCache.Get().(*Session)
	ret.id = id
	ret.gateway = gateway
	ret.security = base.GetRandString(32)
	ret.conn = nil
	ret.channels = make([]Channel, gateway.config.numOfChannels)
	ret.activeTimeNS = base.TimeNow().UnixNano()
	ret.prev = nil
	ret.next = nil
	return ret
}

func (p *Session) TimeCheck(nowNS int64) {
	p.Lock()
	defer p.Unlock()

	if gw := p.gateway; gw != nil {
		if p.conn != nil {
			// conn timeout
			if !p.conn.IsActive(nowNS, gw.config.heartbeatTimeout) {
				p.conn.Close()
			}
		} else {
			// session timeout
			if nowNS-p.activeTimeNS > int64(gw.config.serverSessionTimeout) {
				gw.Remove(p.id)

				// release session
				p.id = 0
				p.gateway = nil
				p.security = ""
				for i := 0; i < len(p.channels); i++ {
					(&p.channels[i]).Clean()
				}
				p.activeTimeNS = 0
				p.prev = nil
				p.next = nil
				sessionCache.Put(p)
			}
		}

		// channel timeout
		for i := 0; i < len(p.channels); i++ {
			p.channels[i].TimeCheck(
				nowNS,
				int64(gw.config.serverCacheTimeout),
			)
		}
	}
}

// OutStream ...
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
		p.activeTimeNS = base.TimeNow().UnixNano()
	} else {
		p.OnConnError(streamConn, errors.ErrStream)
	}
}

// OnConnError ...
func (p *Session) OnConnError(streamConn *adapter.StreamConn, err *base.Error) {
	// Route to gateway
	p.gateway.OnConnError(streamConn, err)
}

// OnConnClose ...
func (p *Session) OnConnClose(_ *adapter.StreamConn) {
	p.Lock()
	p.conn = nil
	p.Unlock()
}

type SessionPool struct {
	gateway *GateWay
	idMap   map[uint64]*Session
	head    *Session
	sync.Mutex
}

func NewSessionMap(gateway *GateWay) *SessionPool {
	return &SessionPool{
		gateway: gateway,
		idMap:   map[uint64]*Session{},
		head:    nil,
	}
}

func (p *SessionPool) Get(id uint64) (*Session, bool) {
	p.Lock()
	defer p.Unlock()

	ret, ok := p.idMap[id]
	return ret, ok
}

func (p *SessionPool) Add(session *Session) bool {
	p.Lock()
	defer p.Unlock()

	if _, exist := p.idMap[session.id]; !exist {
		p.idMap[session.id] = session

		if p.head != nil {
			p.head.prev = session
		}

		session.prev = nil
		session.next = p.head
		p.head = session

		atomic.AddInt64(&p.gateway.totalSessions, 1)
		return true
	}

	return false
}

func (p *SessionPool) Remove(id uint64) bool {
	p.Lock()
	defer p.Unlock()

	if session, exist := p.idMap[id]; exist {
		delete(p.idMap, id)

		if session.prev != nil {
			session.prev.next = session.next
		}

		if session.next != nil {
			session.next.prev = session.prev
		}

		if session == p.head {
			p.head = session.next
		}

		session.prev = nil
		session.next = nil

		atomic.AddInt64(&p.gateway.totalSessions, -1)
		return true
	}

	return false
}

func (p *SessionPool) TimeCheck(nowNS int64) {
	p.Lock()
	defer p.Unlock()

	node := p.head
	for node != nil {
		node.TimeCheck(nowNS)
		node = node.next
	}
}
