package gateway

import (
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

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
	id           uint64
	isFree       bool
	sessionSeed  uint32
	routerSender router.IRouteSender
	closeCH      chan bool
	config       *Config
	sessionMap   map[uint64]*Session
	onError      func(sessionID uint64, err *base.Error)
	adapters     []*adapter.Adapter
	orcManager   *base.ORCManager
	sync.Mutex
}

// NewGateWay ...
func NewGateWay(
	id uint64,
	config *Config,
	router router.IRouter,
	onError func(sessionID uint64, err *base.Error),
) *GateWay {
	if id > maxGatewayID {
		panic(fmt.Sprintf("illegal gateway ID %d", id))
	}

	ret := &GateWay{
		id:           id,
		isFree:       true,
		routerSender: nil,
		closeCH:      make(chan bool, 1),
		config:       config,
		sessionMap:   map[uint64]*Session{},
		onError:      onError,
		adapters:     make([]*adapter.Adapter, 0),
		orcManager:   base.NewORCManager(),
	}

	ret.routerSender = router.Plug(ret)

	return ret
}

func (p *GateWay) generateSessionID() (uint64, *base.Error) {
	if len(p.sessionMap) > p.config.serverMaxSessions {
		return 0, errors.ErrGateWaySeedOverflows
	}

	for {
		// Case:
		//      some extreme case may make session id concentration in a special
		// area. consider the following: first allocate a million long sessions,
		// then allocate lots of short sessions (=2^32 - 1million). because of
		// the long sessions were still alive. the search of seed will take a
		// long time. this will case the system jitter
		// ---------------------------------------------------------------------
		// Solution:
		//      seed increases several intervals (must be prime number)
		// each time, when the seed is conflict we simple add 1 to avoid
		// conflict area.
		// ---------------------------------------------------------------------
		// Notice:
		//      zero seed is not illegal
		// ---------------------------------------------------------------------
		if v := atomic.AddUint32(&p.sessionSeed, 941); v != 0 {
			ret := p.id<<40 | uint64(v)
			if _, ok := p.sessionMap[ret]; !ok {
				return ret, nil
			}
		}

		atomic.AddUint32(&p.sessionSeed, 1)
	}
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
	p.orcManager.Run(func(_ func() bool) {
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

		for waitCount > 0 {
			<-waitCH
			waitCount--
		}
	})
}

// Close ...
func (p *GateWay) Close() {
	p.orcManager.Close(func() {
		for _, item := range p.adapters {
			go func(adapter *adapter.Adapter) {
				adapter.Close()
			}(item)
		}
	}, func() {
		p.Lock()
		defer p.Unlock()
		p.isFree = true
	})
}

// ReceiveStreamFromRouter ...
func (p *GateWay) ReceiveStreamFromRouter(stream *core.Stream) *base.Error {
	p.Lock()
	defer p.Unlock()

	if !stream.IsDirectionOut() {
		stream.Release()
		return errors.ErrStream
	} else if session, ok := p.sessionMap[stream.GetSessionID()]; ok {
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
				p.Lock()
				if s, ok := p.sessionMap[id]; ok && s.security == strArray[1] {
					session = s
				}
				p.Unlock()
			}
		}

		// if session not find by session string, create a new session
		if session == nil {
			if sessionID, err := p.generateSessionID(); err != nil {
				p.OnConnError(streamConn, errors.ErrStream)
			} else {
				session = newSession(sessionID, p)
				p.Lock()
				p.sessionMap[sessionID] = session
				p.Unlock()

				session.Initialized(streamConn)
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
