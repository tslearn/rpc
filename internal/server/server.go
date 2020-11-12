package server

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/gateway"
	"github.com/rpccloud/rpc/internal/router"
	"path"
	"runtime"
	"sync"
	"time"
)

type Server struct {
	isRunning        bool
	processor        *RPCProcessor
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

func NewServer() *Server {
	return &Server{
		isRunning:        false,
		processor:        nil,
		gateway:          nil,
		numOfThreads:     runtime.NumCPU() * 16384,
		maxNodeDepth:     64,
		maxCallDepth:     64,
		threadBufferSize: 2048,
		actionCache:      nil,
		closeTimeout:     5 * time.Second,
		mountServices:    make([]*core.ServiceMeta, 0),
	}
}

func (p *Server) onError(sessionID uint64, err *base.Error) {

}

// SetNumOfThreads ...
func (p *Server) SetNumOfThreads(numOfThreads int) *Server {
	p.Lock()
	defer p.Unlock()

	if p.isRunning {
		panic("server has already running")
	} else if numOfThreads <= 0 {
		panic("numOfThreads must be greater than zero")
	} else {
		p.numOfThreads = numOfThreads
		return p
	}
}

// SetThreadBufferSize ...
func (p *Server) SetThreadBufferSize(threadBufferSize uint32) *Server {
	p.Lock()
	defer p.Unlock()

	if p.isRunning {
		panic("server has already running")
	} else if threadBufferSize <= 0 {
		panic("threadBufferSize must be greater than zero")
	} else {
		p.threadBufferSize = threadBufferSize
		return p
	}
}

// SetActionCache ...
func (p *Server) SetActionCache(actionCache core.ActionCache) *Server {
	p.Lock()
	defer p.Unlock()

	if p.isRunning {
		panic("server has already running")
	} else {
		p.actionCache = actionCache
		return p
	}
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
		panic("server has already running")
	} else {
		p.mountServices = append(p.mountServices, core.NewServiceMeta(
			name,
			service,
			base.GetFileLine(1),
			data,
		))
		return p
	}
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
		panic(err)
	}

	return p
}

// Serve ...
func (p *Server) Serve() *base.Error {
	err := func() *base.Error {
		p.Lock()
		defer p.Unlock()

		directRouter := router.NewDirectRouter()

		if p.isRunning {
			panic("server has already running")
		} else if gateway, err := gateway.NewGateWay(
			directRouter,
			p.onError,
		); err != nil {
			return err
		} else if processor, err := NewRPCProcessor(
			directRouter,
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
			p.gateway = gateway
			p.processor = processor
			return nil
		}
	}()

	if err == nil {
		return err
	}

	p.gateway.Serve()
	return nil
}

func (p *Server) Close() {
	p.Lock()
	defer p.Unlock()

	if !p.isRunning {
		panic("server has not running")
	} else {
		p.gateway.Close()
		p.processor.Close()
		p.isRunning = false
	}
}
