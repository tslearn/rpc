package gateway

import (
	"github.com/rpccloud/rpc/internal"
	"github.com/rpccloud/rpc/internal/adapter/websocket"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
	"sync"
	"time"
)

type GateWay struct {
	isRunning  bool
	slot       internal.IStreamRouterSlot
	closeCH    chan bool
	config     *SessionConfig
	sessionMap map[uint64]*Session
	onError    func(sessionID uint64, err *base.Error)
	adapters   []internal.IServerAdapter
	sync.Mutex
}

func NewGateWay(
	router internal.IStreamRouter,
	onError func(sessionID uint64, err *base.Error),
) (*GateWay, *base.Error) {
	ret := &GateWay{
		isRunning: false,
		slot:      nil,
		closeCH:   make(chan bool, 1),
		config:    getDefaultSessionConfig(),
		onError:   onError,
		adapters:  make([]internal.IServerAdapter, 0),
	}

	slot, err := router.Plug(ret)

	if err != nil {
		return nil, err
	}

	ret.slot = slot
	return ret, nil
}

func (p *GateWay) ListenWebSocket(addr string) *GateWay {
	p.Lock()
	defer p.Unlock()

	if !p.isRunning {
		p.adapters = append(p.adapters, websocket.NewWebsocketServerAdapter(addr))
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
				go func(adapter internal.IServerAdapter) {
					for {
						adapter.Open(p.onConnRun, p.onError)
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

func (p *GateWay) onConnRun(conn internal.IStreamConn, addr net.Addr) {

}

func (p *GateWay) OnStream(stream *core.Stream) *base.Error {
	if stream.IsDirectionOut() {
		return func() *base.Error {
			defer stream.Release()

			if session := p.getSessionById(stream.GetSessionID()); session != nil {
				return session.StreamOut(stream)
			}

			return errors.ErrGateWaySessionNotFound
		}()
	} else {
		return p.slot.SendStream(stream)
	}
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
			go func(adapter internal.IServerAdapter) {
				adapter.Close(p.onError)
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
