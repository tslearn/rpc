// +build linux

package metrics

import (
	"github.com/rpccloud/rpc/internal/base"
	"testing"
)

func TestParseProcStat(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(parseProcStat("")).IsNil()
		assert(parseProcStat("no 1 1 1 1")).IsNil()
		assert(parseProcStat("cpu false 1 1 1")).IsNil()
		assert(parseProcStat("cpu 1 false 1 1")).IsNil()
		assert(parseProcStat("cpu 1 1 false 1")).IsNil()
		assert(parseProcStat("cpu 1 1 1 false")).IsNil()
		assert(parseProcStat("cpu 1 1 1 1")).IsNotNil()
	})
}
