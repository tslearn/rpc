package gateway

import (
	"fmt"
	"github.com/rpccloud/rpc/internal"
	"github.com/rpccloud/rpc/internal/adapter/websocket"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type GateWay struct {
	isRunning   bool
	idGenerator SessionIDGenerator
	slot        internal.IStreamRouterSlot
	closeCH     chan bool
	config      *SessionConfig
	sessionMap  map[uint64]*Session
	onError     func(sessionID uint64, err *base.Error)
	adapters    []internal.IServerAdapter
	sync.Mutex
}

func NewGateWay(
	idGenerator SessionIDGenerator,
	config *SessionConfig,
	router internal.IStreamRouter,
	onError func(sessionID uint64, err *base.Error),
) *GateWay {
	ret := &GateWay{
		isRunning:   false,
		idGenerator: idGenerator,
		slot:        nil,
		closeCH:     make(chan bool, 1),
		config:      config,
		sessionMap:  map[uint64]*Session{},
		onError:     onError,
		adapters:    make([]internal.IServerAdapter, 0),
	}
	ret.slot = router.Plug(ret)
	return ret
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
	session := (*Session)(nil)
	initStream, runError := conn.ReadStream(
		p.config.readTimeout,
		p.config.transLimit,
	)

	defer func() {
		sessionID := uint64(0)
		if session != nil {
			sessionID = session.id
		}
		if runError != errors.ErrStreamConnIsClosed {
			p.onError(sessionID, runError)
		}
		if err := conn.Close(); err != nil {
			p.onError(sessionID, err)
		}
	}()

	// init conn
	if runError != nil {
		return
	} else if initStream.GetCallbackID() != 0 {
		initStream.Release()
		runError = errors.ErrStream
	} else if kind, err := initStream.ReadInt64(); err != nil {
		initStream.Release()
		runError = err
	} else if kind != core.ControlStreamConnectRequest {
		initStream.Release()
		runError = errors.ErrStream
	} else if sessionString, err := initStream.ReadString(); err != nil {
		initStream.Release()
		runError = errors.ErrStream
	} else if !initStream.IsReadFinish() {
		initStream.Release()
		runError = errors.ErrStream
	} else {
		// try to find session by session string
		sessionArray := strings.Split(sessionString, "-")
		if len(sessionArray) == 2 && len(sessionArray[1]) == 32 {
			if id, err := strconv.ParseUint(sessionArray[0], 10, 64); err == nil {
				p.Lock()
				if s, ok := p.sessionMap[id]; ok && s.security == sessionArray[1] {
					session = s
				}
				p.Unlock()
			}
		}
		// if session not find by session string, create a new session
		if session == nil {
			sessionID, err := p.idGenerator.GetID()

			if err != nil {
				runError = err
			} else {
				session = newSession(sessionID, p.config, p.slot)
				p.Lock()
				p.sessionMap[sessionID] = session
				p.Unlock()
			}
		}

		if session != nil {
			initStream.SetWritePosToBodyStart()
			initStream.WriteInt64(core.ControlStreamConnectResponse)
			initStream.WriteString(fmt.Sprintf("%d-%s", session.id, session.security))
			p.config.WriteToStream(initStream)

			if err := conn.WriteStream(
				initStream,
				p.config.writeTimeout,
			); err != nil {
				initStream.Release()
				runError = err
				return
			} else {
				initStream.Release()
			}

			// Pump message from client
			session.SetConn(conn)
			defer session.SetConn(nil)

			for runError == nil {
				if stream, err := conn.ReadStream(
					p.config.readTimeout,
					p.config.transLimit,
				); err != nil {
					runError = err
				} else {
					runError = session.StreamIn(stream)
				}
			}
		} else {
			initStream.Release()
		}
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
