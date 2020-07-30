package rpc

import (
	"fmt"
	"github.com/rpccloud/rpc/internal"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Begin ***** serverSessionRecord ***** //
const serverSessionRecordStatusNone = 0
const serverSessionRecordStatusRunning = 1
const serverSessionRecordStatusBack = 2
const serverSessionRecordStatusClosed = 3

type serverSessionRecord struct {
	id     uint64
	status int32
	mark   bool
	stream *Stream
}

var serverSessionRecordCache = &sync.Pool{
	New: func() interface{} {
		return &serverSessionRecord{
			id:     0,
			status: 0,
			mark:   false,
			stream: nil,
		}
	},
}

func newServerSessionRecord(id uint64) *serverSessionRecord {
	ret := serverSessionRecordCache.Get().(*serverSessionRecord)
	ret.id = id
	atomic.StoreInt32(&ret.status, serverSessionRecordStatusNone)
	ret.mark = false
	return ret
}

func (p *serverSessionRecord) SetRunning() bool {
	return atomic.CompareAndSwapInt32(
		&p.status,
		serverSessionRecordStatusNone,
		serverSessionRecordStatusRunning,
	)
}

func (p *serverSessionRecord) BackStream(stream *Stream) {
	if atomic.CompareAndSwapInt32(
		&p.status,
		serverSessionRecordStatusRunning,
		serverSessionRecordStatusBack,
	) {
		p.stream = stream
	}
}

func (p *serverSessionRecord) Release() {
	atomic.StoreInt32(&p.status, serverSessionRecordStatusClosed)

	if p.stream != nil {
		p.stream.Release()
		p.stream = nil
	}
	serverSessionRecordCache.Put(p)
}

// End ***** serverSessionRecord ***** //

// Begin ***** serverSession ***** //
type serverSession struct {
	id           uint64
	security     string
	conn         internal.IStreamConn
	dataSeed     uint64
	controlSeed  uint64
	callMap      map[uint64]*serverSessionRecord
	readLimit    int64
	writeLimit   int64
	readTimeout  time.Duration
	writeTimeout time.Duration
	maxStreams   int64
	internal.Lock
}

var serverSessionCache = &sync.Pool{
	New: func() interface{} {
		return &serverSession{}
	},
}

func newServerSession(
	id uint64,
	maxStreams int64,
	readLimit int64,
	writeLimit int64,
	readTimeout time.Duration,
	writeTimeout time.Duration,
) *serverSession {
	ret := serverSessionCache.Get().(*serverSession)
	ret.id = id
	ret.security = internal.GetRandString(32)
	ret.dataSeed = 0
	ret.controlSeed = 0
	ret.callMap = make(map[uint64]*serverSessionRecord)
	ret.readLimit = readLimit
	ret.writeLimit = writeLimit
	ret.readTimeout = readTimeout
	ret.writeTimeout = writeTimeout
	ret.maxStreams = maxStreams
	return ret
}

func (p *serverSession) WriteStream(stream *Stream) Error {
	return internal.ConvertToError(p.CallWithLock(func() interface{} {
		if p.conn != nil {
			return p.conn.WriteStream(
				stream,
				p.writeTimeout,
			)
		} else {
			return internal.NewBaseError(
				"serverSession: WriteStream: conn is nil",
			)
		}
	}))
}

func (p *serverSession) OnDataStream(
	stream *Stream,
	processor *internal.Processor,
) Error {
	if stream == nil {
		return internal.NewBaseError("stream is nil")
	}

	if processor == nil {
		return internal.NewBaseError(
			"serverSession: OnDataStream: processor is nil",
		)
	}

	record, ok := p.callMap[stream.GetCallbackID()]

	if !ok {
		return internal.NewBaseError(
			"serverSession: OnDataStream: stream callbackID error",
		)
	}

	if !record.SetRunning() {
		// it not error, it is just redundant
		return nil
	}

	stream.GetCallbackID()
	stream.SetSessionID(p.id)

	if !processor.PutStream(stream) {
		return internal.NewBaseError(
			"serverSession: OnDataStream: processor can not deal with stream",
		)
	}

	return nil
}

func (p *serverSession) OnControlStream(stream *Stream) Error {
	return internal.ConvertToError(p.CallWithLock(func() interface{} {
		if stream == nil {
			return internal.NewBaseError(
				"Server: OnControlStream: stream is nil",
			)
		}
		defer stream.Release()

		if p.conn == nil {
			return internal.NewBaseError(
				"Server: OnControlStream: conn is nil",
			)
		}

		controlSequence := stream.GetSequence()
		if controlSequence <= p.controlSeed {
			return internal.NewBaseError(
				"Server: OnControlStream: sequence is omit",
			)
		}
		p.controlSeed = controlSequence

		kind, ok := stream.ReadInt64()
		if !ok {
			return internal.NewBaseError(
				"Server: OnControlStream: stream format error",
			)
		}

		switch kind {
		case SystemStreamKindInit:
			stream.Reset()
			stream.SetCallbackID(0)
			stream.SetSequence(controlSequence)
			stream.WriteInt64(SystemStreamKindInitBack)
			stream.WriteString(fmt.Sprintf("%d-%s", p.id, p.security))
			stream.WriteInt64(int64(p.readTimeout / time.Millisecond))
			stream.WriteInt64(int64(p.writeTimeout / time.Millisecond))
			stream.WriteInt64(p.writeLimit)
			stream.WriteInt64(p.readLimit)
			stream.WriteInt64(p.maxStreams)
			return p.conn.WriteStream(stream, p.writeTimeout)
		case SystemStreamKindRequestIds:
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
			for count < p.maxStreams {
				p.dataSeed++
				p.callMap[p.dataSeed] = newServerSessionRecord(p.dataSeed)
				count++
			}
			// return stream
			stream.Reset()
			stream.SetCallbackID(0)
			stream.SetSequence(controlSequence)
			stream.WriteInt64(SystemStreamKindRequestIdsBack)
			stream.WriteUint64(p.dataSeed)
			return p.conn.WriteStream(stream, time.Second)
		default:
			return internal.NewProtocolError(internal.ErrStringBadStream)
		}
	}))
}

func (p *serverSession) Release() {
	p.DoWithLock(func() {
		if p.callMap != nil {
			for _, v := range p.callMap {
				v.Release()
			}
			p.callMap = nil
		}
		p.conn = nil
	})

	p.id = 0
	p.security = ""
	p.dataSeed = 0
	p.controlSeed = 0
	p.maxStreams = 0
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
	isDebug      bool
	runSeed      uint64
	listens      []*listenItem
	adapters     []internal.IAdapter
	cacheDir     string
	processor    unsafe.Pointer
	numOfThreads int
	sessionMap   sync.Map
	sessionSize  int64
	sessionSeed  uint64
	fnCache      internal.ReplyCache
	services     []*internal.ServiceMeta
	readLimit    int64
	writeLimit   int64
	readTimeout  time.Duration
	writeTimeout time.Duration
	internal.Lock
}

func NewServer() *Server {
	return &Server{
		isDebug:      false,
		listens:      make([]*listenItem, 0),
		adapters:     nil,
		cacheDir:     "",
		processor:    nil,
		numOfThreads: runtime.NumCPU() * 16384,
		sessionMap:   sync.Map{},
		sessionSize:  64,
		sessionSeed:  0,
		readLimit:    int64(1024 * 1024),
		writeLimit:   int64(1024 * 1024),
		readTimeout:  10 * time.Second,
		writeTimeout: 1 * time.Second,
		fnCache:      nil,
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
	if stream != nil {
		sessionID := stream.GetSessionID()
		stream.SetReadPosToBodyStart()

		if errKind, ok := stream.ReadUint64(); ok {
			switch internal.ErrorKind(errKind) {
			case internal.ErrorKindNone:
				fallthrough
			case internal.ErrorKindProtocol:
				fallthrough
			case internal.ErrorKindTransport:
				fallthrough
			case internal.ErrorKindReply:
				if v, ok := p.sessionMap.Load(sessionID); ok {
					if session, ok := v.(*serverSession); ok && session != nil {
						if err := session.WriteStream(stream); err != nil {
							p.onSessionError(stream.GetSessionID(), err)
						}
					}
				}
			case internal.ErrorKindReplyPanic:
				fallthrough
			case internal.ErrorKindRuntimePanic:
				fallthrough
			case internal.ErrorKindKernelPanic:
				if message, ok := stream.ReadString(); !ok {
					// stream.SetReadPosToBodyStart()
				} else if debug, ok := stream.ReadString(); !ok {
					// stream.SetReadPosToBodyStart()
				} else {
					p.onSessionError(stream.GetSessionID(), internal.NewError(
						internal.ErrorKind(errKind),
						message,
						debug,
					))
				}
			}
		}
		stream.Release()
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
			p.onError(internal.NewBaseError("it is not running"))
		} else if !atomic.CompareAndSwapPointer(
			&p.processor,
			unsafe.Pointer(processor),
			nil,
		) {
			p.onError(internal.NewBaseError("it is not running"))
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
		return nil, internal.NewBaseError(
			"Server: getSession: conn is nil",
		)
	} else if stream, err := conn.ReadStream(
		p.readTimeout,
		p.readLimit,
	); err != nil {
		return nil, err
	} else if stream.GetCallbackID() != 0 {
		return nil, internal.NewProtocolError(internal.ErrStringBadStream)
	} else if stream.GetSequence() == 0 {
		return nil, internal.NewProtocolError(internal.ErrStringBadStream)
	} else if kind, ok := stream.ReadInt64(); !ok ||
		kind != SystemStreamKindInit {
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
			session = newServerSession(
				atomic.AddUint64(&p.sessionSeed, 1),
				p.sessionSize,
				p.readLimit,
				p.writeLimit,
				p.readTimeout,
				p.writeTimeout,
			)
			p.sessionMap.Store(session.id, session)
		}

		// set the session conn
		session.conn = conn

		// Set stream read pos to start
		stream.SetReadPosToBodyStart()
		if err := session.OnControlStream(stream); err != nil {
			return nil, err
		} else {
			return session, nil
		}
	}
}

func (p *Server) onConnRun(conn internal.IStreamConn) {
	if conn == nil {
		p.onError(internal.NewBaseError("Server: onConnRun: conn is nil"))
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
				p.readLimit,
			); err != nil {
				p.onError(err)
				return
			} else {
				cbID := stream.GetCallbackID()
				sequence := stream.GetSequence()

				if cbID == 0 && sequence == 0 {
					return
				} else if cbID == 0 {
					if err := session.OnControlStream(stream); err != nil {
						p.onError(err)
						return
					}
				} else {
					if err := session.OnDataStream(stream, processor); err != nil {
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
