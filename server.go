package rpc

import (
	"github.com/rpccloud/rpc/internal"
	"path"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

type Server struct {
	services     []*internal.ServiceMeta
	numOfThreads int
	baseServer
}

func NewServer() *Server {
	return &Server{
		services:     make([]*internal.ServiceMeta, 0),
		numOfThreads: runtime.NumCPU() * 8192,
		baseServer: baseServer{
			isDebug:            false,
			listens:            make([]*listenItem, 0),
			adapters:           nil,
			hub:                nil,
			sessionMap:         sync.Map{},
			sessionConcurrency: 64,
			sessionSeed:        0,
			transportLimit:     1024 * 1024,
			readTimeout:        10 * time.Second,
			writeTimeout:       1 * time.Second,
			replyCache:         nil,
		},
	}
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

// ListenWebSocket ...
func (p *Server) ListenWebSocket(addr string) *Server {
	p.listenWebSocket(addr, internal.GetFileLine(1))
	return p
}

// AddService ...
func (p *Server) AddService(name string, service *Service) *Server {
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
