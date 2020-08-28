package util

import (
	"github.com/rpccloud/rpc"
	"time"
)

func TimeNowMS() int64 {
	return rpc.TimeNow().UnixNano() / int64(time.Millisecond)
}
