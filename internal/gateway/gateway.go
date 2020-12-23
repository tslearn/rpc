package gateway

import (
	"crypto/tls"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rpccloud/rpc/internal"
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
)

const (
	maxGatewayID         = 0xFFFFFF
	gatewayStatusRunning = int32(1)
	gatewayStatusClosing = int32(2)
	gatewayStatusClosed  = int32(0)
)

// GateWay ...
type GateWay struct {
	id          uint64
	sessionSeed uint32
	status      int32
	slot        internal.IStreamRouterSlot
	closeCH     chan bool
	config      *Config
	sessionMap  map[uint32]*Session
	onError     func(sessionID uint64, err *base.Error)
	adapters    []*adapter.RunnableService
	sync.Mutex
}

// NewGateWay ...
func NewGateWay(
	id uint64,
	config *Config,
	router internal.IStreamRouter,
	onError func(sessionID uint64, err *base.Error),
) (*GateWay, *base.Error) {
	if id > maxGatewayID {
		return nil, errors.ErrGateWayIDOverflows
	}

	ret := &GateWay{
		id:         id,
		status:     gatewayStatusClosed,
		slot:       nil,
		closeCH:    make(chan bool, 1),
		config:     config,
		sessionMap: map[uint32]*Session{},
		onError:    onError,
		adapters:   make([]*adapter.RunnableService, 0),
	}

	ret.slot = router.Plug(ret)

	return ret, nil
}

func (p *GateWay) generateSessionID() (uint64, *base.Error) {
	if len(p.sessionMap) > p.config.maxSessions {
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
			if _, ok := p.sessionMap[v]; !ok {
				return p.id<<40 | uint64(v), nil
			}
		}

		atomic.AddUint32(&p.sessionSeed, 1)
	}
}

// GetConfig ...
func (p *GateWay) GetConfig() *Config {
	return p.config
}

// ListenTCP ...
func (p *GateWay) ListenTCP(
	network string,
	addr string,
	tlsConfig *tls.Config,
) *GateWay {
	p.Lock()
	defer p.Unlock()

	if atomic.LoadInt32(&p.status) == gatewayStatusClosed {
		p.adapters = append(p.adapters, adapter.NewSyncServerAdapter(
			network, addr, tlsConfig, p.config.rBufSize, p.config.wBufSize, p,
		))
	} else {
		p.onError(0, errors.ErrGatewayAlreadyRunning)
	}

	return p
}

// // IsRunning ...
// func (p *GateWay) IsRunning() bool {
// 	return
// }

// func (p *GateWay) onConnRun(conn internal.IStreamConn, addr net.Addr) {
//     session := (*Session)(nil)
//     initStream, runError := conn.ReadStream(
//         p.config.readTimeout,
//         p.config.transLimit,
//     )

//     defer func() {
//         sessionID := uint64(0)
//         if session != nil {
//             sessionID = session.id
//         }
//         if runError != errors.ErrStreamConnIsClosed {
//             p.onError(sessionID, runError)
//         }
//         if err := conn.Close(); err != nil {
//             p.onError(sessionID, err)
//         }
//     }()

//     // init conn
//     if runError != nil {
//         return
//     } else if initStream.GetCallbackID() != 0 {
//         initStream.Release()
//         runError = errors.ErrStream
//     } else if kind, err := initStream.ReadInt64(); err != nil {
//         initStream.Release()
//         runError = err
//     } else if kind != core.ControlStreamConnectRequest {
//         initStream.Release()
//         runError = errors.ErrStream
//     } else if sessionString, err := initStream.ReadString(); err != nil {
//         initStream.Release()
//         runError = errors.ErrStream
//     } else if !initStream.IsReadFinish() {
//         initStream.Release()
//         runError = errors.ErrStream
//     } else {
//         // try to find session by session string
//         sessionArray := strings.Split(sessionString, "-")
//         if len(sessionArray) == 2 && len(sessionArray[1]) == 32 {
//             if id, err := strconv.ParseUint(sessionArray[0], 10, 64); err == nil {
//                 p.Lock()
//                 if s, ok := p.sessionMap[id]; ok && s.security == sessionArray[1] {
//                     session = s
//                 }
//                 p.Unlock()
//             }
//         }
//         // if session not find by session string, create a new session
//         if session == nil {
//             sessionID, err := p.idGenerator.GetID()

//             if err != nil {
//                 runError = err
//             } else {
//                 session = newSession(sessionID, p.config, p.slot)
//                 p.Lock()
//                 p.sessionMap[sessionID] = session
//                 p.Unlock()
//             }
//         }

//         if session != nil {
//             initStream.SetWritePosToBodyStart()
//             initStream.WriteInt64(core.ControlStreamConnectResponse)
//             initStream.WriteString(fmt.Sprintf("%d-%s", session.id, session.security))
//             p.config.WriteToStream(initStream)

//             if err := conn.WriteStream(
//                 initStream,
//                 p.config.writeTimeout,
//             ); err != nil {
//                 initStream.Release()
//                 runError = err
//                 return
//             } else {
//                 initStream.Release()
//             }

//             // Pump message from client
//             session.SetConn(conn)
//             defer session.SetConn(nil)

//             for runError == nil {
//                 if stream, err := conn.ReadStream(
//                     p.config.readTimeout,
//                     p.config.transLimit,
//                 ); err != nil {
//                     runError = err
//                 } else {
//                     runError = session.StreamIn(stream)
//                 }
//             }
//         } else {
//             initStream.Release()
//         }
//     }
// }

// Serve ...
func (p *GateWay) Serve() {
	waitCH := make(chan bool)

	numOfRunning, err := func() (int, *base.Error) {
		p.Lock()
		defer p.Unlock()

		if atomic.LoadInt32(&p.status) != gatewayStatusClosed {
			return 0, errors.ErrGatewayAlreadyRunning
		} else if len(p.adapters) <= 0 {
			return 0, errors.ErrGatewayNoAvailableAdapters
		} else {
			atomic.StoreInt32(&p.status, gatewayStatusRunning)
			for _, item := range p.adapters {
				go func(adapter *adapter.RunnableService) {
					for {
						adapter.Open()

						if atomic.LoadInt32(&p.status) == gatewayStatusRunning {
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

// Close ...
func (p *GateWay) Close() {
	err := func() *base.Error {
		p.Lock()
		defer p.Unlock()

		if atomic.LoadInt32(&p.status) != gatewayStatusRunning {
			return errors.ErrGatewayNotRunning
		}

		atomic.StoreInt32(&p.status, gatewayStatusClosing)

		for _, item := range p.adapters {
			go func(adapter *adapter.RunnableService) {
				adapter.Close()
			}(item)
		}

		return nil
	}()

	if err != nil {
		p.onError(0, err)
	} else {
		<-p.closeCH
		atomic.StoreInt32(&p.status, gatewayStatusClosed)
	}
}

// OnStream ...
func (p *GateWay) OnStream(stream *core.Stream) *base.Error {
	p.Lock()
	defer p.Unlock()
	defer stream.Release()

	if !stream.IsDirectionOut() {
		return errors.ErrStream
	}

	if session, ok := p.sessionMap[uint32(stream.GetSessionID())]; ok {
		return session.StreamOut(stream)
	}

	return errors.ErrGateWaySessionNotFound
}

// OnConnOpen ...
func (p *GateWay) OnConnOpen(streamConn *adapter.StreamConn) {

}

// OnConnClose ...
func (p *GateWay) OnConnClose(streamConn *adapter.StreamConn) {

}

// OnConnReadStream ...
func (p *GateWay) OnConnReadStream(streamConn *adapter.StreamConn, stream *core.Stream) {

}

// OnConnError ...
func (p *GateWay) OnConnError(streamConn *adapter.StreamConn, err *base.Error) {
	p.onError(0, err)
}
