package rpc

import (
	"fmt"
	"github.com/rpccloud/rpc/internal"
	"net"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const maxSessionConcurrency = 1024
const minTransportLimit = 10240
const serverSessionRecordStatusNotRunning = 0
const serverSessionRecordStatusRunning = 1

type serverSessionRecord struct {
	id     uint64
	status int32
	mark   bool
	stream unsafe.Pointer
}

var serverSessionRecordCache = &sync.Pool{
	New: func() interface{} {
		return &serverSessionRecord{}
	},
}

func newServerSessionRecord(id uint64) *serverSessionRecord {
	ret := serverSessionRecordCache.Get().(*serverSessionRecord)
	ret.id = id
	ret.status = serverSessionRecordStatusNotRunning
	ret.mark = false
	ret.stream = nil
	return ret
}

func (p *serverSessionRecord) SetRunning() bool {
	return atomic.CompareAndSwapInt32(
		&p.status,
		serverSessionRecordStatusNotRunning,
		serverSessionRecordStatusRunning,
	)
}

func (p *serverSessionRecord) GetReturn() *Stream {
	return (*Stream)(atomic.LoadPointer(&p.stream))
}

func (p *serverSessionRecord) SetReturn(stream *Stream) bool {
	return atomic.CompareAndSwapPointer(&p.stream, nil, unsafe.Pointer(stream))
}

func (p *serverSessionRecord) Release() {
	if stream := p.GetReturn(); stream != nil {
		stream.Release()
	}
	atomic.StorePointer(&p.stream, nil)
	serverSessionRecordCache.Put(p)
}

type serverSession struct {
	id           uint64
	server       *baseServer
	security     string
	conn         internal.IStreamConn
	dataSequence uint64
	ctrlSequence uint64
	callMap      map[uint64]*serverSessionRecord
	sync.Mutex
}

var serverSessionCache = &sync.Pool{
	New: func() interface{} {
		return &serverSession{}
	},
}

func newServerSession(id uint64, server *baseServer) *serverSession {
	ret := serverSessionCache.Get().(*serverSession)
	ret.id = id
	ret.server = server
	ret.security = internal.GetRandString(32)
	ret.conn = nil
	ret.dataSequence = 0
	ret.ctrlSequence = 0
	ret.callMap = make(map[uint64]*serverSessionRecord)
	return ret
}

func (p *serverSession) SetConn(conn internal.IStreamConn) {
	p.Lock()
	defer p.Unlock()

	p.conn = conn
}

func (p *serverSession) OnControlStream(
	conn internal.IStreamConn,
	stream *Stream,
) Error {
	defer stream.Release()

	if kind, ok := stream.ReadInt64(); !ok {
		return internal.NewTransportError(internal.ErrStringBadStream)
	} else if kind != controlStreamKindRequestIds {
		return internal.NewProtocolError(internal.ErrStringBadStream)
	} else if seq := stream.GetSequence(); seq <= p.ctrlSequence {
		return nil
	} else if currCallbackId, ok := stream.ReadUint64(); !ok {
		return internal.NewProtocolError(internal.ErrStringBadStream)
	} else {
		// update sequence
		p.ctrlSequence = seq
		// mark
		for stream.CanRead() {
			if markId, ok := stream.ReadUint64(); ok {
				if v, ok := p.callMap[markId]; ok {
					v.mark = true
				}
			} else {
				return internal.NewProtocolError(internal.ErrStringBadStream)
			}
		}
		if !stream.IsReadFinish() {
			return internal.NewProtocolError(internal.ErrStringBadStream)
		}
		// do swipe and alloc with lock
		func() {
			p.Lock()
			defer p.Unlock()
			// swipe
			count := int64(0)
			for k, v := range p.callMap {
				if v.id <= currCallbackId && !v.mark {
					delete(p.callMap, k)
					v.Release()
				} else {
					v.mark = false
					count++
				}
			}
			// alloc
			for count < p.server.sessionConcurrency {
				p.dataSequence++
				p.callMap[p.dataSequence] = newServerSessionRecord(p.dataSequence)
				count++
			}
		}()
		// return stream
		stream.SetWritePosToBodyStart()
		stream.WriteInt64(controlStreamKindRequestIdsBack)
		stream.WriteUint64(p.dataSequence)
		return conn.WriteStream(stream, p.server.writeTimeout)
	}
}

func (p *serverSession) OnDataStream(
	conn internal.IStreamConn,
	stream *Stream,
	hub streamHub,
) Error {
	if record, ok := p.callMap[stream.GetCallbackID()]; !ok {
		// Cant find record by callbackID
		stream.Release()
		return internal.NewProtocolError("client callbackID error")
	} else if record.SetRunning() {
		// Run the stream. Dont release stream because it will manage by processor
		stream.SetSessionID(p.id)
		hub.PutStream(stream)
		return nil
	} else if retStream := record.GetReturn(); retStream != nil {
		// Write return stream directly if record is finish
		stream.Release()
		return conn.WriteStream(retStream, p.server.writeTimeout)
	} else {
		// Wait if record is not finish
		stream.Release()
		return nil
	}
}

func (p *serverSession) OnReturnStream(stream *Stream) (ret Error) {
	if errKind, ok := stream.ReadUint64(); !ok {
		stream.Release()
		return internal.NewKernelPanic(
			"stream error",
		).AddDebug(string(debug.Stack()))
	} else {
		// Transform panic message for client
		switch internal.ErrorKind(errKind) {
		case internal.ErrorKindReplyPanic:
			fallthrough
		case internal.ErrorKindRuntimePanic:
			fallthrough
		case internal.ErrorKindKernelPanic:
			if message, ok := stream.ReadString(); !ok {
				stream.Release()
				return internal.NewKernelPanic(
					"stream error",
				).AddDebug(string(debug.Stack()))
			} else if dbgMessage, ok := stream.ReadString(); !ok {
				stream.Release()
				return internal.NewKernelPanic(
					"stream error",
				).AddDebug(string(debug.Stack()))
			} else {
				stream.SetWritePosToBodyStart()
				stream.WriteUint64(errKind)
				stream.WriteString("internal error")
				stream.WriteString("")
				// Report error
				ret = internal.NewError(
					internal.ErrorKind(errKind),
					message,
					dbgMessage,
				)
			}
		}
		// SetReturn and get conn with lock
		conn, needRelease := func() (internal.IStreamConn, bool) {
			p.Lock()
			defer p.Unlock()
			if item, ok := p.callMap[stream.GetCallbackID()]; ok {
				return p.conn, !item.SetReturn(stream)
			}
			return p.conn, true
		}()
		// WriteStream
		if conn != nil {
			_ = conn.WriteStream(stream, p.server.writeTimeout)
		}
		// Release
		if needRelease {
			stream.Release()
		}

		return
	}
}

func (p *serverSession) Release() {
	func() {
		p.Lock()
		defer p.Unlock()

		for _, v := range p.callMap {
			v.Release()
		}
		p.callMap = nil
		p.conn = nil
	}()

	p.id = 0
	p.server = nil
	p.security = ""
	p.dataSequence = 0
	p.ctrlSequence = 0
	serverSessionCache.Put(p)
}

type baseServer struct {
	adapters           []internal.IServerAdapter
	hub                streamHub
	sessionMap         sync.Map
	sessionSeed        uint64
	sessionConcurrency int64
	transportLimit     int64
	readTimeout        time.Duration
	writeTimeout       time.Duration
	internal.StatusManager
	sync.Mutex
}

func (p *baseServer) setTransportLimit(maxTransportBytes int, dbg string) {
	p.Lock()
	defer p.Unlock()

	if p.IsRunning() {
		p.onError(0, internal.NewRuntimePanic(
			"SetTransportLimit must be called before Serve",
		).AddDebug(dbg))
	} else if maxTransportBytes < minTransportLimit {
		p.onError(0, internal.NewRuntimePanic(fmt.Sprintf(
			"maxTransportBytes must be greater than or equal to %d",
			minTransportLimit,
		)).AddDebug(dbg))
	} else {
		p.transportLimit = int64(maxTransportBytes)
	}
}

func (p *baseServer) setSessionConcurrency(sessionConcurrency int, dbg string) {
	p.Lock()
	defer p.Unlock()

	if p.IsRunning() {
		p.onError(0, internal.NewRuntimePanic(
			"SetSessionConcurrency must be called before Serve",
		).AddDebug(dbg))
	} else if sessionConcurrency <= 0 {
		p.onError(0, internal.NewRuntimePanic(
			"sessionConcurrency be greater than 0",
		).AddDebug(dbg))
	} else if sessionConcurrency > maxSessionConcurrency {
		p.onError(0, internal.NewRuntimePanic(fmt.Sprintf(
			"sessionConcurrency be less than or equal to %d",
			maxSessionConcurrency,
		)).AddDebug(dbg))
	} else {
		p.sessionConcurrency = int64(sessionConcurrency)
	}
}

func (p *baseServer) listenWebSocket(addr string, dbg string) {
	p.Lock()
	defer p.Unlock()

	if p.IsRunning() {
		p.onError(0, internal.NewRuntimePanic(
			"ListenWebSocket must be called before Serve",
		).AddDebug(dbg))
	} else {
		p.adapters = append(
			p.adapters,
			internal.NewWebSocketServerAdapter(addr),
		)
	}
}

func (p *baseServer) onReturnStream(stream *internal.Stream) {
	if item, ok := p.sessionMap.Load(stream.GetSessionID()); !ok {
		stream.Release()
	} else if session, ok := item.(*serverSession); !ok {
		stream.Release()
		p.onError(stream.GetSessionID(), internal.NewKernelPanic(
			"serverSession is nil",
		).AddDebug(string(debug.Stack())))
	} else {
		if err := session.OnReturnStream(stream); err != nil {
			p.onError(stream.GetSessionID(), err)
		}
	}
}

func (p *baseServer) serve(onGetStreamHub func() streamHub) {
	waitCount := 0
	waitCH := make(chan bool)

	func() {
		p.Lock()
		defer p.Unlock()

		if len(p.adapters) <= 0 {
			p.onError(0, internal.NewRuntimePanic(
				"no valid listener was found on the server",
			))
		} else if hub := onGetStreamHub(); hub == nil {
			p.onError(0, internal.NewKernelPanic(
				"hub is nil",
			).AddDebug(string(debug.Stack())))
		} else if !p.SetRunning(func() {
			p.hub = hub
		}) {
			hub.Close()
			p.onError(0, internal.NewRuntimePanic("it is already running"))
		} else {
			for _, item := range p.adapters {
				waitCount++
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
		}
	}()

	for i := 0; i < waitCount; i++ {
		<-waitCH
	}

	p.SetClosed(func() {
		p.hub = nil
	})
}

func (p *baseServer) Close() {
	waitCH := chan bool(nil)

	if !p.SetClosing(func(ch chan bool) {
		waitCH = ch
		p.hub.Close()
		for _, item := range p.adapters {
			go func(adapter internal.IServerAdapter) {
				adapter.Close(p.onError)
			}(item)
		}
	}) {
		p.onError(0, internal.NewKernelPanic(
			"it is not running",
		).AddDebug(string(debug.Stack())))
	} else {
		select {
		case <-waitCH:
		case <-time.After(5 * time.Second):
			p.onError(0, internal.NewRuntimePanic(
				"it cannot be closed within 5 seconds",
			).AddDebug(string(debug.Stack())))
		}
	}
}

func (p *baseServer) getSession(
	conn internal.IStreamConn,
) (*serverSession, Error) {
	if stream, err := conn.ReadStream(
		p.readTimeout,
		p.transportLimit,
	); err != nil {
		return nil, err
	} else {
		defer stream.Release()

		if stream.GetCallbackID() != 0 {
			return nil, internal.NewProtocolError(internal.ErrStringBadStream)
		} else if stream.GetSequence() == 0 {
			return nil, internal.NewProtocolError(internal.ErrStringBadStream)
		} else if kind, ok := stream.ReadInt64(); !ok ||
			kind != controlStreamKindInit {
			return nil, internal.NewProtocolError(internal.ErrStringBadStream)
		} else if sessionString, ok := stream.ReadString(); !ok {
			return nil, internal.NewProtocolError(internal.ErrStringBadStream)
		} else if !stream.IsReadFinish() {
			return nil, internal.NewProtocolError(internal.ErrStringBadStream)
		} else {
			// try to find session by session string
			session := (*serverSession)(nil)
			sessionArray := strings.Split(sessionString, "-")
			if len(sessionArray) == 2 && len(sessionArray[1]) == 32 {
				if id, err := strconv.ParseUint(sessionArray[0], 10, 64); err == nil {
					if v, ok := p.sessionMap.Load(id); ok {
						if s, ok := v.(*serverSession); ok && s != nil {
							if s.security == sessionArray[1] {
								session = s
							}
						}
					}
				}
			}
			// if session not find by session string, create a new session
			if session == nil {
				session = newServerSession(atomic.AddUint64(&p.sessionSeed, 1), p)
				p.sessionMap.Store(session.id, session)
			}
			// write respond stream
			stream.SetWritePosToBodyStart()
			stream.WriteInt64(controlStreamKindInitBack)
			stream.WriteString(fmt.Sprintf("%d-%s", session.id, session.security))
			stream.WriteInt64(int64(p.readTimeout / time.Millisecond))
			stream.WriteInt64(int64(p.writeTimeout / time.Millisecond))
			stream.WriteInt64(p.transportLimit)
			stream.WriteInt64(p.sessionConcurrency)
			if err := conn.WriteStream(stream, p.writeTimeout); err != nil {
				return nil, err
			}
			// return session
			return session, nil
		}
	}
}

func (p *baseServer) onConnRun(conn internal.IStreamConn, addr net.Addr) {
	runError := Error(nil)
	sessionId := uint64(0)

	defer func() {
		if runError != internal.ErrTransportStreamConnIsClosed {
			p.onError(sessionId, runError)
		}
		if err := conn.Close(); err != nil {
			p.onError(sessionId, err)
		}
	}()

	if session, err := p.getSession(conn); err != nil {
		runError = err
	} else {
		sessionId = session.id
		session.SetConn(conn)
		defer session.SetConn(nil)
		for runError == nil {
			if stream, err := conn.ReadStream(
				p.readTimeout,
				p.transportLimit,
			); err != nil {
				runError = err
			} else {
				cbID := stream.GetCallbackID()
				sequence := stream.GetSequence()
				if cbID == 0 && sequence == 0 {
					return
				} else if cbID == 0 {
					runError = session.OnControlStream(conn, stream)
				} else {
					runError = session.OnDataStream(conn, stream, p.hub)
				}
			}
		}
	}
}

func (p *baseServer) onError(sessionID uint64, err Error) {
	fmt.Println(sessionID, err)
}
