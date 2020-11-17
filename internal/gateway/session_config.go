package gateway

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"time"
)

type SessionConfig struct {
	numOfChannels   int64
	transLimit      int64
	readTimeout     time.Duration
	writeTimeout    time.Duration
	heartbeat       time.Duration
	cacheTimeout    time.Duration
	requestTimeout  time.Duration
	requestInterval time.Duration
}

func getDefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		numOfChannels:   32,
		transLimit:      4 * 1024 * 1024,
		readTimeout:     8 * time.Second,
		writeTimeout:    3 * time.Second,
		heartbeat:       5 * time.Second,
		cacheTimeout:    10 * time.Second,
		requestTimeout:  16 * time.Second,
		requestInterval: 3 * time.Second,
	}
}

func (p *SessionConfig) NumOfChannels() int64 {
	return p.numOfChannels
}

func (p *SessionConfig) TransLimit() int64 {
	return p.transLimit
}

func (p *SessionConfig) ReadTimeout() time.Duration {
	return p.readTimeout
}

func (p *SessionConfig) WriteTimeout() time.Duration {
	return p.writeTimeout
}

func (p *SessionConfig) Heartbeat() time.Duration {
	return p.heartbeat
}

func (p *SessionConfig) CacheTimeout() time.Duration {
	return p.cacheTimeout
}

func (p *SessionConfig) RequestTimeout() time.Duration {
	return p.requestTimeout
}

func (p *SessionConfig) RequestInterval() time.Duration {
	return p.requestInterval
}

func (p *SessionConfig) Copy() SessionConfig {
	return SessionConfig{
		numOfChannels:   p.numOfChannels,
		transLimit:      p.transLimit,
		readTimeout:     p.readTimeout,
		writeTimeout:    p.writeTimeout,
		heartbeat:       p.heartbeat,
		cacheTimeout:    p.cacheTimeout,
		requestTimeout:  p.requestTimeout,
		requestInterval: p.requestInterval,
	}
}

func (p *SessionConfig) Equals(o *SessionConfig) bool {
	return p.numOfChannels == o.numOfChannels &&
		p.transLimit == o.transLimit &&
		p.readTimeout == o.readTimeout &&
		p.writeTimeout == o.writeTimeout &&
		p.heartbeat == o.heartbeat &&
		p.cacheTimeout == o.cacheTimeout &&
		p.requestTimeout == o.requestTimeout &&
		p.requestInterval == o.requestInterval
}

func (p *SessionConfig) WriteToStream(stream *core.Stream) {
	stream.WriteInt64(p.numOfChannels)
	stream.WriteInt64(p.transLimit)
	stream.WriteInt64(int64(p.readTimeout))
	stream.WriteInt64(int64(p.writeTimeout))
	stream.WriteInt64(int64(p.heartbeat))
	stream.WriteInt64(int64(p.cacheTimeout))
	stream.WriteInt64(int64(p.requestTimeout))
	stream.WriteInt64(int64(p.requestInterval))
}

func ReadSessionConfigFromStream(
	stream *core.Stream,
) (*SessionConfig, *base.Error) {
	if numOfChannels, err := stream.ReadInt64(); err != nil {
		return nil, err
	} else if numOfChannels <= 0 {
		return nil, errors.ErrStream
	} else if transLimit, err := stream.ReadInt64(); err != nil {
		return nil, err
	} else if transLimit <= 0 {
		return nil, errors.ErrStream
	} else if readTimeout, err := stream.ReadInt64(); err != nil {
		return nil, err
	} else if readTimeout <= 0 {
		return nil, errors.ErrStream
	} else if writeTimeout, err := stream.ReadInt64(); err != nil {
		return nil, err
	} else if writeTimeout <= 0 {
		return nil, errors.ErrStream
	} else if heartbeat, err := stream.ReadInt64(); err != nil {
		return nil, err
	} else if heartbeat <= 0 {
		return nil, errors.ErrStream
	} else if cacheTimeout, err := stream.ReadInt64(); err != nil {
		return nil, err
	} else if cacheTimeout <= 0 {
		return nil, errors.ErrStream
	} else if requestTimeout, err := stream.ReadInt64(); err != nil {
		return nil, err
	} else if requestTimeout <= 0 {
		return nil, errors.ErrStream
	} else if requestInterval, err := stream.ReadInt64(); err != nil {
		return nil, err
	} else if requestInterval <= 0 {
		return nil, errors.ErrStream
	} else {
		return &SessionConfig{
			numOfChannels:   numOfChannels,
			transLimit:      transLimit,
			readTimeout:     time.Duration(readTimeout),
			writeTimeout:    time.Duration(writeTimeout),
			heartbeat:       time.Duration(heartbeat),
			cacheTimeout:    time.Duration(cacheTimeout),
			requestTimeout:  time.Duration(requestTimeout),
			requestInterval: time.Duration(requestInterval),
		}, nil
	}
}
