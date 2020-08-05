package rpc

import (
	"github.com/rpccloud/rpc/internal"
	"path"
	"runtime"
	"sync"
	"time"
)

type Server struct {
	isDebug      bool
	services     []*internal.ServiceMeta
	numOfThreads int
	replyCache   internal.ReplyCache
	baseServer
}

func NewServer() *Server {
	return &Server{
		isDebug:      false,
		services:     make([]*internal.ServiceMeta, 0),
		numOfThreads: runtime.NumCPU() * 8192,
		replyCache:   nil,
		baseServer: baseServer{
			adapters:           nil,
			hub:                nil,
			sessionMap:         sync.Map{},
			sessionConcurrency: 64,
			sessionSeed:        0,
			transportLimit:     1024 * 1024,
			readTimeout:        10 * time.Second,
			writeTimeout:       1 * time.Second,
		},
	}
}

func (p *Server) SetDebug() *Server {
	p.Lock()
	defer p.Unlock()

	if p.IsRunning() {
		p.onSessionError(0, internal.NewRuntimePanic(
			"SetDebug must be called before Serve",
		).AddDebug(internal.GetFileLine(1)))
	} else {
		p.isDebug = true
	}

	return p
}

func (p *Server) setRelease() *Server {
	p.Lock()
	defer p.Unlock()

	if p.IsRunning() {
		p.onSessionError(0, internal.NewRuntimePanic(
			"SetRelease must be called before Serve",
		).AddDebug(internal.GetFileLine(1)))
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
		p.onSessionError(0, internal.NewRuntimePanic(
			"numOfThreads must be greater than 0",
		).AddDebug(internal.GetFileLine(1)))
	} else if p.IsRunning() {
		p.onSessionError(0, internal.NewRuntimePanic(
			"SetNumOfThreads must be called before Serve",
		).AddDebug(internal.GetFileLine(1)))
	} else {
		p.numOfThreads = numOfThreads
	}

	return p
}

func (p *Server) SetReplyCache(replyCache internal.ReplyCache) *Server {
	p.Lock()
	defer p.Unlock()

	if p.IsRunning() {
		p.onSessionError(0, internal.NewRuntimePanic(
			"SetReplyCache must be called before Serve",
		).AddDebug(internal.GetFileLine(1)))
	} else {
		p.replyCache = replyCache
	}

	return p
}

// ListenWebSocket ...
func (p *Server) ListenWebSocket(addr string) *Server {
	p.listenWebSocket(addr, internal.GetFileLine(1))
	return p
}

// AddService ...
func (p *Server) AddService(name string, service *Service) *Server {
	p.Lock()
	defer p.Unlock()

	if p.IsRunning() {
		p.onSessionError(0, internal.NewRuntimePanic(
			"AddService must be called before Serve",
		).AddDebug(internal.GetFileLine(1)))
	} else {
		p.services = append(p.services, internal.NewServiceMeta(
			name,
			service,
			internal.GetFileLine(1),
		))
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
		p.onSessionError(0, err)
	}

	return p
}

func (p *Server) Serve() {
	p.serve(func() streamHub {
		ret := internal.NewProcessor(
			p.isDebug,
			p.numOfThreads,
			32,
			32,
			p.replyCache,
			20*time.Second,
			p.services,
			p.onReturnStream,
		)

		// fmt.Println(ret)
		return ret
	})
}
