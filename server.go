package rpc

import (
	"github.com/rpccloud/rpc/internal"
	"path"
	"runtime"
	"time"
)

type Server struct {
	services      []*internal.ServiceMeta
	serverConfig  *serverConfig
	sessionConfig *sessionConfig
	serverCore
}

func NewServer() *Server {
	return &Server{
		services:      nil,
		serverConfig:  newServerConfig(),
		sessionConfig: newSessionConfig(),
	}
}

// SetDebug ...
func (p *Server) SetDebug() *Server {
	p.serverConfig.setDebug(true, internal.GetFileLine(1), p.onError)
	return p
}

// SetRelease ...
func (p *Server) SetRelease() *Server {
	p.serverConfig.setDebug(false, internal.GetFileLine(1), p.onError)
	return p
}

// SetNumOfThreads ...
func (p *Server) SetNumOfThreads(numOfThreads int) *Server {
	p.serverConfig.setNumOfThreads(
		numOfThreads,
		internal.GetFileLine(1),
		p.onError,
	)

	return p
}

func (p *Server) SetReplyCache(replyCache ReplyCache) *Server {
	p.serverConfig.setReplyCache(
		replyCache,
		internal.GetFileLine(1),
		p.onError,
	)

	return p
}

func (p *Server) SetTransportLimit(maxTransportBytes int) *Server {
	p.sessionConfig.setTransportLimit(
		maxTransportBytes,
		internal.GetFileLine(1),
		p.onError,
	)
	return p
}

func (p *Server) SetSessionConcurrency(sessionConcurrency int) *Server {
	p.sessionConfig.setConcurrency(
		sessionConcurrency,
		internal.GetFileLine(1),
		p.onError,
	)
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
		p.onError(0, internal.NewRuntimePanic(
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
		false,
		1,
		256,
		256,
		nil,
		time.Second,
		services,
		func(stream *internal.Stream) {},
	)
	defer processor.Close()

	if err := processor.BuildCache(
		"replycache",
		path.Join(buildDir, "replycache", "cache.go"),
	); err != nil {
		p.onError(0, err)
	}

	return p
}

func (p *Server) Serve() {
	p.sessionConfig.LockConfig()
	p.serverConfig.LockConfig()
	defer func() {
		p.sessionConfig.UnlockConfig()
		p.serverConfig.UnlockConfig()
	}()

	p.serve(
		p.sessionConfig.Clone(),
		func() streamHub {
			return internal.NewProcessor(
				p.serverConfig.isDebug,
				p.serverConfig.numOfThreads,
				256,
				256,
				p.serverConfig.replyCache,
				20*time.Second,
				p.services,
				p.onReturnStream,
			)
		},
	)
}
