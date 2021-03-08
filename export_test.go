package rpc

import (
	"github.com/rpccloud/rpc/internal/base"
	"testing"
)

func TestNewService(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(NewService()).IsNotNil()
	})
}

func TestNewServer(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(NewServer()).IsNotNil()
	})
}

func TestDialTLS(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		cert, _ := GetTLSClientConfig(true, nil)
		v := DialTLS("ws", "127.0.0.1", cert)
		defer v.Close()
		assert(v).IsNotNil()
	})
}

func TestDial(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := Dial("ws", "127.0.0.1")
		defer v.Close()
		assert(v).IsNotNil()
	})
}

func TestGetTLSServerConfig(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(GetTLSServerConfig("errorFile", "errorFile")).
			Equal(base.GetTLSServerConfig("errorFile", "errorFile"))
	})
}

func TestGetTLSClientConfig(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(GetTLSClientConfig(true, nil)).
			Equal(base.GetTLSClientConfig(true, nil))
	})
}
