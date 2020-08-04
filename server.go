package rpc

import (
	"fmt"
	"github.com/rpccloud/rpc/internal"
	"path"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

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
	id          uint64
	server      *Server
	security    string
	conn        internal.IStreamConn
	dataSeed    uint64
	controlSeed uint64
	callMap     map[uint64]*serverSessionRecord
	sync.Mutex
}

var serverSessionCache = &sync.Pool{
	New: func() interface{} {
		return &serverSession{}
	},
}

func newServerSession(id uint64, server *Server) *serverSession {
	ret := serverSessionCache.Get().(*serverSession)
	ret.id = id
	ret.server = server
	ret.security = internal.GetRandString(32)
	ret.conn = nil
	ret.dataSeed = 0
	ret.controlSeed = 0
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
	if kind, ok := stream.ReadInt64(); !ok {
		return internal.NewTransportError(internal.ErrStringBadStream)
	} else if seq := stream.GetSequence(); seq <= p.controlSeed {
		return nil
	} else {
		p.controlSeed = seq
		switch kind {
		case controlStreamKindInit:
			stream.Reset()
			stream.SetCallbackID(0)
			stream.SetSequence(seq)
			stream.WriteInt64(controlStreamKindInitBack)
			stream.WriteString(fmt.Sprintf("%d-%s", p.id, p.security))
			stream.WriteInt64(int64(p.server.readTimeout / time.Millisecond))
			stream.WriteInt64(int64(p.server.writeTimeout / time.Millisecond))
			stream.WriteInt64(p.server.transportLimit)
			stream.WriteInt64(p.server.sessionMaxConcurrency)
			return conn.WriteStream(stream, p.server.writeTimeout)
		case controlStreamKindRequestIds:
			// get client currCallbackId
			currCallbackId, ok := stream.ReadUint64()
			if !ok {
				return internal.NewProtocolError(internal.ErrStringBadStream)
			}
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
				for count < p.server.sessionMaxConcurrency {
					p.dataSeed++
					p.callMap[p.dataSeed] = newServerSessionRecord(p.dataSeed)
					count++
				}
			}()
			// return stream
			stream.Reset()
			stream.SetCallbackID(0)
			stream.SetSequence(seq)
			stream.WriteInt64(controlStreamKindRequestIdsBack)
			stream.WriteUint64(p.dataSeed)
			return conn.WriteStream(stream, p.server.writeTimeout)
		default:
			return internal.NewProtocolError(internal.ErrStringBadStream)
		}
	}
}

func (p *serverSession) OnDataStream(
	conn internal.IStreamConn,
	stream *Stream,
	processor *internal.Processor,
) Error {
	if record, ok := p.callMap[stream.GetCallbackID()]; !ok {
		return internal.NewProtocolError("client callbackID error")
	} else if !record.SetRunning() {
		if stream := record.GetReturn(); stream != nil {
			return conn.WriteStream(stream, p.server.writeTimeout)
		}
		return nil
	} else {
		stream.GetCallbackID()
		stream.SetSessionID(p.id)
		processor.PutStream(stream)
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
		// mask panic message for client
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
				// report error to server
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
			if item, ok := p.callMap[stream.GetCallbackID()]; !ok {
				return p.conn, true
			} else if !item.SetReturn(stream) {
				return p.conn, true
			} else {
				return p.conn, false
			}
		}()

		if conn != nil {
			_ = conn.WriteStream(stream, p.server.writeTimeout)
		}

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
	p.security = ""
	p.dataSeed = 0
	p.controlSeed = 0
	p.server = nil
	serverSessionCache.Put(p)
}

const listenItemSchemeNone = uint32(0)
const listenItemSchemeTCP = uint32(1)
const listenItemSchemeUDP = uint32(2)
const listenItemSchemeHTTP = uint32(3)
const listenItemSchemeHTTPS = uint32(4)
const listenItemSchemeWS = uint32(5)
const listenItemSchemeWSS = uint32(6)

type listenItem struct {
	scheme   uint32
	addr     string
	certFile string
	keyFile  string
	fileLine string
}

type Server struct {
	isDebug               bool
	runSeed               uint64
	listens               []*listenItem
	adapters              []internal.IAdapter
	cacheDir              string
	processor             unsafe.Pointer
	numOfThreads          int
	sessionMap            sync.Map
	sessionMaxConcurrency int64
	sessionSeed           uint64
	fnCache               internal.ReplyCache
	services              []*internal.ServiceMeta
	transportLimit        int64
	writeLimit            int64
	readTimeout           time.Duration
	writeTimeout          time.Duration
	internal.Lock
}

func NewServer() *Server {
	return &Server{
		isDebug:               false,
		listens:               make([]*listenItem, 0),
		adapters:              nil,
		cacheDir:              "",
		processor:             nil,
		numOfThreads:          runtime.NumCPU() * 16384,
		sessionMap:            sync.Map{},
		sessionMaxConcurrency: 64,
		sessionSeed:           0,
		transportLimit:        int64(1024 * 1024),
		writeLimit:            int64(1024 * 1024),
		readTimeout:           10 * time.Second,
		writeTimeout:          1 * time.Second,
		fnCache:               nil,
	}
}

// AddChildService ...
func (p *Server) AddService(
	name string,
	service *Service,
) *Server {
	fileLine := internal.GetFileLine(1)
	p.DoWithLock(func() {
		if atomic.LoadPointer(&p.processor) == nil {
			p.services = append(p.services, internal.NewServiceMeta(
				name,
				service,
				fileLine,
			))
		} else {
			p.onError(internal.NewRuntimePanic(
				"AddService must be before Serve",
			).AddDebug(fileLine))
		}
	})
	return p
}

// BuildReplyCache ...
func (p *Server) BuildReplyCache() *Server {
	_, file, _, _ := runtime.Caller(1)
	fileLine := internal.GetFileLine(1)
	p.DoWithLock(func() {
		if atomic.LoadPointer(&p.processor) == nil {
			p.cacheDir = path.Join(path.Dir(file))
		} else {
			p.onError(internal.NewRuntimePanic(
				"ListenWebSocket must be before Serve",
			).AddDebug(fileLine))
		}
	})
	return p
}

// ListenWebSocket ...
func (p *Server) ListenWebSocket(addr string) *Server {
	fileLine := internal.GetFileLine(1)
	p.DoWithLock(func() {
		if atomic.LoadPointer(&p.processor) == nil {
			p.listens = append(p.listens, &listenItem{
				scheme:   listenItemSchemeWS,
				addr:     addr,
				certFile: "",
				keyFile:  "",
				fileLine: fileLine,
			})
		} else {
			p.onError(internal.NewRuntimePanic(
				"ListenWebSocket must be before Serve",
			).AddDebug(fileLine))
		}
	})
	return p
}

func (p *Server) OnReturnStream(stream *Stream) {
	if item, ok := p.sessionMap.Load(stream.GetSessionID()); !ok {
		stream.Release()
	} else if session, ok := item.(*serverSession); !ok {
		stream.Release()
		p.onSessionError(stream.GetSessionID(), internal.NewKernelPanic(
			"serverSession is nil",
		).AddDebug(string(debug.Stack())))
	} else {
		if err := session.OnReturnStream(stream); err != nil {
			p.onSessionError(stream.GetSessionID(), err)
		}
	}
}

func (p *Server) Serve() {
	waitCount := 0
	waitCH := make(chan struct{})

	p.DoWithLock(func() {
		p.adapters = make([]internal.IAdapter, 0)
		for _, listener := range p.listens {
			switch listener.scheme {
			case listenItemSchemeWS:
				p.adapters = append(
					p.adapters,
					internal.NewWebSocketServerAdapter(listener.addr),
				)
			default:
			}
		}

		if len(p.adapters) <= 0 {
			p.onError(internal.NewRuntimePanic(
				"no valid listener was found on the server",
			))
		} else if processor := internal.NewProcessor(
			p.isDebug,
			p.numOfThreads,
			32,
			32,
			p.fnCache,
			20*time.Second,
			p.services,
			p.OnReturnStream,
		); processor == nil {
			// ignore
		} else if !atomic.CompareAndSwapPointer(
			&p.processor,
			nil,
			unsafe.Pointer(processor),
		) {
			processor.Close()
			p.onError(internal.NewRuntimePanic("it is already running"))
		} else {
			for _, item := range p.adapters {
				waitCount++
				go func(serverAdapter internal.IAdapter) {
					for atomic.CompareAndSwapPointer(
						&p.processor,
						unsafe.Pointer(processor),
						unsafe.Pointer(processor),
					) {
						serverAdapter.Open(p.onConnRun, p.onError)
					}
					waitCH <- struct{}{}
				}(item)
			}
		}
	})

	for i := 0; i < waitCount; i++ {
		<-waitCH
	}
}

func (p *Server) Close() {
	p.DoWithLock(func() {
		processor := atomic.LoadPointer(&p.processor)
		if processor == nil {
			p.onError(internal.NewRuntimePanic("it is not running"))
		} else if !atomic.CompareAndSwapPointer(
			&p.processor,
			unsafe.Pointer(processor),
			nil,
		) {
			p.onError(internal.NewRuntimePanic("it is not running"))
		} else {
			for _, item := range p.adapters {
				go func(adapter internal.IAdapter) {
					adapter.Close(p.onError)
				}(item)
			}
		}
	})
}

func (p *Server) getSession(conn internal.IStreamConn) (*serverSession, Error) {
	if conn == nil {
		return nil, internal.NewKernelPanic(
			"Server: getSession: conn is nil",
		)
	} else if stream, err := conn.ReadStream(
		p.readTimeout,
		p.transportLimit,
	); err != nil {
		return nil, err
	} else if stream.GetCallbackID() != 0 {
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
		session := (*serverSession)(nil)
		// try to find session by session string
		sessionArray := strings.Split(sessionString, "-")
		if len(sessionArray) == 2 && len(sessionArray[1]) == 32 {
			if id, err := strconv.ParseUint(
				sessionArray[0],
				10,
				64,
			); err == nil && id > 0 {
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

		// set the session conn
		session.conn = conn

		// Set stream read pos to start
		stream.SetReadPosToBodyStart()
		if err := session.OnControlStream(conn, stream); err != nil {
			return nil, err
		} else {
			return session, nil
		}
	}
}

func (p *Server) onConnRun(conn internal.IStreamConn) {
	if conn == nil {
		p.onError(internal.NewKernelPanic("Server: onConnRun: conn is nil"))
	} else if session, err := p.getSession(conn); err != nil {
		p.onError(err)
	} else {
		defer func() {
			session.conn = nil
			if err := conn.Close(); err != nil {
				p.onError(err)
			}
		}()

		processor := (*internal.Processor)(p.processor)

		for {
			if stream, err := conn.ReadStream(
				p.readTimeout,
				p.transportLimit,
			); err != nil {
				if err != internal.ErrTransportStreamConnIsClosed {
					p.onError(err)
				}
				return
			} else {
				cbID := stream.GetCallbackID()
				sequence := stream.GetSequence()

				if cbID == 0 && sequence == 0 {
					return
				} else if cbID == 0 {
					if err := session.OnControlStream(conn, stream); err != nil {
						p.onError(err)
						return
					}
				} else {
					if err := session.OnDataStream(conn, stream, processor); err != nil {
						p.onError(err)
						return
					}
				}
			}
		}
	}
}

func (p *Server) onError(err Error) {
	p.onSessionError(0, err)
}

func (p *Server) onSessionError(sessionID uint64, err Error) {
	fmt.Println(sessionID, err)
}

// End ***** Server ***** //
