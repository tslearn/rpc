package rpc

import (
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/util"
	"path"
	"runtime"
	"time"
)

// Server ...
type Server struct {
	services      []*core.ServiceMeta
	serverConfig  *serverConfig
	sessionConfig *sessionConfig
	serverCore
}

// NewServer ...
func NewServer() *Server {
	return &Server{
		services:      nil,
		serverConfig:  newServerConfig(),
		sessionConfig: newSessionConfig(),
	}
}

// SetDebug ...
func (p *Server) SetDebug() *Server {
	p.serverConfig.setDebug(true, util.GetFileLine(1), p.onError)
	return p
}

// SetRelease ...
func (p *Server) SetRelease() *Server {
	p.serverConfig.setDebug(false, util.GetFileLine(1), p.onError)
	return p
}

// SetNumOfThreads ...
func (p *Server) SetNumOfThreads(numOfThreads int) *Server {
	p.serverConfig.setNumOfThreads(
		numOfThreads,
		util.GetFileLine(1),
		p.onError,
	)

	return p
}

// SetReplyCache ...
func (p *Server) SetReplyCache(replyCache ReplyCache) *Server {
	p.serverConfig.setReplyCache(
		replyCache,
		util.GetFileLine(1),
		p.onError,
	)

	return p
}

// SetTransportLimit ...
func (p *Server) SetTransportLimit(maxTransportBytes int) *Server {
	p.sessionConfig.setTransportLimit(
		maxTransportBytes,
		util.GetFileLine(1),
		p.onError,
	)
	return p
}

// SetSessionConcurrency ...
func (p *Server) SetSessionConcurrency(sessionConcurrency int) *Server {
	p.sessionConfig.setConcurrency(
		sessionConcurrency,
		util.GetFileLine(1),
		p.onError,
	)
	return p
}

// ListenWebSocket ...
func (p *Server) ListenWebSocket(addr string) *Server {
	p.listenWebSocket(addr, util.GetFileLine(1))
	return p
}

// AddService ...
func (p *Server) AddService(
	name string,
	service *Service,
	data interface{},
) *Server {
	p.Lock()
	defer p.Unlock()

	if p.IsRunning() {
		p.onError(0, core.NewRuntimePanic(
			"AddService must be called before Serve",
		).AddDebug(util.GetFileLine(1)))
	} else {
		p.services = append(p.services, core.NewServiceMeta(
			name,
			service,
			util.GetFileLine(1),
			data,
		))
	}

	return p
}

// BuildReplyCache ...
func (p *Server) BuildReplyCache() *Server {
	_, file, _, _ := runtime.Caller(1)
	buildDir := path.Join(path.Dir(file))

	services := func() []*core.ServiceMeta {
		p.Lock()
		defer p.Unlock()
		return p.services
	}()

	processor := core.NewProcessor(
		false,
		1,
		64,
		64,
		nil,
		time.Second,
		services,
		func(stream *core.Stream) {},
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

// Serve ...
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
			return core.NewProcessor(
				p.serverConfig.isDebug,
				p.serverConfig.numOfThreads,
				64,
				64,
				p.serverConfig.replyCache,
				20*time.Second,
				p.services,
				p.onReturnStream,
			)
		},
	)
}
