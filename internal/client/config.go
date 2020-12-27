package client

import (
	"time"
)

// Config ...
type Config struct {
	numOfChannels    int
	transLimit       int
	rBufSize         int
	wBufSize         int
	heartbeat        time.Duration
	heartbeatTimeout time.Duration
	requestTimeout   time.Duration
	requestInterval  time.Duration
}
