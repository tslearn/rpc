package rpc

import (
	"fmt"
	"github.com/rpccloud/rpc/internal"
	"runtime"
	"sync"
	"time"
)

const sessionConfigMaxConcurrency = 1024
const sessionConfigMinTransportLimit = 10240

type baseConfig struct {
	locked bool
	sync.Mutex
}

func (p *baseConfig) LockConfig() {
	p.Lock()
	defer p.Unlock()
	p.locked = true
}

func (p *baseConfig) UnlockConfig() {
	p.Lock()
	defer p.Unlock()
	p.locked = false
}

type sessionConfig struct {
	concurrency    int64
	transportLimit int64
	readTimeout    time.Duration
	writeTimeout   time.Duration
	baseConfig
}

func newSessionConfig() *sessionConfig {
	return &sessionConfig{
		concurrency:    64,
		transportLimit: 4 * 1024 * 1024,
		readTimeout:    12 * time.Second,
		writeTimeout:   2 * time.Second,
	}
}

func (p *sessionConfig) setTransportLimit(
	maxTransportBytes int,
	dbg string,
	onError func(uint64, Error),
) {
	p.Lock()
	defer p.Unlock()

	if p.locked {
		onError(0, internal.NewRuntimePanic("config is locked").AddDebug(dbg))
	} else if maxTransportBytes < sessionConfigMinTransportLimit {
		onError(0, internal.NewRuntimePanic(fmt.Sprintf(
			"maxTransportBytes must be greater than or equal to %d",
			sessionConfigMinTransportLimit,
		)).AddDebug(dbg))
	} else {
		p.transportLimit = int64(maxTransportBytes)
	}
}

func (p *sessionConfig) setConcurrency(
	concurrency int,
	dbg string,
	onError func(uint64, Error),
) {
	p.Lock()
	defer p.Unlock()

	if p.locked {
		onError(0, internal.NewRuntimePanic("config is locked").AddDebug(dbg))
	} else if concurrency <= 0 {
		onError(0, internal.NewRuntimePanic(
			"sessionConcurrency be greater than 0",
		).AddDebug(dbg))
	} else if concurrency > sessionConfigMaxConcurrency {
		onError(0, internal.NewRuntimePanic(fmt.Sprintf(
			"sessionConcurrency be less than or equal to %d",
			sessionConfigMaxConcurrency,
		)).AddDebug(dbg))
	} else {
		p.concurrency = int64(concurrency)
	}
}

func (p *sessionConfig) Clone() *sessionConfig {
	p.Lock()
	defer p.Unlock()

	return &sessionConfig{
		concurrency:    p.concurrency,
		transportLimit: p.transportLimit,
		readTimeout:    p.readTimeout,
		writeTimeout:   p.writeTimeout,
	}
}

type serverConfig struct {
	isDebug      bool
	numOfThreads int
	replyCache   ReplyCache
	baseConfig
}

func newServerConfig() *serverConfig {
	return &serverConfig{
		isDebug:      false,
		numOfThreads: runtime.NumCPU() * 8192,
		replyCache:   nil,
	}
}

func (p *serverConfig) setDebug(
	isDebug bool,
	dbg string,
	onError func(uint64, Error),
) {
	p.Lock()
	defer p.Unlock()

	if p.locked {
		onError(0, internal.NewRuntimePanic("config is locked").AddDebug(dbg))
	} else {
		p.isDebug = isDebug
	}
}

func (p *serverConfig) setNumOfThreads(
	numOfThreads int,
	dbg string,
	onError func(uint64, Error),
) {
	p.Lock()
	defer p.Unlock()

	if p.locked {
		onError(0, internal.NewRuntimePanic("config is locked").AddDebug(dbg))
	} else if numOfThreads <= 0 {
		onError(0, internal.NewRuntimePanic(
			"sessionConcurrency be greater than 0",
		).AddDebug(dbg))
	} else {
		p.numOfThreads = numOfThreads
	}
}

func (p *serverConfig) setReplyCache(
	replyCache ReplyCache,
	dbg string,
	onError func(uint64, Error),
) {
	p.Lock()
	defer p.Unlock()

	if p.locked {
		onError(0, internal.NewRuntimePanic("config is locked").AddDebug(dbg))
	} else {
		p.replyCache = replyCache
	}
}

func (p *serverConfig) Clone() *serverConfig {
	p.Lock()
	defer p.Unlock()

	return &serverConfig{
		isDebug:      p.isDebug,
		numOfThreads: p.numOfThreads,
		replyCache:   p.replyCache,
	}
}
