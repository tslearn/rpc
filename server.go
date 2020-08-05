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
	server       *Server
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

func newServerSession(id uint64, server *Server) *serverSession {
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
	processor *internal.Processor,
) Error {
	if record, ok := p.callMap[stream.GetCallbackID()]; !ok {
		// Cant find record by callbackID
		stream.Release()
		return internal.NewProtocolError("client callbackID error")
	} else if record.SetRunning() {
		// Run the stream. Dont release stream because it will manage by processor
		stream.SetSessionID(p.id)
		processor.PutStream(stream)
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
	p.security = ""
	p.dataSequence = 0
	p.ctrlSequence = 0
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
	isDebug            bool
	listens            []*listenItem
	adapters           []internal.IAdapter
	processor          *internal.Processor
	numOfThreads       int
	sessionMap         sync.Map
	sessionConcurrency int64
	sessionSeed        uint64
	services           []*internal.ServiceMeta
	transportLimit     int64
	readTimeout        time.Duration
	writeTimeout       time.Duration
	replyCache         internal.ReplyCache
	internal.StatusManager
	sync.Mutex
}

func NewServer() *Server {
	return &Server{
		isDebug:            false,
		listens:            make([]*listenItem, 0),
		adapters:           nil,
		processor:          nil,
		numOfThreads:       runtime.NumCPU() * 8192,
		sessionMap:         sync.Map{},
		sessionConcurrency: 64,
		sessionSeed:        0,
		transportLimit:     1024 * 1024,
		readTimeout:        10 * time.Second,
		writeTimeout:       1 * time.Second,
		replyCache:         nil,
	}
}

// SetDebug ...
func (p *Server) SetDebug() *Server {
	p.Lock()
	defer p.Unlock()

	if p.IsRunning() {
		p.onError(internal.NewRuntimePanic(
			"SetDebug must be called before Serve",
		).AddDebug(string(debug.Stack())))
	} else {
		p.isDebug = true
	}

	return p
}

// SetRelease ...
func (p *Server) SetRelease() *Server {
	p.Lock()
	defer p.Unlock()

	if p.IsRunning() {
		p.onError(internal.NewRuntimePanic(
			"SetRelease must be called before Serve",
		).AddDebug(string(debug.Stack())))
	} else {
		p.isDebug = false
	}

	return p
}

// SetNumOfThreads ...
func (p *Server) SetNumOfThreads(numOfThreads int) *Server {
	p.Lock()
	defer p.Unlock()

	if numOfThreads <= 0 {
		p.onError(internal.NewRuntimePanic(
			"numOfThreads must be greater than 0",
		).AddDebug(string(debug.Stack())))
	} else if p.IsRunning() {
		p.onError(internal.NewRuntimePanic(
			"SetNumOfThreads must be called before Serve",
		).AddDebug(string(debug.Stack())))
	} else {
		p.numOfThreads = numOfThreads
	}

	return p
}

// SetTransportLimit ...
func (p *Server) SetTransportLimit(maxTransportBytes int) *Server {
	p.Lock()
	defer p.Unlock()

	if maxTransportBytes < minTransportLimit {
		p.onError(internal.NewRuntimePanic(fmt.Sprintf(
			"maxTransportBytes must be greater than or equal to %d",
			minTransportLimit,
		)).AddDebug(string(debug.Stack())))
	} else if p.IsRunning() {
		p.onError(internal.NewRuntimePanic(
			"SetTransportLimit must be called before Serve",
		).AddDebug(string(debug.Stack())))
	} else {
		p.transportLimit = int64(maxTransportBytes)
	}

	return p
}

func (p *Server) SetSessionConcurrency(sessionConcurrency int) *Server {
	p.Lock()
	defer p.Unlock()

	if sessionConcurrency > maxSessionConcurrency {
		p.onError(internal.NewRuntimePanic(fmt.Sprintf(
			"sessionConcurrency be less than or equal to %d",
			maxSessionConcurrency,
		)).AddDebug(string(debug.Stack())))
	} else if p.IsRunning() {
		p.onError(internal.NewRuntimePanic(
			"SetSessionConcurrency must be called before Serve",
		).AddDebug(string(debug.Stack())))
	} else {
		p.sessionConcurrency = int64(sessionConcurrency)
	}

	return p
}

// SetReplyCache ...
func (p *Server) SetReplyCache(replyCache internal.ReplyCache) *Server {
	p.Lock()
	defer p.Unlock()

	if p.IsRunning() {
		p.onError(internal.NewRuntimePanic(
			"SetReplyCache must be called before Serve",
		).AddDebug(string(debug.Stack())))
	} else {
		p.replyCache = replyCache
	}

	return p
}

// AddService ...
func (p *Server) AddService(
	name string,
	service *Service,
) *Server {
	p.Lock()
	defer p.Unlock()

	fileLine := internal.GetFileLine(1)
	if p.IsRunning() {
		p.onError(internal.NewRuntimePanic(
			"AddService must be called before Serve",
		).AddDebug(fileLine))
	} else {
		p.services = append(p.services, internal.NewServiceMeta(
			name,
			service,
			fileLine,
		))
	}

	return p
}

// ListenWebSocket ...
func (p *Server) ListenWebSocket(addr string) *Server {
	p.Lock()
	defer p.Unlock()

	fileLine := internal.GetFileLine(1)
	if p.IsRunning() {
		p.onError(internal.NewRuntimePanic(
			"ListenWebSocket must be called before Serve",
		).AddDebug(fileLine))
	} else {
		p.listens = append(p.listens, &listenItem{
			scheme:   listenItemSchemeWS,
			addr:     addr,
			certFile: "",
			keyFile:  "",
			fileLine: fileLine,
		})
	}

	return p
}

// BuildReplyCache ...
func (p *Server) BuildReplyCache() *Server {
	_, file, _, _ := runtime.Caller(1)
	buildDir := path.Join(path.Dir(file))

	services := func() []*internal.ServiceMeta {
		p.Lock()
		defer p.Unlock()
		return p.services
	}()

	processor := internal.NewProcessor(
		p.isDebug,
		1,
		32,
		32,
		nil,
		time.Second,
		services,
		func(stream *internal.Stream) {},
	)
	defer processor.Close()

	if err := processor.BuildCache(
		"cache",
		path.Join(buildDir, "cache", "reply_cache.go"),
	); err != nil {
		p.onError(err)
	}

	return p
}

func (p *Server) OnReturnStream(stream *internal.Stream) {
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
	waitCH := make(chan bool)

	func() {
		p.Lock()
		defer p.Unlock()

		// create adapters by listens
		adapters := make([]internal.IAdapter, 0)
		for _, ln := range p.listens {
			switch ln.scheme {
			case listenItemSchemeWS:
				adapters = append(
					adapters,
					internal.NewWebSocketServerAdapter(ln.addr),
				)
			default:
			}
		}

		if len(adapters) <= 0 {
			p.onError(internal.NewRuntimePanic(
				"no valid listener was found on the server",
			))
		} else if processor := internal.NewProcessor(
			p.isDebug,
			p.numOfThreads,
			32,
			32,
			p.replyCache,
			20*time.Second,
			p.services,
			p.OnReturnStream,
		); processor == nil {
			p.onError(internal.NewKernelPanic(
				"processor is nil",
			).AddDebug(string(debug.Stack())))
		} else if !p.SetRunning(func() {
			p.adapters = adapters
			p.processor = processor
		}) {
			processor.Close()
			p.onError(internal.NewRuntimePanic("it is already running"))
		} else {
			for _, item := range p.adapters {
				waitCount++
				go func(serverAdapter internal.IAdapter) {
					for p.IsRunning() {
						serverAdapter.Open(p.onConnRun, p.onError)
					}
					waitCH <- true
				}(item)
			}
		}
	}()

	for i := 0; i < waitCount; i++ {
		<-waitCH
	}

	p.SetClosed(func() {
		p.adapters = nil
		p.processor = nil
	})
}

func (p *Server) Close() {
	waitCH := chan bool(nil)

	if !p.SetClosing(func(ch chan bool) {
		waitCH = ch
		p.processor.Close()
		for _, item := range p.adapters {
			go func(adapter internal.IAdapter) {
				adapter.Close(p.onError)
			}(item)
		}
	}) {
		p.onError(internal.NewKernelPanic(
			"it is not running",
		).AddDebug(string(debug.Stack())))
	} else {
		select {
		case <-waitCH:
		case <-time.After(5 * time.Second):
			p.onError(internal.NewRuntimePanic(
				"it cannot be closed within 5 seconds",
			).AddDebug(string(debug.Stack())))
		}
	}
}

func (p *Server) getSession(conn internal.IStreamConn) (*serverSession, Error) {
	if conn == nil {
		return nil, internal.NewKernelPanic(
			"conn is nil",
		).AddDebug(string(debug.Stack()))
	} else if stream, err := conn.ReadStream(
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
			session.SetConn(conn)
			return session, nil
		}
	}
}

func (p *Server) onConnRun(conn internal.IStreamConn) {
	runError := Error(nil)

	if session, err := p.getSession(conn); err != nil {
		runError = err
	} else {
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
					runError = session.OnDataStream(conn, stream, p.processor)
				}
			}
		}
	}

	if runError != internal.ErrTransportStreamConnIsClosed {
		p.onError(runError)
	}
}

func (p *Server) onError(err Error) {
	p.onSessionError(0, err)
}

func (p *Server) onSessionError(sessionID uint64, err Error) {
	fmt.Println(sessionID, err)
}

// End ***** Server ***** //
