package gateway

import (
	"fmt"
	"sync"
	"sync/atomic"

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

type SessionList struct {
	manager    *SessionManager
	sessionMap map[uint64]*Session
	head       *Session
	sync.Mutex
}

func NewSessionList(manager *SessionManager) *SessionList {
	return &SessionList{
		manager:    manager,
		sessionMap: map[uint64]*Session{},
		head:       nil,
	}
}

func (p *SessionList) Get(id uint64) (*Session, bool) {
	p.Lock()
	defer p.Unlock()

	ret, ok := p.sessionMap[id]
	return ret, ok
}

func (p *SessionList) Add(session *Session) bool {
	p.Lock()
	defer p.Unlock()

	if _, exist := p.sessionMap[session.id]; !exist {
		p.sessionMap[session.id] = session

		if p.head != nil {
			p.head.prev = session
		}

		session.prev = nil
		session.next = p.head
		p.head = session

		atomic.AddInt64(&p.manager.totalSessions, 1)
		return true
	}

	return false
}

func (p *SessionList) Remove(id uint64) bool {
	p.Lock()
	defer p.Unlock()

	if session, exist := p.sessionMap[id]; exist {
		delete(p.sessionMap, id)

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

		atomic.AddInt64(&p.manager.totalSessions, -1)
		return true
	}

	return false
}

func (p *SessionList) TimeCheck(nowNS int64) {
	p.Lock()
	defer p.Unlock()

	node := p.head
	for node != nil {
		node.TimeCheck(nowNS)
		node = node.next
	}
}

const sessionManagerVectorSize = 1024

type SessionManager struct {
	totalSessions int64
	listVector    []*SessionList
	sync.Mutex
}

func NewSessionManager() *SessionManager {
	ret := &SessionManager{
		totalSessions: 0,
		listVector:    make([]*SessionList, sessionManagerVectorSize),
	}

	for i := 0; i < sessionManagerVectorSize; i++ {
		ret.listVector[i] = NewSessionList(ret)
	}

	return ret
}

func (p *SessionManager) TotalSessions() int64 {
	return atomic.LoadInt64(&p.totalSessions)
}

func (p *SessionManager) Get(id uint64) (*Session, bool) {
	return p.listVector[id%sessionManagerVectorSize].Get(id)
}

func (p *SessionManager) Add(session *Session) bool {
	return p.listVector[session.id%sessionManagerVectorSize].Add(session)
}

func (p *SessionManager) Remove(id uint64) bool {
	return p.listVector[id%sessionManagerVectorSize].Remove(id)
}

// thread unsafe
func (p *SessionManager) TimeCheck(nowNS int64) {
	for i := 0; i < sessionManagerVectorSize; i++ {
		p.listVector[i].TimeCheck(nowNS)
	}
}

// Session ...
type Session struct {
	id       uint64
	security string
	conn     *adapter.StreamConn
	gateway  *GateWay
	channels []Channel
	prev     *Session
	next     *Session
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

func (p *Session) TimeCheck(nowNS int64) {
	fmt.Println("Session TimeCheck")
}

// Initialized ...
func (p *Session) Initialized(conn *adapter.StreamConn) {
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
	stream.Write(int64(config.heartbeatTimeout))
	stream.Write(int64(config.clientRequestInterval))
	p.conn.WriteStreamAndRelease(stream)
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
	// Route to gateway
	p.gateway.OnConnOpen(streamConn)
}

// OnConnError ...
func (p *Session) OnConnError(streamConn *adapter.StreamConn, err *base.Error) {
	// Route to gateway
	p.gateway.OnConnError(streamConn, err)
}

// OnConnClose ...
func (p *Session) OnConnClose(_ *adapter.StreamConn) {
	//p.gateway.OnConnClose(streamConn)
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
