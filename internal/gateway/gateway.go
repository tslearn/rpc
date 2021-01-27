package gateway

import (
	"crypto/tls"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"github.com/rpccloud/rpc/internal/route"
)

const (
	sessionManagerVectorSize = 1024
)

// GateWay ...
type GateWay struct {
	id             uint32
	isRunning      bool
	sessionSeed    uint64
	totalSessions  int64
	sessionMapList []*SessionPool
	routeSender    route.IRouteSender
	closeCH        chan bool
	config         *Config
	onError        func(sessionID uint64, err *base.Error)
	adapters       []*adapter.Adapter
	orcManager     *base.ORCManager
	sync.Mutex
}

// NewGateWay ...
func NewGateWay(
	id uint32,
	config *Config,
	router route.IRouter,
	onError func(sessionID uint64, err *base.Error),
) *GateWay {
	ret := &GateWay{
		id:             id,
		isRunning:      false,
		sessionSeed:    1,
		totalSessions:  0,
		sessionMapList: make([]*SessionPool, sessionManagerVectorSize),
		routeSender:    nil,
		closeCH:        make(chan bool, 1),
		config:         config,
		onError:        onError,
		adapters:       make([]*adapter.Adapter, 0),
		orcManager:     base.NewORCManager(),
	}

	for i := 0; i < sessionManagerVectorSize; i++ {
		ret.sessionMapList[i] = NewSessionPool(ret)
	}

	ret.routeSender = router.Plug(ret)

	return ret
}

func (p *GateWay) TotalSessions() int64 {
	return atomic.LoadInt64(&p.totalSessions)
}

func (p *GateWay) addSession(session *Session) bool {
	return p.sessionMapList[session.id%sessionManagerVectorSize].Add(session)
}

func (p *GateWay) getSession(id uint64) (*Session, bool) {
	return p.sessionMapList[id%sessionManagerVectorSize].Get(id)
}

// thread unsafe
func (p *GateWay) TimeCheck(nowNS int64) {
	for i := 0; i < sessionManagerVectorSize; i++ {
		p.sessionMapList[i].TimeCheck(nowNS)
	}
}

// Listen ...
func (p *GateWay) Listen(
	network string,
	addr string,
	tlsConfig *tls.Config,
) *GateWay {
	p.Lock()
	defer p.Unlock()

	if !p.isRunning {
		p.adapters = append(p.adapters, adapter.NewServerAdapter(
			network,
			addr,
			tlsConfig,
			p.config.serverReadBufferSize,
			p.config.serverWriteBufferSize,
			p,
		))
	} else {
		p.onError(0, errors.ErrGatewayAlreadyRunning)
	}

	return p
}

// Open ...
func (p *GateWay) Open() {
	p.orcManager.Open(func() bool {
		p.Lock()
		defer p.Unlock()

		if p.isRunning {
			p.onError(0, errors.ErrGatewayAlreadyRunning)
			return false
		} else if len(p.adapters) <= 0 {
			p.onError(0, errors.ErrGatewayNoAvailableAdapter)
			return false
		} else {
			p.isRunning = true
			return true
		}
	})

	// -------------------------------------------------------------------------
	// Notice:
	//      if p.orcManager.Close() is called between Open and Run. Run will not
	// execute at all.
	// -------------------------------------------------------------------------
	p.orcManager.Run(func(isRunning func() bool) bool {
		waitCH := make(chan bool)
		waitCount := 0

		for _, item := range p.adapters {
			waitCount++
			item.Open()
			go func(adapter *adapter.Adapter) {
				adapter.Run()
				waitCH <- true
			}(item)
		}

		for isRunning() {
			startNS := base.TimeNow().UnixNano()
			p.TimeCheck(startNS)
			base.WaitAtLeastDurationWhenRunning(startNS, isRunning, time.Second)
		}

		for waitCount > 0 {
			<-waitCH
			waitCount--
		}

		return true
	})
}

// Close ...
func (p *GateWay) Close() {
	p.orcManager.Close(func() bool {
		for _, item := range p.adapters {
			item.Close()
		}
		return true
	}, func() {
		p.Lock()
		defer p.Unlock()
		p.isRunning = false
	})
}

func (p *GateWay) ReceiveStreamFromRouter(stream *core.Stream) *base.Error {
	if session, ok := p.getSession(stream.GetSessionID()); ok {
		session.OutStream(stream)
		return nil
	} else {
		stream.Release()
		return errors.ErrGateWaySessionNotFound
	}
}

// OnConnOpen ...
func (p *GateWay) OnConnOpen(_ *adapter.StreamConn) {
	// ignore this message
	// we will add some security checks here
}

// OnConnReadStream ...
func (p *GateWay) OnConnReadStream(
	streamConn *adapter.StreamConn,
	stream *core.Stream,
) {
	defer stream.Release()

	if stream.GetCallbackID() != 0 {
		p.OnConnError(streamConn, errors.ErrStream)
	} else if kind, err := stream.ReadInt64(); err != nil {
		p.OnConnError(streamConn, errors.ErrStream)
	} else if kind != core.ControlStreamConnectRequest {
		p.OnConnError(streamConn, errors.ErrStream)
	} else if sessionString, err := stream.ReadString(); err != nil {
		p.OnConnError(streamConn, errors.ErrStream)
	} else if !stream.IsReadFinish() {
		p.OnConnError(streamConn, errors.ErrStream)
	} else {
		session := (*Session)(nil)

		// try to find session by session string
		strArray := strings.Split(sessionString, "-")
		if len(strArray) == 2 && len(strArray[1]) == 32 {
			if id, err := strconv.ParseUint(strArray[0], 10, 64); err == nil {
				if s, ok := p.getSession(id); ok && s.security == strArray[1] {
					session = s
				}
			}
		}

		// if session not find by session string, create a new session
		if session == nil {
			if p.TotalSessions() >= int64(p.config.serverMaxSessions) {
				p.OnConnError(streamConn, errors.ErrGateWaySeedOverflows)
			} else {
				session = NewSession(atomic.AddUint64(&p.sessionSeed, 1), p)
				p.addSession(session)
				session.OnConnOpen(streamConn)
			}
		}
	}
}

// OnConnError ...
func (p *GateWay) OnConnError(streamConn *adapter.StreamConn, err *base.Error) {
	p.onError(0, err)

	if streamConn != nil {
		streamConn.Close()
	}
}

// OnConnClose ...
func (p *GateWay) OnConnClose(_ *adapter.StreamConn) {
	panic("kernel error: it should not be called")
}
