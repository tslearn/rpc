package gateway

import (
	"github.com/rpccloud/rpc/internal"
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/adapter/xtcp"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"sync"
	"sync/atomic"
	"time"
)

const MaxID = 0xFFFFFF
const MaxSeed = 0xFFFFFFFFFF

type GateWay struct {
	sessionHead uint64
	sessionSeed uint64
	isRunning   bool
	slot        internal.IStreamRouterSlot
	closeCH     chan bool
	config      *SessionConfig
	sessionMap  map[uint64]*Session
	onError     func(sessionID uint64, err *base.Error)
	adapters    []adapter.IAdapter
	sync.Mutex
}

func NewGateWay(
	id uint64,
	config *SessionConfig,
	router internal.IStreamRouter,
	onError func(sessionID uint64, err *base.Error),
) (*GateWay, *base.Error) {
	if id > MaxID {
		return nil, errors.ErrGateWayIDOverflows
	}

	ret := &GateWay{
		sessionHead: id << 40,
		isRunning:   false,
		slot:        nil,
		closeCH:     make(chan bool, 1),
		config:      config,
		sessionMap:  map[uint64]*Session{},
		onError:     onError,
		adapters:    make([]adapter.IAdapter, 0),
	}
	ret.slot = router.Plug(ret)
	return ret, nil
}

func (p *GateWay) GetSessionID() (uint64, *base.Error) {
	seed := atomic.AddUint64(&p.sessionSeed, 1)
	if seed > MaxSeed {
		return 0, errors.ErrGateWaySeedOverflows
	}
	return p.sessionHead | seed, nil
}

func (p *GateWay) ListenTCP(addr string) *GateWay {
	p.Lock()
	defer p.Unlock()

	if !p.isRunning {
		p.adapters = append(p.adapters, xtcp.NewTCPServerAdapter(addr))
	} else {
		p.onError(0, errors.ErrGatewayAlreadyRunning)
	}

	return p
}

func (p *GateWay) IsRunning() bool {
	p.Lock()
	defer p.Unlock()
	return p.isRunning
}

func (p *GateWay) Serve() {
	waitCH := make(chan bool)

	numOfRunning, err := func() (int, *base.Error) {
		p.Lock()
		defer p.Unlock()

		if p.isRunning {
			return 0, errors.ErrGatewayAlreadyRunning
		} else if len(p.adapters) <= 0 {
			return 0, errors.ErrGatewayNoAvailableAdapters
		} else {
			p.isRunning = true
			for _, item := range p.adapters {
				go func(adapter adapter.IAdapter) {
					for {
						adapter.Open(p)

						if p.IsRunning() {
							time.Sleep(time.Second)
						} else {
							waitCH <- true
							return
						}
					}
				}(item)
			}
			return len(p.adapters), nil
		}
	}()

	if err != nil {
		p.onError(0, err)
	} else {
		for i := 0; i < numOfRunning; i++ {
			<-waitCH
		}

		p.closeCH <- true
	}
}

func (p *GateWay) getSessionById(id uint64) *Session {
	p.Lock()
	defer p.Unlock()
	return p.sessionMap[id]
}

func (p *GateWay) Close() {
	err := func() *base.Error {
		p.Lock()
		defer p.Unlock()

		if !p.isRunning {
			return errors.ErrGatewayNotRunning
		}

		p.isRunning = false

		for _, item := range p.adapters {
			go func(adapter adapter.IAdapter) {
				adapter.Close(p)
			}(item)
		}

		return nil
	}()

	if err != nil {
		p.onError(0, err)
	} else {
		<-p.closeCH
	}
}

func (p *GateWay) OnStream(stream *core.Stream) *base.Error {
	defer stream.Release()

	if !stream.IsDirectionOut() {
		return errors.ErrStream
	}

	if session := p.getSessionById(stream.GetSessionID()); session != nil {
		return session.StreamOut(stream)
	}

	return errors.ErrGateWaySessionNotFound
}

func (p *GateWay) OnEventConnStream(
	eventConn *adapter.EventConn,
	stream *core.Stream,
) {

}

func (p *GateWay) OnEventConnError(
	eventConn *adapter.EventConn,
	err *base.Error,
) {

}

func (p *GateWay) OnEventConnClose(eventConn *adapter.EventConn) {

}
