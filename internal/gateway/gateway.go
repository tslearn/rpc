// Package gateway ...
package gateway

import (
	"crypto/tls"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
)

const (
	sessionManagerVectorSize = 1024
)

// GateWay ...
type GateWay struct {
	id             uint64
	isRunning      bool
	sessionSeed    uint64
	totalSessions  int64
	sessionMapList []*SessionPool
	streamHub      rpc.IStreamHub
	closeCH        chan bool
	config         *Config
	adapters       []*adapter.Adapter
	orcManager     *base.ORCManager
	sync.Mutex
}

// NewGateWay ...
func NewGateWay(
	id uint64,
	config *Config,
	streamHub rpc.IStreamHub,
) *GateWay {
	if streamHub == nil {
		panic("streamHub is nil")
	}

	ret := &GateWay{
		id:             id,
		isRunning:      false,
		sessionSeed:    0,
		totalSessions:  0,
		sessionMapList: make([]*SessionPool, sessionManagerVectorSize),
		streamHub:      streamHub,
		closeCH:        make(chan bool, 1),
		config:         config,
		adapters:       make([]*adapter.Adapter, 0),
		orcManager:     base.NewORCManager(),
	}

	for i := 0; i < sessionManagerVectorSize; i++ {
		ret.sessionMapList[i] = NewSessionPool(ret)
	}

	return ret
}

// TotalSessions ...
func (p *GateWay) TotalSessions() int64 {
	return atomic.LoadInt64(&p.totalSessions)
}

// AddSession ...
func (p *GateWay) AddSession(session *Session) bool {
	return p.sessionMapList[session.id%sessionManagerVectorSize].Add(session)
}

// GetSession ...
func (p *GateWay) GetSession(id uint64) (*Session, bool) {
	return p.sessionMapList[id%sessionManagerVectorSize].Get(id)
}

// CreateSessionID ...
func (p *GateWay) CreateSessionID() uint64 {
	return atomic.AddUint64(&p.sessionSeed, 1)
}

// TimeCheck ...
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
			false,
			network,
			addr,
			tlsConfig,
			p.config.serverReadBufferSize,
			p.config.serverWriteBufferSize,
			p,
		))
	} else {
		p.streamHub.OnReceiveStream(
			rpc.MakeSystemErrorStream(base.ErrGatewayAlreadyRunning),
		)
	}

	return p
}

// ListenWithDebug ...
func (p *GateWay) ListenWithDebug(
	network string,
	addr string,
	tlsConfig *tls.Config,
) *GateWay {
	p.Lock()
	defer p.Unlock()

	if !p.isRunning {
		p.adapters = append(p.adapters, adapter.NewServerAdapter(
			true,
			network,
			addr,
			tlsConfig,
			p.config.serverReadBufferSize,
			p.config.serverWriteBufferSize,
			p,
		))
	} else {
		p.streamHub.OnReceiveStream(
			rpc.MakeSystemErrorStream(base.ErrGatewayAlreadyRunning),
		)
	}

	return p
}

// Open ...
func (p *GateWay) Open() {
	p.orcManager.Open(func() bool {
		p.Lock()
		defer p.Unlock()

		if p.isRunning {
			p.streamHub.OnReceiveStream(
				rpc.MakeSystemErrorStream(base.ErrGatewayAlreadyRunning),
			)
			return false
		} else if len(p.adapters) <= 0 {
			p.streamHub.OnReceiveStream(
				rpc.MakeSystemErrorStream(base.ErrGatewayNoAvailableAdapter),
			)
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

// OutStream ...
func (p *GateWay) OutStream(stream *rpc.Stream) {
	if session, ok := p.GetSession(stream.GetSessionID()); ok {
		session.OutStream(stream)
	} else {
		errStream := rpc.MakeSystemErrorStream(base.ErrGateWaySessionNotFound)
		errStream.SetGatewayID(p.id)
		errStream.SetSessionID(stream.GetSessionID())
		p.streamHub.OnReceiveStream(errStream)
		stream.Release()
	}
}

// OnConnOpen ...
func (p *GateWay) OnConnOpen(_ *adapter.StreamConn) {
	// ignore
	// we will add some security checks here
}

// OnConnReadStream ...
func (p *GateWay) OnConnReadStream(
	streamConn *adapter.StreamConn,
	stream *rpc.Stream,
) {
	InitSession(p, streamConn, stream)
}

// OnConnError ...
func (p *GateWay) OnConnError(streamConn *adapter.StreamConn, err *base.Error) {
	p.streamHub.OnReceiveStream(rpc.MakeSystemErrorStream(err))

	if streamConn != nil {
		streamConn.Close()
	}
}

// OnConnClose ...
func (p *GateWay) OnConnClose(_ *adapter.StreamConn) {
	// ignore
	// streamConn is not attached to a session
	// If it happens multiple times on one ip, it may be an attack
}
