package gateway

import (
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"github.com/rpccloud/rpc/internal/router"
)

const (
	maxGatewayID = 0xFFFFFF
)

// GateWay ...
type GateWay struct {
	id             uint32
	isFree         bool
	sessionSeed    uint64
	routerSender   router.IRouteSender
	closeCH        chan bool
	config         *Config
	sessionManager *SessionManager
	onError        func(sessionID uint64, err *base.Error)
	adapters       []*adapter.Adapter
	orcManager     *base.ORCManager
	sync.Mutex
}

// NewGateWay ...
func NewGateWay(
	id uint32,
	config *Config,
	router router.IRouter,
	onError func(sessionID uint64, err *base.Error),
) *GateWay {
	if id > maxGatewayID {
		panic(fmt.Sprintf("illegal gateway ID %d", id))
	}

	ret := &GateWay{
		id:             id,
		isFree:         true,
		sessionSeed:    1,
		routerSender:   nil,
		closeCH:        make(chan bool, 1),
		config:         config,
		sessionManager: NewSessionManager(),
		onError:        onError,
		adapters:       make([]*adapter.Adapter, 0),
		orcManager:     base.NewORCManager(),
	}

	ret.routerSender = router.Plug(ret)

	return ret
}

// GetConfig ...
func (p *GateWay) GetConfig() *Config {
	return p.config
}

// Listen ...
func (p *GateWay) Listen(
	network string,
	addr string,
	tlsConfig *tls.Config,
) *GateWay {
	p.Lock()
	defer p.Unlock()

	if p.isFree {
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

// Serve ...
func (p *GateWay) Serve() {
	p.orcManager.Open(func() bool {
		p.Lock()
		defer p.Unlock()

		if !p.isFree {
			p.OnConnError(nil, errors.ErrGatewayAlreadyRunning)
			return false
		} else if len(p.adapters) <= 0 {
			p.OnConnError(nil, errors.ErrGatewayNoAvailableAdapters)
			return false
		} else {
			p.isFree = false
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
			go func(adapter *adapter.Adapter) {
				adapter.Open()
				adapter.Run()
				waitCH <- true
			}(item)
		}

		for isRunning() {
			startNS := base.TimeNow().UnixNano()
			p.sessionManager.TimeCheck(startNS)
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
		for _, adapter := range p.adapters {
			adapter.Close()
		}
		return true
	}, func() {
		p.Lock()
		defer p.Unlock()
		p.isFree = true
	})
}

// ReceiveStreamFromRouter ...
func (p *GateWay) ReceiveStreamFromRouter(stream *core.Stream) *base.Error {
	if !stream.IsDirectionOut() {
		stream.Release()
		return errors.ErrStream
	} else if session, ok := p.sessionManager.Get(stream.GetSessionID()); ok {
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

// OnConnClose ...
func (p *GateWay) OnConnClose(_ *adapter.StreamConn) {
	// All streamConn that is successfully initialized will reset the receiver
	// to session. If it runs here, it means the streamConn is not correctly
	// initialized. so ignore this message
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
				if s, ok := p.sessionManager.Get(id); ok && s.security == strArray[1] {
					session = s
				}
			}
		}

		// if session not find by session string, create a new session
		if session == nil {
			if p.sessionManager.TotalSessions() >= int64(p.config.serverMaxSessions) {
				p.OnConnError(streamConn, errors.ErrGateWaySeedOverflows)
			} else {
				session = newSession(atomic.AddUint64(&p.sessionSeed, 1), p)
				p.sessionManager.Add(session)
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
