package server

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
}

// GetDefaultConfig ...
func GetDefaultConfig() *Config {
	return &Config{
		numOfChannels:    32,
		transLimit:       4 * 1024 * 1024,
		heartbeat:        4 * time.Second,
		heartbeatTimeout: 8 * time.Second,

		serverMaxSessions:     10240000,
		serverSessionTimeout:  120 * time.Second,
		serverReadBufferSize:  1200,
		serverWriteBufferSize: 1200,
		serverCacheTimeout:    10 * time.Second,
	}
}
