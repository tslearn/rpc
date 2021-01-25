package gateway

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/router"
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		gateway := NewGateWay(
			43,
			GetDefaultConfig(),
			router.NewDirectRouter(),
			func(sessionID uint64, err *base.Error) {},
		)
		v := NewSession(3, gateway)
		assert(v.id).Equal(uint64(3))
		assert(v.gateway).Equal(gateway)
		assert(len(v.security)).Equal(32)
		assert(v.conn).IsNil()
		assert(len(v.channels)).Equal(GetDefaultConfig().numOfChannels)
		assert(cap(v.channels)).Equal(GetDefaultConfig().numOfChannels)
		assert(base.TimeNow().UnixNano()-v.activeTimeNS < int64(time.Second)).
			IsTrue()
		assert(v.prev).IsNil()
		assert(v.next).IsNil()
	})
}
