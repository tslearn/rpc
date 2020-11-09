package gateway

import (
	"github.com/rpccloud/rpc/internal"
	"github.com/rpccloud/rpc/internal/base"
	"sync"
	"time"
)

var sessionCache = &sync.Pool{
	New: func() interface{} {
		return &Session{}
	},
}

type SessionConfig struct {
	channels     int64
	readLimit    int64
	writeLimit   int64
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func getDefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		channels:     16,
		readLimit:    4 * 1024 * 1024,
		writeLimit:   4 * 1024 * 1024,
		readTimeout:  12 * time.Second,
		writeTimeout: 3 * time.Second,
	}
}

type Session struct {
	id           uint64
	security     string
	conn         internal.IStreamConn
	gateway      *GateWay
	channels     []serverChannel
	prevChannels []serverChannel
	sync.Mutex
}

func newSession(id uint64, gateway *GateWay) *Session {
	ret := sessionCache.Get().(*Session)
	ret.id = id
	ret.security = base.GetRandString(32)
	ret.conn = nil
	ret.gateway = gateway
	ret.channels = make([]serverChannel, gateway.config.channels)
	ret.prevChannels = nil
	return ret
}
