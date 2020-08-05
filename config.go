package rpc

import (
	"fmt"
	"github.com/rpccloud/rpc/internal"
	"sync"
	"time"
)

const sessionConfigMaxConcurrency = 1024
const sessionConfigMinTransportLimit = 10240

type sessionConfig struct {
	concurrency    int64
	transportLimit int64
	readTimeout    time.Duration
	writeTimeout   time.Duration
	locked         bool
	sync.Mutex
}

func newSessionConfig() *sessionConfig {
	return &sessionConfig{
		concurrency:    64,
		transportLimit: 4 * 1024 * 1024,
		readTimeout:    12 * time.Second,
		writeTimeout:   2 * time.Second,
		locked:         false,
	}
}

func (p *sessionConfig) LockConfig() *sessionConfig {
	p.Lock()
	defer p.Unlock()
	p.locked = true

	return &sessionConfig{
		concurrency:    p.concurrency,
		transportLimit: p.transportLimit,
		readTimeout:    p.readTimeout,
		writeTimeout:   p.writeTimeout,
		locked:         false,
	}
}

func (p *sessionConfig) UnlockConfig() {
	p.Lock()
	defer p.Unlock()
	p.locked = false
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
