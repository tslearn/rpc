package adapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"net"
	"strings"
	"testing"
)

func TestAdapter(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(ErrNetClosingSuffix).Equal("use of closed network connection")

		// check ErrNetClosingSuffix on all platform
		waitCH := make(chan bool)
		go func() {
			ln, e := net.Listen("tcp", "0.0.0.0:65432")
			if e != nil {
				panic(e)
			}

			waitCH <- true
			_, _ = ln.Accept()
			_ = ln.Close()
		}()

		<-waitCH
		conn, e := net.Dial("tcp", "0.0.0.0:65432")
		if e != nil {
			panic(e)
		}
		_ = conn.Close()
		e = conn.Close()
		assert(e).IsNotNil()
		assert(strings.HasSuffix(e.Error(), ErrNetClosingSuffix)).IsTrue()
	})
}
