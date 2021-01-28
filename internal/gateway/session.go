package gateway

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/adapter"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
)

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

func InitSession(
	gw *GateWay,
	streamConn *adapter.StreamConn,
	stream *core.Stream,
) {
	if stream.GetCallbackID() != 0 {
		stream.Release()
		gw.OnConnError(streamConn, errors.ErrStream)
	} else if kind, err := stream.ReadInt64(); err != nil {
		stream.Release()
		gw.OnConnError(streamConn, errors.ErrStream)
	} else if kind != core.ControlStreamConnectRequest {
		stream.Release()
		gw.OnConnError(streamConn, errors.ErrStream)
	} else if sessionString, err := stream.ReadString(); err != nil {
		stream.Release()
		gw.OnConnError(streamConn, errors.ErrStream)
	} else if !stream.IsReadFinish() {
		stream.Release()
		gw.OnConnError(streamConn, errors.ErrStream)
	} else {
		session := (*Session)(nil)
		config := gw.config

		// try to find session by session string
		strArray := strings.Split(sessionString, "-")
		if len(strArray) == 2 && len(strArray[1]) == 32 {
			if id, err := strconv.ParseUint(strArray[0], 10, 64); err == nil {
				if s, ok := gw.GetSession(id); ok && s.security == strArray[1] {
					session = s
				}
			}
		}

		// if session not find by session string, create a new session
		if session == nil {
			if gw.TotalSessions() >= int64(config.serverMaxSessions) {
				stream.Release()
				gw.OnConnError(streamConn, errors.ErrGateWaySeedOverflows)
				return
			}

			session = &Session{
				id:           gw.CreateSessionID(),
				gateway:      gw,
				security:     base.GetRandString(32),
				conn:         nil,
				channels:     make([]Channel, config.numOfChannels),
				activeTimeNS: base.TimeNow().UnixNano(),
				prev:         nil,
				next:         nil,
			}

			gw.AddSession(session)
		}

		streamConn.SetReceiver(session)

		stream.SetWritePosToBodyStart()
		stream.WriteInt64(core.ControlStreamConnectResponse)
		stream.WriteString(fmt.Sprintf("%d-%s", session.id, session.security))
		stream.WriteInt64(int64(config.numOfChannels))
		stream.WriteInt64(int64(config.transLimit))
		stream.WriteInt64(int64(config.heartbeat))
		stream.WriteInt64(int64(config.heartbeatTimeout))
		stream.WriteInt64(int64(config.clientRequestInterval))
		streamConn.WriteStreamAndRelease(stream)

		session.OnConnOpen(streamConn)
	}
}

func (p *Session) TimeCheck(nowNS int64) {
	p.Lock()
	defer p.Unlock()

	if gw := p.gateway; gw != nil {
		config := gw.config

		if p.conn != nil {
			// conn timeout
			if !p.conn.IsActive(nowNS, config.heartbeatTimeout) {
				p.conn.Close()
			}
		} else {
			// session timeout
			if nowNS-p.activeTimeNS > int64(config.serverSessionTimeout) {
				p.activeTimeNS = 0
			}
		}

		// channel timeout
		timeoutNS := int64(config.serverCacheTimeout)
		for i := 0; i < len(p.channels); i++ {
			if channel := &p.channels[i]; channel.IsTimeout(nowNS, timeoutNS) {
				channel.Clean()
			}
		}
	}
}

// OutStream ...
func (p *Session) OutStream(stream *core.Stream) {
	p.Lock()
	defer p.Unlock()

	if stream != nil && p.conn != nil {
		// record stream
		channel := &p.channels[stream.GetCallbackID()%uint64(len(p.channels))]
		if channel.Out(stream) {
			p.conn.WriteStreamAndRelease(stream.Clone())
		} else {
			stream.Release()
		}
	}
}

// OnConnOpen ...
func (p *Session) OnConnOpen(streamConn *adapter.StreamConn) {
	p.Lock()
	defer p.Unlock()
	p.conn = streamConn
}

// OnConnReadStream ...
func (p *Session) OnConnReadStream(
	streamConn *adapter.StreamConn,
	stream *core.Stream,
) {
	p.Lock()
	defer p.Unlock()

	if cbID := stream.GetCallbackID(); cbID > 0 {
		channel := &p.channels[cbID%uint64(len(p.channels))]
		if accepted, backStream := channel.In(cbID); accepted {
			stream.SetGatewayID(p.gateway.id)
			stream.SetSessionID(p.id)
			sender := p.gateway.routeSender
			// who receives the stream is responsible for releasing it
			if err := sender.SendStreamToRouter(stream); err != nil {
				p.OnConnError(streamConn, err)
			}
		} else if backStream != nil {
			// do not release the backStream, so we need to clone it
			streamConn.WriteStreamAndRelease(backStream.Clone())
			stream.Release()
		} else {
			// ignore the stream
			stream.Release()
		}
	} else if kind, err := stream.ReadInt64(); err != nil {
		p.OnConnError(streamConn, err)
		stream.Release()
	} else if kind == core.ControlStreamPing && stream.IsReadFinish() {
		p.activeTimeNS = base.TimeNow().UnixNano()
		stream.SetWritePosToBodyStart()
		stream.WriteInt64(core.ControlStreamPong)
		streamConn.WriteStreamAndRelease(stream)
	} else {
		p.OnConnError(streamConn, errors.ErrStream)
		stream.Release()
	}
}

// OnConnError ...
func (p *Session) OnConnError(streamConn *adapter.StreamConn, err *base.Error) {
	p.gateway.onError(p.id, err)

	if streamConn != nil {
		streamConn.Close()
	}
}

// OnConnClose ...
func (p *Session) OnConnClose(_ *adapter.StreamConn) {
	p.Lock()
	defer p.Unlock()
	p.conn = nil
}

type SessionPool struct {
	gateway *GateWay
	idMap   map[uint64]*Session
	head    *Session
	sync.Mutex
}

func NewSessionPool(gateway *GateWay) *SessionPool {
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

	if session == nil {
		return false
	}

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

func (p *SessionPool) TimeCheck(nowNS int64) {
	p.Lock()
	defer p.Unlock()

	node := p.head
	for node != nil {
		node.TimeCheck(nowNS)

		// remove it from the list
		if node.activeTimeNS == 0 {
			delete(p.idMap, node.id)

			if node.prev != nil {
				node.prev.next = node.next
			}

			if node.next != nil {
				node.next.prev = node.prev
			}

			if node == p.head {
				p.head = node.next
			}

			atomic.AddInt64(&p.gateway.totalSessions, -1)
		}

		node = node.next
	}
}
