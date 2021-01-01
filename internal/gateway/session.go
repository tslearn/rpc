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

type SessionList struct {
    head *Session
}

func (p *SessionList) Add(session *Session) {
    if p.head != nil {
        p.head.prev = session
    }

    session.prev = nil
    session.next = p.head
    p.head = session
}

func (p *SessionList) Remove(session *Session) {
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
}

type SessionManager struct {
    sessionMap map[uint64]*Session
    listVector [16]SessionList
    sync.Mutex
}

func NewSessionManager() *SessionManager {
    return &SessionManager{
        sessionMap: map[uint64]*Session{},
    }
}

func (p *SessionManager) Add(session *Session) bool {
    p.Lock()
    defer p.Unlock()

    if _, ok := p.sessionMap[session.id]; ok {
        return false
    }

    p.sessionMap[session.id] = session
    p.listVector[session.id % 16].Add(session)
    return true
}

func (p *SessionManager) Remove(id uint64) bool {
    p.Lock()
    defer p.Unlock()

    if session, ok := p.sessionMap[id]; ok {
        p.listVector[id % 16].Remove(session)
        return true
    }

    return false
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
