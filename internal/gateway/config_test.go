package gateway

import (
	"github.com/rpccloud/rpc/internal/base"
	"testing"
	"time"
)

func TestGetDefaultConfig(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		cfg := GetDefaultConfig()
		assert(cfg.numOfChannels).Equal(32)
		assert(cfg.transLimit).Equal(4 * 1024 * 1024)
		assert(cfg.heartbeat).Equal(5 * time.Second)
		assert(cfg.heartbeatTimeout).Equal(8 * time.Second)
		assert(cfg.serverMaxSessions).Equal(10240000)
		assert(cfg.serverSessionTimeout).Equal(120 * time.Second)
		assert(cfg.serverReadBufferSize).Equal(1200)
		assert(cfg.serverWriteBufferSize).Equal(1200)
		assert(cfg.serverCacheTimeout).Equal(10 * time.Second)
		assert(cfg.clientRequestInterval).Equal(3 * time.Second)
	})
}
