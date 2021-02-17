// Package server ...
package server

import (
	"crypto/tls"
	"log"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/gateway"
	"github.com/rpccloud/rpc/internal/route"
)

const (
	defaultMaxNumOfThreads  = 1024 * 1024
	defaultThreadsPerCPU    = 16384
	defaultThreadBufferSize = 2048
	defaultCloseTimeout     = 5 * time.Second
	defaultMaxNodeDepth     = 128
	defaultMaxCallDepth     = 128
)

var fnNumCPU = runtime.NumCPU

// Server ...
type Server struct {
	isRunning        bool
	processor        *core.Processor
	router           route.IRouter
	gateway          *gateway.GateWay
	numOfThreads     int
	maxNodeDepth     int16
	maxCallDepth     int16
	threadBufferSize uint32
	actionCache      core.ActionCache
	closeTimeout     time.Duration
	mountServices    []*core.ServiceMeta
	logHub           core.IStreamReceiver
	sync.Mutex
}

// NewServer ...
func NewServer() *Server {
	ret := &Server{
		isRunning:        false,
		processor:        nil,
		router:           route.NewDirectRouter(),
		gateway:          nil,
		numOfThreads:     fnNumCPU() * defaultThreadsPerCPU,
		maxNodeDepth:     defaultMaxNodeDepth,
		maxCallDepth:     defaultMaxCallDepth,
		threadBufferSize: defaultThreadBufferSize,
		actionCache:      nil,
		closeTimeout:     defaultCloseTimeout,
		mountServices:    make([]*core.ServiceMeta, 0),
	}

	if ret.numOfThreads > defaultMaxNumOfThreads {
		ret.numOfThreads = defaultMaxNumOfThreads
	}

	ret.gateway = gateway.NewGateWay(
		0,
		gateway.GetDefaultConfig(),
		ret,
	)

	return ret
}

// Listen ...
func (p *Server) Listen(
	network string,
	addr string,
	tlsConfig *tls.Config,
) *Server {
	p.gateway.Listen(network, addr, tlsConfig)
	return p
}

// ListenWithDebug ...
func (p *Server) ListenWithDebug(
	network string,
	addr string,
	tlsConfig *tls.Config,
) *Server {
	p.gateway.ListenWithDebug(network, addr, tlsConfig)
	return p
}

// SetNumOfThreads ...
func (p *Server) SetNumOfThreads(numOfThreads int) *Server {
	p.Lock()
	defer p.Unlock()

	if p.isRunning {
		p.OnReceiveStream(core.MakeSystemErrorStream(
			base.ErrServerAlreadyRunning.AddDebug(base.GetFileLine(1)),
		))
	} else if numOfThreads <= 0 {
		p.OnReceiveStream(core.MakeSystemErrorStream(
			base.ErrNumOfThreadsIsWrong.AddDebug(base.GetFileLine(1)),
		))
	} else {
		p.numOfThreads = numOfThreads
	}

	return p
}

// SetThreadBufferSize ...
func (p *Server) SetThreadBufferSize(threadBufferSize uint32) *Server {
	p.Lock()
	defer p.Unlock()

	if p.isRunning {
		p.OnReceiveStream(core.MakeSystemErrorStream(
			base.ErrServerAlreadyRunning.AddDebug(base.GetFileLine(1)),
		))
	} else if threadBufferSize <= 0 {
		p.OnReceiveStream(core.MakeSystemErrorStream(
			base.ErrThreadBufferSizeIsWrong.AddDebug(base.GetFileLine(1)),
		))
	} else {
		p.threadBufferSize = threadBufferSize
	}

	return p
}

// SetActionCache ...
func (p *Server) SetActionCache(actionCache core.ActionCache) *Server {
	p.Lock()
	defer p.Unlock()

	if p.isRunning {
		p.OnReceiveStream(core.MakeSystemErrorStream(
			base.ErrServerAlreadyRunning.AddDebug(base.GetFileLine(1)),
		))
	} else {
		p.actionCache = actionCache
	}

	return p
}

// SetLogHub ...
func (p *Server) SetLogHub(logHub core.IStreamReceiver) *Server {
	p.Lock()
	defer p.Unlock()

	if p.isRunning {
		p.OnReceiveStream(core.MakeSystemErrorStream(
			base.ErrServerAlreadyRunning.AddDebug(base.GetFileLine(1)),
		))
	} else {
		p.logHub = logHub
	}

	return p
}

// AddService ...
func (p *Server) AddService(
	name string,
	service *core.Service,
	data core.Map,
) *Server {
	p.Lock()
	defer p.Unlock()

	if p.isRunning {
		p.OnReceiveStream(core.MakeSystemErrorStream(
			base.ErrServerAlreadyRunning.AddDebug(base.GetFileLine(1)),
		))
	} else {
		p.mountServices = append(p.mountServices, core.NewServiceMeta(
			name,
			service,
			base.GetFileLine(1),
			data,
		))
	}

	return p
}

// BuildReplyCache ...
func (p *Server) BuildReplyCache() *Server {
	p.Lock()
	defer p.Unlock()

	_, file, _, _ := runtime.Caller(1)
	buildDir := path.Join(path.Dir(file))

	processor := core.NewProcessor(
		1,
		64,
		64,
		1024,
		nil,
		time.Second,
		p.mountServices,
		core.NewTestStreamReceiver(),
	)
	defer processor.Close()

	if err := processor.BuildCache(
		"cache",
		path.Join(buildDir, "cache", "rpc_action_cache.go"),
	); err != nil {
		p.OnReceiveStream(core.MakeSystemErrorStream(err))
	}

	return p
}

// OnReceiveStream ...
func (p *Server) OnReceiveStream(stream *core.Stream) {
	if stream != nil {
		switch stream.GetKind() {
		case core.DataStreamInternalRequest:
			fallthrough
		case core.DataStreamExternalRequest:
			p.processor.PutStream(stream)
		case core.DataStreamResponseOK:
			fallthrough
		case core.DataStreamResponseError:
			fallthrough
		case core.DataStreamBoardCast:
			p.gateway.OutStream(stream)
		default:
			if stream.GetKind() == core.SystemStreamReportError {
				if p.logHub != nil {
					p.logHub.OnReceiveStream(stream)
				} else {
					// log to screen
					if _, err := core.ParseResponseStream(stream); err != nil {
						log.Printf(
							"[Server Error (%d)]: %s",
							stream.GetSessionID(),
							err.Error(),
						)
					}
					stream.Release()
				}
			} else {
				stream.Release()
			}
		}
	}
}

// Open ...
func (p *Server) Open() bool {
	source := base.GetFileLine(1)

	ret := func() bool {
		p.Lock()
		defer p.Unlock()

		if p.isRunning {
			p.OnReceiveStream(core.MakeSystemErrorStream(
				base.ErrServerAlreadyRunning.AddDebug(source),
			))
			return false
		} else if processor := core.NewProcessor(
			p.numOfThreads,
			p.maxNodeDepth,
			p.maxCallDepth,
			p.threadBufferSize,
			p.actionCache,
			p.closeTimeout,
			p.mountServices,
			p,
		); processor == nil {
			return false
		} else {
			p.isRunning = true
			p.processor = processor
			return true
		}
	}()

	if ret {
		p.gateway.Open()
	}

	return ret
}

// IsRunning ...
func (p *Server) IsRunning() bool {
	p.Lock()
	defer p.Unlock()

	return p.isRunning
}

// Close ...
func (p *Server) Close() bool {
	p.Lock()
	defer p.Unlock()

	if !p.isRunning {
		p.OnReceiveStream(core.MakeSystemErrorStream(
			base.ErrServerNotRunning.AddDebug(base.GetFileLine(1)),
		))
		return false
	}

	p.gateway.Close()
	p.processor.Close()
	p.processor = nil
	p.isRunning = false
	return true
}
