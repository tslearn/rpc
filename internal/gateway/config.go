package gateway

import (
	"time"
)

// Config ...
type Config struct {
	numOfChannels    int
	transLimit       int
	heartbeat        time.Duration
	heartbeatTimeout time.Duration

	serverMaxSessions     int
	serverSessionTimeout  time.Duration
	serverReadBufferSize  int
	serverWriteBufferSize int
	serverCacheTimeout    time.Duration

	clientRequestInterval time.Duration
}

// GetDefaultConfig ...
func GetDefaultConfig() *Config {
	return &Config{
		numOfChannels:    48,
		transLimit:       4 * 1024 * 1024,
		heartbeat:        5 * time.Second,
		heartbeatTimeout: 8 * time.Second,

		serverMaxSessions:     10240000,
		serverSessionTimeout:  60 * time.Second,
		serverReadBufferSize:  1200,
		serverWriteBufferSize: 1200,
		serverCacheTimeout:    10 * time.Second,
		clientRequestInterval: 3 * time.Second,
	}
}
