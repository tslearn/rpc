package server

import (
	"crypto/tls"
	"fmt"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"github.com/rpccloud/rpc/internal/gateway"
	"github.com/rpccloud/rpc/internal/route"
)

// Server ...
type Server struct {
	isRunning        bool
	processor        *RPCProcessor
	router           route.IRouter
	gateway          *gateway.GateWay
	numOfThreads     int
	maxNodeDepth     int16
	maxCallDepth     int16
	threadBufferSize uint32
	actionCache      core.ActionCache
	closeTimeout     time.Duration
	mountServices    []*core.ServiceMeta
	sync.Mutex
}

// NewServer ...
func NewServer() *Server {
	ret := &Server{
		isRunning:        false,
		processor:        nil,
		router:           route.NewDirectRouter(),
		gateway:          nil,
		numOfThreads:     runtime.NumCPU() * 16384,
		maxNodeDepth:     64,
		maxCallDepth:     64,
		threadBufferSize: 2048,
		actionCache:      nil,
		closeTimeout:     5 * time.Second,
		mountServices:    make([]*core.ServiceMeta, 0),
	}
	ret.gateway = gateway.NewGateWay(
		0,
		gateway.GetDefaultConfig(),
		ret.router,
		ret.onError,
	)
	return ret
}

func (p *Server) onError(sessionID uint64, err *base.Error) {
	fmt.Println("server onError: ", sessionID, err)
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

// SetNumOfThreads ...
func (p *Server) SetNumOfThreads(numOfThreads int) *Server {
	p.Lock()
	defer p.Unlock()

	if p.isRunning {
		p.onError(0, errors.ErrServerAlreadyRunning.AddDebug(base.GetFileLine(1)))
	} else if numOfThreads <= 0 {
		p.onError(0, errors.ErrNumOfThreadsIsWrong.AddDebug(base.GetFileLine(1)))
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
		p.onError(0, errors.ErrServerAlreadyRunning.AddDebug(base.GetFileLine(1)))
	} else if threadBufferSize <= 0 {
		p.onError(
			0,
			errors.ErrThreadBufferSizeIsWrong.AddDebug(base.GetFileLine(1)),
		)
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
		p.onError(0, errors.ErrServerAlreadyRunning.AddDebug(base.GetFileLine(1)))
	} else {
		p.actionCache = actionCache
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
		p.onError(0, errors.ErrServerAlreadyRunning.AddDebug(base.GetFileLine(1)))
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

	processor, _ := core.NewProcessor(
		1,
		64,
		64,
		1024,
		nil,
		time.Second,
		p.mountServices,
		func(stream *core.Stream) {},
	)
	defer processor.Close()

	if err := processor.BuildCache(
		"cache",
		path.Join(buildDir, "cache", "rpc_action_cache.go"),
	); err != nil {
		p.onError(0, err)
	}

	return p
}

// Serve ...
func (p *Server) Serve() {
	err := func() *base.Error {
		p.Lock()
		defer p.Unlock()

		if p.isRunning {
			return errors.ErrServerAlreadyRunning.AddDebug(base.GetFileLine(1))
		} else if processor, err := NewRPCProcessor(
			p.router,
			p.numOfThreads,
			p.maxNodeDepth,
			p.maxCallDepth,
			p.threadBufferSize,
			p.actionCache,
			p.closeTimeout,
			p.mountServices,
		); err != nil {
			return err
		} else {
			p.isRunning = true
			p.processor = processor
			return nil
		}
	}()

	if err != nil {
		p.onError(0, err)
	} else {
		p.gateway.Open()
	}
}

// Close ...
func (p *Server) Close() {
	p.Lock()
	defer p.Unlock()

	if !p.isRunning {
		p.onError(0, errors.ErrServerNotRunning.AddDebug(base.GetFileLine(1)))
	} else {
		p.gateway.Close()
		p.processor.Close()
		p.isRunning = false
	}
}
