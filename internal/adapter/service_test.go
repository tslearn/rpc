package adapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"path"
	"runtime"
	"testing"
)

func TestSyncTCPServerService_Open(t *testing.T) {
	_, curFile, _, _ := runtime.Caller(0)
	curDir := path.Dir(curFile)

	t.Run("addr error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		v := &syncTCPServerService{
			adapter: NewClientAdapter(
				"tcp", "error", nil, 1200, 1200, receiver,
			),
			ln:         nil,
			orcManager: base.NewORCManager(),
		}
		assert(v.Open()).IsFalse()
		assert(receiver.GetError()).IsNotNil()
		v.Close()
	})

	t.Run("test tcp", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		v := &syncTCPServerService{
			adapter: NewClientAdapter(
				"tcp", "0.0.0.0:65432", nil, 1200, 1200, receiver,
			),
			ln:         nil,
			orcManager: base.NewORCManager(),
		}
		assert(v.Open()).IsTrue()
		assert(v.ln).IsNotNil()

		v.Close()
	})

	t.Run("test tls", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		tlsConfig, e := base.GetTLSServerConfig(
			path.Join(curDir, "_cert_", "server.crt"),
			path.Join(curDir, "_cert_", "server.key"),
		)
		if e != nil {
			panic(e)
		}
		v := &syncTCPServerService{
			adapter: NewClientAdapter(
				"tcp", "0.0.0.0:65432", tlsConfig, 1200, 1200, receiver,
			),
			ln:         nil,
			orcManager: base.NewORCManager(),
		}
		assert(v.Open()).IsTrue()
		assert(v.ln).IsNotNil()
		v.Close()
	})
}

func TestSyncTCPServerService_Run(t *testing.T) {

}

func TestSyncTCPServerService_Close(t *testing.T) {

}
