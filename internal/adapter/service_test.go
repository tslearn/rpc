package adapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"testing"
)

func TestSyncTCPServerService_Open(t *testing.T) {
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
	})

	t.Run("test tls", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		tlsConfig, e := base.GetTLSServerConfig(
			"../cert/server.pem",
			"../cert/server-key.pem",
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
	})
}

func TestSyncTCPServerService_Run(t *testing.T) {

}

func TestSyncTCPServerService_Close(t *testing.T) {

}
