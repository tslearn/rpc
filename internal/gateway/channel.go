package gateway

import (
	"github.com/rpccloud/rpc/internal/core"
	"sync"
)

type Channel struct {
	seq       uint64
	retTimeNS uint64
	retStream *core.Stream
	sync.Mutex
}
