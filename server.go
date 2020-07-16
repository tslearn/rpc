package rpcc

import (
	"fmt"
	"github.com/tslearn/rpcc/internal"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
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
	stream Stream
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

func (p *serverSessionRecord) BackStream(stream Stream) {
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
	id          uint64
	security    string
	conn        IStreamConnection
	dataSeed    uint64
	controlSeed uint64
	callMap     map[uint64]*serverSessionRecord
	size        int64
	internal.Lock
}

var serverSessionCache = &sync.Pool{
	New: func() interface{} {
		return &serverSession{
			id:          0,
			security:    "",
			conn:        nil,
			dataSeed:    0,
			controlSeed: 0,
			callMap:     nil,
			size:        0,
		}
	},
}

func newServerSession(id uint64, size int64) *serverSession {
	ret := serverSessionCache.Get().(*serverSession)
	ret.id = id
	ret.security = internal.GetRandString(32)
	ret.dataSeed = 0
	ret.controlSeed = 0
	ret.callMap = make(map[uint64]*serverSessionRecord)
	ret.size = size
	return ret
}

func (p *serverSession) WriteStream(stream Stream) Error {
	return internal.ConvertToError(p.CallWithLock(func() interface{} {
		if p.conn != nil {
			return p.conn.WriteStream(
				stream,
				configWriteTimeout,
				configServerWriteLimit,
			)
		} else {
			return internal.NewError(
				"serverSession: WriteStream: conn is nil",
			)
		}
	}))
}

func (p *serverSession) OnDataStream(
	stream Stream,
	processor *internal.Processor,
) Error {
	if stream == nil {
		return internal.NewError(
			"serverSession: OnDataStream: stream is nil",
		)
	}

	if processor == nil {
		return internal.NewError(
			"serverSession: OnDataStream: processor is nil",
		)
	}

	record, ok := p.callMap[stream.GetCallbackID()]

	if !ok {
		return internal.NewError(
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
		return internal.NewError(
			"serverSession: OnDataStream: processor can not deal with stream",
		)
	}

	return nil
}

func (p *serverSession) OnControlStream(
	stream Stream,
) Error {
	ret := internal.ConvertToError(p.CallWithLock(func() interface{} {
		if stream == nil {
			return internal.NewError(
				"Server: OnControlStream: stream is nil",
			)
		}

		if p.conn == nil {
			return internal.NewError(
				"Server: OnControlStream: conn is nil",
			)
		}

		controlSequence := stream.GetSequence()
		if controlSequence <= p.controlSeed {
			return internal.NewError(
				"Server: OnControlStream: sequence is omit",
			)
		}
		p.controlSeed = controlSequence

		kind, ok := stream.ReadInt64()
		if !ok {
			return internal.NewError(
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
			stream.WriteInt64(int64(configReadTimeout / time.Millisecond))
			stream.WriteInt64(int64(configWriteTimeout / time.Millisecond))
			stream.WriteInt64(configServerWriteLimit)
			stream.WriteInt64(configServerReadLimit)
			stream.WriteInt64(p.size)
			return p.conn.WriteStream(
				stream,
				configWriteTimeout,
				configServerWriteLimit,
			)
		case SystemStreamKindRequestIds:
			currCallbackId, ok := stream.ReadUint64()
			if !ok {
				return internal.NewError(
					"Server: OnControlStream: stream format error",
				)
			}

			// mark
			for stream.CanRead() {
				if markId, ok := stream.ReadUint64(); ok {
					if v, ok := p.callMap[markId]; ok {
						v.mark = true
					}
				} else {
					return internal.NewError(
						"Server: OnControlStream: stream format error",
					)
				}
			}
			if !stream.IsReadFinish() {
				return internal.NewError(
					"Server: OnControlStream: stream format error",
				)
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
			for count < p.size {
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
			return p.conn.WriteStream(
				stream,
				configWriteTimeout,
				configServerWriteLimit,
			)
		default:
			return internal.NewError(
				"Server: OnControlStream: stream format error",
			)
		}
	}))

	if stream != nil {
		stream.Release()
	}

	return ret
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
	p.size = 0
	serverSessionCache.Put(p)
}

// End ***** serverSession ***** //

// Begin ***** Server ***** //
type Server struct {
	isOpen      bool
	logger      *internal.Logger
	endPoints   []IAdapter
	processor   *internal.Processor
	sessionMap  sync.Map
	sessionSize int64
	sessionSeed uint64
	internal.Lock
}

func NewServer(isDebug bool, numOfThreads uint, sessionSize int64, fnCache internal.ReplyCache) *Server {
	server := &Server{
		isOpen:      false,
		logger:      internal.NewLogger(nil),
		endPoints:   make([]IAdapter, 0),
		processor:   nil,
		sessionMap:  sync.Map{},
		sessionSize: sessionSize,
		sessionSeed: 0,
	}

	server.processor = internal.NewProcessor(
		isDebug,
		numOfThreads,
		32,
		32,
		fnCache,
	)

	return server
}

func (p *Server) Start() bool {
	return p.CallWithLock(func() interface{} {
		if p.isOpen {
			p.onError(internal.NewError("Server: Start: it is already opened"))
			return false
		} else if err := p.processor.Start(
			func(stream Stream, success bool) {
				if v, ok := p.sessionMap.Load(stream.GetSessionID()); ok {
					if session, ok := v.(*serverSession); ok && session != nil {
						if err := session.WriteStream(stream); err != nil {
							p.logger.Error(err.Error())
						}
					}
				}
			},
			func(v interface{}, debug string) {

			},
		); err != nil {
			p.onError(err)
			return false
		} else {
			openList := make([]IAdapter, 0)
			defer func() {
				if openList != nil {
					for _, v := range openList {
						v.Close(p.onError)
					}
				}
			}()
			for _, endPoint := range p.endPoints {
				if endPoint.Open(p.onConnRun, p.onError) {
					openList = append(openList, endPoint)
				} else {
					return false
				}
			}
			openList = nil
			p.isOpen = true
			return true
		}
	}).(bool)
}

func (p *Server) Stop() {
	p.DoWithLock(func() {
		if !p.isOpen {
			p.onError(internal.NewError("Server: Stop: it is not opened"))
		} else {
			p.isOpen = false

			for _, endPoint := range p.endPoints {
				endPoint.Close(p.onError)
			}

			if err := p.processor.Stop(); err != nil {
				p.onError(err)
			}
		}
	})
}

func (p *Server) GetLogger() *internal.Logger {
	return p.logger
}

func (p *Server) BuildFuncCache(
	pkgName string,
	relativePath string,
) Error {
	_, file, _, _ := runtime.Caller(1)
	return p.processor.BuildCache(
		pkgName,
		path.Join(path.Dir(file), relativePath),
	)
}

// AddChild ...
func (p *Server) AddService(
	name string,
	service Service,
) *Server {
	if err := p.processor.AddService(
		name,
		service,
		internal.GetStackString(1),
	); err != nil {
		p.onError(err)
	}
	return p
}

func (p *Server) AddAdapter(endPoint IAdapter) *Server {
	if endPoint == nil {
		p.onError(internal.NewError("Server: AddAdapter: endpoint is nil"))
	} else if endPoint.IsRunning() {
		p.onError(internal.NewError(fmt.Sprintf(
			"Server: AddAdapter: endpoint %s has already served",
			endPoint.ConnectString(),
		)))
	} else {
		p.DoWithLock(func() {
			p.endPoints = append(p.endPoints, endPoint)
			if p.isOpen {
				endPoint.Open(p.onConnRun, p.onError)
			}
		})
	}

	return p
}

func (p *Server) getSession(conn IStreamConnection) (*serverSession, Error) {
	if conn == nil {
		return nil, internal.NewError(
			"Server: getSession: conn is nil",
		)
	} else if stream, err := conn.ReadStream(
		configReadTimeout,
		configServerReadLimit,
	); err != nil {
		return nil, err
	} else if stream.GetCallbackID() != 0 {
		return nil, internal.NewError(
			"Server: getSession: stream format error",
		)
	} else if stream.GetSequence() == 0 {
		return nil, internal.NewError(
			"Server: getSession: stream format error",
		)
	} else if kind, ok := stream.ReadInt64(); !ok ||
		kind != SystemStreamKindInit {
		return nil, internal.NewError(
			"Server: getSession: stream format error",
		)
	} else if sessionString, ok := stream.ReadString(); !ok {
		return nil, internal.NewError(
			"Server: getSession: stream format error",
		)
	} else if !stream.IsReadFinish() {
		return nil, internal.NewError(
			"Server: getSession: stream format error",
		)
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
			)
			p.sessionMap.Store(session.id, session)
		}

		// set the session conn
		session.conn = conn

		// Set stream read pos to start
		stream.SetReadPos(internal.StreamBodyPos)
		if err := session.OnControlStream(stream); err != nil {
			return nil, err
		} else {
			return session, nil
		}
	}
}

func (p *Server) onConnRun(conn IStreamConnection) {
	if conn == nil {
		p.onError(internal.NewError("Server: onConnRun: conn is nil"))
	} else if session, err := p.getSession(conn); err != nil {
		p.onError(err)
	} else {
		defer func() {
			session.conn = nil
			if err := conn.Close(); err != nil {
				p.onError(err)
			}
		}()

		for {
			if stream, err := conn.ReadStream(
				configReadTimeout,
				configServerReadLimit,
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
					if err := session.OnDataStream(stream, p.processor); err != nil {
						p.onError(err)
						return
					}
				}
			}
		}
	}
}

func (p *Server) onError(err Error) {
	p.logger.Error(err.Error())
	//fmt.Println(err.Error())
}

// End ***** Server ***** //
